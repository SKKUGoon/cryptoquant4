package binancetrade_test

import (
	"os"
	"testing"
	"time"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	"github.com/joho/godotenv"
)

const ENV_PATH = "../../../.env.local"

func TestSetLeverageTestServer(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.SetTestPubKey(os.Getenv("BINANCE_TESTNET_API_KEY"))
	trader.SetTestPriKey(os.Getenv("BINANCE_TESTNET_SECRET_KEY"))
	trader.UpdateRateLimit(1000)

	levReq := &binancerest.LeverageRequest{
		Symbol:    "BTCUSDT",
		Leverage:  10,
		Timestamp: time.Now().UnixMilli(),
	}
	levResp, err := trader.SetLeverageTestServer(levReq)
	if err != nil {
		t.Fatalf("Failed to set leverage: %v", err)
	}
	t.Logf("Leverage set: %+v", levResp)
}
