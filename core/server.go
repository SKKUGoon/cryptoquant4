package core

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"cryptoquant.com/m/core/account"
	pb "cryptoquant.com/m/gen/traderpb"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/shopspring/decimal"
)

// Trader gRPC server
type Server struct {
	pb.UnimplementedTraderServer
	Account *account.AccountSource
}

func (s *Server) SubmitTrade(ctx context.Context, req *pb.TradeRequest) (*pb.OrderResponse, error) {
	s.Account.Mu.Lock()
	defer s.Account.Mu.Unlock()

	switch order := req.OrderType.(type) {
	case *pb.TradeRequest_PairOrder:
		// Generate upbit order sheet
		upbitOrderSheet, err := generateUpbitOrderSheet(order.PairOrder.UpbitOrder)
		if err != nil {
			return nil, err
		}

		// Generate binance order sheet
		binanceOrderSheet, err := generateBinanceOrderSheet(order.PairOrder.BinanceOrder)
		if err != nil {
			return nil, err
		}

		log.Printf("Pair order: %v", upbitOrderSheet)
		log.Printf("Binance order: %v", binanceOrderSheet)

	case *pb.TradeRequest_SingleOrder:
		// TODO: Implement single order
		return &pb.OrderResponse{Success: false, Message: "Not implemented"}, nil
	default:
		return &pb.OrderResponse{Success: false, Message: "Invalid order type"}, nil
	}

	return &pb.OrderResponse{Success: true, Message: "Order submitted"}, nil
}

func generateUpbitOrderSheet(order *pb.ExchangeOrder) (*upbitrest.OrderSheet, error) {
	switch order.Side {
	case "buy":
		pq := order.Amount * order.Price
		return &upbitrest.OrderSheet{
			Symbol:  order.Symbol,
			Side:    "bid",
			Price:   strconv.FormatFloat(pq, 'f', -1, 64),
			OrdType: "market",
		}, nil

	case "sell":
		// Only need to specify amount
		return &upbitrest.OrderSheet{
			Symbol:  order.Symbol,
			Side:    "ask",
			Volume:  strconv.FormatFloat(order.Amount, 'f', -1, 64),
			OrdType: "market",
		}, nil
	default:
		return nil, fmt.Errorf("invalid order side: %v", order.Side)
	}
}

func generateBinanceOrderSheet(order *pb.ExchangeOrder) (*binancerest.OrderSheet, error) {
	switch order.Side {
	case "buy":
		// To binance, buy order is exit order
		return &binancerest.OrderSheet{}, nil
	case "sell":
		timestamp := time.Now().UnixMilli()
		quantity := decimal.NewFromFloat(order.Amount)
		return &binancerest.OrderSheet{
			Symbol:    order.Symbol,
			Side:      "SELL",
			Quantity:  quantity,
			Type:      "MARKET",
			Timestamp: timestamp,
		}, nil
	default:
		return nil, fmt.Errorf("invalid order side: %v", order.Side)
	}
}
