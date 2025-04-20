//go:build server && !init && !trader
// +build server,!init,!trader

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	database "cryptoquant.com/m/data/database"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	sig "cryptoquant.com/m/signal"
)

func init() {
	// Set leverage here
	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		panic(err)
	}
	symbol := os.Getenv("BINANCE_SYMBOL")
	if symbol == "" {
		log.Fatalf("BINANCE_SYMBOL is not set")
		panic(err)
	}

	// Trade setup
	// 1. Set leverage
	maxLeverageKey := fmt.Sprintf("binance_%v_max_leverage", strings.ToLower(symbol))
	maxLeverage, err := db.GetTradeMetadata(maxLeverageKey, 1)
	if err != nil {
		log.Printf("Failed to get max leverage: %v", err)
		panic(err)
	}
	trader := binancetrade.NewTrader()
	trader.UpdateRateLimit(2000)
	levReq := &binancerest.LeverageRequest{
		Symbol:    symbol,
		Leverage:  maxLeverage.(int),
		Timestamp: time.Now().UnixMilli(),
	}
	log.Printf("Leverage request: %+v", levReq)
	levResp, err := trader.SetLeverage(levReq)
	if err != nil {
		log.Printf("Failed to set leverage: %v", err)
		panic(err)
	}
	log.Printf("Leverage set: %+v", levResp)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize engine with production settings
	engine := sig.New(ctx)
	engine.ConfirmTargetSymbols()
	engine.ConfirmTradeParameters()

	// Start all necessary components
	engine.StartAssetPair()
	engine.StartAssetStreams()
	engine.Run()
	engine.StartTSLog()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	engine.Stop()
}
