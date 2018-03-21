package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	consul "github.com/armon/consul-api"
	pb "github.com/grey-noise/eventsgrids/events"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/minio/cli"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/oklog/ulid"
	"github.com/prometheus/client_golang/prometheus"
	schema "github.com/xeipuuv/gojsonschema"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var msgSent = make(chan counter)
var msgRCV = make(chan counter)

type counter struct {
	clientID string
	groupID  string
	eventID  string
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

func (e *eventServer) checkEvent(eventID string) bool {
	config := consul.Config{Address: e.storageEP}

	// Get a new client, with KV endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Fatalf("Can't connect to Consul: check Consul  is running with address %s", e.storageEP)
	}
	kv := client.KV()
	path := "events/" + eventID + "/schema.json"
	log.Println(path)
	pair, _, err := kv.Get(path, nil)
	if err != nil || pair == nil {
		log.Println(err)
		return false
	}

	return true

}

func validate(eventID, endpoint string, json []byte) (valid bool) {

	config := consul.Config{Address: endpoint}

	// Get a new client, with KV endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Printf("Can't connect to Consul: check Consul  is running with address %s", endpoint)
		return false
	}
	kv := client.KV()
	path := "events/" + eventID + "/schema.json"
	log.Printf("validating using schema %s", path)
	pair, _, err := kv.Get(path, nil)
	if err != nil {
		log.Printf("error while getting schema : %v", err)
		return false
	}

	schemaLoader := schema.NewBytesLoader(pair.Value)
	jsonLoader := schema.NewStringLoader(string(json))
	result, err := schema.Validate(schemaLoader, jsonLoader)

	if err != nil {
		log.Printf("error while validating document %v", err)
		return false
	}

	if result.Valid() {
		log.Printf("The document is valid\n")
		return true
	}

	log.Printf("The document is not valid. see errors :\n")
	for _, desc := range result.Errors() {
		log.Printf("- %s\n", desc)
	}
	log.Printf("got an error while validating document")

	return false
}

func (e *eventServer) SendEvents(stream pb.EventsSender_SendEventsServer) error {
	messages := make(chan *pb.Event)
	ctx := stream.Context()

	go func() {
		for {
			select {
			case msg := <-messages:
				log.Printf("receive a call from %s with groupID %s to store an event %s", msg.GetHeader().GetClientId(), msg.GetHeader().GetGroupId(), msg.GetHeader().GetEventid())
				var dst bytes.Buffer
				json.Compact(&dst, msg.GetPayload())
				log.Printf("message is received %+v", msg)
				clientID := msg.GetHeader().GetClientId()
				eventID := msg.GetHeader().GetEventid()
				sc, err := stan.Connect(e.clusterID, clientID, stan.NatsURL(e.natsEP))
				if err != nil {
					log.Printf("Can't connect to nats server %s \n Make sure a NATS Streaming Server %v  is running :%v", e.clusterID, clientID, stan.NatsURL(e.natsEP))
					return
				}
				log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", e.natsEP, e.clusterID, clientID)

				err = sc.Publish(eventID, dst.Bytes())
				if err != nil {
					log.Fatalf("Error during publish: %v\n", err)
				}
				log.Printf("Published [%s] : '%s'\n", eventID, dst.Bytes())
				sc.Close()

			case <-ctx.Done():
				log.Println("ctx done")
				return
			}
		}
	}()

	for {
		in, err := stream.Recv()
		if in == nil {
			break
		}
		msgRCV <- counter{clientID: in.GetHeader().GetClientId(), groupID: in.GetHeader().GetGroupId(), eventID: in.GetHeader().GetEventid()}
		if err == io.EOF {
			close(messages)
			return nil

		}
		if err != nil {
			return err
		}
		var s pb.Status
		if validate(in.GetHeader().GetEventid(), e.storageEP, in.GetPayload()) {
			log.Printf("in sent to backend %+v", in)
			messages <- in
			s = pb.Status{Code: pb.Status_OK, Description: "Valid"}

		} else {
			s = pb.Status{Code: pb.Status_NOK_VALID, Description: "not valid"}
		}
		h := pb.Header{
			ClientId: in.GetHeader().GetClientId(),
			GroupId:  in.GetHeader().GetGroupId(),
			Eventid:  in.GetHeader().GetEventid()}
		c := pb.Cursor{
			Ts: time.Now().Unix(),
			Id: 0}
		a := &pb.Acknowledge{Header: &h, Cursor: &c, Status: &s}

		if err := stream.Send(a); err != nil {
			log.Printf("error while sending to the client")
		}

	}
	return nil
}

