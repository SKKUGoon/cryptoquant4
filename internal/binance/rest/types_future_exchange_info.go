package binancerest

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
)

type FutureExchange struct {
	Timezone    string `json:"timezone"`
	ServerTime  int64  `json:"serverTime"`
	FuturesType string `json:"futuresType"`
	RateLimits  []struct {
		RateLimitType string `json:"rateLimitType"`
		Interval      string `json:"interval"`
		IntervalNum   int    `json:"intervalNum"`
		Limit         int    `json:"limit"`
	} `json:"rateLimits"`
	ExchangeFilters []any `json:"exchangeFilters"`
	Assets          []struct {
		Asset             string `json:"asset"`
		MarginAvailable   bool   `json:"marginAvailable"`
		AutoAssetExchange string `json:"autoAssetExchange"`
	} `json:"assets"`
	Symbols []FutureSymbolInfo `json:"symbols"`

	// Custom field
	symbolsMap map[string]*FutureSymbolInfo `json:"-"` // -: ignore this field during JSON marshaling
}

type FutureSymbolInfo struct {
	Symbol                string               `json:"symbol"`
	Pair                  string               `json:"pair"`
	ContractType          string               `json:"contractType"`
	DeliveryDate          int64                `json:"deliveryDate"`
	OnboardDate           int64                `json:"onboardDate"`
	Status                string               `json:"status"`
	MaintMarginPercent    string               `json:"maintMarginPercent"`
	RequiredMarginPercent string               `json:"requiredMarginPercent"`
	BaseAsset             string               `json:"baseAsset"`
	QuoteAsset            string               `json:"quoteAsset"`
	MarginAsset           string               `json:"marginAsset"`
	PricePrecision        int                  `json:"pricePrecision"`
	QuantityPrecision     int                  `json:"quantityPrecision"`
	BaseAssetPrecision    int                  `json:"baseAssetPrecision"`
	QuotePrecision        int                  `json:"quotePrecision"`
	UnderlyingType        string               `json:"underlyingType"`
	UnderlyingSubType     []string             `json:"underlyingSubType"`
	TriggerProtect        string               `json:"triggerProtect"`
	LiquidationFee        string               `json:"liquidationFee"`
	MarketTakeBound       string               `json:"marketTakeBound"`
	MaxMoveOrderLimit     int                  `json:"maxMoveOrderLimit"`
	Filters               []FutureSymbolFilter `json:"filters"`
	OrderTypes            []string             `json:"orderTypes"`
	TimeInForce           []string             `json:"timeInForce"`
}

type FutureSymbolFilter struct {
	MaxPrice          string `json:"maxPrice,omitempty"`
	MinPrice          string `json:"minPrice,omitempty"`
	TickSize          string `json:"tickSize,omitempty"`
	FilterType        string `json:"filterType"`
	MaxQty            string `json:"maxQty,omitempty"`
	MinQty            string `json:"minQty,omitempty"`
	StepSize          string `json:"stepSize,omitempty"`
	Limit             int    `json:"limit,omitempty"`
	Notional          string `json:"notional,omitempty"`
	MultiplierDecimal string `json:"multiplierDecimal,omitempty"`
	MultiplierUp      string `json:"multiplierUp,omitempty"`
	MultiplierDown    string `json:"multiplierDown,omitempty"`
}

func NewFutureExchange() (*FutureExchange, error) {
	var exchangeInfo FutureExchange

	// Binance API url + endpoint
	url := "https://fapi.binance.com/fapi/v1/exchangeInfo"

	// Make the API request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &exchangeInfo)
	if err != nil {
		return nil, err
	}

	// Initialize the map and populate it
	exchangeInfo.symbolsMap = make(map[string]*FutureSymbolInfo)
	for i := range exchangeInfo.Symbols {
		exchangeInfo.symbolsMap[exchangeInfo.Symbols[i].Symbol] = &exchangeInfo.Symbols[i]
	}

	return &exchangeInfo, nil
}

func (e *FutureExchange) GetRequestRateLimit() int {
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "REQUEST_WEIGHT" {
			return rateLimit.Limit
		}
	}
	return 0
}

func (e *FutureExchange) GetMinuteOrderRateLimit() int {
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "ORDERS" && rateLimit.Interval == "MINUTE" {
			return rateLimit.Limit
		}
	}
	return 0
}

func (e *FutureExchange) GetSecondOrderRateLimit() int {
	for _, rateLimit := range e.RateLimits {
		if rateLimit.RateLimitType == "ORDERS" && rateLimit.Interval == "SECOND" {
			return rateLimit.Limit
		}
	}
	return 0
}

func (e *FutureExchange) GetAvailableSymbols() []string {
	symbols := make([]string, 0, len(e.symbolsMap))
	for symbol := range e.symbolsMap {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols) // ensures deterministic order
	return symbols
}

func (e *FutureExchange) IsSymbolAvailable(symbol string) bool {
	_, ok := e.symbolsMap[symbol]
	return ok
}

// GetSymbolInfo returns the symbol information for the given symbol
func (e *FutureExchange) GetSymbolInfo(symbol string) *FutureSymbolInfo {
	return e.symbolsMap[symbol]
}

// GetSymbolFilter returns the filter of specified type for the given symbol
func (s *FutureSymbolInfo) GetSymbolFilter(filterType string) *FutureSymbolFilter {
	for _, f := range s.Filters {
		if f.FilterType == filterType {
			return &f
		}
	}
	return nil
}

// GetSymbolPricePrecision returns the price precision for the given symbol
func (s *FutureSymbolInfo) GetSymbolPricePrecision() int {
	return s.PricePrecision
}

func (s *FutureSymbolInfo) GetSymbolQuantityPrecision() int {
	return s.QuantityPrecision
}
