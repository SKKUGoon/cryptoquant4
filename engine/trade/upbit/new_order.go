package upbittrade

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

func (t *Trader) SendOrder(orderSheet upbitrest.OrderSheet) (upbitrest.OrderResult, error) {
	const weight = 5
	const urlBase = "https://api.upbit.com/v1/orders"
	t.checkRateLimit(weight)

	// Make byteQuery + signature
	byteQuery, authToken, err := t.GetSignature(orderSheet)
	if err != nil {
		return upbitrest.OrderResult{}, err
	}

	// Convert orderSheet to JSON and send as request body
	req, err := http.NewRequest("POST", urlBase, bytes.NewReader(byteQuery))
	if err != nil {
		return upbitrest.OrderResult{}, err
	}

	// Add Authorization header
	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Send request
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return upbitrest.OrderResult{}, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return upbitrest.OrderResult{}, err
	}

	// Parse response body
	var result upbitrest.OrderResult
	if err := json.Unmarshal(body, &result); err != nil {
		return upbitrest.OrderResult{}, err
	}

	return result, nil
}
