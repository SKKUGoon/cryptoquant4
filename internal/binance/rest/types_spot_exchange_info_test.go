package binancerest_test

import (
	"fmt"
	"testing"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func TestGetSpotSymbolInfo(t *testing.T) {
	_, err := binancerest.NewSpotExchange()
	if err != nil {
		t.Fatalf("Failed to create exchange info: %v", err)
	}
}

func TestGetSymbolFilter(t *testing.T) {
	exchangeInfo, err := binancerest.NewSpotExchange()
	if err != nil {
		t.Fatalf("Failed to create exchange info: %v", err)
	}

	symbolInfo := exchangeInfo.GetSymbolInfo("BTCUSDT")
	if symbolInfo == nil {
		t.Fatalf("Failed to get symbol info: %v", err)
	}

	symbolFilter := symbolInfo.GetSymbolFilter("LOT_SIZE")
	fmt.Println(symbolFilter)
}

func TestGetSymbolQuotePrecision(t *testing.T) {
	exchangeInfo, err := binancerest.NewSpotExchange()
	if err != nil {
		t.Fatalf("Failed to create exchange info: %v", err)
	}

	symbolInfo := exchangeInfo.GetSymbolInfo("XRPUSDT")
	if symbolInfo == nil {
		t.Fatalf("Failed to get symbol info: %v", err)
	}

	quotePrecision := symbolInfo.GetSymbolQuotePrecision()
	fmt.Println(quotePrecision)
}
