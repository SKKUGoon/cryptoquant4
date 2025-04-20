package config_test

import (
	"testing"

	"cryptoquant.com/m/config"
	"github.com/go-playground/assert"
	"github.com/joho/godotenv"
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
	upbit.SetPrincipalCurrency()
	upbit.SetMinimumTradeAmount()
	upbit.SetUpbitPrecisionSpecial()

	assert.Equal(t, upbit.PrincipalCurrency, "KRW")
	assert.Equal(t, upbit.MinimumTradeAmount, 5000)
	assert.Equal(t, upbit.UpbitPrecision.KrwOneSymbols, []string{"KRW-ADA", "KRW-ALGO", "KRW-BLUR", "KRW-CELO", "KRW-ELF", "KRW-EOS", "KRW-GRS", "KRW-GRT", "KRW-ICX", "KRW-MANA", "KRW-MINA", "KRW-POL", "KRW-SAND", "KRW-SEI", "KRW-STG", "KRW-TRX"})
	assert.Equal(t, upbit.UpbitPrecision.KrwPointFiveSymbols, []string{"KRW-USDT", "KRW-USDC"})
}
