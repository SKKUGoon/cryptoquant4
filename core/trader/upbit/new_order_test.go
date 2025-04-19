package upbittrade_test

import (
	"testing"

	upbittrade "cryptoquant.com/m/core/trader/upbit"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/joho/godotenv"
)

func TestSendOrder(t *testing.T) {
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := upbittrade.NewTrader()
	trader.UpdateRateLimit(2000)
	orderSheet := upbitrest.NewTestOrderSheetLong()

	result, err := trader.SendOrder(*orderSheet)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result.Success)
	t.Logf("result: %+v", result.Error)
}
