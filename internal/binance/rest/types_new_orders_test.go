package binancerest_test

import (
	"encoding/json"
	"testing"

	binancerest "cryptoquant.com/m/internal/binance/rest"
	"github.com/shopspring/decimal"
)

func TestOrderSheet_ToParam(t *testing.T) {
	quantity, err := decimal.NewFromString("1")
	if err != nil {
		t.Fatalf("Failed to create quantity: %v", err)
	}
	price, err := decimal.NewFromString("9000")
	if err != nil {
		t.Fatalf("Failed to create price: %v", err)
	}

	order := &binancerest.OrderSheet{
		Symbol:      "BTCUSDT",
		Side:        "BUY",
		Type:        "LIMIT",
		TimeInForce: "GTC",
		Quantity:    quantity,
		Price:       price,
		Timestamp:   1591702613943,
	}

	params := order.ToParamsMap()

	expected := map[string]string{
		"symbol":      "BTCUSDT",
		"side":        "BUY",
		"type":        "LIMIT",
		"timeInForce": "GTC",
		"quantity":    "1",
		"price":       "9000",
		"timestamp":   "1591702613943",
	}
	for key, expectedValue := range expected {
		actualValue, exists := params[key]
		if !exists {
			t.Errorf("missing key: %s", key)
		}
		if actualValue != expectedValue {
			t.Errorf("key %s: expected %s, got %s", key, expectedValue, actualValue)
		}
	}

	if len(params) != len(expected) {
		t.Errorf("expected %d params, got %d", len(expected), len(params))
	}
}

func TestSingleOrder_Success(t *testing.T) {
	input := []byte(`{
		"clientOrderId": "abc123",
		"orderId": 12345,
		"symbol": "BTCUSDT",
		"status": "NEW"
	}`)

	var result binancerest.OrderResult
	if err := json.Unmarshal(input, &result); err != nil {
		t.Fatal("unmarshal failed:", err)
	}

	if result.Success == nil {
		t.Error("expected success, got nil")
	}
	if result.Error != nil {
		t.Error("expected no error, but got one")
	}
	if result.Success.OrderID != 12345 {
		t.Errorf("expected orderId 12345, got %d", result.Success.OrderID)
	}
}

func TestSingleOrder_Error(t *testing.T) {
	input := []byte(`{
		"code": -2022,
		"msg": "ReduceOnly Order is rejected."
	}`)

	var result binancerest.OrderResult
	if err := json.Unmarshal(input, &result); err != nil {
		t.Fatal("unmarshal failed:", err)
	}

	if result.Error == nil {
		t.Error("expected error, got nil")
	}
	if result.Success != nil {
		t.Error("expected no success, but got one")
	}
	if result.Error.Code != -2022 {
		t.Errorf("expected code -2022, got %d", result.Error.Code)
	}
}

func TestBatchOrders_AllSuccess(t *testing.T) {
	input := []byte(`[
		{"clientOrderId": "a1", "orderId": 1, "symbol": "BTCUSDT", "status": "NEW"},
		{"clientOrderId": "a2", "orderId": 2, "symbol": "ETHUSDT", "status": "FILLED"}
	]`)

	var results binancerest.BatchOrderResponse
	if err := json.Unmarshal(input, &results); err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Success == nil {
			t.Errorf("expected success at index %d", i)
		}
	}
}

func TestBatchOrders_AllError(t *testing.T) {
	input := []byte(`[
		{"code": -2022, "msg": "Error A"},
		{"code": -1100, "msg": "Error B"}
	]`)

	var results binancerest.BatchOrderResponse
	if err := json.Unmarshal(input, &results); err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for i, r := range results {
		if r.Error == nil {
			t.Errorf("expected error at index %d", i)
		}
	}
}

func TestBatchOrders_Hybrid(t *testing.T) {
	input := []byte(`[
		{"clientOrderId": "hybrid1", "orderId": 999, "symbol": "BTCUSDT", "status": "NEW"},
		{"code": -2011, "msg": "Unknown order"}
	]`)

	var results binancerest.BatchOrderResponse
	if err := json.Unmarshal(input, &results); err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Success == nil || results[0].Error != nil {
		t.Errorf("expected success at index 0")
	}
	if results[1].Error == nil || results[1].Success != nil {
		t.Errorf("expected error at index 1")
	}
}
