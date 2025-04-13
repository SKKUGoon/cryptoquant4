package upbittrade_test

import (
	"testing"

	"github.com/joho/godotenv"

	upbittrade "cryptoquant.com/m/engine/trade/upbit"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

const ENVLOC = "../../../.env"

func TestGetSignature(t *testing.T) {
	if err := godotenv.Load(ENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := upbittrade.NewTrader()
	orderSheet := &upbitrest.OrderSheet{
		Symbol:  "KRW-BTC",
		Side:    "bid",
		Volume:  "0.001",
		Price:   "50000000",
		OrdType: "limit",
	}

	signature, authToken, err := trader.GetSignature(*orderSheet)
	if err != nil {
		t.Fatalf("Error getting signature: %v", err)
	}

	t.Logf("signature: %s", signature)
	t.Logf("authToken: %s", authToken)
}
