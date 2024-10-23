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

type ChittyChatServer struct {
	proto.UnimplementedChittyChatServiceServer
	clients []grpc.ServerStreamingServer[proto.Message]
	name    string
}

func (s *ChittyChatServer) broadcastMessage(message *proto.Message) {
	for index, cli := range s.clients {
		err := cli.Send(message)
		if err != nil {
			// Remove unresponsive client stream.
			// TODO: add log message
			s.clients = append(s.clients[:index], s.clients[index+1:]...)
		}
	}
}

func (s *ChittyChatServer) JoinMessageBoard(confirm *proto.Confirm, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("Incoming client: %v\n", confirm)

	if err := s.handshake(stream, confirm.Author); err != nil {
		return err
	}

	log.Printf("Good client, adding\n")
	s.clients = append(s.clients, stream)
	// Else we start a broadcast
	log.Printf("Broadcasting join message...")
	s.broadcastMessage(&proto.Message{
		Content:   "Everyone please welcome '" + confirm.Author + "' to the chat!",
		Author:    s.name,
		LamportTs: getTime(),
	})
	log.Printf("JoinMessageBoard was executed successfully.\n")
	return nil
}

func (s *ChittyChatServer) PostMessage(ctx context.Context, in *proto.Message) (*proto.Confirm, error) {
	s.broadcastMessage(&proto.Message{})
	return &proto.Confirm{
		Author:    s.name,
		LamportTs: getTime(),
	}, nil
}

// Sends an initial message to client and returns nil.
// If an error occurs, it is logged, and a status error for the RPC is returned.
func (s *ChittyChatServer) handshake(stream grpc.ServerStreamingServer[proto.Message], name string) error {
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

func getTime() int64 {
	return 123
}

func main() {
	server := ChittyChatServer{
		clients: make([]grpc.ServerStreamingServer[proto.Message], 0),
		name:    "ChittyServer",
	}
	go server.start()

	wait := make(chan bool)
	<-wait
}

func (s *ChittyChatServer) start() {
	listener, err := net.Listen("tcp", "localhost:5050")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterChittyChatServiceServer(grpcServer, s)
	fmt.Printf("server listening at %v", listener.Addr())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
