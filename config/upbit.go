package config

import (
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

type UpbitSpotTradeConfig struct {
	ExchangeInfo *upbitrest.SpotExchanges

	// List of symbols to exclude from trading
	ExcludeTrades map[string]bool

	// Trading Parameters
	FundSize     float32
	QuotingAsset string
}

func NewUpbitSpotTradeConfig() (*UpbitSpotTradeConfig, error) {
	if exchangeInfo, err := upbitrest.NewSpotExchange(); err == nil {
		return &UpbitSpotTradeConfig{
			ExchangeInfo: exchangeInfo,
		}, nil
	} else {
		return nil, err
	}
}

func (e *UpbitSpotTradeConfig) UpdateExchangeInfo() {
	if exchangeInfo, err := upbitrest.NewSpotExchange(); err != nil {
		panic(err)
	} else {
		e.ExchangeInfo = exchangeInfo
	}
}

func (e *UpbitSpotTradeConfig) UpdateQuotingAsset(quote string) {
	e.QuotingAsset = quote
}

func (e *UpbitSpotTradeConfig) IsAvailableSymbol(symbol string) bool {
	return e.ExchangeInfo.IsSymbolAvailable(symbol)
}
