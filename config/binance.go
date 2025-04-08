package config

import (
	binancerest "cryptoquant.com/m/internal/binance/rest"
)

// BinanceFutureTradeConfig implements the Exchange interface
type BinanceFutureTradeConfig struct {
	ExchangeInfo *binancerest.FutureExchange

	// List of symbols to exclude from trading
	ExcludeTrades map[string]bool

	// Trading Parameters
	FundSize        float32
	MaximumLeverage int
	QuotingAsset    string
}

func NewBinanceFutureTradeConfig() (*BinanceFutureTradeConfig, error) {
	if exchangeInfo, err := binancerest.NewFutureExchange(); err == nil {
		return &BinanceFutureTradeConfig{
			ExchangeInfo:    exchangeInfo,
			MaximumLeverage: 1,
		}, nil
	} else {
		return nil, err
	}
}

func (e *BinanceFutureTradeConfig) UpdateExchangeInfo() {
	if exchangeInfo, err := binancerest.NewFutureExchange(); err != nil {
		panic(err)
	} else {
		e.ExchangeInfo = exchangeInfo
	}
}

func (e *BinanceFutureTradeConfig) UpdateMaximumLeverage(lev int) {
	e.MaximumLeverage = lev
}

func (e *BinanceFutureTradeConfig) UpdateQuotingAsset(quote string) {
	e.QuotingAsset = quote
}

func (e *BinanceFutureTradeConfig) IsAvailableSymbol(symbol string) bool {
	return e.ExchangeInfo.IsSymbolAvailable(symbol)
}

func (e *BinanceFutureTradeConfig) GetSymbolPricePrecision(symbol string) int {
	return e.ExchangeInfo.GetSymbolInfo(symbol).GetSymbolPricePrecision()
}

func (e *BinanceFutureTradeConfig) GetSymbolQuantityPrecision(symbol string) int {
	return e.ExchangeInfo.GetSymbolInfo(symbol).GetSymbolQuantityPrecision()
}
