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

	sig "cryptoquant.com/m/signal"
	"github.com/joho/godotenv"
)

// For local development
func init() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize engine with development settings
	engine := sig.New(ctx)
	engine.ConfirmTargetSymbols()
	engine.ConfirmTradeParameters()

	// Start all components with development logging
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
