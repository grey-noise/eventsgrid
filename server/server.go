package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	consul "github.com/armon/consul-api"
	pb "github.com/grey-noise/eventsgrid/events"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/oklog/ulid"

	openzipkin "github.com/openzipkin/zipkin-go"
	zhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli"
	schema "github.com/xeipuuv/gojsonschema"
	"go.opencensus.io/exporter/zipkin"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

var msgSent = make(chan counter)
var msgRCV = make(chan counter)

const (
	ok                  = "OK"
	valid               = "VALID"
	unableTovalidate    = "UNABLE_TO_VALIDATE"
	notValid            = "NOT_VALID"
	notPublished        = "NOT_PUBLISHED"
	unabletoPublish     = "UNABLE_TO_PUBLISH"
	unknown             = "UNKNOWN"
	received            = "RECEIVED"
	unableToAcknowledge = "UNABLE_TO_ACKNOWLEDGED"
	acknowledged        = "ACKNOWLEDGED"
)

type counter struct {
	clientID string
	groupID  string
	eventID  string
	status   string
}

type eventServer struct {
	natsEP          string
	clusterID       string
	storageEP       string
	monitoringPort  int
	port            int
	serviceRegistry string
	prom            *http.Server
	grpcServer      *grpc.Server
}

func (e *eventServer) acceptEvent(eventID string) bool {
	config := consul.Config{Address: e.storageEP}

	// Get a new client, with KV endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Fatalf("Can't connect to Consul: check Consul  is running with address %s", e.storageEP)
	}
	kv := client.KV()
	path := "events/" + eventID + "/schema"
	log.Println(path)
	pair, _, err := kv.Get(path, nil)
	if err != nil || pair == nil {
		log.Println(err)
		return false
	}

	return true

}

//validate : validating of the document
func validate(eventID, endpoint string, json []byte) (*ValidationError, error) {

	config := consul.Config{Address: endpoint}
	// Get a new client, with KV endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Printf("Can't connect to Consul: check Consul  is running with address %s", endpoint)
		return nil, err
	}
	kv := client.KV()
	path := "events/" + eventID + "/schema"
	pair, _, err := kv.Get(path, nil)
	if err != nil {
		log.Printf("error while getting schema : %v", err)
		return nil, err
	}
	schemaLoader := schema.NewBytesLoader(pair.Value)

	jsonLoader := schema.NewStringLoader(string(json))
	result, err := schema.Validate(schemaLoader, jsonLoader)

	if err != nil {
		log.Printf("error while validating document %v", err)
		return nil, err
	}

	if result.Valid() {
		return nil, nil
	}
	log.Printf("The document is not valid. see errors :\n")
	for _, desc := range result.Errors() {
		log.Printf("- %s\n", desc)
	}
	log.Printf("got an error while validating document")
	return &ValidationError{Message: "json invalid"}, nil

}

func (e *eventServer) SendEvents(stream pb.EventsSender_SendEventsServer) error {
	log.Println("client connection open")
	messages := make(chan *pb.Event)
	ctx := stream.Context()
	headers, _ := metadata.FromIncomingContext(ctx)

	clientID := headers["clientid"][0] + "-emit"
	groupID := headers["groupid"][0]
	eventID := headers["eventid"][0]

	go func() {
		for {
			select {
			case msg := <-messages:
				sc, err := stan.Connect(e.clusterID, clientID, stan.NatsURL(e.natsEP))
				if err != nil {
					log.Printf("Can't connect to nats server %s \n Make sure a NATS Streaming Server %v  is running :%v", e.clusterID, clientID, e.natsEP)
					msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: notPublished}
					log.Print(err)
					break
				}
				var dst bytes.Buffer
				json.Compact(&dst, msg.GetPayload())
				err = sc.Publish(eventID, dst.Bytes())
				if err != nil {
					log.Fatalf("Error during publish: %v\n", err)
					msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: notPublished}
				}
				msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: ok}
				sc.Close()
			case <-ctx.Done():
				log.Println("client connection closed")
				return
			}
		}
	}()

	for {
		//_, span := trace.StartSpan(stream.Context(), "server.receive.message")
		//span.End()
		in, err := stream.Recv()

		if in == nil {
			break
		}
		if err == io.EOF {
			close(messages)
			return nil
		}
		if err != nil {
			log.Printf("unable to open stream")
			return grpc.Errorf(codes.Internal, "Unable to open stream: %v.\n", err)
		}
		msgRCV <- counter{clientID: clientID, groupID: groupID, eventID: eventID, status: received}
		var s pb.Status
		verr, err := validate(eventID, e.storageEP, in.GetPayload())
		if err != nil {
			s = pb.Status{Code: pb.Status_NOK_UNKOWN, Description: err.Error()}
			msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: unableTovalidate}
		} else {
			if verr != nil {
				s = pb.Status{Code: pb.Status_NOK_VALID, Description: "Not valid"}
				msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: notValid}
			} else {
				msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: valid}
				messages <- in
				s = pb.Status{Code: pb.Status_OK, Description: "Valid"}
			}
		}
		c := pb.Cursor{
			Ts: time.Now().Unix(),
			Id: 0}
		a := &pb.Acknowledge{Cursor: &c, Status: &s}
		if err := stream.Send(a); err != nil {
			log.Printf("error while sending to the client")
			msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: unableToAcknowledge}
		}
		msgRCV <- counter{eventID: eventID, clientID: clientID, groupID: groupID, status: acknowledged}
	}
	return nil
}

