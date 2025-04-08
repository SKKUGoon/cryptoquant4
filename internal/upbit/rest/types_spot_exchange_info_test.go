package upbitrest_test

import (
	"testing"

	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

func TestNewSpotExchange(t *testing.T) {
	spotExchanges, err := upbitrest.NewSpotExchange()
	if err != nil {
		t.Fatalf("Failed to create spot exchanges: %v", err)
	}

	t.Log(spotExchanges)
}

func TestIsSymbolAvailable(t *testing.T) {
	spotExchanges, err := upbitrest.NewSpotExchange()
	if err != nil {
		t.Fatalf("Failed to create spot exchanges: %v", err)
	}

	ok := spotExchanges.IsSymbolAvailable("KRW-USDT")
	if !ok {
		t.Fatalf("Symbol KRW-USDT is not available")
	}

	ok = spotExchanges.IsSymbolAvailable("KRW-BTC")
	if !ok {
		t.Fatalf("Symbol KRW-BTC is not available")
	}

	ok = spotExchanges.IsSymbolAvailable("KRW-ETH")
	if ok {
		t.Log("Symbol KRW-ETH is available")
	}
}
