package engine_test

import (
	"context"
	"testing"

	"cryptoquant.com/m/engine"
	"github.com/joho/godotenv"
)

func TestAccountSource_OnInit(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := engine.NewAccountSource(context.Background())

	// Init account source
	if err := as.OnInit(); err != nil {
		t.Fatalf("failed to init account source: %v", err)
	}
}

func TestAccountSource_Sync(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Create account source
	as := engine.NewAccountSourceSync(context.Background())

	// Check if redis's information is inserted

	t.Log("Upbit Fund")
	t.Logf("%+v", as.UpbitFund)
	t.Log("Binance Fund")
	t.Logf("%+v", as.BinanceFund)
}
