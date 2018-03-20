package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	consul "github.com/armon/consul-api"
	prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	stan "github.com/nats-io/go-nats-streaming"
	"github.com/oklog/ulid"
	"github.com/prometheus/client_golang/prometheus"
	pb "github.com/tixu/events-web-receivers-grpc/events"
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
	natsEP       string
	clusterID    string
	consulEP     string
	monitoringEP string
	address      string
}

func (e *eventServer) checkEvent(eventID string) bool {
	config := consul.Config{Address: e.consulEP}

	// Get a new client, with KV endpoints
	client, err := consul.NewClient(&config)
	if err != nil {
		log.Fatalf("Can't connect to Consul: check Consul  is running with address %s", e.consulEP)
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

// ListFeatures implements
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
		if validate(in.GetHeader().GetEventid(), e.consulEP, in.GetPayload()) {
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

func newServer() *eventServer {

	natsEndpoint := getEnv("NATS-URL", "nats://open-faas.cloud.smals.be:4222")
	clusterID := getEnv("CLUSTER-ID", "test-cluster")
	consulEndpoint := getEnv("CONSUL-URL", "consul-ea.cloud.smals.be")
	monitoring := getEnv("MONITORING-URL", "localhost:8000")
	address := getEnv("URL", "localhost:8080")
	return &eventServer{natsEP: natsEndpoint, clusterID: clusterID, consulEP: consulEndpoint, monitoringEP: monitoring, address: address}

}
func main() {

	msgSentCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_sent",
		Help: "Number of message sent to customer."},
		[]string{"clientID", "groupID", "eventID"},
	)

	err := prometheus.Register(msgSentCounter)
	if err != nil {
		log.Println("MSG counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}

	go func() {
		for m := range msgSent {
			msgSentCounter.WithLabelValues(m.clientID, m.groupID, m.eventID).Inc()
		}
	}()
	s := newServer()
	var opts []grpc.ServerOption
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Printf("error is %s", err)
		panic(err)
	}
	msgRCVCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_receive",
		Help: "Number of message receive customer."},
		[]string{"clientID", "groupID", "eventID"},
	)

	go func() {
		for m := range msgRCV {
			msgRCVCounter.WithLabelValues(m.clientID, m.groupID, m.eventID).Inc()
		}
	}()

	opts = append(opts, grpc.StreamInterceptor(prom.StreamServerInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(prom.UnaryServerInterceptor))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventsSenderServer(grpcServer, s)
	prom.Register(grpcServer)

	http.Handle("/metrics", prometheus.Handler())
	h := &http.Server{Addr: s.monitoringEP}
	go func() {
		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	log.Fatal(grpcServer.Serve(lis))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		log.Printf("%s:%s", key, value)
		return value
	}
	log.Printf("%s:%s", key, fallback)
	return fallback
}
