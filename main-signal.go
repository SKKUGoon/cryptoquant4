//go:build server && !init && !trader
// +build server,!init,!trader

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	sig "cryptoquant.com/m/signal"
)

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
