package binancemarket_test

import (
	"context"
	"testing"
	"time"

	binancews "cryptoquant.com/m/internal/binance/ws"
	binancemarket "cryptoquant.com/m/streams/binance/market"
)

func TestSubscribeBook(t *testing.T) {
	t.Log("Starting Stream BTCUSDT")
	ch1 := make(chan [2][2]float64)
	handler1 := binancemarket.NewOrderbookHandler(ch1)
	handlers := []func(binancews.FutureBookTicker) error{handler1}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go binancemarket.SubscribeBook(ctx, "BTCUSDT", handlers)

	// Create done channel to signal test completion
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(ch1)

		timeout := time.After(10 * time.Second)
		received := make(map[string]bool)

		for {
			select {
			case <-timeout:
				return
			case price, ok := <-ch1:
				if !ok {
					return
				}
				t.Log("Orderbook:", price)
				received["orderbook"] = true
			}
		}
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(11 * time.Second):
		t.Fatal("Test timed out")
	}
}
