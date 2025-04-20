package upbittrade_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"cryptoquant.com/m/core/account"
	upbittrade "cryptoquant.com/m/core/trader/upbit"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/joho/godotenv"
)

const ENV_PATH = "../../../.env.local"

// WARNING: This test goes straight to the Upbit non-test API !!

func TestSendOrder(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
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

func TestSendOrderMarketLong(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := upbittrade.NewTrader()
	trader.UpdateRateLimit(2000)
	orderSheet := upbitrest.OrderSheet{
		Symbol:  "KRW-XRP",
		Side:    "bid",
		Price:   "5000.99552",
		OrdType: "price",
	}

	result, err := trader.SendOrder(orderSheet)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result.Success)
	t.Logf("result: %+v", result.Error)

	time.Sleep(10 * time.Second)
}

func TestAccountSource_Sync(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	trader := upbittrade.NewTrader()
	trader.UpdateRateLimit(2000)

	// Create account source
	as := account.NewAccountSource(context.Background())
	as.SetPrincipalCurrency("upbit", "KRW")
	as.SetPrincipalCurrency("binance", "USDT")
	_ = as.UpdateRedis()
	_ = as.Sync()

	t.Logf("binance reserved fund: %v", as.GetBinanceWalletSnapshot())
	t.Logf("upbit reserved fund: %v", as.GetUpbitWalletSnapshot())

	symbol := "USDT"
	upbitAmount, ok := as.GetUpbitWalletSnapshot()[symbol]
	if !ok {
		t.Fatalf("upbit amount not found")
	}
	t.Logf("upbit amount: %v", upbitAmount)

	orderSheet := upbitrest.OrderSheet{
		Symbol:  "KRW-USDT",
		Side:    "ask",
		Volume:  strconv.FormatFloat(upbitAmount, 'f', -1, 64),
		OrdType: "market",
	}

	result, err := trader.SendOrder(orderSheet)
	if err != nil {
		t.Fatalf("Failed to send order: %v", err)
	}
	t.Logf("result: %+v", result.Success)
	t.Logf("result: %+v", result.Error)
}
