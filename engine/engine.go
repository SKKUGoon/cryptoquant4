package engine

import (
	"context"
	"io"
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
	TimeScale        *database.TimeScale
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

	// logger
	logger *log.Logger
	tsLog  chan database.PremiumLog
}

func setupLog() *log.Logger {
	// Ensure the log directory exists
	if err := os.MkdirAll("log", 0755); err != nil {
		panic(err)
	}

	// Create a log file with today's date in the name
	today := time.Now().Format("20060102")
	logPath := filepath.Join("log", "engine_"+today+".log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// Create a logger that writes to the file and stdout
	multiWriter := io.MultiWriter(logFile, os.Stdout)
	return log.New(multiWriter, "", log.LstdFlags)
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

	// Connect to TimeScale
	ts, err := database.ConnectTS()
	if err != nil {
		logger.Printf("Failed to connect to TimeScale: %v", err)
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
		TimeScale:             ts,
		logger:                logger,
		AnchorAssetSymbol:     anchor,

		kimchiTradeChan:  make(chan float64),
		binanceTradeChan: make(chan float64),
		anchorTradeChan:  make(chan float64),

		kimchiBestBidChan:  make(chan float64),
		binanceBestBidChan: make(chan float64),
		anchorBestBidChan:  make(chan float64),

		kimchiBestAskChan:  make(chan float64),
		binanceBestAskChan: make(chan float64),
		anchorBestAskChan:  make(chan float64),

		tsLog: make(chan database.PremiumLog),
	}

	logger.Printf("Engine initialized")

	return engine
}

func (e *EngineContext) ConfirmTargetSymbols() {
	csymbol := os.Getenv("BINANCE_SYMBOL")
	ksymbol := os.Getenv("UPBIT_SYMBOL")

	if csymbol == "" || ksymbol == "" {
		e.logger.Printf("Failed to confirm target symbols: %v", "Environment variables not set")
		panic("Environment variables not set")
	}

	// Confirm trading symbols in cefi and kimchi
	if !e.BinanceExchangeConfig.IsAvailableSymbol(csymbol) {
		e.logger.Printf("Failed to confirm target symbols: %v", "Binance symbol not available")
		panic("Binance symbol not available")
	}

	if !e.KimchiExchangeConfig.IsAvailableSymbol(ksymbol) {
		e.logger.Printf("Failed to confirm target symbols: %v", "Kimchi symbol not available")
		panic("Kimchi symbol not available")
	}

	// Confirm anchor symbol
	if !e.KimchiExchangeConfig.IsAvailableSymbol(e.AnchorAssetSymbol) {
		e.logger.Printf("Failed to confirm target symbols: %v", "Kimchi anchor symbol not available")
		panic("Kimchi anchor symbol not available")
	}

	e.KimchiAssetSymbol = ksymbol
	e.CefiAssetSymbol = csymbol
}

func (e *EngineContext) StartAsset() {
	e.logger.Printf("Starting asset for %v", e.KimchiAssetSymbol)
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

	e.logger.Printf("Asset Pairs(%v, %v, %v) initialized", e.KimchiAssetSymbol, e.CefiAssetSymbol, e.AnchorAssetSymbol)
}

func (e *EngineContext) StartMonitor() {
	e.logger.Printf("Engine started")

	// Start stream
	e.logger.Printf("Starting stream for %v", e.KimchiAssetSymbol)
	h1 := upbitmarket.NewPriceHandler(e.kimchiTradeChan)
	h1s := []func(upbitws.SpotTrade) error{h1}
	go upbitmarket.SubscribeTrade(e.ctx, e.KimchiAssetSymbol, h1s)

	// Start stream
	e.logger.Printf("Starting stream for %v", e.CefiAssetSymbol)
	h2 := binancemarket.NewPriceHandler(e.binanceTradeChan)
	h2s := []func(binancews.FutureAggTrade) error{h2}
	go binancemarket.SubscribeAggtrade(e.ctx, e.CefiAssetSymbol, h2s)

	// Start stream
	e.logger.Printf("Starting stream for %v", e.AnchorAssetSymbol)
	h3 := upbitmarket.NewPriceHandler(e.anchorTradeChan)
	h3s := []func(upbitws.SpotTrade) error{h3}
	go upbitmarket.SubscribeTrade(e.ctx, e.AnchorAssetSymbol, h3s)
}

func (e *EngineContext) StartStrategy() {
	e.logger.Printf("Starting strategy")
	go e.KimchiPairs.Run(e.ctx)

	// Track consecutive failures
	consecutiveFailures := 0
	const maxConsecutiveFailures = 5000 // Adjust this threshold as needed

	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return
			case <-time.Tick(100 * time.Millisecond):
				ok, msg := e.KimchiPairs.Status()
				if !ok {
					consecutiveFailures++
					e.logger.Printf("Strategy is not ready: %v (consecutive failures: %d)", msg, consecutiveFailures)

					if consecutiveFailures >= maxConsecutiveFailures {
						e.logger.Printf("Too many consecutive failures (%d), initiating container restart", consecutiveFailures)
						// Trigger graceful shutdown
						e.Stop()
						// Exit with non-zero status to trigger container restart
						os.Exit(1)
					}
					continue
				} else {
					// Reset failure counter on success
					consecutiveFailures = 0
					e.logger.Println(msg)
				}

				log := e.KimchiPairs.ToPremiumLog()
				e.tsLog <- log
			}
		}
	}()
}

func (e *EngineContext) StartTSLog() {
	e.logger.Printf("Starting DB log")
	buffer := make([]database.PremiumLog, 0, 100) // Insert 100 at a time
	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return
			case row := <-e.tsLog:
				buffer = append(buffer, row)
				if len(buffer) >= 100 {
					bufferCopy := make([]database.PremiumLog, len(buffer))
					copy(bufferCopy, buffer)
					go func(logs []database.PremiumLog) {
						if err := e.TimeScale.InsertPremiumLog(logs); err != nil {
							e.logger.Printf("Failed to insert premium log: %v", err)
						} else {
							e.logger.Printf("Inserted %v rows to TimeScale", len(logs))
						}
					}(bufferCopy)
					buffer = make([]database.PremiumLog, 0, 100)
				}
			}
		}
	}()
}

func (e *EngineContext) Stop() {
	e.logger.Printf("Engine stopping...")
	e.cancel()
	e.wg.Wait()
	e.logger.Printf("Engine stopped")
}

func (e *EngineContext) Context() context.Context {
	return e.ctx
}
