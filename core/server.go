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
)

const SAFE_MARGIN = 0.9

// Trader gRPC server
type Server struct {
	pb.UnimplementedTraderServer

	ctx context.Context

	// Unified Account Manager
	Account *account.AccountSource

	// Exchange configurations - Vet precision
	UpbitPrecision   *config.UpbitSpotTradeConfig
	BinancePrecision *config.BinanceFutureTradeConfig

	// Traders
	UpbitTrader   *upbittrade.Trader
	BinanceTrader *binancetrade.Trader

	// Data
	Database  *database.Database  // Get trade parameters
	TimeScale *database.TimeScale // Log premium data

	// Logging channel
	kimchiTradeLog chan database.KimchiOrderLog
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
		ctx:              ctx,
		Account:          as,
		UpbitPrecision:   upbitConfig,
		BinancePrecision: binanceConfig,
		UpbitTrader:      upbitTrader,
		BinanceTrader:    binanceTrader,
		Database:         db,
		TimeScale:        ts,
		kimchiTradeLog:   make(chan database.KimchiOrderLog, 100),
		walletLog:        make(chan database.AccountSnapshot, 100),
	}, nil
}

func (s *Server) SubmitTrade(ctx context.Context, req *pb.TradeRequest) (*pb.OrderResponse, error) {
	var upbitOrderSheet *upbitrest.OrderSheet
	var binanceOrderSheet *binancerest.OrderSheet
	var orderTime time.Time
	var executionTime time.Time
	var err error
	s.Account.Mu.Lock()
	defer s.Account.Mu.Unlock()
	defer func() {
		log.Println("--------------------------------")
		log.Printf("upbitOrderSheet: %+v\n", upbitOrderSheet)
		log.Printf("binanceOrderSheet: %+v\n", binanceOrderSheet)
		log.Println("--------------------------------")
	}()

	switch order := req.OrderType.(type) {
	// TODO: Add update exchange config method
	case *pb.TradeRequest_PairOrder:

		switch order.PairOrder.PairOrderType {
		case pb.PairOrderType_PairOrderEnter: // Enter Upbit Long and Binance Short
			// Calculate the amount of the order for upbit and binance
			fmt.Println("upbitOrder: ", order.PairOrder.UpbitOrder)
			fmt.Println("binanceOrder: ", order.PairOrder.BinanceOrder)

			upbitAmount, binanceAmount, err := s.calculateOrderAmount(
				order.PairOrder.UpbitOrder,
				order.PairOrder.BinanceOrder,
				s.Account.UpbitPrincipalCurrency,
				s.Account.BinancePrincipalCurrency,
				order.PairOrder.ExchangeRate,
			)
			if err != nil {
				return nil, err
			}
			log.Printf("upbitAmount: %f, binanceAmount: %f", upbitAmount, binanceAmount)

			// Generate upbit order sheet (buy) + binance order sheet (sell)
			if upbitOrderSheet = generateUpbitBuyOrderSheet(order.PairOrder.UpbitOrder, upbitAmount); upbitOrderSheet == nil {
				return nil, fmt.Errorf("failed to generate upbit order sheet")
			}

			if binanceOrderSheet, err = generateBinanceSellOrderSheet(order.PairOrder.BinanceOrder, binanceAmount); err != nil {
				return nil, err
			}

		case pb.PairOrderType_PairOrderExit: // Exit Upbit Short and Binance Long
			// Get the amount of the order for upbit and binance
			upbitWallet := s.Account.GetUpbitWalletSnapshot()
			binanceWallet := s.Account.GetBinanceWalletSnapshot()

			// Misc job to generate the order sheet
			upbitWalletSymbol := strings.Replace(order.PairOrder.UpbitOrder.Symbol, "KRW-", "", 1) // e.g.) KRW-XRP -> XRP
			upbitAmount, ok := upbitWallet[upbitWalletSymbol]
			if !ok {
				return nil, fmt.Errorf("upbit amount not found")
			}
			binanceAmount, ok := binanceWallet[order.PairOrder.BinanceOrder.Symbol]
			if !ok {
				return nil, fmt.Errorf("binance amount not found")
			}

			// Generate the order sheet
			if upbitOrderSheet = generateUpbitSellOrderSheet(order.PairOrder.UpbitOrder, upbitAmount); upbitOrderSheet == nil {
				return nil, fmt.Errorf("failed to generate upbit order sheet")
			}
			if binanceOrderSheet, err = generateBinanceBuyOrderSheet(order.PairOrder.BinanceOrder, binanceAmount); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("invalid pair order type: %v", order.PairOrder.PairOrderType)
		}

		// Send orders
		orderTime = time.Now()
		upbitResult, err := s.UpbitTrader.SendOrder(*upbitOrderSheet)
		if err != nil {
			s.KimchiPremiumEject()
			return nil, err
		}
		binanceResult, err := s.BinanceTrader.SendSingleOrder(binanceOrderSheet)
		if err != nil {
			s.KimchiPremiumEject()
			return nil, err
		}
		executionTime = time.Now()

		// Log orders
		kimchiOrderLogs, err := s.CreateKimchiOrderLog(
			order.PairOrder.UpbitOrder,
			order.PairOrder.BinanceOrder,
			order.PairOrder.ExchangeRate,
			&upbitResult,
			&binanceResult,
			orderTime,
			executionTime,
		)
		if err != nil {
			s.KimchiPremiumEject()
			return nil, err
		}
		s.kimchiTradeLog <- kimchiOrderLogs[0]

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
	upbitPrincipalCurrency, binancePrincipalCurrency string,
	exchangeRate float64, // KRW for USDT. e.g.) 1400KRW for 1USDT
) (float64, float64, error) { // upbitAmount, binanceAmount, error

	// Necessary information
	upbitWallet := s.Account.GetUpbitWalletSnapshot()
	binanceWallet := s.Account.GetBinanceWalletSnapshot()
	binancePrecision := s.BinancePrecision.GetSymbolPricePrecision(binanceOrder.Symbol)
	upbitMinNotional := float64(s.UpbitPrecision.MinimumTradeAmount)
	binanceMinNotional := float64(s.BinancePrecision.MinimumTradeAmount)

	// Calculate the maximum amount of the order for upbit
	upbitBookAvailable := upbitOrder.Amount * upbitOrder.Price * SAFE_MARGIN
	upbitFund := upbitWallet[upbitPrincipalCurrency]

	// Maximum amount of the order for binance
	binanceBookAvailable := binanceOrder.Amount * binanceOrder.Price * SAFE_MARGIN
	binanceFund := binanceWallet[binancePrincipalCurrency]

	// Maximum available amount (minimum of the 4 values (bookAvailable, upbitFund, binanceFund, binanceAvailable))
	maxNotional := math.Min(
		math.Min(upbitBookAvailable/exchangeRate, upbitFund/exchangeRate),
		math.Min(binanceBookAvailable, binanceFund),
	)

	if maxNotional <= 0 {
		return 0, 0, fmt.Errorf("max notional is less than 0: %f", maxNotional)
	}

	rawBinanceQty := maxNotional / binanceOrder.Price    // Step 2: Convert to raw Binance quantity (before rounding)
	stepSize := math.Pow(10, -float64(binancePrecision)) // Step 3: Round down Binance quantity using step size
	roundedBinanceQty := math.Floor(rawBinanceQty/stepSize) * stepSize

	actualBinanceNotional := roundedBinanceQty * binanceOrder.Price // Step 4: Recalculate Binance notional based on rounded quantity
	upbitNotional := actualBinanceNotional * exchangeRate           // Step 5: Calculate matching Upbit amount in KRW

	if upbitNotional < upbitMinNotional {
		return 0, 0, fmt.Errorf("upbit notional is less than minimum notional: %f", upbitNotional)
	}

	if roundedBinanceQty < binanceMinNotional {
		return 0, 0, fmt.Errorf("binance notional is less than minimum notional: %f", roundedBinanceQty)
	}

	return upbitNotional, roundedBinanceQty, nil
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

func generateBinanceBuyOrderSheet(order *pb.ExchangeOrder, binanceAmount float64) (*binancerest.OrderSheet, error) {
	safeDecimal, err := utils.SafeDecimalFromFloat(binanceAmount * 1.0005)
	if err != nil {
		return nil, err
	}

	return &binancerest.OrderSheet{
		Symbol:     order.Symbol,
		Side:       "BUY",
		Quantity:   safeDecimal, // NOTE: Buffer of 0.05% ensure all position closed (Reduce only)
		ReduceOnly: "true",
		Type:       "MARKET",
		Timestamp:  time.Now().UnixMilli(),
	}, nil
}

func generateBinanceSellOrderSheet(order *pb.ExchangeOrder, binanceAmount float64) (*binancerest.OrderSheet, error) {
	safeDecimal, err := utils.SafeDecimalFromFloat(binanceAmount)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().UnixMilli()
	return &binancerest.OrderSheet{
		Symbol:    order.Symbol,
		Side:      "SELL",
		Quantity:  safeDecimal,
		Type:      "MARKET",
		Timestamp: timestamp,
	}, nil
}
