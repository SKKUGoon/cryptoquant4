//go:build trader && !server && !init
// +build trader,!server,!init

package main

import (
	"context"
	"log"
	"net"

	core "cryptoquant.com/m/core"
	pb "cryptoquant.com/m/gen/traderpb"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tradingServer, err := core.NewTraderServer(ctx)
	if err != nil {
		log.Fatalf("failed to create trader server: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterTraderServer(server, tradingServer)

	log.Printf("Server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
