package engine

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	config "cryptoquant.com/m/config"
	database "cryptoquant.com/m/data/database"
	binancesource "cryptoquant.com/m/data/sources/binance"
	binancews "cryptoquant.com/m/internal/binance/ws"
	upbitws "cryptoquant.com/m/internal/upbit/ws"
	kimchiarb "cryptoquant.com/m/strategy/arbitrage/kimchi"
	binancemarket "cryptoquant.com/m/streams/binance/market"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

// EngineContext represents the core trading engine context that manages configurations
// for both Upbit (Korean) and Binance exchanges for cross-exchange arbitrage
// TODO: Add more exchanges support
type EngineContext struct {
	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Exchange configurations
	KimchiExchangeConfig  *config.UpbitSpotTradeConfig
	BinanceExchangeConfig *config.BinanceFutureTradeConfig

	// Data
	Database         *database.Database
	FutureMarketData *binancesource.BinanceFutureMarketData

	// Strategy
	KimchiPairs *kimchiarb.KimchiPremium

	// Target Symbol
	KimchiAssetSymbol string
	CefiAssetSymbol   string
	AnchorAssetSymbol string // USDT

	// Trading Parameters
	kimchiTradeChan  chan float64
	binanceTradeChan chan float64
	anchorTradeChan  chan float64 // KRW-USDT

	kimchiBestBidChan  chan float64
	binanceBestBidChan chan float64
	anchorBestBidChan  chan float64

	kimchiBestAskChan  chan float64
	binanceBestAskChan chan float64
	anchorBestAskChan  chan float64

	// Logger
	Logger *log.Logger
}

func setupLog() *log.Logger {
	// Create log directory with today's date
	today := time.Now().Format("20060102")
	logDir := filepath.Join("log", today)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// Open log file in dated directory
	logPath := filepath.Join(logDir, "engine.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return log.New(logFile, "", log.LstdFlags)
}

func New(ctx context.Context) *EngineContext {
	// Create a new context with cancellation
	engineCtx, cancel := context.WithCancel(ctx)

	// Set up logging
	logger := setupLog()

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		logger.Printf("Failed to connect to database: %v", err)
		panic(err)
	}

	// Create exchange configs
	kimchiConfig, err := config.NewUpbitSpotTradeConfig()
	if err != nil {
		logger.Printf("Failed to create Upbit config: %v", err)
		panic(err)
	}
	binanceConfig, err := config.NewBinanceFutureTradeConfig()
	if err != nil {
		logger.Printf("Failed to create Binance config: %v", err)
		panic(err)
	}

	anchor := os.Getenv("ANCHOR_SYMBOL")
	if anchor == "" {
		logger.Printf("Failed to confirm target symbols: %v", "Environment variables not set")
		panic("Environment variables not set")
	}

	engine := &EngineContext{
		ctx:                   engineCtx,
		cancel:                cancel,
		KimchiExchangeConfig:  kimchiConfig,
		BinanceExchangeConfig: binanceConfig,
		Database:              db,
		Logger:                logger,
		AnchorAssetSymbol:     anchor,

		kimchiTradeChan:  make(chan float64),
		binanceTradeChan: make(chan float64),
		anchorTradeChan:  make(chan float64),
	}

	logger.Printf("Engine initialized")

	return engine
}

