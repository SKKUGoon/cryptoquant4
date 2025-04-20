package account_test

import (
	"context"
	"testing"

	"cryptoquant.com/m/core/account"
	"github.com/joho/godotenv"
)

const ENV_PATH = "../../.env.local"

func TestAccountSource_OnInit(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := account.NewAccountSource(context.Background())
	as.SetPrincipalCurrency("upbit", "KRW")
	as.SetPrincipalCurrency("binance", "USDT")

	// Init account source
	if err := as.OnInit(); err != nil {
		t.Fatalf("failed to init account source: %v", err)
	}
}

func TestAccountSource_Sync(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := account.NewAccountSource(context.Background())

	err := as.UpdateRedis()
	if err != nil {
		t.Fatalf("failed to update account source: %v", err)
	}
	err = as.Sync()
	if err != nil {
		t.Fatalf("failed to sync account source: %v", err)
	}

	t.Logf("binance reserved fund: %v", as.GetBinanceWalletSnapshot())
	t.Logf("upbit reserved fund: %v", as.GetUpbitWalletSnapshot())
}
