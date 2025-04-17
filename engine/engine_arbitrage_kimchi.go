package engine

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	config "cryptoquant.com/m/config"
	database "cryptoquant.com/m/data/database"
	binancesource "cryptoquant.com/m/data/sources/binance"
	account "cryptoquant.com/m/engine/account"
	binancetrade "cryptoquant.com/m/engine/trade/binance"
	upbittrade "cryptoquant.com/m/engine/trade/upbit"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	binancews "cryptoquant.com/m/internal/binance/ws"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	upbitws "cryptoquant.com/m/internal/upbit/ws"
	kimchiarbv1 "cryptoquant.com/m/strategy/arbitrage/kimchi_v1"
	binancemarket "cryptoquant.com/m/streams/binance/market"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

// EngineContext represents the core trading engine context that manages configurations
// for both Upbit (Korean) and Binance exchanges for cross-exchange arbitrage
// TODO: Add more exchanges support
type EngineContext struct {
	EngineName string

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Exchange configurations
	UpbitExchangeConfig   *config.UpbitSpotTradeConfig
	BinanceExchangeConfig *config.BinanceFutureTradeConfig

	// Traders
	UpbitTrader   *upbittrade.Trader
	BinanceTrader *binancetrade.Trader

	// Accounts - One true source
	AccountSource *account.AccountSource

	// Data
	Database         *database.Database
	TimeScale        *database.TimeScale
	FutureMarketData *binancesource.BinanceFutureMarketData

	// Strategy
	UpbitBinancePairs    *kimchiarbv1.UpbitBinancePair
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
	inPosition       bool
	upbitOrderChan   chan upbitrest.OrderSheet
	binanceOrderChan chan binancerest.OrderSheet

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
	// 1. Get engine name
	engineName := os.Getenv("ENGINE_NAME")
	if engineName == "" {
		panic("ENGINE_NAME is not set")
	}

	// 2. Create a new context with cancellation
	// Controls the lifecycle of the whole engine and daemon structs and streams
	engineCtx, cancel := context.WithCancel(ctx)

	// 3. Set up logging
	logger := setupLog()

	// 4. Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		logger.Printf("Failed to connect to database: %v", err)
		panic(err)
	}
	ts, err := database.ConnectTS()
	if err != nil {
		logger.Printf("Failed to connect to TimeScale: %v", err)
		panic(err)
	}

	// 5. Create exchange configs
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

	// 6. Confirm anchor symbol
	anchor := os.Getenv("ANCHOR_SYMBOL")
	if anchor == "" {
		logger.Printf("Failed to confirm anchor symbol: %v", "Environment variables not set")
		panic("Environment variables not set")
	}

	// 7. Create traders
	kimchiTrader := upbittrade.NewTrader()
	binanceTrader := binancetrade.NewTrader()

	// 8. Create struct with order channels
	engine := &EngineContext{
		EngineName:            engineName,
		ctx:                   engineCtx,
		cancel:                cancel,
		UpbitExchangeConfig:   kimchiConfig,
		BinanceExchangeConfig: binanceConfig,
		UpbitTrader:           kimchiTrader,
		BinanceTrader:         binanceTrader,
		Database:              db,
		TimeScale:             ts,
		logger:                logger,
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

		inPosition:       false,
		upbitOrderChan:   make(chan upbitrest.OrderSheet),
		binanceOrderChan: make(chan binancerest.OrderSheet),

		tsLog: make(chan database.PremiumLog),
	}

	logger.Printf("Engine initialized")

	return engine
}

// ConfirmTargetSymbols verifies that the trading symbols specified in environment variables
// are available for trading on both Binance and Upbit exchanges. It checks:
// 1. BINANCE_SYMBOL and UPBIT_SYMBOL environment variables are set
// 2. The symbols are available for trading on their respective exchanges
// 3. The anchor asset symbol is available on Upbit
// If any check fails, it logs an error and panics. Otherwise, it sets the confirmed
// symbols in the engine context.
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

	if !e.UpbitExchangeConfig.IsAvailableSymbol(ksymbol) {
		e.logger.Printf("Failed to confirm target symbols: %v", "Kimchi symbol not available")
		panic("Kimchi symbol not available")
	}

	// Confirm anchor symbol
	if !e.UpbitExchangeConfig.IsAvailableSymbol(e.AnchorAssetSymbol) {
		e.logger.Printf("Failed to confirm target symbols: %v", "Kimchi anchor symbol not available")
		panic("Kimchi anchor symbol not available")
	}

	e.UpbitAssetSymbol = ksymbol
	e.BinanceAssetSymbol = csymbol
}

