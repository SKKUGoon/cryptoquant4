package core

import (
	"context"
	"log"

	pb "cryptoquant.com/m/gen/traderpb"
)

// Trader gRPC server
type Server struct {
	pb.UnimplementedTraderServer
}

func (s *Server) SubmitTrade(ctx context.Context, req *pb.TradeRequest) (*pb.OrderResponse, error) {
	switch order := req.OrderType.(type) {
	case *pb.TradeRequest_PairOrder:
		log.Printf("Pair order: %v", order)
	case *pb.TradeRequest_SingleOrder:
		// TODO: Implement single order
		return &pb.OrderResponse{Success: false, Message: "Not implemented"}, nil
	default:
		return &pb.OrderResponse{Success: false, Message: "Invalid order type"}, nil
	}

	return &pb.OrderResponse{Success: true, Message: "Order submitted"}, nil
}
