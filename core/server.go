package core

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	config "cryptoquant.com/m/config"
	account "cryptoquant.com/m/core/account"
	binancetrade "cryptoquant.com/m/core/trader/binance"
	upbittrade "cryptoquant.com/m/core/trader/upbit"
	database "cryptoquant.com/m/data/database"
	pb "cryptoquant.com/m/gen/traderpb"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"cryptoquant.com/m/utils"
	"github.com/shopspring/decimal"
)

const SAFE_MARGIN = 0.9
const USE_FUND_UPPER_BOUND = 0.4

// Trader gRPC server
type Server struct {
	pb.UnimplementedTraderServer

	ctx context.Context

	// Unified Account Manager
	Account *account.AccountSource

	// Exchange configurations - Vet precision
	UpbitConfig   *config.UpbitSpotTradeConfig
	BinanceConfig *config.BinanceFutureTradeConfig

	// Traders
	UpbitTrader   *upbittrade.Trader
	BinanceTrader *binancetrade.Trader

	// Data
	Database  *database.Database  // Get trade parameters
	TimeScale *database.TimeScale // Log premium data

	// Logging channel
	kimchiTradeLog chan []database.KimchiOrderLog
	walletLog      chan database.AccountSnapshot
}

func NewTraderServer(ctx context.Context) (*Server, error) {
	as := account.NewAccountSource(ctx)
	err := as.Sync()
	if err != nil {
		return nil, err
	}

	// Create exchange configs
	upbitConfig, err := config.NewUpbitSpotTradeConfig()
	if err != nil {
		log.Printf("Failed to create Upbit config: %v", err)
		panic(err)
	}
	binanceConfig, err := config.NewBinanceFutureTradeConfig()
	if err != nil {
		log.Printf("Failed to create Binance config: %v", err)
		panic(err)
	}

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		panic(err)
	}
	ts, err := database.ConnectTS()
	if err != nil {
		log.Printf("Failed to connect to TimeScale: %v", err)
		panic(err)
	}

	// Create trader
	upbitTrader := upbittrade.NewTrader()
	upbitTrader.UpdateRateLimit(1000)
	binanceTrader := binancetrade.NewTrader()
	binanceTrader.UpdateRateLimit(1000)

	return &Server{
		ctx:     ctx,
		Account: as,

		// Configurations
		UpbitConfig:   upbitConfig,
		BinanceConfig: binanceConfig,

		// Traders
		UpbitTrader:   upbitTrader,
		BinanceTrader: binanceTrader,

		// Utility
		Database:  db,
		TimeScale: ts,

		// Logging channels
		kimchiTradeLog: make(chan []database.KimchiOrderLog, 100),
		walletLog:      make(chan database.AccountSnapshot, 100),
	}, nil
}

