package binancetrade_test

import (
	"os"
	"testing"
	"time"

	binancetrade "cryptoquant.com/m/core/trader/binance"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

func TestTraderCycle(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := binancetrade.NewTrader()
	trader.SetTestPubKey(os.Getenv("BINANCE_TESTNET_API_KEY"))
	trader.SetTestPriKey(os.Getenv("BINANCE_TESTNET_SECRET_KEY"))
	trader.UpdateRateLimit(1000)

	// Set leverage
	levReq := &binancerest.LeverageRequest{
		Symbol:    "BTCUSDT",
		Leverage:  1,
		Timestamp: time.Now().UnixMilli(),
	}
	levResp, err := trader.SetLeverageTestServer(levReq)
	if err != nil {
		t.Fatalf("Failed to set leverage: %v", err)
	}
	t.Logf("Leverage set: %+v", levResp)

	// Send order - market short
	orderSheet := binancerest.NewTestOrderSheetMarketShort()
	result, err := trader.SendSingleOrderTestServer(orderSheet)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("Order sent: %+v", result.Success)
	t.Logf("Order sent: %+v", result.Success.OrderID)
	t.Logf("Order sent: %+v", result.Error)

	// Sleep for 5 seconds
	time.Sleep(5 * time.Second)

	// Close order - exit position
	quantity, _ := decimal.NewFromString("10")
	test := &binancerest.OrderSheet{
		Symbol:     "BTCUSDT",
		Side:       "BUY",
		Type:       "MARKET",
		ReduceOnly: "true",
		Quantity:   quantity,
		Timestamp:  time.Now().UnixMilli(),
	}
	result, err = trader.SendSingleOrderTestServer(test)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("Order sent: %+v", result.Success)
	t.Logf("Order sent: %+v", result.Success.OrderID)
}
