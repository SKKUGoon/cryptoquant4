package config

import (
	"log"
	"strings"

	database "cryptoquant.com/m/data/database"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

// UpbitSpotTradeConfig is a struct that contains the configuration for the Upbit spot trade
// It contains the exchange info and the trading parameters
// It also contains the trading parameters
// 1. UpbitPrecision
//   - KrwOneSymbols: List of symbols with precision 1
//   - KrwPointFiveSymbols: List of symbols with precision 0.5
type UpbitSpotTradeConfig struct {
	ExchangeInfo *upbitrest.SpotExchanges
	db           *database.Database

	// List of symbols to exclude from trading
	ExcludeTrades map[string]bool

	// Trading Parameters
	// Upbit needs UpbitPrecision to be set - Binance does not need it
	UpbitPrecision struct {
		KrwOneSymbols       map[string]bool
		KrwPointFiveSymbols map[string]bool
	}
	MinimumTradeAmount int // In KRW
	PrincipalCurrency  string
}

func NewUpbitSpotTradeConfig() (*UpbitSpotTradeConfig, error) {
	exchangeInfo, err := upbitrest.NewSpotExchange()
	if err != nil {
		return nil, err
	}

	db, err := database.ConnectDB()
	if err != nil {
		return nil, err
	}

	return &UpbitSpotTradeConfig{
		ExchangeInfo: exchangeInfo,
		db:           db,
		UpbitPrecision: struct {
			KrwOneSymbols       map[string]bool
			KrwPointFiveSymbols map[string]bool
		}{
			KrwOneSymbols:       make(map[string]bool),
			KrwPointFiveSymbols: make(map[string]bool),
		},
	}, nil
}

func (e *UpbitSpotTradeConfig) UpdateExchangeInfo() {
	if exchangeInfo, err := upbitrest.NewSpotExchange(); err != nil {
		log.Println("Failed to update exchange info", err)
	} else {
		e.ExchangeInfo = exchangeInfo
	}
}

func (e *UpbitSpotTradeConfig) SetPrincipalCurrency() {
	principalCurrency, err := e.db.GetTradeMetadata("upbit_principal_currency", "KRW")
	if err != nil {
		panic(err)
	}

	e.PrincipalCurrency = principalCurrency.(string)
}

func (e *UpbitSpotTradeConfig) SetMinimumTradeAmount() {
	amount, err := e.db.GetTradeMetadata("upbit_minimum_trade_amount", 1000)
	if err != nil {
		panic(err)
	}

	amountInt, ok := amount.(int)
	if !ok {
		panic("upbit_minimum_trade_amount is not an int")
	}

	e.MinimumTradeAmount = amountInt
}

func (e *UpbitSpotTradeConfig) SetUpbitPrecisionSpecial() {
	e.upbitPricePrecisionKrwOne()
	e.upbitPricePrecisionKrwPointFive()
}

func (e *UpbitSpotTradeConfig) upbitPricePrecisionKrwOne() {
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Precision: 1
	symbolsStr, err := db.GetTradeMetadata("upbit_price_precision_krw_one", "")
	if err != nil {
		panic(err)
	}

	if symbolsStr == "" {
		return
	}

	symbols := make(map[string]bool)
	for _, s := range strings.Split(symbolsStr.(string), ",") {
		symbols[strings.TrimSpace(s)] = true
	}
	e.UpbitPrecision.KrwOneSymbols = symbols
}

func (e *UpbitSpotTradeConfig) upbitPricePrecisionKrwPointFive() {
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Precision: 0.5
	symbolsStr, err := db.GetTradeMetadata("upbit_price_precision_krw_point_five", "")
	if err != nil {
		panic(err)
	}

	if symbolsStr == "" {
		return
	}

	symbols := make(map[string]bool)
	for _, s := range strings.Split(symbolsStr.(string), ",") {
		symbols[strings.TrimSpace(s)] = true
	}
	e.UpbitPrecision.KrwPointFiveSymbols = symbols
}

func (e *UpbitSpotTradeConfig) IsAvailableSymbol(symbol string) bool {
	return e.ExchangeInfo.IsSymbolAvailable(symbol)
}

func (e *UpbitSpotTradeConfig) GetSymbolPricePrecision(symbol string, price float64) float32 {
	switch e.PrincipalCurrency {
	case "KRW":
		if _, ok := e.UpbitPrecision.KrwOneSymbols[symbol]; ok {
			return 1
		}

		if _, ok := e.UpbitPrecision.KrwPointFiveSymbols[symbol]; ok {
			return 0.5
		}

		// Else
		switch true {
		case price >= 2_000_000:
			return 1000
		case price < 2_000_000 && price >= 1_000_000:
			return 500
		case price < 1_000_000 && price >= 500_000:
			return 100
		case price < 500_000 && price >= 100_000:
			return 50
		case price < 100_000 && price >= 10_000:
			return 10
		case price < 10_000 && price >= 1_000:
			return 1
		case price < 1_000 && price >= 100:
			return 0.1
		case price < 100 && price >= 10:
			return 0.01
		case price < 10 && price >= 1:
			return 0.001
		case price < 1 && price > 0.1:
			return 0.0001
		case price < 0.1 && price >= 0.01:
			return 0.00001
		case price < 0.01 && price >= 0.001:
			return 0.000001
		case price < 0.001 && price >= 0.0001:
			return 0.0000001
		default:
			return 0.00000001
		}

	default:
		log.Fatalln("Quoting asset is not supported", e.PrincipalCurrency)
		return 0
	}
}
