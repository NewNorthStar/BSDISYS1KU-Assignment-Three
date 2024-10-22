package main

import (
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ChittyChatServer struct {
	proto.UnimplementedChittyChatServiceServer
	clients []grpc.ServerStreamingServer[proto.Message]
}

func (s *ChittyChatServer) JoinMessageBoard(confirm *proto.Confirm, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("JoinMessageBoard call: %v\n", confirm)

	err := stream.Send(&proto.Message{
		Content:   "Welcome to ChittyChat, " + confirm.Author + "!",
		Author:    "ChittyService",
		LamportTs: 123,
	})
	if err != nil {
		log.Printf("JoinMessageBoard error: %v\n", err)
		return err
	}

	log.Printf("Good client, adding\n")
	s.clients = append(s.clients, stream)

	log.Printf("Broadcasting join message...")
	// Else we start a broadcast
	for index, cli := range s.clients {
		err := cli.Send(&proto.Message{
			Content:   confirm.Author + " has arrived!",
			Author:    "ChittyService",
			LamportTs: 123,
		})
		if err != nil {
			// Remove unresponsive client stream.
			s.clients = append(s.clients[:index], s.clients[index+1:]...)
		}
	}
	return nil
}

func (s *ChittyChatServer) PostMessage(ctx context.Context, in *proto.Message) (*proto.Confirm, error) {
	return &proto.Confirm{
		Author:    "ChittyService",
		LamportTs: 123,
	}, nil
}

func main() {
	server := ChittyChatServer{
		clients: make([]grpc.ServerStreamingServer[proto.Message], 0),
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
