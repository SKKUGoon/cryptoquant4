package binancerest

import (
	"encoding/json"
	"log"
	"time"

	"cryptoquant.com/m/utils"
	"github.com/shopspring/decimal"
)

// Mandatory fiels
// 1. LIMIT
//   - timeInForce
//   - quantity
//   - price
// 2. MARKET
//   - quantity
// 3. STOP/TAKE_PROFIT
//   - quantity
//   - price
//   - stopPrice
// 4. STOP_MARKET/TAKE_PROFIT_MARKET
//   - stopPrice
// 5. TRAILING_STOP_MARKET
//   - callbackRate

type OrderSheet struct {
	Symbol              string          `json:"symbol"`
	Side                string          `json:"side"`
	PositionSide        string          `json:"positionSide,omitempty"` // "BOTH", "LONG", "SHORT". Default is "BOTH"
	Type                string          `json:"type"`
	TimeInForce         string          `json:"timeInForce,omitempty"`
	Quantity            decimal.Decimal `json:"quantity"`
	ReduceOnly          string          `json:"reduceOnly,omitempty"` // "true" or "false". Default is "false"
	Price               decimal.Decimal `json:"price,omitempty"`
	NewClientOrderID    string          `json:"newClientOrderId,omitempty"`
	StopPrice           decimal.Decimal `json:"stopPrice,omitempty"`     // `STOP/STOP_MARKET` or `TAKE_PROFIT/TAKE_PROFIT_MARKET`
	ClosePosition       string          `json:"closePosition,omitempty"` // "true" or "false". Used with `STOP_MARKET` or `TAKE_PROFIT_MARKET`
	ActivationPrice     decimal.Decimal `json:"activationPrice,omitempty"`
	CallbackRate        decimal.Decimal `json:"callbackRate,omitempty"`
	WorkingType         string          `json:"workingType,omitempty"`         // "MARK_PRICE" or "CONTRACT_PRICE". Default is "CONTRACT_PRICE"
	PriceProtect        bool            `json:"priceProtect,omitempty"`        // "true" or "false". Default is "false"
	NewOrderRespType    string          `json:"newOrderRespType,omitempty"`    // "ACK" or "RESULT". Default is "ACK"
	SelfTradePrevention string          `json:"selfTradePrevention,omitempty"` // "SELL" or "BUY". Default is "SELL"
	Timestamp           int64           `json:"timestamp"`
}

func NewTestOrderSheetLong() *OrderSheet {
	quantity, err := decimal.NewFromString("0.002")
	if err != nil {
		log.Fatalf("Failed to create quantity: %v", err)
	}
	price, err := decimal.NewFromString("80000") // NOTE: Make sure to use the price that's never going to be hit
	if err != nil {
		log.Fatalf("Failed to create price: %v", err)
	}
	timestamp := time.Now().UnixMilli()

	test := &OrderSheet{
		Symbol:      "BTCUSDT",
		Side:        "BUY",
		Type:        "LIMIT",
		TimeInForce: "GTC",
		Quantity:    quantity,
		Price:       price,
		Timestamp:   timestamp,
	}
	return test
}

func NewTestOrderSheetShort() *OrderSheet {
	quantity, err := decimal.NewFromString("0.05")
	if err != nil {
		log.Fatalf("Failed to create quantity: %v", err)
	}
	price, err := decimal.NewFromString("3800.35") // NOTE: Make sure to use the price that's never going to be hit
	if err != nil {
		log.Fatalf("Failed to create price: %v", err)
	}
	timestamp := time.Now().UnixMilli()

	test := &OrderSheet{
		Symbol:      "ETHUSDT",
		Side:        "SELL",
		Type:        "LIMIT",
		TimeInForce: "GTC",
		Quantity:    quantity,
		Price:       price,
		Timestamp:   timestamp,
	}
	return test
}

func (o *OrderSheet) RemoveTimestamp() {
	o.Timestamp = 0
}

func (o *OrderSheet) ToParamsMap() map[string]string {
	return utils.StructToParamsMap(o)
}

type OrderResponse struct {
	ClientOrderID           string `json:"clientOrderId"`
	CumQty                  string `json:"cumQty"`
	CumQuote                string `json:"cumQuote"`
	ExecutedQty             string `json:"executedQty"`
	OrderID                 int64  `json:"orderId"`
	AvgPrice                string `json:"avgPrice"`
	OrigQty                 string `json:"origQty"`
	Price                   string `json:"price"`
	ReduceOnly              bool   `json:"reduceOnly"`
	Side                    string `json:"side"`
	PositionSide            string `json:"positionSide"`
	Status                  string `json:"status"`
	StopPrice               string `json:"stopPrice"`
	ClosePosition           bool   `json:"closePosition"`
	Symbol                  string `json:"symbol"`
	TimeInForce             string `json:"timeInForce"`
	Type                    string `json:"type"`
	OrigType                string `json:"origType"`
	ActivatePrice           string `json:"activatePrice"`
	PriceRate               string `json:"priceRate"`
	UpdateTime              int64  `json:"updateTime"`
	WorkingType             string `json:"workingType"`
	PriceProtect            bool   `json:"priceProtect"`
	PriceMatch              string `json:"priceMatch"`
	SelfTradePreventionMode string `json:"selfTradePreventionMode"`
	GoodTillDate            int64  `json:"goodTillDate"`
}

type OrderError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type OrderResult struct {
	Success *OrderResponse
	Error   *OrderError
}

type BatchOrderResponse []OrderResult

func (r *OrderResult) UnmarshalJSON(data []byte) error {
	var errObj OrderError
	if err := json.Unmarshal(data, &errObj); err == nil && errObj.Code != 0 {
		r.Error = &errObj
		return nil
	}

	var successObj OrderResponse
	if err := json.Unmarshal(data, &successObj); err == nil && successObj.OrderID != 0 {
		r.Success = &successObj
		return nil
	}

	// Unknown format
	return nil
}