// ConfirmTradeParameters retrieves trade parameters from the database and sets them in the engine context.
// It retrieves: (as a default)
// 1. EnterPremiumBoundary: 0.9980
// 2. ExitPremiumBoundary: 1.0035
// These parameters are used to determine the entry and exit points for the arbitrage strategy.
func (e *EngineContext) ConfirmTradeParameters() {
	e.logger.Printf("Confirming trade parameters")

	// Premium Calculation Parameters key value
	enterPremiumBoundaryKey := fmt.Sprintf("%v_enter_premium_boundary", e.EngineName)
	exitPremiumBoundaryKey := fmt.Sprintf("%v_exit_premium_boundary", e.EngineName)

	// Premium Calculation Parameters

	// EnterPremiumBoundary: 0.9980
	enterPremiumBoundary, err := e.Database.GetTradeMetadata(enterPremiumBoundaryKey, 0.9980)
	if err != nil {
		e.logger.Printf("Failed to get enter premium boundary: %v", err)
		panic(err)
	}
	e.EnterPremiumBoundary = enterPremiumBoundary.(float64)

	// ExitPremiumBoundary: 1.0035
	exitPremiumBoundary, err := e.Database.GetTradeMetadata(exitPremiumBoundaryKey, 1.0035)
	if err != nil {
		e.logger.Printf("Failed to get exit premium boundary: %v", err)
		panic(err)
	}
	e.ExitPremiumBoundary = exitPremiumBoundary.(float64)

	e.logger.Printf("Trade parameters confirmed: enterPremiumBoundary: %v, exitPremiumBoundary: %v", e.EnterPremiumBoundary, e.ExitPremiumBoundary)
}

// StartAssetPair initializes the asset objects for the trading pairs and sets up their channels.
// It creates:
// 1. KimchiAsset: Represents the Kimchi trading pair
// 2. ForexAsset: Represents the Cefi trading pair
// 3. AnchorAsset: Represents the anchor trading pair
func (e *EngineContext) StartAssetPair() {
	e.logger.Printf("Starting asset for %v", e.UpbitAssetSymbol)
	kimchiAsset := kimchiarbv1.NewAsset(e.UpbitAssetSymbol)
	forexAsset := kimchiarbv1.NewAsset(e.BinanceAssetSymbol)
	anchorAsset := kimchiarbv1.NewAsset(e.AnchorAssetSymbol)

	kimchiAsset.SetPriceChannel(e.upbitTradeChan)
	forexAsset.SetPriceChannel(e.binanceTradeChan)
	anchorAsset.SetPriceChannel(e.anchorTradeChan)

	kimchiAsset.SetBestBidPrcChannel(e.upbitBestBidPrcChan)
	forexAsset.SetBestBidPrcChannel(e.binanceBestBidPrcChan)
	anchorAsset.SetBestBidPrcChannel(e.anchorBestBidPrcChan)

	kimchiAsset.SetBestAskPrcChannel(e.upbitBestAskPrcChan)
	forexAsset.SetBestAskPrcChannel(e.binanceBestAskPrcChan)
	anchorAsset.SetBestAskPrcChannel(e.anchorBestAskPrcChan)

	kimchiAsset.SetBestBidQtyChannel(e.upbitBestBidQtyChan)
	forexAsset.SetBestBidQtyChannel(e.binanceBestBidQtyChan)
	anchorAsset.SetBestBidQtyChannel(e.anchorBestBidQtyChan)

	kimchiAsset.SetBestAskQtyChannel(e.upbitBestAskQtyChan)
	forexAsset.SetBestAskQtyChannel(e.binanceBestAskQtyChan)
	anchorAsset.SetBestAskQtyChannel(e.anchorBestAskQtyChan)

	e.UpbitBinancePairs = &kimchiarbv1.UpbitBinancePair{
		AnchorAsset: anchorAsset,
		CefiAsset:   forexAsset,
		KimchiAsset: kimchiAsset,
	}

	e.logger.Printf("Asset Pairs(%v, %v, %v) initialized", e.UpbitAssetSymbol, e.BinanceAssetSymbol, e.AnchorAssetSymbol)
}

