package core

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"cryptoquant.com/m/core/account"
	pb "cryptoquant.com/m/gen/traderpb"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/shopspring/decimal"
)

const SAFE_MARGIN = 0.9

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
		binanceWallet := s.Account.GetBinanceWalletSnapshot()
		upbitWallet := s.Account.GetUpbitWalletSnapshot()

		// Calculate the amount of the order for upbit and binance
		upbitAmount, binanceAmount, err := calculateOrderAmount(
			order.PairOrder.UpbitOrder,
			order.PairOrder.BinanceOrder,
			upbitWallet,
			binanceWallet,
			s.Account.UpbitPrincipalCurrency,
			s.Account.BinancePrincipalCurrency,
			order.PairOrder.ExchangeRate,
		)
		if err != nil {
			return nil, err
		}

		// Generate upbit order sheet
		upbitOrderSheet, err := generateUpbitOrderSheet(order.PairOrder.UpbitOrder, upbitAmount)
		if err != nil {
			return nil, err
		}

		// Generate binance order sheet
		binanceOrderSheet, err := generateBinanceOrderSheet(order.PairOrder.BinanceOrder, binanceAmount)
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

// calculateOrderAmount calculates the amount of the order for upbit and binance
// based on the order and wallet information.
// It returns the amount of the order for upbit and binance, and an error if there is one.
func calculateOrderAmount(
	upbitOrder, binanceOrder *pb.ExchangeOrder,
	upbitWallet, binanceWallet map[string]float64,
	upbitPrincipalCurrency, binancePrincipalCurrency string,
	exchangeRate float64, // KRW for USDT. e.g.) 1400KRW for 1USDT
) (float64, float64, error) { // upbitAmount, binanceAmount, error
	// Calculate the maximum amount of the order for upbit
	upbitBookAvailable := upbitOrder.Amount * upbitOrder.Price * SAFE_MARGIN
	upbitFund := upbitWallet[upbitPrincipalCurrency]

	// Maximum amount of the order for binance
	binanceBookAvailable := binanceOrder.Amount * binanceOrder.Price * SAFE_MARGIN
	binanceFund := binanceWallet[binancePrincipalCurrency]

	// Maximum available amount (minimum of the 4 values (bookAvailable, upbitFund, binanceFund, binanceAvailable))
	maxAmount := math.Min(
		math.Min(upbitBookAvailable/exchangeRate, upbitFund/exchangeRate),
		math.Min(binanceBookAvailable, binanceFund),
	)

	// Calculate the amount of the order for upbit and binance
	upbitAmount := maxAmount * exchangeRate         // Upbit needs Price * Quantity value
	binanceAmount := maxAmount / binanceOrder.Price // Binance needs Quantity value

	return upbitAmount, binanceAmount, nil
}

func generateUpbitOrderSheet(order *pb.ExchangeOrder, orderSpec float64) (*upbitrest.OrderSheet, error) {
	switch order.Side {
	case "buy":
		return &upbitrest.OrderSheet{
			Symbol:  order.Symbol,
			Side:    "bid",
			Price:   strconv.FormatFloat(orderSpec, 'f', -1, 64),
			OrdType: "market",
		}, nil

	case "sell":
		// Only need to specify amount
		return &upbitrest.OrderSheet{
			Symbol:  order.Symbol,
			Side:    "ask",
			Volume:  strconv.FormatFloat(orderSpec, 'f', -1, 64),
			OrdType: "market",
		}, nil
	default:
		return nil, fmt.Errorf("invalid order side: %v", order.Side)
	}
}

func generateBinanceOrderSheet(order *pb.ExchangeOrder, orderSpec float64) (*binancerest.OrderSheet, error) {
	switch order.Side {
	case "buy":
		// To binance, buy order is exit order
		return &binancerest.OrderSheet{}, nil
	case "sell":
		timestamp := time.Now().UnixMilli()
		return &binancerest.OrderSheet{
			Symbol:    order.Symbol,
			Side:      "SELL",
			Quantity:  decimal.NewFromFloat(orderSpec),
			Type:      "MARKET",
			Timestamp: timestamp,
		}, nil
	default:
		return nil, fmt.Errorf("invalid order side: %v", order.Side)
	}
}
