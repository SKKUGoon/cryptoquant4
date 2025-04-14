package upbituser_test

import (
	"context"
	"testing"
	"time"

	upbitws "cryptoquant.com/m/internal/upbit/ws"
	upbituser "cryptoquant.com/m/streams/upbit/user"
	"github.com/joho/godotenv"
)

const PGENVLOC = "../../../.env"

func TestSubscribeMyAsset(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	ch := make(chan float64)
	handler := upbituser.NewAssetBalanceKRWHandler(ch)
	handlers := []func(upbitws.MyAssetResponse) error{handler}
	ctx, cancel := context.WithCancel(context.Background())
	go upbituser.SubscribeMyAsset(ctx, handlers)

	go func() {
		time.Sleep(10 * time.Second)
		close(ch)
		cancel()
	}()

	for balance := range ch {
		t.Log(balance)
	}
}
