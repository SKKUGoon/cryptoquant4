package binancetrade_test

import (
	"testing"

	"github.com/joho/godotenv"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func TestGetSignature(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	orderSheet := binancerest.NewTestOrderSheetLong()
	queryString, signature, err := trader.GenerateSignature(orderSheet)
	if err != nil {
		t.Fatalf("Error getting signature: %v", err)
	}

	t.Logf("queryString: %s", queryString)
	t.Logf("signature: %s", signature)
}