func (s *Server) SubmitTrade(ctx context.Context, req *pb.TradeRequest) (*pb.OrderResponse, error) {
	var isPairEnter bool
	var upbitOrderSheet *upbitrest.OrderSheet
	var binanceOrderSheet *binancerest.OrderSheet
	var orderTime time.Time
	var executionTime time.Time

	defer func() {
		// Log the order sheet in docker logs
		log.Println("--------------------------------")
		log.Printf("upbitOrderSheet: %+v\n", upbitOrderSheet)
		log.Printf("binanceOrderSheet: %+v\n", binanceOrderSheet)
		log.Println("--------------------------------")
	}()

	// Lock account - Prevent race condition
	s.Account.Mu.Lock()
	defer s.Account.Mu.Unlock()

	switch order := req.OrderType.(type) {
	case *pb.TradeRequest_PairOrder:
		// Pair Order
		// - Deals with pair entry and pair exit
		switch order.PairOrder.PairOrderType {
		case pb.PairOrderType_PairOrderEnter: // Enter Upbit Long and Binance Short. Loading my position
			isPairEnter = true

			// Calculate the entry amount
			upbitAmountF64, binanceAmountDec, err := s.calculateOrderAmount(
				order.PairOrder.UpbitOrder,
				order.PairOrder.BinanceOrder,
				order.PairOrder.ExchangeRate,
			) // Float64, Decimal, error
			if err != nil {
				return &pb.OrderResponse{Success: false, Message: "Failed to calculate order amount"}, nil
			}

			upbitOrderSheet = generateUpbitBuyOrderSheet(order.PairOrder.UpbitOrder, upbitAmountF64)
			binanceOrderSheet = generateBinanceSellOrderSheet(order.PairOrder.BinanceOrder, binanceAmountDec)

		case pb.PairOrderType_PairOrderExit: // Exit Upbit Short and Binance Long. Unloading my position
			var upbitAmount float64
			var binanceAmount float64
			var ok bool

			isPairEnter = false

			// Retreive the selling amount from the wallet
			upbitWallet := s.Account.GetUpbitWalletSnapshot()
			binanceWallet := s.Account.GetBinanceWalletSnapshot()

			upbitWalletSymbol := strings.Replace(order.PairOrder.UpbitOrder.Symbol, "KRW-", "", 1) // e.g.) KRW-XRP -> XRP
			if upbitAmount, ok = upbitWallet[upbitWalletSymbol]; !ok {
				return &pb.OrderResponse{Success: false, Message: "Upbit amount not found"}, nil
			}
			if binanceAmount, ok = binanceWallet[order.PairOrder.BinanceOrder.Symbol]; !ok {
				return &pb.OrderResponse{Success: false, Message: "Binance amount not found"}, nil
			}
			binanceAmount = math.Abs(binanceAmount) // NOTE: Binance amount is negative (Short)
			binanceAmountDec, err := utils.SafeDecimalFromFloat(binanceAmount)
			if err != nil {
				return &pb.OrderResponse{Success: false, Message: "Failed to convert binance amount to decimal"}, nil
			}

			// Generate the order sheet
			upbitOrderSheet = generateUpbitSellOrderSheet(order.PairOrder.UpbitOrder, upbitAmount)
			binanceOrderSheet = generateBinanceBuyOrderSheet(order.PairOrder.BinanceOrder, binanceAmountDec)

		default:
			return &pb.OrderResponse{Success: false, Message: "Invalid pair order type"}, nil
		}

		// Send orders
		orderTime = time.Now()
		upbitResult, err := s.UpbitTrader.SendOrder(*upbitOrderSheet)
		if err != nil || upbitResult.Error != nil {
			log.Printf("upbitResult: %+v", upbitResult.Error)
			return &pb.OrderResponse{Success: false, Message: "Upbit order failed"}, nil
		}
		binanceResult, err := s.BinanceTrader.SendSingleOrder(binanceOrderSheet)
		if err != nil || binanceResult.Error != nil {
			log.Printf("binanceResult: %+v", binanceResult.Error)
			return &pb.OrderResponse{Success: false, Message: "Binance order failed"}, nil
		}
		executionTime = time.Now()

		// Log orders - Order time
		kimchiOrderLogs, err := s.CreateKimchiOrderLog(
			isPairEnter,
			order.PairOrder,
			upbitResult,
			binanceResult,
			orderTime,
			executionTime,
		)
		if err != nil {
			return &pb.OrderResponse{Success: false, Message: "Failed to create kimchi order log"}, nil
		}

		// Insert Upbit - Binance Trading Log
		select {
		case s.kimchiTradeLog <- kimchiOrderLogs:
		default:
			log.Printf("Warning: kimchiTradeLog channel is full, skipping log")
		}

	case *pb.TradeRequest_SingleOrder:
		// TODO: Implement single order
		return &pb.OrderResponse{Success: false, Message: "Not implemented"}, nil
	default:
		return &pb.OrderResponse{Success: false, Message: "Invalid order type"}, nil
	}

	// Update account
	s.Account.UpdateRedis()
	s.Account.Sync()

	return &pb.OrderResponse{Success: true, Message: "Order submitted"}, nil
}

