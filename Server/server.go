package main

import (
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Represents a running ChittyChat server.
type ChittyChatServer struct {
	proto.UnimplementedChittyChatServiceServer
	clients     []client
	name        string
	lamportTime int64
}

// Channels for the connection to a client.
// Each connection is a running coroutine.
type client struct {
	name     string
	feed     chan *proto.Message
	isClosed chan bool
}

// The client obtains a stream of the chat. Method returns when the stream terminates.
func (s *ChittyChatServer) JoinMessageBoard(confirm *proto.Confirm, stream grpc.ServerStreamingServer[proto.Message]) error {
	s.setTime(confirm.LamportTs)
	log.Printf("%d JoinMessageBoard: %v\n", s.getTime(), confirm)

	err := s.welcomeClient(stream, confirm.Author)
	if err != nil {
		return err
	}

	cli := s.addNewClient(confirm)
	s.enteredChatMessage(cli.name)

	cli.streamToClientRoutine(stream) // Continues until connection terminates.

	s.leftChatMessage(cli.name)

	return nil
}

// The incoming message is broadcasted; queued in the feed of all clients.
// The server returns a confirm message with a timestamp.
func (s *ChittyChatServer) PostMessage(ctx context.Context, in *proto.Message) (*proto.Confirm, error) {
	s.setTime(in.LamportTs)
	log.Printf("%d PostMessage: %v\n", s.getTime(), in)

	s.broadcastMessage(in)
	return &proto.Confirm{
		Author:    s.name,
		LamportTs: s.getTime(),
	}, nil
}

func (s *ChittyChatServer) enteredChatMessage(name string) {
	time := s.getTime()
	s.broadcastMessage(&proto.Message{
		Content:   "Participant " + name + " joined Chitty-Chat at Lamport time " + strconv.FormatInt(time, 10),
		Author:    s.name,
		LamportTs: time,
	})
}

func (s *ChittyChatServer) leftChatMessage(name string) {
	time := s.getTime()
	s.broadcastMessage(&proto.Message{
		Content:   "Participant " + name + " left Chitty-Chat at Lamport time " + strconv.FormatInt(time, 10),
		Author:    s.name,
		LamportTs: time,
	})
}

// Sends an initial message to client and returns nil.
// If an error occurs, it is logged, and a status error for the RPC is returned.
func (s *ChittyChatServer) welcomeClient(stream grpc.ServerStreamingServer[proto.Message], name string) error {
	msg := proto.Message{
		Content:   "\U0001F680 Welcome to ChittyChat, " + name + "! \U0001F680",
		Author:    s.name,
		LamportTs: s.getTime(),
	}
	err := stream.Send(&msg)
	if err != nil {
		log.Printf("Handshake error: %v\n", err)
		return status.Error(codes.Aborted, err.Error())
	} else {
		return nil
	}
}

// Add a new channel struct for control and feed from server to active client stream.
func (s *ChittyChatServer) addNewClient(confirm *proto.Confirm) client {
	cli := client{
		name:     confirm.Author,
		feed:     make(chan *proto.Message, 20),
		isClosed: make(chan bool, 1),
	}
	s.clients = append(s.clients, cli)
	return cli
}

// Routine call that handles the stream to a client.
// Runs for the duration of each client connection.
func (cli *client) streamToClientRoutine(stream grpc.ServerStreamingServer[proto.Message]) {
	done := stream.Context().Done()

main:
	for {
		select {
		case message := <-cli.feed:
			err := stream.Send(message)
			if err != nil {
				log.Printf("Client '%s': Stream error. Closing...\n", cli.name)
				break main
			}
		case <-done:
			log.Printf("Client '%s': Stream terminated. Closing...\n", cli.name)
			break main
		}
	}

	cli.isClosed <- true
}

// Adds a message to the feed channel of each client connection.
// Closed connections are pruned as messages are sent.
func (s *ChittyChatServer) broadcastMessage(message *proto.Message) {
	message.LamportTs = s.getTime()
	for i := 0; i < len(s.clients); i++ {
		cli := s.clients[i]
		select {
		case <-cli.isClosed:
			s.removeClient(i)
			i--
		case cli.feed <- message:
		default:
			log.Printf("Warning: Feed overflow to %s\n", cli.name)
		}
	}
}

// Dereferences a clients slices from the channel.
// Only do this when communication to the client has been terminated.
func (s *ChittyChatServer) removeClient(i int) {
	cli := s.clients[i]
	s.clients = append(s.clients[:i], s.clients[i+1:]...)
	log.Printf("Client '%s': Removed from connections.\n", cli.name)
}

// Gets the next Lamport timestamp.
func (s *ChittyChatServer) getTime() int64 {
	s.lamportTime++
	return s.lamportTime
}

// Updates the Lamport timestamp to reflect an incoming timestamp.
func (s *ChittyChatServer) setTime(in int64) {
	if s.lamportTime < in {
		s.lamportTime = in
	}
}

// Start point for program.
func main() {
	server := ChittyChatServer{
		clients: make([]client, 0),
		name:    "ChittyServer",
	}

	listener := listenOn("localhost:5050")
	server.startService(listener)

	wait := make(chan bool)
	<-wait // Run until terminated manually or by error.
}

// Obtains a TCP listener on a given network address.
func listenOn(address string) net.Listener {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	return listener
}

// Begins serving ChittyChat gRPC service as a goroutine and returns.
func (s *ChittyChatServer) startService(listener net.Listener) {
	grpcServer := grpc.NewServer()
	proto.RegisterChittyChatServiceServer(grpcServer, s)
	fmt.Printf("server listening at %v\n", listener.Addr())
	err := grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
