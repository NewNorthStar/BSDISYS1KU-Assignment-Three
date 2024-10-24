package main

import (
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Represents a running ChittyChat server.
type ChittyChatServer struct {
	proto.UnimplementedChittyChatServiceServer
	clients []client
	name    string
}

// Channels for the connection to a client.
// Each connection is a running coroutine.
type client struct {
	name   string
	feed   chan *proto.Message
	closed chan bool
}

// The client obtains a stream of the chat. Method returns when the stream terminates.
func (s *ChittyChatServer) JoinMessageBoard(confirm *proto.Confirm, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("Incoming client: %v\n", confirm)

	err := s.welcomeClient(stream, confirm.Author)
	if err != nil {
		return err
	}

	cli := s.addNewClient(confirm)

	s.broadcastMessage(&proto.Message{
		Content:   "Everyone please welcome '" + confirm.Author + "' to the chat!",
		Author:    s.name,
		LamportTs: getTime(),
	})

	cli.streamToClientRoutine(stream) // Continues until connection terminates.

	log.Printf("Terminated: %v\n", confirm.Author)
	return nil
}

// The incoming message is broadcasted; queued in the feed of all clients.
// The server returns a confirm message with a timestamp.
func (s *ChittyChatServer) PostMessage(ctx context.Context, in *proto.Message) (*proto.Confirm, error) {
	s.broadcastMessage(&proto.Message{})
	return &proto.Confirm{
		Author:    s.name,
		LamportTs: getTime(),
	}, nil
}

// Sends an initial message to client and returns nil.
// If an error occurs, it is logged, and a status error for the RPC is returned.
func (s *ChittyChatServer) welcomeClient(stream grpc.ServerStreamingServer[proto.Message], name string) error {
	err := stream.Send(&proto.Message{
		Content:   "\U0001F680 Welcome to ChittyChat, " + name + "! \U0001F680",
		Author:    s.name,
		LamportTs: getTime(),
	})
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
		name:   confirm.Author,
		feed:   make(chan *proto.Message, 20),
		closed: make(chan bool),
	}
	s.clients = append(s.clients, cli)
	return cli
}

// Routine call that handles the stream to a client.
// Runs for the duration of each client connection.
func (cli *client) streamToClientRoutine(stream grpc.ServerStreamingServer[proto.Message]) {
	log.Printf("Client '%s': stream open.\n", cli.name)
	for {
		select {
		case <-cli.closed:
			return
		case message := <-cli.feed:
			if err := stream.Send(message); err != nil {
				log.Printf("Client '%s': Stream terminated. Will close.\n", cli.name)
				cli.closed <- true
				return
			}
		}
	}
}

// Adds a message to the feed channel of each client connection.
// Closed connections are pruned as messages are sent.
func (s *ChittyChatServer) broadcastMessage(message *proto.Message) {
	for i := 0; i < len(s.clients); i++ {
		cli := s.clients[i]
		select {
		case <-cli.closed:
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

func getTime() int64 {
	return 123
}

// Start point for program.
func main() {
	server := ChittyChatServer{
		clients: make([]client, 0),
		name:    "ChittyServer",
	}
	server.startService()

	wait := make(chan bool)
	<-wait // Run until terminated manually or by error.
}

// Put a TCP listener on the network and begin serving ChittyChat gRPC service on a goroutine.
func (s *ChittyChatServer) startService() {
	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatalf("failed to listen: %v\n", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterChittyChatServiceServer(grpcServer, s)
	fmt.Printf("server listening at %v\n", listener.Addr())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}