func (e *eventServer) GetEvents(a *pb.Acknowledge, stream pb.EventsSender_GetEventsServer) error {
	ctx := stream.Context()

	headers, _ := metadata.FromIncomingContext(ctx)
	log.Printf("%+v", headers)

	clientID := headers["clientid"][0] + "-listen"
	groupID := headers["groupid"][0]
	eventID := headers["eventid"][0]

	log.Println("looking for event : ", eventID)
	if eventID == "" {
		log.Println("got no event to subscribe")
		return grpc.Errorf(codes.InvalidArgument, "event can not be empty ")

	}
	if !e.acceptEvent(eventID) {
		log.Println("event unknown to subscribe")
		return grpc.Errorf(codes.InvalidArgument, "eventid %s is unknown ", eventID)
	}

	if clientID == "" {
		log.Println("Go no  X-ClientID generating one")
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
		clientID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	}
	if groupID == "" {
		log.Println("Go no  X-ClientID generating one")
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
		groupID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	}
	log.Printf("groupID is %s", groupID)

	sc, err := stan.Connect(e.clusterID, clientID, stan.NatsURL(e.natsEP))

	if err != nil {
		log.Printf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, e.natsEP)
		return grpc.Errorf(codes.Internal, "Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, e.natsEP)
	}

	messages := make(chan *pb.Event)
	mcb := func(msg *stan.Msg) {

		c := &pb.Cursor{Id: msg.Sequence, Ts: msg.Timestamp}
		e := &pb.Event{Cursor: c, Payload: msg.Data}
		msgSent <- counter{clientID: clientID, groupID: groupID, eventID: eventID}
		messages <- e

	}
	log.Printf("subsribed  to queue %s with group %ss starting at %d", eventID, groupID, a.GetCursor().GetId())
	sub, err := sc.QueueSubscribe(eventID, groupID, mcb, stan.StartAtSequence(a.Cursor.GetId()), stan.DurableName("dur"))
	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	if err != nil {
		return grpc.Errorf(codes.Internal, "unable to create subcription %v", err)
	}
	log.Printf("subsribed  to queue %s with group %ss starting at %d", eventID, groupID, a.GetCursor().GetId())
	defer sub.Close()
	for {
		select {
		case e := <-messages:
			if err := stream.Send(e); err != nil {
				log.Println("error processing stream")
				log.Println("closing subscription for ", clientID)
				sub.Close()
				log.Println("closing connexion for ", clientID)
				sc.Close()
				return nil
			}
		case <-ctx.Done():
			log.Println("ctx done")
			log.Println("closing subscription for ", clientID)
			sub.Close()
			log.Println("closing connexion for ", clientID)
			sc.Close()
			return nil
		}
	}

}

func main() {
	s := eventServer{}
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "seriveport",
			Value:       8080,
			Usage:       "server port endpoint",
			Destination: &s.port,
			EnvVar:      "EVENT_PORT"},
		cli.IntFlag{
			Name:        "monitoringport",
			Value:       8000,
			Usage:       "prometheus port endpoint",
			Destination: &s.monitoringPort,
			EnvVar:      "EVENT_MONITORING_PORT"},
		cli.StringFlag{
			Name:        "service-Registry",
			Value:       "consul-ea.cloud.smals.be",
			Usage:       "consul endpoint",
			Destination: &s.serviceRegistry,
			EnvVar:      "CONSUL-URL"},
		cli.StringFlag{
			Name:        "clusterid",
			Value:       "test-cluster",
			Usage:       "cluster",
			Destination: &s.clusterID,
			EnvVar:      "CLUSTER-ID"},
	}

	app.Action = func(c *cli.Context) error {
		s.run()
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nReceived an interrupt, stopping services...\n")
			s.deregister()
			s.stopMonitoring()
			s.grpcServer.GracefulStop()

			cleanupDone <- true
		}
	}()
	<-cleanupDone
	os.Exit(0)

}
func (e *eventServer) run() {
	// The localEndpoint stores the name and address of the local service
	localEndpoint, err := openzipkin.NewEndpoint("example-server", "192.168.1.5:5454")
	if err != nil {
		log.Println(err)
	}

	// The Zipkin reporter takes collected spans from the app and reports them to the backend
	// http://localhost:9411/api/v2/spans is the default for the Zipkin Span v2
	reporter := zhttp.NewReporter("http://localhost:9411/api/v2/spans")
	defer reporter.Close()

	// The OpenCensus exporter wraps the Zipkin reporter
	exporter := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(exporter)

	// For example purposes, sample every trace.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	ctx := context.Background()
	foo(ctx)
	time.Sleep(15 * time.Second)
	os.Exit(0)
	/* 	log.Println("trace config")
	   	go func() {
	   		http.Handle("/debug/", http.StripPrefix("/debug", zpages.Handler))
	   		log.Fatal(http.ListenAndServe(":8081", nil))
	   	}()  */

	/* 	e.natsEP, err = e.getConfig("nats")
	   	if err != nil {
	   		log.Fatalf("Error while getting queing system %s", err)
	   	}

	   	e.storageEP, err = e.getConfig("storage")
	   	if err != nil {
	   		log.Fatalf("Error while getting  storage %s", err)
	   	}

	   	var opts []grpc.ServerOption
	   	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", e.port))
	   	if err != nil {
	   		log.Printf("error is %s", err)
	   		panic(err)
	   	}

	   	//Set up a new server with the OpenCensus
	   	// stats handler to enable stats and tracing.

	   	opts = append(opts, grpc.StreamInterceptor(prom.StreamServerInterceptor))
	   	opts = append(opts, grpc.UnaryInterceptor(prom.UnaryServerInterceptor))
	   	opts = append(opts, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	   	e.grpcServer = grpc.NewServer(opts...)
	   	pb.RegisterEventsSenderServer(e.grpcServer, e)
	   	reflection.Register(e.grpcServer)

	   	e.startMonitoring()

	   	go e.grpcServer.Serve(lis)
	   	log.Printf("server %+v", e)
	   	e.register() */

}

