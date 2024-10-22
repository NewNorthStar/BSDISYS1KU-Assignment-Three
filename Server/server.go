package main

import (
	"context"
	proto "example/chittychat/grpc"
	"fmt"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
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
	log.Printf("JoinMessageBoard call: %v\n", confirm)

	err := stream.Send(&proto.Message{
		Content:   "Welcome to ChittyChat, " + confirm.Author + "!",
		Author:    s.name,
		LamportTs: getTime(),
	})
	if err != nil {
		log.Printf("JoinMessageBoard error: %v\n", err)
		return err
	}

	log.Printf("Good client, adding\n")
	s.clients = append(s.clients, stream)
	// Else we start a broadcast
	log.Printf("Broadcasting join message...")
	s.broadcastMessage(&proto.Message{
		Content:   "Participant " + confirm.Author + " joined Chitty-Chat at Lamport time " + strconv.Itoa(int(getTime())),
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

func getTime() int64 {
	return 123
}

func main() {
	server := ChittyChatServer{
		clients: make([]grpc.ServerStreamingServer[proto.Message], 0),
		name:    "ChittyService",
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
