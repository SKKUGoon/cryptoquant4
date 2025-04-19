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

func TestSendOrder(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.UpdateRateLimit(1000)
	orderSheet := binancerest.NewTestOrderSheetLong()
	result, err := trader.SendSingleOrder(*orderSheet)
	t.Log("Order sent")
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result)
}

func TestSendOrderTest(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.SetTestPubKey(os.Getenv("BINANCE_TESTNET_API_KEY"))
	trader.SetTestPriKey(os.Getenv("BINANCE_TESTNET_SECRET_KEY"))
	trader.UpdateRateLimit(1000)

	start := time.Now()
	orderSheet := binancerest.NewTestOrderSheetMarketShort()
	result, err := trader.SendSingleOrderTestServer(orderSheet)
	elapsed := time.Since(start)
	t.Logf("Order execution time: %v", elapsed)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result.Success)
}
