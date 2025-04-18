package engine

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

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
	premiumChan      chan [2]float64 // [EnterPremium, ExitPremium]
	upbitOrderChan   chan upbitrest.OrderSheet
	binanceOrderChan chan binancerest.OrderSheet

	// logger
	tsLog chan database.PremiumLog
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
	kimchiConfig, err := config.NewUpbitSpotTradeConfig()
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

	// 6. Create traders
	kimchiTrader := upbittrade.NewTrader()
	binanceTrader := binancetrade.NewTrader()

	// 7. Create account source
	accountSource := account.NewAccountSource(engineCtx)

	// 8. Create struct with order channels
	engine := &EngineContext{
		EngineName:            engineName,
		ctx:                   engineCtx,
		cancel:                cancel,
		UpbitExchangeConfig:   kimchiConfig,
		BinanceExchangeConfig: binanceConfig,
		UpbitTrader:           kimchiTrader,
		BinanceTrader:         binanceTrader,
		AccountSource:         accountSource,
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

		inPosition:       false,
		premiumChan:      make(chan [2]float64),
		upbitOrderChan:   make(chan upbitrest.OrderSheet),
		binanceOrderChan: make(chan binancerest.OrderSheet),

		tsLog: make(chan database.PremiumLog),
	}

	log.Println("Engine initialized")
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
		log.Println("Failed to confirm target symbols: Environment variables not set")
		panic("Environment variables not set")
	}

	// Confirm trading symbols in cefi and kimchi
	if !e.BinanceExchangeConfig.IsAvailableSymbol(csymbol) {
		log.Println("Failed to confirm target symbols: Binance symbol not available")
		panic("Binance symbol not available")
	}

	if !e.UpbitExchangeConfig.IsAvailableSymbol(ksymbol) {
		log.Println("Failed to confirm target symbols: Kimchi symbol not available")
		panic("Kimchi symbol not available")
	}

	// Confirm anchor symbol
	if !e.UpbitExchangeConfig.IsAvailableSymbol(e.AnchorAssetSymbol) {
		log.Println("Failed to confirm target symbols: Kimchi anchor symbol not available")
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
	log.Println("Confirming trade parameters")

	// Premium Calculation Parameters key value
	enterPremiumBoundaryKey := fmt.Sprintf("%v_enter_premium_boundary", e.EngineName)
	exitPremiumBoundaryKey := fmt.Sprintf("%v_exit_premium_boundary", e.EngineName)

	// Premium Calculation Parameters

	// EnterPremiumBoundary: 0.9980
	enterPremiumBoundary, err := e.Database.GetTradeMetadata(enterPremiumBoundaryKey, 0.9980)
	if err != nil {
		log.Printf("Failed to get enter premium boundary: %v", err)
		panic(err)
	}
	e.EnterPremiumBoundary = enterPremiumBoundary.(float64)

	// ExitPremiumBoundary: 1.0035
	exitPremiumBoundary, err := e.Database.GetTradeMetadata(exitPremiumBoundaryKey, 1.0035)
	if err != nil {
		log.Printf("Failed to get exit premium boundary: %v", err)
		panic(err)
	}
	e.ExitPremiumBoundary = exitPremiumBoundary.(float64)

	log.Printf("Trade parameters confirmed: enterPremiumBoundary: %v, exitPremiumBoundary: %v", e.EnterPremiumBoundary, e.ExitPremiumBoundary)
}

// StartAssetPair initializes the asset objects for the trading pairs and sets up their channels.
// It creates:
// 1. KimchiAsset: Represents the Kimchi trading pair
// 2. ForexAsset: Represents the Cefi trading pair
// 3. AnchorAsset: Represents the anchor trading pair
func (e *EngineContext) StartAssetPair() {
	log.Printf("Starting asset for %v", e.UpbitAssetSymbol)
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
		PremiumChan: e.premiumChan,
	}

	log.Printf("Asset Pairs(%v, %v, %v) initialized", e.UpbitAssetSymbol, e.BinanceAssetSymbol, e.AnchorAssetSymbol)
}

