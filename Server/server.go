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
	clients []client
	name    string
}

type client struct {
	name   string
	feed   chan *proto.Message
	closed chan bool
}

func (s *ChittyChatServer) JoinMessageBoard(confirm *proto.Confirm, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("Incoming client: %v\n", confirm)

	if err := s.handshake(stream, confirm.Author); err != nil {
		return err
	}

	cli := client{
		name:   confirm.Author,
		feed:   make(chan *proto.Message, 20),
		closed: make(chan bool),
	}
	s.clients = append(s.clients, cli)
	s.broadcastMessage(&proto.Message{
		Content:   "Everyone please welcome '" + confirm.Author + "' to the chat!",
		Author:    s.name,
		LamportTs: getTime(),
	})

	pushStream(stream, &cli)

	log.Printf("Terminated: %v\n", confirm.Author)
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

func pushStream(stream grpc.ServerStreamingServer[proto.Message], cli *client) {
	for {
		select {
		case <-cli.closed:
			return
		case message := <-cli.feed:
			if err := stream.Send(message); err != nil {
				log.Printf("Stream terminated for %s. Will close.\n", cli.name)
				cli.closed <- true
				return
			}
		}
	}
}

func (s *ChittyChatServer) broadcastMessage(message *proto.Message) {
	for i := 0; i < len(s.clients); i++ {
		cli := s.clients[i]
		select {
		case <-cli.closed:
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			i--
			log.Printf("Removed %s from client slice\n", cli.name)
		case cli.feed <- message:
		default:
			log.Printf("Feed overflow to %s\n", cli.name)
		}
	}
}

func getTime() int64 {
	return 123
}

func main() {
	server := ChittyChatServer{
		clients: make([]client, 0),
		name:    "ChittyServer",
	}
	go server.start()

	wait := make(chan bool)
	<-wait
}

func (s *ChittyChatServer) start() {
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
