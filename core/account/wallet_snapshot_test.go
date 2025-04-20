package account_test

import (
	"context"
	"testing"

	"cryptoquant.com/m/core/account"
	"github.com/joho/godotenv"
)

func TestAccountSource_SyncWalletSnapsho(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := account.NewAccountSource(context.Background())
	as.SetPrincipalCurrency("upbit", "KRW")
	as.SetPrincipalCurrency("binance", "USDT")
	err := as.Sync()
	if err != nil {
		t.Fatalf("failed to sync account source: %v", err)
	}

	snapshot := as.GetUpbitWalletSnapshot()
	t.Logf("upbit snapshot: %v", snapshot)

	snapshot = as.GetBinanceWalletSnapshot()
	t.Logf("binance snapshot: %v", snapshot)
}
