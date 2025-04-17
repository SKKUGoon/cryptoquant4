package config

import (
	"log"
	"strings"

	"cryptoquant.com/m/data/database"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/shopspring/decimal"
)

// UpbitSpotTradeConfig is a struct that contains the configuration for the Upbit spot trade
// It contains the exchange info and the trading parameters
// It also contains the trading parameters
// 1. UpbitPrecision
//   - KrwOneSymbols: List of symbols with precision 1
//   - KrwPointFiveSymbols: List of symbols with precision 0.5
type UpbitSpotTradeConfig struct {
	ExchangeInfo *upbitrest.SpotExchanges

	// List of symbols to exclude from trading
	ExcludeTrades map[string]bool

	// Trading Parameters
	// Upbit needs UpbitPrecision to be set - Binance does not need it
	UpbitPrecition struct {
		KrwOneSymbols       map[string]bool
		KrwPointFiveSymbols map[string]bool
	}
	MinimumTradeAmount decimal.Decimal
	FundSize           float32
	QuotingAsset       string
}

func NewUpbitSpotTradeConfig() (*UpbitSpotTradeConfig, error) {
	exchangeInfo, err := upbitrest.NewSpotExchange()
	if err != nil {
		return nil, err
	}

	return &UpbitSpotTradeConfig{
		ExchangeInfo: exchangeInfo,
		UpbitPrecition: struct {
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

func (e *UpbitSpotTradeConfig) UpdateQuotingAsset(quote string) {
	e.QuotingAsset = quote
}

func (e *UpbitSpotTradeConfig) UpdateMinimumTradeAmount() {
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	amount, err := db.GetTradeMetadata("upbit_minimum_trade_amount", decimal.NewFromInt(1000))
	if err != nil {
		panic(err)
	}

	amountInt, ok := amount.(int)
	if !ok {
		panic("upbit_minimum_trade_amount is not an int")
	}

	amountInt64 := int64(amountInt)

	e.MinimumTradeAmount = decimal.NewFromInt(amountInt64)
}

func (e *UpbitSpotTradeConfig) UpdateUpbitPrecision() {
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
	e.UpbitPrecition.KrwOneSymbols = symbols
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
	e.UpbitPrecition.KrwPointFiveSymbols = symbols
}

func (e *UpbitSpotTradeConfig) IsAvailableSymbol(symbol string) bool {
	return e.ExchangeInfo.IsSymbolAvailable(symbol)
}

func (e *UpbitSpotTradeConfig) AuditOrderSheetPrecision(orderSheet *upbitrest.OrderSheet) error {
	// NOTE: Upbit does not have precision for each symbol
	// NOTE: Upbit ordersheet only requires total amount of money you want to spend

	// symbolAffiliation := func(symbol string) string {
	// 	if e.UpbitPrecition.KrwOneSymbols[symbol] {
	// 		return "UPBIT_KRW_ONE" // Special case for Upbit. Its precision is set
	// 	}

	// 	if e.UpbitPrecition.KrwPointFiveSymbols[symbol] {
	// 		return "UPBIT_KRW_POINT_FIVE"
	// 	}

	// 	return "NORMAL"
	// }

	// // Audit symbol precision
	// switch symbolAffiliation(orderSheet.Symbol) {
	// case "UPBIT_KRW_ONE":
	// 	pricePrecision := 1
	// 	if orderSheet.Price != "" {
	// 		price, err := decimal.NewFromString(orderSheet.Price)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		roundedPrice := price.Round(int32(pricePrecision))
	// 		orderSheet.Price = roundedPrice.String()
	// 	}
	// case "UPBIT_KRW_POINT_FIVE":
	// 	if orderSheet.Price != "" {
	// 		// Multiply by 2 to work with whole numbers
	// 		price, err := decimal.NewFromString(orderSheet.Price)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		scaled := price.Mul(decimal.NewFromFloat(2))
	// 		// Round to nearest integer
	// 		rounded := scaled.Round(0)
	// 		// Divide by 2 to get back to 0.5 precision
	// 		orderSheet.Price = rounded.Div(decimal.NewFromFloat(2)).String()
	// 	}
	// default:
	// 	return nil
	// }

	return nil
}
