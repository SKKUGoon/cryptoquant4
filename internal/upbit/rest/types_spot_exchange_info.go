package upbitrest

import (
	"encoding/json"
	"io"
	"net/http"
)

// Spot represents the market information for a trading pair
type SpotExchanges []Spot

type Spot struct {
	Market      string      `json:"market"`
	KoreanName  string      `json:"korean_name"`
	EnglishName string      `json:"english_name"`
	MarketEvent MarketEvent `json:"market_event"`
}

// MarketEvent represents the market event information including warnings and cautions
type MarketEvent struct {
	Warning bool    `json:"warning"`
	Caution Caution `json:"caution"`
}

// Caution represents various caution flags for a market
type Caution struct {
	PriceFluctuations            bool `json:"PRICE_FLUCTUATIONS"`
	TradingVolumeSoaring         bool `json:"TRADING_VOLUME_SOARING"`
	DepositAmountSoaring         bool `json:"DEPOSIT_AMOUNT_SOARING"`
	GlobalPriceDifferences       bool `json:"GLOBAL_PRICE_DIFFERENCES"`
	ConcentrationOfSmallAccounts bool `json:"CONCENTRATION_OF_SMALL_ACCOUNTS"`
}

type SpotSymbolInfo struct {
	Market string
	Symbol string
}

func NewSpotExchange() (*SpotExchanges, error) {
	var spotExchanges SpotExchanges

	// Upbit API url + endpoint
	response, err := http.Get("https://api.upbit.com/v1/market/all")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &spotExchanges)
	if err != nil {
		return nil, err
	}

	return &spotExchanges, nil
}

func (e *SpotExchanges) IsSymbolAvailable(symbol string) bool {
	var ok bool = false
	for _, market := range *e {
		if market.Market == symbol {
			ok = true
			break
		}
	}
	return ok
}
