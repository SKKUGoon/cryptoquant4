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
	ch1 := make(chan float64)
	ch2 := make(chan float64)
	ch3 := make(chan float64)
	ch4 := make(chan float64)
	handler1 := binancemarket.NewBestBidPrcHandler(ch1)
	handler2 := binancemarket.NewBestAskPrcHandler(ch2)
	handler3 := binancemarket.NewBestBidQtyHandler(ch3)
	handler4 := binancemarket.NewBestAskQtyHandler(ch4)
	handlers := []func(binancews.FutureBookTicker) error{handler1, handler2, handler3, handler4}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go binancemarket.SubscribeBook(ctx, "BTCUSDT", handlers)

	// Create done channel to signal test completion
	done := make(chan struct{})

	go func() {
		defer close(done)
		defer close(ch1)
		defer close(ch2)
		defer close(ch3)
		defer close(ch4)

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
				t.Log("Best Bid Price:", price)
				received["bid_price"] = true
			case price, ok := <-ch2:
				if !ok {
					return
				}
				t.Log("Best Ask Price:", price)
				received["ask_price"] = true
			case qty, ok := <-ch3:
				if !ok {
					return
				}
				t.Log("Best Bid Qty:", qty)
				received["bid_qty"] = true
			case qty, ok := <-ch4:
				if !ok {
					return
				}
				t.Log("Best Ask Qty:", qty)
				received["ask_qty"] = true
			}

			// Check if we've received at least one of each type
			if received["bid_price"] && received["ask_price"] &&
				received["bid_qty"] && received["ask_qty"] {
				return
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
