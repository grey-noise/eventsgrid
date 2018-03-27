package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	consul "github.com/armon/consul-api"
	pb "github.com/grey-noise/eventsgrids/events"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var registryEP string
var timeout *int
var request *int

func main() {
	app := cli.NewApp()
	log.Println("\U0001f197")
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "serviceRegistry",
			Value:       "consul-ea.cloud.smals.be",
			Usage:       "service registry",
			Destination: &registryEP},
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
				cli.StringFlag{Name: "clientID", EnvVar: "CLIENTID", Value: "ddd"},
				cli.StringFlag{Name: "groupID", EnvVar: "GROUPID", Value: "grp"},
				cli.StringFlag{Name: "eventID", EnvVar: "EVENTID", Value: "55f464c8-a702-45e8-ad78-ddab4892e8e2"},
			},
			Action: func(c *cli.Context) error {
				sc := sconn{ClientID: c.String("clientID"), GroupID: c.String("groupID"), EventID: c.String("eventID")}
				sc.sendEvents(1000)
				log.Println("number of request")
				return nil
			},
		}, {
			Name:    "listen",
			Aliases: []string{"c"},
			Usage:   "listen for events",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "clientID", EnvVar: "CLIENTID", Value: "ddd"},
				cli.StringFlag{Name: "groupID", EnvVar: "GROUPID", Value: "grp"},
				cli.StringFlag{Name: "eventID", EnvVar: "EVENTID", Value: "c41666ad-a1c0-4d5e-ae51-30c58d42deb7"},
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

type sconn struct {
	ClientID string
	GroupID  string
	EventID  string
}

func listen(eventID, clientID, groupID string) error {
	serverEp, err := discover(registryEP, "Event-Server")
	if err != nil {
		log.Fatalf("error while looking up service")
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(serverEp, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventsSenderClient(conn)
	c := pb.Cursor{
		Ts: time.Now().Unix(),
		Id: 0}
	a := &pb.Acknowledge{Cursor: &c}
	con := sconn{ClientID: clientID, GroupID: groupID, EventID: eventID}
	con.printEvents(client, a, 100)
	return nil
}

// printFeatures lists all the features within the given bounding Rectangle.
func (c sconn) printEvents(client pb.EventsSenderClient, a *pb.Acknowledge, timeout int) {
	log.Printf("Looking for events within %v", a)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	header := metadata.New(map[string]string{"clientID": c.ClientID,
		"groupID": c.GroupID,
		"eventID": c.EventID})
	sctx := metadata.NewOutgoingContext(ctx, header)

	defer cancel()

	stream, err := client.GetEvents(sctx, a)
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

func (c sconn) sendEvents(timeout int) {

	serverEp, err := discover(registryEP, "Event-Server")
	if err != nil {
		log.Fatalf("error while looking up service")
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	log.Printf("serverEP:%s", serverEp)
	conn, err := grpc.Dial(serverEp, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewEventsSenderClient(conn)

	waitc := make(chan struct{})
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)

	header := metadata.New(map[string]string{"clientID": c.ClientID,
		"groupID": c.GroupID,
		"eventID": c.EventID})
	sctx := metadata.NewOutgoingContext(ctx, header)
	stream, err := client.SendEvents(sctx)
	if err != nil {
		log.Fatalf("%v.SendEvents(_) = _, %v", client, err)
	}

	go func() {
		events := []*pb.Event{
			{Payload: []byte("{\"function\" : \"dddd\"}")},
			//			{Payload: []byte("{\"function\" : \"edddd\"}")},
		}

		go func() {
			for _, event := range events {
				log.Printf("sending event %v", event)
				if err := stream.Send(event); err != nil {
					log.Fatalf("Failed to send a note: %v", err)
				}
				log.Printf("sent event %v", event)
			}
			if err := stream.CloseSend(); err != nil {
				log.Print(err)
			}
		}()

	}()

	go func() {
		log.Println("receiving message")
		for {
			log.Println("waiting  message")

			in, err := stream.Recv()
			log.Print("received")
			if err == io.EOF {
				// read done.
				close(waitc)
				break
			}
			if err != nil {
				log.Fatalf("Failed to receive an akckonwledgement : %v", err)
				break
			}
			log.Printf("receive %+v", in)

		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(waitc)
	}()
	<-waitc
	return
}

func discover(registryEP, serviceName string) (endpoint string, err error) {
	log.Printf("looking up for service %s \n", serviceName)
	config := consul.Config{Address: registryEP}

	client, err := consul.NewClient(&config)
	if err != nil {
		return "", err
	}
	log.Println("got a client")
	catalog := client.Catalog()
	log.Println("got a catalog")
	services, _, err := catalog.Service(serviceName, "", &consul.QueryOptions{})
	if len(services) == 0 {
		return "", fmt.Errorf("no  services with key %s found", serviceName)
	}
	log.Printf("%s:%d", services[0].Address, services[0].ServicePort)
	return fmt.Sprintf("%s:%d", services[0].Address, services[0].ServicePort), nil

}