func (e *EngineContext) ConfirmTargetSymbols() {
	csymbol := os.Getenv("BINANCE_SYMBOL")
	ksymbol := os.Getenv("UPBIT_SYMBOL")

	if csymbol == "" || ksymbol == "" {
		e.Logger.Printf("Failed to confirm target symbols: %v", "Environment variables not set")
		panic("Environment variables not set")
	}

	// Confirm trading symbols in cefi and kimchi
	if !e.BinanceExchangeConfig.IsAvailableSymbol(csymbol) {
		e.Logger.Printf("Failed to confirm target symbols: %v", "Binance symbol not available")
		panic("Binance symbol not available")
	}

	if !e.KimchiExchangeConfig.IsAvailableSymbol(ksymbol) {
		e.Logger.Printf("Failed to confirm target symbols: %v", "Kimchi symbol not available")
		panic("Kimchi symbol not available")
	}

	// Confirm anchor symbol
	if !e.KimchiExchangeConfig.IsAvailableSymbol(e.AnchorAssetSymbol) {
		e.Logger.Printf("Failed to confirm target symbols: %v", "Kimchi anchor symbol not available")
		panic("Kimchi anchor symbol not available")
	}

	e.KimchiAssetSymbol = ksymbol
	e.CefiAssetSymbol = csymbol
}

func (e *EngineContext) StartAsset() {
	e.Logger.Printf("Starting asset for %v", e.KimchiAssetSymbol)
	kimchiAsset := kimchiarb.NewAsset(e.KimchiAssetSymbol)
	binanceAsset := kimchiarb.NewAsset(e.CefiAssetSymbol)
	anchorAsset := kimchiarb.NewAsset(e.AnchorAssetSymbol)

	kimchiAsset.SetPriceChannel(e.kimchiTradeChan)
	binanceAsset.SetPriceChannel(e.binanceTradeChan)
	anchorAsset.SetPriceChannel(e.anchorTradeChan)

	kimchiAsset.SetBestBidChannel(e.kimchiBestBidChan)
	binanceAsset.SetBestBidChannel(e.binanceBestBidChan)
	anchorAsset.SetBestBidChannel(e.anchorBestBidChan)

	kimchiAsset.SetBestAskChannel(e.kimchiBestAskChan)
	binanceAsset.SetBestAskChannel(e.binanceBestAskChan)
	anchorAsset.SetBestAskChannel(e.anchorBestAskChan)

	e.KimchiPairs = &kimchiarb.KimchiPremium{
		AnchorAsset: anchorAsset,
		CefiAsset:   binanceAsset,
		KimchiAsset: kimchiAsset,
	}

	e.Logger.Printf("Asset Pairs(%v, %v, %v) initialized", e.KimchiAssetSymbol, e.CefiAssetSymbol, e.AnchorAssetSymbol)
}

func (e *EngineContext) StartMonitor() {
	e.Logger.Printf("Engine started")

	// Start stream
	e.Logger.Printf("Starting stream for %v", e.KimchiAssetSymbol)
	h1 := upbitmarket.NewPriceHandler(e.kimchiTradeChan)
	h1s := []func(upbitws.SpotTrade) error{h1}
	go upbitmarket.SubscribeTrade(e.ctx, e.KimchiAssetSymbol, h1s)

	// Start stream
	e.Logger.Printf("Starting stream for %v", e.CefiAssetSymbol)
	h2 := binancemarket.NewPriceHandler(e.binanceTradeChan)
	h2s := []func(binancews.FutureAggTrade) error{h2}
	go binancemarket.SubscribeAggtrade(e.ctx, e.CefiAssetSymbol, h2s)

	// Start stream
	e.Logger.Printf("Starting stream for %v", e.AnchorAssetSymbol)
	h3 := upbitmarket.NewPriceHandler(e.anchorTradeChan)
	h3s := []func(upbitws.SpotTrade) error{h3}
	go upbitmarket.SubscribeTrade(e.ctx, e.AnchorAssetSymbol, h3s)
}

func (e *EngineContext) StartStrategy() {
	e.Logger.Printf("Starting strategy")
	go e.KimchiPairs.Run(e.ctx)

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-time.Tick(1 * time.Second):
			e.Logger.Println(e.KimchiPairs.Status())
		}
	}
}

func (e *EngineContext) Stop() {
	e.Logger.Printf("Engine stopping...")
	e.cancel()
	e.wg.Wait()
	e.Logger.Printf("Engine stopped")
}

func (e *EngineContext) Context() context.Context {
	return e.ctx
}
