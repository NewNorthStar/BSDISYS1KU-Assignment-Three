package main

import (
	"bufio"
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var stdIn = setScanner()
var name string

func main() {
	fmt.Print("Enter your callsign and press ENTER: ")
	name = nextLine()

	run()
}

func run() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not work")
	}
	defer conn.Close()

	client := proto.NewChittyChatServiceClient(conn)

	ctx := context.Background()
	stream, err := client.JoinMessageBoard(ctx, &proto.Confirm{
		Author:    name,
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

func setScanner() *bufio.Scanner {
	var sc = bufio.NewScanner(os.Stdin)
	return sc
}

func nextLine() string {
	stdIn.Scan()
	return stdIn.Text()
}
