package account_test

import (
	"context"
	"testing"

	"cryptoquant.com/m/core/account"
	"github.com/joho/godotenv"
)

func TestAccountSource_SyncFromRedisPosition(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := account.NewAccountSource(context.Background())
	as.SetPrincipalCurrency("upbit", "KRW")
	as.SetPrincipalCurrency("binance", "USDT")
	err := as.Sync()
	if err != nil {
		t.Fatalf("failed to update account source: %v", err)
	}

	amount, err := as.SyncFromRedisPosition("upbit", as.UpbitPrincipalCurrency)
	if err != nil {
		t.Fatalf("failed to sync from redis position: %v", err)
	}
	t.Logf("KRW amount: %v", amount)

	amount, err = as.SyncFromRedisPosition("binance", as.BinancePrincipalCurrency)
	if err != nil {
		t.Fatalf("failed to sync from redis position: %v", err)
	}
	t.Logf("USDT amount: %v", amount)
}
