//go:build trader && !server && !init
// +build trader,!server,!init

package main

import (
	"context"
	"log"
	"net"

	core "cryptoquant.com/m/core"
	account "cryptoquant.com/m/core/account"
	pb "cryptoquant.com/m/gen/traderpb"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Account source
	as := account.NewAccountSourceSync(ctx)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterTraderServer(server, &core.Server{Account: as})

	log.Printf("Server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