// calculateOrderAmount calculates the amount of the order for upbit and binance
// based on the order and wallet information.
// It returns the amount of the order for upbit and binance, and an error if there is one.
func (s *Server) calculateOrderAmount(
	upbitOrder, binanceOrder *pb.ExchangeOrder,
	exchangeRate float64, // KRW for USDT. e.g.) 1400KRW for 1USDT
) (float64, decimal.Decimal, error) { // upbitAmount, binanceAmount, error
	// Necessary information
	upbitWallet := s.Account.GetUpbitWalletSnapshot()
	binanceWallet := s.Account.GetBinanceWalletSnapshot()

	// Get minimum notional
	upbitMinNotional := float64(s.UpbitConfig.MinimumTradeAmount)
	binanceMinNotional := float64(s.BinanceConfig.MinimumTradeAmount)

	binancePrecision := s.BinanceConfig.GetSymbolQuantityPrecision(binanceOrder.Symbol)

	// Calculate the maximum amount of the order for upbit
	upbitBookAvailable := upbitOrder.Amount * upbitOrder.Price * SAFE_MARGIN
	upbitFund := upbitWallet[s.Account.UpbitPrincipalCurrency] * USE_FUND_UPPER_BOUND

	// Maximum amount of the order for binance
	binanceBookAvailable := binanceOrder.Amount * binanceOrder.Price * SAFE_MARGIN
	binanceFund := binanceWallet[s.Account.BinancePrincipalCurrency] * USE_FUND_UPPER_BOUND

	// Maximum available amount (minimum of the 4 values (bookAvailable, upbitFund, binanceFund, binanceAvailable))
	maxNotional := math.Min(
		math.Min(upbitBookAvailable/exchangeRate, upbitFund/exchangeRate),
		math.Min(binanceBookAvailable, binanceFund),
	)

	if maxNotional <= 0 {
		return 0, decimal.Zero, fmt.Errorf("max notional is less than 0: %f", maxNotional)
	}

	rawBinanceQty := maxNotional / binanceOrder.Price // Convert to raw Binance quantity (before rounding)
	decimalQty, err := utils.SafeDecimalFromFloatTruncate(rawBinanceQty, binancePrecision)
	if err != nil {
		log.Printf("Invalid binance quantity: %v", err)
		return 0, decimal.Zero, err
	}

	actualBinanceNotional := decimalQty.InexactFloat64() * binanceOrder.Price // Recalculate Binance notional based on rounded quantity
	upbitNotional := actualBinanceNotional * exchangeRate                     // Calculate matching Upbit amount in KRW

	if upbitNotional < upbitMinNotional {
		return 0, decimal.Zero, fmt.Errorf("upbit notional is less than minimum notional: %f", upbitNotional)
	}

	if decimalQty.InexactFloat64() < binanceMinNotional {
		return 0, decimal.Zero, fmt.Errorf("binance notional is less than minimum notional: %v", decimalQty)
	}

	return upbitNotional, decimalQty, nil
}

func generateUpbitBuyOrderSheet(order *pb.ExchangeOrder, upbitAmount float64) *upbitrest.OrderSheet {
	return &upbitrest.OrderSheet{
		Symbol:  order.Symbol,
		Side:    "bid",
		Price:   strconv.FormatFloat(upbitAmount, 'f', -1, 64),
		OrdType: "price",
	}
}

func generateUpbitSellOrderSheet(order *pb.ExchangeOrder, upbitAmount float64) *upbitrest.OrderSheet {
	return &upbitrest.OrderSheet{
		Symbol:  order.Symbol,
		Side:    "ask",
		Volume:  strconv.FormatFloat(upbitAmount, 'f', -1, 64),
		OrdType: "market",
	}
}

func generateBinanceBuyOrderSheet(order *pb.ExchangeOrder, binanceAmount decimal.Decimal) *binancerest.OrderSheet {
	return &binancerest.OrderSheet{
		Symbol:     order.Symbol,
		Side:       "BUY",
		Quantity:   binanceAmount,
		ReduceOnly: "true",
		Type:       "MARKET",
		Timestamp:  time.Now().UnixMilli(),
	}
}

func generateBinanceSellOrderSheet(order *pb.ExchangeOrder, binanceAmount decimal.Decimal) *binancerest.OrderSheet {
	timestamp := time.Now().UnixMilli()
	return &binancerest.OrderSheet{
		Symbol:    order.Symbol,
		Side:      "SELL",
		Quantity:  binanceAmount,
		Type:      "MARKET",
		Timestamp: timestamp,
	}
}
