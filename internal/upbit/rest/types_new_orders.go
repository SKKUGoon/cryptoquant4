package upbitrest

import (
	"encoding/json"
	"log"
	"time"

	"cryptoquant.com/m/utils"
	"github.com/shopspring/decimal"
)

type OrderSheet struct {
	Symbol      string `json:"market"`           // e.g. KRW-BTC
	Side        string `json:"side"`             // "bid" or "ask"
	Volume      string `json:"volume,omitempty"` // e.g. "0.001", Essential for limit order + market sell order
	Price       string `json:"price,omitempty"`
	OrdType     string `json:"ord_type,omitempty"`      // 'limit', 'price', 'market', 'best'
	Identifier  string `json:"identifier,omitempty"`    // Unique value
	TimeInForce string `json:"time_in_force,omitempty"` // 'ioc', 'fok'
}

func (o *OrderSheet) ToParamsMap() map[string]string {
	return utils.StructToParamsMap(o)
}

type OrderResponse struct {
	UUID            string    `json:"uuid"`
	Side            string    `json:"side"`
	OrdType         string    `json:"ord_type"`
	Price           string    `json:"price"`
	State           string    `json:"state"`
	Market          string    `json:"market"`
	CreatedAt       time.Time `json:"created_at"`
	Volume          string    `json:"volume"`
	RemainingVolume string    `json:"remaining_volume"`
	ReservedFee     string    `json:"reserved_fee"`
	RemainingFee    string    `json:"remaining_fee"`
	PaidFee         string    `json:"paid_fee"`
	Locked          string    `json:"locked"`
	ExecutedVolume  string    `json:"executed_volume"`
	TradesCount     int       `json:"trades_count"`
}

func NewTestOrderSheetLong() *OrderSheet {
	quantity, err := decimal.NewFromString("0.001")
	if err != nil {
		log.Fatalf("Failed to create quantity: %v", err)
	}
	quantityStr := quantity.String()

	price, err := decimal.NewFromString("100000000.0") // NOTE: Make sure to use the price that's never going to be hit
	if err != nil {
		log.Fatalf("Failed to create price: %v", err)
	}
	priceStr := price.String()

	return &OrderSheet{
		Symbol:  "KRW-BTC",
		Side:    "bid",
		Volume:  quantityStr,
		Price:   priceStr,
		OrdType: "limit",
	}
}

type OrderError struct {
	Error struct {
		Name    string `json:"name,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"error,omitempty"`
}

type OrderResult struct {
	Success *OrderResponse
	Error   *OrderError
}

func (o *OrderResult) UnmarshalJSON(data []byte) error {
	var errObj OrderError
	if err := json.Unmarshal(data, &errObj); err == nil && errObj.Error.Name != "" {
		o.Error = &errObj
		return nil
	}

	var successObj OrderResponse
	if err := json.Unmarshal(data, &successObj); err == nil && successObj.UUID != "" {
		o.Success = &successObj
		return nil
	}

	// Unknown format
	return nil
}
