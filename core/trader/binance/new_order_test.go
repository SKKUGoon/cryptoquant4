package binancetrade_test

import (
	"os"
	"testing"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	"github.com/joho/godotenv"
)

func TestSendOrder(t *testing.T) {
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.UpdateRateLimit(1000)
	orderSheet := binancerest.NewTestOrderSheetLong()
	result, err := trader.SendOrder(*orderSheet)
	t.Log("Order sent")
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result)
}

func TestSendOrders(t *testing.T) {
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.UpdateRateLimit(1000)

	orderSheet1 := binancerest.NewTestOrderSheetLong()
	orderSheet2 := binancerest.NewTestOrderSheetShort()

	orderSheets := []binancerest.OrderSheet{*orderSheet1, *orderSheet2}
	result, err := trader.SendOrders(orderSheets)
	if err != nil {
		t.Fatalf("Failed to send orders: %v", err)
	}
	t.Logf("result: %+v", result)
}

func TestSendOrderTest(t *testing.T) {
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.SetTestPubKey(os.Getenv("BINANCE_TESTNET_API_KEY"))
	trader.SetTestPriKey(os.Getenv("BINANCE_TESTNET_SECRET_KEY"))
	trader.UpdateRateLimit(1000)

	orderSheet := binancerest.NewTestOrderSheetMarketShort()
	result, err := trader.SendOrderTest(*orderSheet)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result.Success)
}

func TestSendMarketOrder(t *testing.T) {
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.SetTestPubKey(os.Getenv("BINANCE_TESTNET_API_KEY"))
	trader.SetTestPriKey(os.Getenv("BINANCE_TESTNET_SECRET_KEY"))
	trader.UpdateRateLimit(1000)

	orderSheet := binancerest.NewTestOrderSheetMarketShort()

	ordersheets := []binancerest.OrderSheet{*orderSheet}
	result, err := trader.SendOrdersTest(ordersheets)
	if err != nil {
		t.Fatalf("Failed to send orders: %v", err)
	}
	t.Logf("result: %+v", result[0].Success)
}
