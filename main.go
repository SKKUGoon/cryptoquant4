package main

import (
	"net/http"
	_ "net/http/pprof"

	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"cryptoquant.com/m/engine"
	"github.com/joho/godotenv"
)

// For local development
func init() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	go func() {
		log.Println("pprof listening at :6060")
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
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
