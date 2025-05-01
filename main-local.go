//go:build !server && !init && !trader
// +build !server,!init,!trader

package main

import (
	"net/http"
	_ "net/http/pprof"

	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	database "cryptoquant.com/m/data/database"
	binancews "cryptoquant.com/m/internal/binance/ws"
	upbitws "cryptoquant.com/m/internal/upbit/ws"
	s01signal "cryptoquant.com/m/signal/s01"
	strategybase "cryptoquant.com/m/strategy/base"
	s01 "cryptoquant.com/m/strategy/s01"
	binancemarket "cryptoquant.com/m/streams/binance/market"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

// For local development
func init() {
	// Load environment variables
	if err := godotenv.Load(".env.local"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Start pprof server for debugging
	go func() {
		log.Println("pprof listening at :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Enable more verbose logging for development
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting development environment...")
}

func main() {
	var binanceAsset *strategybase.SubscribableAsset
	var upbitAsset *strategybase.SubscribableAsset
	var anchorAsset *strategybase.SubscribableAsset
	var pair *s01.UpbitBinancePair

	ctx := context.Background()

	// Create Channels
	upbitOrderbookChan := make(chan [2][2]float64)
	binanceOrderbookChan := make(chan [2][2]float64)
	defer close(upbitOrderbookChan)
	defer close(binanceOrderbookChan)

	anchorPriceChan := make(chan [2]float64)
	defer close(anchorPriceChan)

	premiumChan := make(chan [3]float64) // Connects to pair and signal engine
	dataLogChan := make(chan database.PremiumLog)
	defer close(premiumChan)
	defer close(dataLogChan)

	symbol, anchorSymbol, binanceSymbol, upbitSymbol := getSignalEngineEnv()

	// Start listening to assets
	binanceAsset = strategybase.NewSubscribableAsset(ctx, binanceSymbol)
	upbitAsset = strategybase.NewSubscribableAsset(ctx, upbitSymbol)
	anchorAsset = strategybase.NewSubscribableAsset(ctx, anchorSymbol)

	binanceAsset.SetOrderbookChan(binanceOrderbookChan)
	upbitAsset.SetOrderbookChan(upbitOrderbookChan)
	anchorAsset.SetTradeChan(anchorPriceChan)

	go binanceAsset.Listen()
	go upbitAsset.Listen()
	go anchorAsset.Listen()

	// Start listening to pair
	pair = s01.NewUpbitBinancePair(ctx, symbol, anchorSymbol)
	pair.SetPremiumChan(premiumChan)
	pair.SubscribeKoreanAsset(upbitAsset)
	pair.SubscribeForeignAsset(binanceAsset)
	pair.SubscribeAnchorPrice(anchorAsset)

	upbitAsset.Check()
	binanceAsset.Check()
	anchorAsset.Check()

	go pair.Run()

	// Start generating signals
	upbitBinanceSignal := s01signal.NewUpbitBinanceSignal(ctx, pair)
	upbitBinanceSignal.UpdateUpbitExchangeConfig()
	upbitBinanceSignal.UpdateBinanceExchangeConfig()
	upbitBinanceSignal.UpdateEnterPremiumBoundary()
	upbitBinanceSignal.UpdateExitPremiumBoundary()
	upbitBinanceSignal.SetPremiumChan(premiumChan)
	upbitBinanceSignal.SetDataLogChan(dataLogChan)
	upbitBinanceSignal.Check()

	go upbitBinanceSignal.Run()

	// Start streams
	upbitHandler1 := upbitmarket.NewOrderbookHandler(upbitOrderbookChan)
	upbitHandlers := []func(upbitws.SpotOrderbook) error{upbitHandler1}
	binanceHandler1 := binancemarket.NewOrderbookHandler(binanceOrderbookChan)
	binanceHandlers := []func(binancews.FutureBookTicker) error{binanceHandler1}
	upbitBinanceAnchorHandler := upbitmarket.NewTradeHandler(anchorPriceChan)
	upbitBinanceAnchorHandlers := []func(upbitws.SpotTrade) error{upbitBinanceAnchorHandler}

	go upbitmarket.SubscribeBook(ctx, upbitSymbol, upbitHandlers)
	go binancemarket.SubscribeBook(ctx, binanceSymbol, binanceHandlers)
	go upbitmarket.SubscribeTrade(ctx, anchorSymbol, upbitBinanceAnchorHandlers)

	// Wait for interrupt signal
	extSigChan := make(chan os.Signal, 1)
	signal.Notify(extSigChan, syscall.SIGINT, syscall.SIGTERM)
	<-extSigChan

	log.Println("Shutting down gracefully...")
}

func getSignalEngineEnv() (string, string, string, string) {
	var symbol string
	var anchorSymbol string
	var binanceSymbol string
	var upbitSymbol string

	if symbol = os.Getenv("SYMBOL"); symbol == "" {
		panic("environment variable `SYMBOL` is not set")
	}

	if anchorSymbol = os.Getenv("ANCHOR_SYMBOL"); anchorSymbol == "" {
		panic("environment variable `ANCHOR_SYMBOL` is not set")
	}

	if binanceSymbol = os.Getenv("BINANCE_SYMBOL"); binanceSymbol == "" {
		panic("environment variable `BINANCE_SYMBOL` is not set")
	}

	if upbitSymbol = os.Getenv("UPBIT_SYMBOL"); upbitSymbol == "" {
		panic("environment variable `UPBIT_SYMBOL` is not set")
	}

	return symbol, anchorSymbol, binanceSymbol, upbitSymbol
}