func (e *eventServer) GetEvents(a *pb.Acknowledge, stream pb.EventsSender_GetEventsServer) error {
	eventID := a.Header.GetEventid()
	log.Println("looking for event : ", eventID)
	if eventID == "" {
		log.Println("got no event to subscribe")
		return grpc.Errorf(codes.InvalidArgument, "event can not be empty ")

	}
	if !e.checkEvent(eventID) {
		log.Println("event unknown to subscribe")
		return grpc.Errorf(codes.InvalidArgument, "eventid %s is unknown ", eventID)
	}

	clientID := a.Header.GetClientId()
	if clientID == "" {
		log.Println("Go no  X-ClientID generating one")
		entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
		clientID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
	}
	groupID := a.Header.GetGroupId()
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
		h := &pb.Header{}
		c := &pb.Cursor{Id: msg.Sequence, Ts: msg.Timestamp}
		e := &pb.Event{Header: h, Cursor: c, Payload: msg.Data}
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
	ctx := stream.Context()
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
			Value:       "localhost:8500",
			Usage:       "consul endpoint",
			Destination: &s.serviceRegistry,
			EnvVar:      "CONSUL-URL"},
		cli.StringFlag{
			Name:        "clusterid",
			Value:       "cluster-test",
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
	var err error
	e.natsEP, err = e.getConfig("nats")
	if err != nil {
		log.Fatalf("Error while getting queing system %s", err)
	}

	e.storageEP, err = e.getConfig("storage")
	if err != nil {
		log.Fatalf("Error while getting queing storage %s", err)
	}

	var opts []grpc.ServerOption
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", e.port))
	if err != nil {
		log.Printf("error is %s", err)
		panic(err)
	}

	opts = append(opts, grpc.StreamInterceptor(prom.StreamServerInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(prom.UnaryServerInterceptor))
	e.grpcServer = grpc.NewServer(opts...)
	pb.RegisterEventsSenderServer(e.grpcServer, e)
	e.startMonitoring()

	go e.grpcServer.Serve(lis)
	log.Printf("server %+v", e)
	e.register()

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
	c := consul.AgentServiceRegistration{ID: "Event-Server1", Name: "Event-Server", Port: 8000}

	agent := client.Agent()
	err = agent.ServiceRegister(&c)
	if err != nil {
		log.Print(err)
	}
}

func (e *eventServer) getConfig(serviceName string) (endpoint string, err error) {
	config := consul.Config{Address: e.serviceRegistry}

	client, err := consul.NewClient(&config)
	if err != nil {
		return "", err
	}
	catalog := client.Catalog()

	queueServices, _, err := catalog.Service(serviceName, "", &consul.QueryOptions{})
	if len(queueServices) == 0 {
		return "", fmt.Errorf("no nats services with key %s found", serviceName)
	}
	return fmt.Sprintf("%s:%d", queueServices[0].Address, queueServices[0].ServicePort), nil

}

func (e *eventServer) startMonitoring() {
	msgSentCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_sent",
		Help: "Number of message sent to customer."},
		[]string{"clientID", "groupID", "eventID"},
	)
	err := prometheus.Register(msgSentCounter)
	if err != nil {
		log.Println("MSG SENT counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}
	msgRCVCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_receive",
		Help: "Number of message receive customer."},
		[]string{"clientID", "groupID", "eventID"},
	)
	err = prometheus.Register(msgRCVCounter)
	if err != nil {
		log.Println("MSG RCV counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}
	go func() {
		for m := range msgSent {
			msgSentCounter.WithLabelValues(m.clientID, m.groupID, m.eventID).Inc()
		}
	}()
	go func() {
		for m := range msgRCV {
			msgRCVCounter.WithLabelValues(m.clientID, m.groupID, m.eventID).Inc()
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
