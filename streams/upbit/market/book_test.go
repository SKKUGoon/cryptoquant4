package upbitmarket_test

import (
	"context"
	"testing"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

func TestSubscribeBook(t *testing.T) {
	t.Log("Starting Stream KRW-BTC")
	ch := make(chan float64)
	handler := upbitmarket.NewBestBidPrcHandler(ch)
	handlers := []func(upbitws.SpotOrderbook) error{handler}
	ctx, cancel := context.WithCancel(context.Background())
	go upbitmarket.SubscribeBook(ctx, "KRW-BTC", handlers)

	go func() {
		time.Sleep(15 * time.Second)
		cancel()
		close(ch)
	}()

	for trade := range ch {
		t.Log(trade)
	}
}
