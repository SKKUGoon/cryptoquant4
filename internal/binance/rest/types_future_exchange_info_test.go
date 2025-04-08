package binancerest_test

import (
	"fmt"
	"testing"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func TestNewFutureExchange(t *testing.T) {
	exchangeInfo, err := binancerest.NewFutureExchange()
	if err != nil {
		t.Fatalf("Failed to create exchange info: %v", err)
	}

	symbolInfo := exchangeInfo.GetSymbolInfo("BTCUSDT")
	if symbolInfo == nil {
		t.Fatalf("Failed to get symbol info: %v", err)
	}
}

func TestGetFutureSymbolFilter(t *testing.T) {
	exchangeInfo, err := binancerest.NewFutureExchange()
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

func TestGetFutureSymbolPricePrecision(t *testing.T) {
	exchangeInfo, err := binancerest.NewFutureExchange()
	if err != nil {
		t.Fatalf("Failed to create exchange info: %v", err)
	}

	symbolInfo := exchangeInfo.GetSymbolInfo("BTCUSDT")
	if symbolInfo == nil {
		t.Fatalf("Failed to get symbol info: %v", err)
	}

	fmt.Println(symbolInfo.GetSymbolPricePrecision())
}