// StartAssetStreams starts the streams for the asset pairs.
// It creates total of 5 streams:
// 1. KimchiStream: Represents the Kimchi trading pair (Upbit, 2 streams: price and book)
// 2. CefiStream: Represents the Cefi trading pair (Binance, 2 streams: price and book)
// 3. AnchorStream: Represents the anchor trading pair (Upbit, 1 stream: price)
func (e *EngineContext) StartAssetStreams() {
	log.Println("Engine started")

	// Start stream - Kimchi - Upbit
	log.Printf("Starting stream for %v", e.UpbitAssetSymbol)
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
	log.Printf("Starting stream for %v", e.BinanceAssetSymbol)
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
	log.Printf("Starting stream for %v", e.AnchorAssetSymbol)
	anchor1 := upbitmarket.NewPriceHandler(e.anchorTradeChan)
	h3s := []func(upbitws.SpotTrade) error{anchor1}
	go upbitmarket.SubscribeTrade(e.ctx, e.AnchorAssetSymbol, h3s)
}

// Run starts the arbitrage strategy.
// It creates a goroutine that runs the arbitrage strategy and tracks consecutive failures.
// If the strategy is not ready, it logs an error and panics. Otherwise, it logs a message and starts the strategy.
func (e *EngineContext) Run() {
	log.Println("Starting strategy")
	go e.UpbitBinancePairs.Run(e.ctx)

	go func() {
		var entryUpbitOrderSheet upbitrest.OrderSheet     // Long
		var exitUpbitOrderSheet upbitrest.OrderSheet      // Exit position
		var entryBinanceOrderSheet binancerest.OrderSheet // Short
		var exitBinanceOrderSheet binancerest.OrderSheet  // Exit position
		var err error

		for {
			select {
			case <-e.ctx.Done():
				return
			case premiums := <-e.premiumChan:
				enter := premiums[0]
				exit := premiums[1]

				switch true {
				case e.inPosition && exit > e.ExitPremiumBoundary:
					log.Printf("Exiting position: %v (standard: %v)", exit, e.ExitPremiumBoundary)

					// Create order sheets
					if exitUpbitOrderSheet, exitBinanceOrderSheet, err = e.UpbitBinancePairs.CreatePremiumShortOrders(
						e.UpbitAssetSymbol,
					); err != nil {
						log.Printf("Failed to create premium short orders: %v", err)
						continue
					}

					// TODO: Make short orders (Exit)
					log.Printf("Exit order sheets: %+v, %+v", exitUpbitOrderSheet, exitBinanceOrderSheet)

					e.AccountSource.Update()
				case !e.inPosition && enter < e.EnterPremiumBoundary:
					log.Printf("Entering position: %v (standard: %v)", enter, e.EnterPremiumBoundary)

					// Create order sheets
					if entryUpbitOrderSheet, entryBinanceOrderSheet, err = e.UpbitBinancePairs.CreatePremiumLongOrders(
						e.AccountSource.UpbitFund.AvailableFund,
						e.AccountSource.BinanceFund.AvailableFund,
					); err != nil {
						log.Printf("Failed to create premium long orders: %v", err)
						continue
					}

					// Audit order sheet precision
					err = e.UpbitExchangeConfig.AuditOrderSheetPrecision(&entryUpbitOrderSheet)
					if err != nil {
						log.Printf("Failed to audit order sheet precision: %v", err)
						continue
					}
					err = e.BinanceExchangeConfig.AuditOrderSheetPrecision(&entryBinanceOrderSheet)
					if err != nil {
						log.Printf("Failed to audit order sheet precision: %v", err)
						continue
					}

					// TODO: Send order sheet

					e.AccountSource.Update()
					e.EnterExitPosition()
				}
				log.Printf("Premium: %v, %v", enter, exit)
			}
		}
	}()
}

// StartTSLog starts the TimeScale log.
// It creates a goroutine that logs the premium data to the TimeScale database.
func (e *EngineContext) StartTSLog() {
	log.Println("Starting DB log")
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
							log.Printf("Failed to insert premium log: %v", err)
						} else {
							log.Printf("Inserted %v rows to TimeScale", len(logs))
						}
					}(bufferCopy)
					buffer = make([]database.PremiumLog, 0, 100)
				}
			}
		}
	}()
}

func (e *EngineContext) Stop() {
	log.Println("Engine stopping...")
	e.cancel()
	e.wg.Wait()
	log.Println("Engine stopped")
}

func (e *EngineContext) Context() context.Context {
	return e.ctx
}

func (e *EngineContext) EnterExitPosition() {
	e.inPosition = !e.inPosition
}
