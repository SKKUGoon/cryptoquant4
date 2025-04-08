package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cryptoquant.com/m/engine"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine := engine.New(ctx)
	engine.ConfirmTargetSymbols()

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