// StartAssetStreams starts the streams for the asset pairs.
// It creates total of 5 streams:
// 1. KimchiStream: Represents the Kimchi trading pair (Upbit, 2 streams: price and book)
// 2. CefiStream: Represents the Cefi trading pair (Binance, 2 streams: price and book)
// 3. AnchorStream: Represents the anchor trading pair (Upbit, 1 stream: price)
func (e *EngineContext) StartAssetStreams() {
	e.logger.Printf("Engine started")

	// Start stream - Kimchi - Upbit
	e.logger.Printf("Starting stream for %v", e.UpbitAssetSymbol)
	kimchi1 := upbitmarket.NewPriceHandler(e.upbitTradeChan)
	kimchi2 := upbitmarket.NewBestBidPrcHandler(e.upbitBestBidPrcChan)
	kimchi3 := upbitmarket.NewBestBidQtyHandler(e.upbitBestBidQtyChan)
	kimchi4 := upbitmarket.NewBestAskPrcHandler(e.upbitBestAskPrcChan)
	kimchi5 := upbitmarket.NewBestAskQtyHandler(e.upbitBestAskQtyChan)
	kimchiHandlerPrice := []func(upbitws.SpotTrade) error{kimchi1}
	kimchiHandlerBook := []func(upbitws.SpotOrderbook) error{kimchi2, kimchi3, kimchi4, kimchi5}
	go upbitmarket.SubscribeTrade(e.ctx, e.UpbitAssetSymbol, kimchiHandlerPrice)
	go upbitmarket.SubscribeBook(e.ctx, e.UpbitAssetSymbol, kimchiHandlerBook)

	// Start stream - Binance
	e.logger.Printf("Starting stream for %v", e.BinanceAssetSymbol)
	cefi1 := binancemarket.NewPriceHandler(e.binanceTradeChan)
	cefi2 := binancemarket.NewBestBidPrcHandler(e.binanceBestBidPrcChan)
	cefi3 := binancemarket.NewBestBidQtyHandler(e.binanceBestBidQtyChan)
	cefi4 := binancemarket.NewBestAskPrcHandler(e.binanceBestAskPrcChan)
	cefi5 := binancemarket.NewBestAskQtyHandler(e.binanceBestAskQtyChan)
	cefiHandlerPrice := []func(binancews.FutureAggTrade) error{cefi1}
	cefiHandlerBook := []func(binancews.FutureBookTicker) error{cefi2, cefi3, cefi4, cefi5}
	go binancemarket.SubscribeAggtrade(e.ctx, e.BinanceAssetSymbol, cefiHandlerPrice)
	go binancemarket.SubscribeBook(e.ctx, e.BinanceAssetSymbol, cefiHandlerBook)

	// Start stream - Anchor - Upbit
	e.logger.Printf("Starting stream for %v", e.AnchorAssetSymbol)
	anchor1 := upbitmarket.NewPriceHandler(e.anchorTradeChan)
	h3s := []func(upbitws.SpotTrade) error{anchor1}
	go upbitmarket.SubscribeTrade(e.ctx, e.AnchorAssetSymbol, h3s)
}

// StartStrategy starts the arbitrage strategy.
// It creates a goroutine that runs the arbitrage strategy and tracks consecutive failures.
// If the strategy is not ready, it logs an error and panics. Otherwise, it logs a message and starts the strategy.
func (e *EngineContext) StartStrategy() {
	e.logger.Printf("Starting strategy")
	go e.UpbitBinancePairs.Run(e.ctx)

	// Track consecutive failures
	consecutiveFailures := 0
	const maxConsecutiveFailures = 5000 // Adjust this threshold as needed

	go func() {
		var entryUpbitOrderSheet upbitrest.OrderSheet // Long
		// var exitUpbitOrderSheet upbitrest.OrderSheet      // Exit position
		var entryBinanceOrderSheet binancerest.OrderSheet // Short
		// var exitBinanceOrderSheet binancerest.OrderSheet  // Exit position
		var err error

		for {
			select {
			case <-e.ctx.Done():
				return
			case <-time.Tick(100 * time.Millisecond):
				ok, msg := e.UpbitBinancePairs.Status()
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

				// Check the premium boundary
				if e.inPosition && e.UpbitBinancePairs.CheckExit(e.ExitPremiumBoundary) {
					e.logger.Printf("Exiting position")
					// e.orderChan <- "exit"
					// TODO: Implement this
					e.AccountSource.Update()
					e.logger.Printf("Account updated")
				} else if !e.inPosition && e.UpbitBinancePairs.CheckEnter(e.EnterPremiumBoundary) {
					e.logger.Printf("Entering position")

					// Create order sheets
					entryUpbitOrderSheet, entryBinanceOrderSheet, err = e.UpbitBinancePairs.CreatePremiumLongOrders(
						e.AccountSource.UpbitFund.AvailableFund,
						e.AccountSource.BinanceFund.AvailableFund,
					)
					if err != nil {
						e.logger.Printf("Failed to create premium long orders: %v", err)
						continue
					}

					// Audit order sheet precision
					err = e.UpbitExchangeConfig.AuditOrderSheetPrecision(&entryUpbitOrderSheet)
					if err != nil {
						e.logger.Printf("Failed to audit order sheet precision: %v", err)
						continue
					}
					err = e.BinanceExchangeConfig.AuditOrderSheetPrecision(&entryBinanceOrderSheet)
					if err != nil {
						e.logger.Printf("Failed to audit order sheet precision: %v", err)
						continue
					}

					fmt.Printf("%+v\n", entryUpbitOrderSheet)
					fmt.Printf("%+v\n", entryBinanceOrderSheet)

					// Send order sheets
					// upbitResult, err := e.KimchiTrader.SendOrder(upbitOrderSheet)
					// if err != nil {
					// 	e.logger.Printf("Failed to send upbit order: %v", err)
					// 	continue
					// }
					// binanceResult, err := e.BinanceTrader.SendOrder(binanceOrderSheet)
					// if err != nil {
					// 	e.logger.Printf("Failed to send binance order: %v", err)
					// 	continue
					// }

					// Update account source
					e.AccountSource.Update()
					e.EnterExitPosition()
					e.logger.Printf("Account updated")
				}

				log := e.UpbitBinancePairs.ToPremiumLog()
				e.tsLog <- log
			}
		}
	}()
}

// StartTSLog starts the TimeScale log.
// It creates a goroutine that logs the premium data to the TimeScale database.
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
					copy(bufferCopy, buffer) // Copy the buffer to avoid race condition
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

func (e *EngineContext) EnterExitPosition() {
	e.inPosition = !e.inPosition
}
