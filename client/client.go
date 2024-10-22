package main

import (
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"strconv"

	"golang.org/x/exp/rand"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type client struct {
	name string
}

func main() {
	client := client{
		name: strconv.Itoa(rand.Intn(100000)),
	}

	client.run()
	// wait := make(chan bool)
	// <-wait

}

func (c *client) run() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not work")
	}
	defer conn.Close()

	client := proto.NewChittyChatServiceClient(conn)

	ctx := context.Background()
	stream, err := client.JoinMessageBoard(ctx, &proto.Confirm{
		Author:    c.name,
		LamportTs: -1,
	})
	if err != nil {
		log.Fatalf("Did not work2")
	}

	for {
		var message proto.Message
		err = stream.RecvMsg(&message)
		if err != nil {
			log.Fatal(err)
		}
		printMessage(&message)
	}
}

func printMessage(message *proto.Message) {
	fmt.Printf("%d %s: %s\n", message.LamportTs, message.Author, message.Content)
}
