package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	pb "github.com/grey-noise/eventsgrids/events"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var serverEP string
var timeout *int
var request *int

func main() {
	app := cli.NewApp()
	log.Println("\U0001f197")
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "endpoint",
			Value:       "localhost:8080",
			Usage:       "server endpoint",
			Destination: &serverEP},
		cli.IntFlag{
			Name:        "timeout",
			Value:       1000,
			Usage:       "time out while listening",
			Destination: timeout},
	}

	app.Commands = []cli.Command{
		{
			Name:    "sent",
			Aliases: []string{"s"},
			Usage:   "send toward the servers events",
			Flags: []cli.Flag{
				cli.IntFlag{Name: "request",
					Value:       50,
					Destination: request},
			},
			Action: func(c *cli.Context) error {
				sendEvents()
				log.Println("number of request")
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func sendEvents() {
	i := 0
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverEP, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventsSenderClient(conn)
	events := []*pb.Event{
		{Header: &pb.Header{ClientId: "123", Eventid: "01C8QCMRNP82WBK2PVFRK9VABV", GroupId: "eeee"}, Payload: []byte("{\"event-id\" : \"dddd\",\"event-type\": \"notification/creation\",\"api-version\":\"1.0.1\",\"content-type\":\"application/json\",\"created-at\":\"now\",\"source\":\"application/e-box\",\"body\": {\"object\" : \"122222\",\"action\": \"update\"}}")},
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	stream, err := client.SendEvents(ctx)
	if err != nil {
		log.Fatalf("%v.SendEvents(_) = _, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive an akckonwledgement : %v", err)
			}
			if in.Status == pb.Status_OK {
				log.Printf("ack : %+v", "", in)
			}

		}
	}()
	for _, event := range events {
		i++
		log.Println(i)
		if err := stream.Send(event); err != nil {
			log.Fatalf("Failed to send a note: %v", err)
		}
	}

	<-waitc
	stream.CloseSend()
	return
}
