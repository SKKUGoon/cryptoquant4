package binancetrade

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func (t *Trader) SetLeverage(levReq *binancerest.LeverageRequest) (binancerest.LeverageResponse, error) {
	const weight = 5
	const urlBase = "https://fapi.binance.com/fapi/v1/leverage"
	t.checkRateLimit(weight)

	// Make query + signature
	queryString, signature, err := t.GenerateSignature(levReq)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}
	fullQuery := queryString + "&signature=" + signature

	// Send request
	req, err := http.NewRequest("POST", urlBase, strings.NewReader(fullQuery))
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}

	req.Header.Set("X-MBX-APIKEY", t.pubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}

	var levResp binancerest.LeverageResponse
	if err := json.Unmarshal(body, &levResp); err != nil {
		return binancerest.LeverageResponse{}, err
	}

	return levResp, nil
}

func (t *Trader) SetLeverageTestServer(levReq *binancerest.LeverageRequest) (binancerest.LeverageResponse, error) {
	const weight = 5
	const urlBase = "https://testnet.binancefuture.com/fapi/v1/leverage"
	t.checkRateLimit(weight)

	// Make query + signature
	queryString, signature, err := t.GenerateSignature(levReq)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}
	fullQuery := queryString + "&signature=" + signature

	// Send request
	req, err := http.NewRequest("POST", urlBase, strings.NewReader(fullQuery))
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}

	req.Header.Set("X-MBX-APIKEY", t.pubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return binancerest.LeverageResponse{}, err
	}

	var levResp binancerest.LeverageResponse
	if err := json.Unmarshal(body, &levResp); err != nil {
		return binancerest.LeverageResponse{}, err
	}

	return levResp, nil
}
