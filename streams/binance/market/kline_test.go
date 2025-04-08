package binancemarket_test

import (
	"context"
	"testing"
	"time"

	binancews "cryptoquant.com/m/internal/binance/ws"
	binancemarket "cryptoquant.com/m/streams/binance/market"
)

func TestSubscribeKline(t *testing.T) {
	t.Log("Starting Stream BTCUSDT 1m")
	ch := make(chan float64)
	handler := binancemarket.NewCloseHandler(ch)
	handlers := []func(binancews.KlineDataStream) error{handler}
	ctx, cancel := context.WithCancel(context.Background())
	go binancemarket.SubscribeKline(ctx, "BTCUSDT", "1m", handlers)

	// Send done signal after 10 seconds
	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		close(ch)
	}()

	for price := range ch {
		t.Log(price)
	}
}
