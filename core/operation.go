package core

import (
	"context"
	"log"

	config "cryptoquant.com/m/config"
	account "cryptoquant.com/m/core/account"
	binancetrade "cryptoquant.com/m/core/trader/binance"
	upbittrade "cryptoquant.com/m/core/trader/upbit"
	database "cryptoquant.com/m/data/database"
	pb "cryptoquant.com/m/gen/traderpb"
)

const SAFE_MARGIN = 0.9
const USE_FUND_UPPER_BOUND = 0.4

// Trader gRPC server
type Operation struct {
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

func NewOperation(ctx context.Context) (*Operation, error) {
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

	return &Operation{
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
