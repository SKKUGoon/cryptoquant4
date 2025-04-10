//go:build server
// +build server

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cryptoquant.com/m/engine"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize engine with production settings
	engine := engine.New(ctx)
	engine.ConfirmTargetSymbols()

	// Start all necessary components
	engine.StartAsset()
	engine.StartMonitor()
	engine.StartStrategy()
	engine.StartTSLog()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down gracefully...")
	engine.Stop()
}
