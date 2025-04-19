package binancetrade_test

import (
	"testing"

	"github.com/joho/godotenv"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func TestGetSignature(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	orderSheet := binancerest.NewTestOrderSheetLong()
	queryString, signature, err := trader.GetSignature(*orderSheet)
	if err != nil {
		t.Fatalf("Error getting signature: %v", err)
	}

	t.Logf("queryString: %s", queryString)
	t.Logf("signature: %s", signature)
}

func TestGetSignatureBatch(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	orderSheet1 := binancerest.NewTestOrderSheetLong()
	orderSheet2 := binancerest.NewTestOrderSheetShort()
	orderSheets := []binancerest.OrderSheet{*orderSheet1, *orderSheet2}
	queryString, signature, err := trader.GetSignatureBatch(orderSheets)
	if err != nil {
		t.Fatalf("Error getting signature batch: %v", err)
	}

	t.Logf("queryString: %s", queryString)
	t.Logf("signature: %s", signature)
}
