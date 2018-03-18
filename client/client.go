package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/tixu/events-web-receivers-grpc/events"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var serverEndpoint string
var timeout string

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "endpoint",
			Value:       "localhost:8080",
			Usage:       "server endpoint",
			Destination: &serverEndpoint},
		cli.StringFlag{
			Name:        "timeout",
			Value:       "1000",
			Usage:       "time out while listening",
			Destination: &timeout},
	}

	app.Commands = []cli.Command{
		{
			Name:    "listen",
			Aliases: []string{"c"},
			Usage:   "listen for events",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "clientID", EnvVar: "CLIENTID", Value: "ddd"},
				cli.StringFlag{Name: "groupID", EnvVar: "GROUPID", Value: "grp"},
				cli.StringFlag{Name: "eventID", EnvVar: "EVENTID", Value: "01C8QCMRNP82WBK2PVFRK9VABV"},
			},
			Action: func(c *cli.Context) error {
				log.Printf("client id : %s", c.String("clientID"))
				log.Printf("group id : %s", c.String("groupID"))
				log.Printf("event id : %s", c.String("eventID"))
				return listen(c.String("eventID"), c.String("clientID"), c.String("groupID"))
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func listen(eventID, clientID, groupID string) error {
	to, err := strconv.Atoi(timeout)
	if err != nil {
		log.Printf("error is %s", err)
		to = 60
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
		ClientId: clientID,
		GroupId:  groupID,
		Eventid:  eventID}
	c := pb.Cursor{
		Ts: time.Now().Unix(),
		Id: 0}
	a := &pb.Acknowledge{Header: &h, Cursor: &c}
	printEvents(client, a, to)
	return nil
}

// printFeatures lists all the features within the given bounding Rectangle.
func printEvents(client pb.EventsSenderClient, a *pb.Acknowledge, timeout int) {
	log.Printf("Looking for events within %v", a)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	stream, err := client.GetEvents(ctx, a)
	if err != nil {
		log.Fatalf("%v.(_) = _, %v", client, err)
	}
	for {
		feature, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.GetEvents(_) = _, %v", client, err)
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
