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
var ctx context.Context = context.Background()

var name string
var lamportTime int64

// Start point for program.
func main() {
	fmt.Print("Enter your callsign and press ENTER: ")
	name = nextLine()

	runChatService()
}

// Overall method for running the chat service.
func runChatService() {
	conn := getConnectionToServer()
	defer conn.Close()

	client := proto.NewChittyChatServiceClient(conn)
	stream := joinChatBoard(client)

	go pollStream(stream)
	handleUserInput(client)
}

// Establishes connection to the server.
func getConnectionToServer() *grpc.ClientConn {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to obtain connection: %v", err)
	}
	return conn
}

// Obtains a proto.Confirm to send.
func confirmMessage() *proto.Confirm {
	return &proto.Confirm{
		Author:    name,
		LamportTs: getTime(),
	}
}

// Joins the chat board. Returns a stream of posted messages from the chat.
func joinChatBoard(client proto.ChittyChatServiceClient) grpc.ServerStreamingClient[proto.Message] {
	stream, err := client.JoinMessageBoard(ctx, confirmMessage())
	if err != nil {
		log.Fatalf("Failed to obtain stream: %v", err)
	}
	return stream
}

// Go-routine for polling stream from server and displaying messages from the chat board.
func pollStream(stream grpc.ServerStreamingClient[proto.Message]) {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Stream closed. Bye!")
			break
		} else if err != nil {
			log.Fatal(err)
		}
		setTime(msg.LamportTs)
		printMessage(msg)
	}
}

// Main routine for accepting terminal input as messages to be posted.
func handleUserInput(client proto.ChittyChatServiceClient) {
	for {
		input := nextLine()
		if len(input) == 0 {
			continue
		}
		msg := proto.Message{
			Content:   input,
			Author:    name,
			LamportTs: getTime(),
		}
		confirm, err := client.PostMessage(ctx, &msg)
		if err != nil {
			log.Fatal(err)
		}
		setTime(confirm.LamportTs)
	}
}

// Gets the next Lamport timestamp.
func getTime() int64 {
	lamportTime++
	return lamportTime
}

// Updates the Lamport timestamp to reflect an incoming timestamp.
func setTime(in int64) {
	if lamportTime < in {
		lamportTime = in
	}
}

// Prints the standard chat message format to console.
func printMessage(message *proto.Message) {
	fmt.Printf("%d %s: %s\n", message.LamportTs, message.Author, message.Content)
}

// Setup for stdIn (input from console). Any scanner settings go here.
func setScanner() *bufio.Scanner {
	var sc = bufio.NewScanner(os.Stdin)
	return sc
}

// Obtains next line from stdIn.
func nextLine() string {
	stdIn.Scan()
	return stdIn.Text()
}
