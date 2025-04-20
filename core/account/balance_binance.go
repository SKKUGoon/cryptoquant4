package account

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func (a *AccountSource) syncRedisFromBinance() error {
	var total float64
	var available float64
	var account binancerest.AccountInfo
	var snapshot = make(map[string]float64)
	const urlBase = "https://fapi.binance.com/fapi/v3/account"

	// Build query string
	params := map[string]int64{"timestamp": time.Now().UnixMilli()}
	query := url.Values{}
	for k, v := range params {
		query.Add(k, strconv.FormatInt(v, 10))
	}
	queryString := query.Encode()
	if queryString == "" {
		return fmt.Errorf("empty query string")
	}

	// Build HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(a.binancePrikey))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	fullQuery := queryString + "&signature=" + signature

	// Send request
	fullURL := urlBase + "?" + fullQuery
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", a.binancePubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, &account); err != nil {
		return err
	}

	// Save to redis - Assets (Quoting)
	for _, asset := range account.Assets {
		var total float64
		if total, err = strconv.ParseFloat(asset.WalletBalance, 64); err != nil {
			return err
		}

		// Store individual currency balance
		if err := a.UpdateRedisPosition("binance", asset.Asset, total); err != nil {
			return err
		}
		log.Printf("Account: Binance %s balance: %f\n", asset.Asset, total)

		snapshot[asset.Asset] = total
	}

	// Save to redis - Positions (Quoting)

	for _, position := range account.Positions {
		if total, err = strconv.ParseFloat(position.PositionAmt, 64); err != nil {
			return err
		}
		available = total // TODO: For now, it's market order only.

		if err := a.UpdateRedisPosition("binance", position.Symbol, total); err != nil {
			return err
		}

		snapshot[position.Symbol] = available
		log.Printf("Position: Binance %s balance: %f, available: %f\n", position.Symbol, total, available)
	}

	if err := a.UpdateRedisWalletSnapshot("binance", snapshot); err != nil {
		return err
	}
	return nil
}
