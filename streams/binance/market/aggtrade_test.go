package binancemarket_test

import (
	"context"
	"testing"
	"time"

	binancews "cryptoquant.com/m/internal/binance/ws"
	binancemarket "cryptoquant.com/m/streams/binance/market"
)

func TestSubscribeAggtrade(t *testing.T) {
	t.Log("Starting Stream BTCUSDT")
	ch := make(chan float64)
	handler := binancemarket.NewPriceHandler(ch)
	handlers := []func(binancews.FutureAggTrade) error{handler}
	ctx, cancel := context.WithCancel(context.Background())
	go binancemarket.SubscribeAggtrade(ctx, "BTCUSDT", handlers)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		close(ch)
	}()

	for price := range ch {
		t.Log(price)
	}
}
