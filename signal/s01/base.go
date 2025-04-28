package s01signal

import (
	"context"
	"log"
	"os"
	"sync"

	config "cryptoquant.com/m/config"
	database "cryptoquant.com/m/data/database"
	signal "cryptoquant.com/m/signal"
	s01 "cryptoquant.com/m/strategy/s01"
)

// SignalContext represents the core trading engine context that manages configurations
// for both Upbit (Korean) and Binance exchanges for cross-exchange arbitrage
// TODO: Add more exchanges support
type SignalContext struct {
	SignalID string

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// To Trader - Submit trade
	traderMessenger *signal.TraderMessenger

	// Exchange configurations - Check if the symbols are available
	UpbitExchangeConfig   *config.UpbitSpotTradeConfig
	BinanceExchangeConfig *config.BinanceFutureTradeConfig

	// Data
	Database  *database.Database  // Get trade parameters
	TimeScale *database.TimeScale // Log premium data

	// Strategy
	UpbitBinancePairs    *s01.UpbitBinancePair
	EnterPremiumBoundary float64
	ExitPremiumBoundary  float64

	// Target Symbol
	UpbitAssetSymbol   string
	BinanceAssetSymbol string
	AnchorAssetSymbol  string // USDT-KRW, Upbit

	// Premium Calculation Parameters
	upbitTradeChan   chan float64
	binanceTradeChan chan float64
	anchorTradeChan  chan float64 // KRW-USDT

	upbitBestBidPrcChan   chan float64
	binanceBestBidPrcChan chan float64
	anchorBestBidPrcChan  chan float64

	upbitBestBidQtyChan   chan float64
	binanceBestBidQtyChan chan float64
	anchorBestBidQtyChan  chan float64

	upbitBestAskPrcChan   chan float64
	binanceBestAskPrcChan chan float64
	anchorBestAskPrcChan  chan float64

	upbitBestAskQtyChan   chan float64
	binanceBestAskQtyChan chan float64
	anchorBestAskQtyChan  chan float64

	// Trading Channel
	inPosition  bool
	premiumChan chan [3]float64 // [EnterPremium, ExitPremium]

	// Logging channel
	premiumLog chan database.PremiumLog
}

func New(ctx context.Context) *SignalContext {
	// 1. Get engine name
	signalEngineName := os.Getenv("ENGINE_NAME")
	if signalEngineName == "" {
		panic("ENGINE_NAME is not set")
	}

	// 2. Create a new context with cancellation
	// Controls the lifecycle of the whole engine and daemon structs and streams
	signalCtx, cancel := context.WithCancel(ctx)

	// 3. Connect to database
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

	// 4. Create exchange configs
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

	// 5. Confirm anchor symbol
	anchor := os.Getenv("ANCHOR_SYMBOL")
	if anchor == "" {
		log.Println("Failed to confirm anchor symbol: Environment variables not set")
		panic("Environment variables not set")
	}

	// 6. Create trader messenger
	traderAddr := os.Getenv("TRADER_ADDRESS")
	if traderAddr == "" {
		log.Println("Failed to confirm trader address: Environment variables not set")
		panic("Environment variables not set")
	}
	traderMessenger := signal.NewTraderMessenger(traderAddr, ctx)

	// 7. Create struct with order channels
	engine := &SignalContext{
		SignalID:              signalEngineName,
		ctx:                   signalCtx,
		cancel:                cancel,
		traderMessenger:       traderMessenger,
		UpbitExchangeConfig:   upbitConfig,
		BinanceExchangeConfig: binanceConfig,
		Database:              db,
		TimeScale:             ts,
		AnchorAssetSymbol:     anchor,

		upbitTradeChan:   make(chan float64),
		binanceTradeChan: make(chan float64),
		anchorTradeChan:  make(chan float64),

		upbitBestBidPrcChan:   make(chan float64),
		binanceBestBidPrcChan: make(chan float64),
		anchorBestBidPrcChan:  make(chan float64),

		upbitBestAskPrcChan:   make(chan float64),
		binanceBestAskPrcChan: make(chan float64),
		anchorBestAskPrcChan:  make(chan float64),

		upbitBestAskQtyChan:   make(chan float64),
		binanceBestAskQtyChan: make(chan float64),
		anchorBestAskQtyChan:  make(chan float64),

		upbitBestBidQtyChan:   make(chan float64),
		binanceBestBidQtyChan: make(chan float64),
		anchorBestBidQtyChan:  make(chan float64),

		inPosition:  false,
		premiumChan: make(chan [3]float64),

		premiumLog: make(chan database.PremiumLog, 100),
	}

	log.Println("Engine initialized")
	return engine
}

func (e *SignalContext) Stop() {
	log.Println("Engine stopping...")
	e.cancel()
	e.wg.Wait()
	log.Println("Engine stopped")
}

func (e *SignalContext) Context() context.Context {
	return e.ctx
}

func (e *SignalContext) ChangePositionStatus() {
	e.inPosition = !e.inPosition
}
