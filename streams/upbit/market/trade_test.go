package upbitmarket_test

import (
	"context"
	"testing"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

func TestSubscribeTrade(t *testing.T) {
	t.Log("Starting Stream KRW-BTC")
	ch := make(chan float64)
	handler := upbitmarket.NewPriceHandler(ch)
	handlers := []func(upbitws.SpotTrade) error{handler}
	ctx, cancel := context.WithCancel(context.Background())
	go upbitmarket.SubscribeTrade(ctx, "KRW-BTC", handlers)

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		close(ch)
	}()

	for trade := range ch {
		t.Log(trade)
	}
}
