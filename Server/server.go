package main

import (
	proto "example/chittychat/grpc"
)

type ChittyChatServer struct {
	proto.UnimplementedChittyChatServiceServer
	messageLog []string
}
