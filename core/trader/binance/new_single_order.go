package binancetrade

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func (t *Trader) SendSingleOrder(orderSheet binancerest.OrderSheet) (binancerest.OrderResult, error) {
	const weight = 5
	const urlBase = "https://fapi.binance.com/fapi/v1/order"
	t.checkRateLimit(weight)

	// Make query + signature
	queryString, signature, err := t.GenerateSignature(orderSheet)
	if err != nil {
		return binancerest.OrderResult{}, err
	}
	fullQuery := queryString + "&signature=" + signature

	// Send request
	req, err := http.NewRequest("POST", urlBase, strings.NewReader(fullQuery))
	if err != nil {
		return binancerest.OrderResult{}, err
	}

	req.Header.Set("X-MBX-APIKEY", t.pubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return binancerest.OrderResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return binancerest.OrderResult{}, err
	}

	var result binancerest.OrderResult
	if err := json.Unmarshal(body, &result); err != nil {
		return binancerest.OrderResult{}, err
	}

	return result, nil
}

func (t *Trader) SendSingleOrderTestServer(orderSheet *binancerest.OrderSheet) (binancerest.OrderResult, error) {
	const weight = 5
	const urlBase = "https://testnet.binancefuture.com/fapi/v1/order"
	t.checkRateLimit(weight)

	// Make query + signature
	queryString, signature, err := t.GenerateSignature(orderSheet)
	if err != nil {
		return binancerest.OrderResult{}, err
	}
	fullQuery := queryString + "&signature=" + signature

	// Send request
	req, err := http.NewRequest("POST", urlBase, strings.NewReader(fullQuery))
	if err != nil {
		return binancerest.OrderResult{}, err
	}

	req.Header.Set("X-MBX-APIKEY", t.pubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return binancerest.OrderResult{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return binancerest.OrderResult{}, err
	}

	var result binancerest.OrderResult
	if err := json.Unmarshal(body, &result); err != nil {
		return binancerest.OrderResult{}, err
	}

	return result, nil
}
