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
	messageLog []string
}

func (s *ChittyChatServer) JoinMessageBoard(*proto.Confirm, grpc.ServerStreamingServer[proto.Message]) error {
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
		messageLog: make([]string, 0),
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
