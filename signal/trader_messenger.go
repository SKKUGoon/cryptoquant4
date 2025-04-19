package signal

import (
	"context"
	"log"

	"cryptoquant.com/m/gen/traderpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TraderMessenger struct {
	client traderpb.TraderClient
	ctx    context.Context
}

func NewTraderMessenger(addr string, ctx context.Context) *TraderMessenger {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	client := traderpb.NewTraderClient(conn)
	return &TraderMessenger{
		client: client,
		ctx:    ctx,
	}
}

func (t *TraderMessenger) SubmitTrade(trade *traderpb.TradeRequest) (*traderpb.OrderResponse, error) {
	resp, err := t.client.SubmitTrade(t.ctx, trade)
	if err != nil {
		return nil, err
	}
	log.Printf("Response: %v", resp)
	return resp, nil
}
