package trade_test

import (
	"testing"

	"github.com/joho/godotenv"

	"cryptoquant.com/m/engine/trade"
	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func TestGetSignature(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := trade.NewTrader()
	orderSheet := binancerest.NewTestOrderSheetLong()
	queryString, signature := trader.GetSignature(*orderSheet)

	t.Logf("queryString: %s", queryString)
	t.Logf("signature: %s", signature)
}

func TestGetSignatureBatch(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := trade.NewTrader()
	orderSheet1 := binancerest.NewTestOrderSheetLong()
	orderSheet2 := binancerest.NewTestOrderSheetShort()
	orderSheets := []binancerest.OrderSheet{*orderSheet1, *orderSheet2}
	queryString, signature := trader.GetSignatureBatch(orderSheets)

	t.Logf("queryString: %s", queryString)
	t.Logf("signature: %s", signature)
}
