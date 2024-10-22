package main

import (
	"bufio"
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"io"
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

	runChatService()
}

func runChatService() {
	conn := getConnectionToServer()
	defer conn.Close()
	client := proto.NewChittyChatServiceClient(conn)
	ctx := context.Background()

	stream, err := client.JoinMessageBoard(ctx, confirmMessage())
	if err != nil {
		log.Fatalf("Failed to obtain stream: %v", err)
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream closed. Bye!")
			return
		} else if err != nil {
			log.Fatal(err)
		}
		printMessage(msg)
	}
}

// Establishes connection to the server.
func getConnectionToServer() *grpc.ClientConn {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to obtain connection: %v", err)
	}
	return conn
}

// Obtains a proto.Confirm to send to the server.
func confirmMessage() *proto.Confirm {
	return &proto.Confirm{
		Author:    name,
		LamportTs: -1,
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
