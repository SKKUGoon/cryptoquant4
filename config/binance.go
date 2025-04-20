package config

import (
	database "cryptoquant.com/m/data/database"
	binancerest "cryptoquant.com/m/internal/binance/rest"
)

// BinanceFutureTradeConfig implements the Exchange interface
type BinanceFutureTradeConfig struct {
	ExchangeInfo *binancerest.FutureExchange
	db           *database.Database

	// List of symbols to exclude from trading
	ExcludeTrades map[string]bool

	// Trading Parameters
	PrincipalCurrency  string // In USD(T)
	MinimumTradeAmount int
}

func NewBinanceFutureTradeConfig() (*BinanceFutureTradeConfig, error) {
	if exchangeInfo, err := binancerest.NewFutureExchange(); err == nil {
		return &BinanceFutureTradeConfig{
			ExchangeInfo: exchangeInfo,
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

func (e *BinanceFutureTradeConfig) SetPrincipalCurrency() {
	principalCurrency, err := e.db.GetTradeMetadata("binance_principal_currency", "USDT")
	if err != nil {
		panic(err)
	}
	e.PrincipalCurrency = principalCurrency.(string)
}

func (e *BinanceFutureTradeConfig) SetMinimumTradeAmount() {
	amount, err := e.db.GetTradeMetadata("binance_minimum_trade_amount", 10)
	if err != nil {
		panic(err)
	}

	amountInt, ok := amount.(int)
	if !ok {
		panic("binance_minimum_trade_amount is not an int")
	}
	e.MinimumTradeAmount = amountInt
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