func foo(ctx context.Context) {
	log.Println("je asse ici")
	// Name the current span "/foo"
	ctx, span := trace.StartSpan(ctx, "/foo")
	defer span.End()

	// Foo calls bar and baz
	bar(ctx)
	baz(ctx)
}

func bar(ctx context.Context) {
	log.Println("yeaa")
	ctx, span := trace.StartSpan(ctx, "/bar")
	defer span.End()

	// Do bar
	time.Sleep(2 * time.Millisecond)
}

func baz(ctx context.Context) {
	log.Print("bang")
	ctx, span := trace.StartSpan(ctx, "/baz")
	defer span.End()

	// Do baz
	time.Sleep(4 * time.Millisecond)
}

func (e *eventServer) deregister() {
	log.Println("deregistering from registry")
	config := consul.Config{Address: e.serviceRegistry}

	// Get Service Endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Println("impossible to access configuration store %s", e.serviceRegistry)
	}
	agent := client.Agent()
	err = agent.ServiceDeregister("Event-Server1")
	if err != nil {
		log.Print(err)
	}

}

func (e *eventServer) register() {
	config := consul.Config{Address: e.serviceRegistry}

	// Get Service Endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Println("impossible to access configuration store %s", e.serviceRegistry)
	}
	host, _ := os.Hostname()
	s := consul.AgentService{ID: "Event-Server1", Service: "Event-Server", Port: e.port, Tags: []string{"events"}}
	c := consul.CatalogRegistration{Node: host, Address: host, Datacenter: "dc1", Service: &s}

	catalog := client.Catalog()
	_, err = catalog.Register(&c, nil)
	if err != nil {
		log.Print(err)
	}
}

func (e *eventServer) getConfig(serviceName string) (endpoint string, err error) {
	config := consul.Config{Address: e.serviceRegistry}
	log.Printf("looking for %s on %s \n", serviceName, e.serviceRegistry)
	client, err := consul.NewClient(&config)
	if err != nil {
		return "", err
	}
	catalog := client.Catalog()

	queueServices, _, err := catalog.Service(serviceName, "", &consul.QueryOptions{})
	if len(queueServices) == 0 {
		return "", fmt.Errorf("no  services with key %s found", serviceName)
	}
	log.Printf("%s:%d", queueServices[0].Address, queueServices[0].ServicePort)
	return fmt.Sprintf("%s:%d", queueServices[0].Address, queueServices[0].ServicePort), nil

}

func (e *eventServer) startMonitoring() {
	msgSentCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_sent",
		Help: "Number of message sent to customer."},
		[]string{"clientID", "groupID", "eventID", "status"},
	)
	err := prometheus.Register(msgSentCounter)
	if err != nil {
		log.Println("MSG SENT counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}
	msgRCVCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_receive",
		Help: "Number of message receive customer."},
		[]string{"clientID", "groupID", "eventID", "status"},
	)
	err = prometheus.Register(msgRCVCounter)
	if err != nil {
		log.Println("MSG RCV counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}
	go func() {
		for m := range msgSent {
			msgSentCounter.WithLabelValues(m.clientID, m.groupID, m.eventID, m.status).Inc()
		}
	}()
	go func() {
		for m := range msgRCV {
			msgRCVCounter.WithLabelValues(m.clientID, m.groupID, m.eventID, m.status).Inc()
		}
	}()
	prom.Register(e.grpcServer)
	http.Handle("/metrics", prometheus.Handler())
	e.prom = &http.Server{Addr: fmt.Sprintf(":%d", e.monitoringPort)}
	go func() {
		if err := e.prom.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

}

func (e *eventServer) stopMonitoring() {
	e.prom.Shutdown(nil)
}
