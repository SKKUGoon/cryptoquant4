package config_test

import (
	"testing"

	"cryptoquant.com/m/config"
	"github.com/go-playground/assert"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
)

const PGENVLOC = "../.env"

func TestUpbitSpotTradeConfig(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	upbit, err := config.NewUpbitSpotTradeConfig()
	if err != nil {
		t.Fatalf("Failed to create upbit spot trade config: %v", err)
	}

	upbit.UpdateExchangeInfo()
	upbit.UpdateQuotingAsset("KRW")
	upbit.UpdateMinimumTradeAmount()
	upbit.UpdateUpbitPrecision()

	assert.Equal(t, upbit.QuotingAsset, "KRW")
	assert.Equal(t, upbit.MinimumTradeAmount, decimal.NewFromInt(5000))
	assert.Equal(t, upbit.UpbitPrecition.KrwOneSymbols, []string{"KRW-ADA", "KRW-ALGO", "KRW-BLUR", "KRW-CELO", "KRW-ELF", "KRW-EOS", "KRW-GRS", "KRW-GRT", "KRW-ICX", "KRW-MANA", "KRW-MINA", "KRW-POL", "KRW-SAND", "KRW-SEI", "KRW-STG", "KRW-TRX"})
	assert.Equal(t, upbit.UpbitPrecition.KrwPointFiveSymbols, []string{"KRW-USDT", "KRW-USDC"})
}
