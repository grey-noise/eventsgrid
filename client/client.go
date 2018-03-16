package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/tixu/events-web-receivers-grpc/events"
	"google.golang.org/grpc"
)

func main() {
	serverEndpoint := getEnv("GRPC-SERVER-URL", "localhost:8080")
	timeout, err := strconv.Atoi(getEnv("GRPC-CLIENT-TIMEOUT", "1000"))
	if err != nil {
		log.Printf("error is %s", err)
		timeout = 60
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverEndpoint, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventsSenderClient(conn)
	h := pb.Header{
		ClientId: "xavier",
		GroupId:  "ea",
		Eventid:  "8b2156ff-89df-4075-8316-4f626f521fad"}
	c := pb.Cursor{
		Ts: time.Now().Unix(),
		Id: 0}
	a := &pb.Acknowledge{Header: &h, Cursor: &c}
	printEvents(client, a, timeout)

}

// printFeatures lists all the features within the given bounding Rectangle.
func printEvents(client pb.EventsSenderClient, a *pb.Acknowledge, timeout int) {
	log.Printf("Looking for events within %v", a)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	stream, err := client.ListFeatures(ctx, a)
	if err != nil {
		log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.ListFeatures(_) = _, %v", client, err)
		}
		log.Println(feature)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		log.Printf("%s:%s", key, value)
		return value
	}
	log.Printf("%s:%s", key, fallback)
	return fallback
}
