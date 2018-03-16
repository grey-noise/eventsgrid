package main

import (
	"fmt"
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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var msgSent = make(chan counter)

type counter struct {
	clientID string
	groupID  string
}

type eventServer struct {
	natsEP    string
	clusterID string
	consulEP  string
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

// ListFeatures implements

func (e *eventServer) ListFeatures(a *pb.Acknowledge, stream pb.EventsSender_ListFeaturesServer) error {
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
		msgSent <- counter{clientID: clientID, groupID: groupID}
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

	return nil
}

func newServer() *eventServer {
	natsEndpoint := getEnv("NATS-URL", "nats://open-faas.cloud.smals.be:4223")
	clusterID := getEnv("CLUSTER-ID", "test-cluster")
	//eventID = getEnv("EVENT-ID", "ae1db066-2153-11e8-b467-0ed5f89f718b")
	consulEndpoint := getEnv("CONSUL-URL", "open-faas.cloud.smals.be:8500")
	monitoring := getEnv("MONITORING-URL", "localhost:8000")
	return &eventServer{natsEP: natsEndpoint, clusterID: clusterID, consulEP: consulEndpoint}

}
func main() {

	msgCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "message_sent",
		Help: "Number of message sent to customer."},
		[]string{"clientID", "groupID"},
	)
	err := prometheus.Register(msgCounter)
	if err != nil {
		log.Println("MSG counter couldn't be registered AGAIN, no counting will happen:", err)
		return
	}

	go func() {
		for m := range msgSent {
			msgCounter.WithLabelValues(m.clientID, m.groupID).Inc()
		}
	}()
	var opts []grpc.ServerOption
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", 8080))
	if err != nil {
		log.Printf("error is %s", err)
		panic(err)
	}
	opts = append(opts, grpc.StreamInterceptor(prom.StreamServerInterceptor))
	opts = append(opts, grpc.UnaryInterceptor(prom.UnaryServerInterceptor))
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterEventsSenderServer(grpcServer, newServer())
	prom.Register(grpcServer)

	http.Handle("/metrics", prometheus.Handler())
	h := &http.Server{Addr: ":8000"}
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
