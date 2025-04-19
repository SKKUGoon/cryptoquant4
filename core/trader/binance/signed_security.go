package binancetrade

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

// GetSignature generates a signed query string for a single order request to Binance API.
// It takes an OrderSheet and returns the URL-encoded query string and HMAC SHA256 signature.
// The function ensures the order has a timestamp, converts the order to URL parameters,
// sorts the parameters for consistency, and signs them using the trader's private key.
func (t *Trader) GetSignature(orderSheet binancerest.OrderSheet) (string, string, error) {
	if orderSheet.Timestamp == 0 {
		orderSheet.Timestamp = time.Now().UnixMilli()
	}

	params := orderSheet.ToParamsMap()
	if len(params) == 0 {
		return "", "", fmt.Errorf("empty parameters map")
	}

	// Sort keys for consistent query string
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build query string
	query := url.Values{}
	for _, k := range keys {
		if params[k] == "" {
			return "", "", fmt.Errorf("empty value for parameter %s", k)
		}
		query.Add(k, params[k])
	}
	queryString := query.Encode()
	if queryString == "" {
		return "", "", fmt.Errorf("empty query string")
	}

	// Generate HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(t.prikey))
	if len(t.prikey) == 0 {
		return "", "", fmt.Errorf("empty private key")
	}
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	return queryString, signature, nil
}

// GetSignatureBatch generates a signed query string for a batch order request to Binance API.
// It takes a slice of OrderSheets and returns the URL-encoded query string and HMAC SHA256 signature.
// The function removes individual timestamps, converts orders to a JSON array, adds a batch timestamp,
// and signs the parameters using the trader's private key. Returns empty strings if JSON marshaling fails.
func (t *Trader) GetSignatureBatch(orderSheets []binancerest.OrderSheet) (string, string, error) {
	if len(orderSheets) == 0 {
		return "", "", fmt.Errorf("empty order sheets")
	}

	params := []map[string]string{}
	for _, orderSheet := range orderSheets {
		orderSheet.RemoveTimestamp() // Remove timestamp for batch order
		pm := orderSheet.ToParamsMap()
		if len(pm) == 0 {
			return "", "", fmt.Errorf("empty parameters map in order sheet")
		}

		// Sort keys for consistent query string
		keys := make([]string, 0, len(pm))
		for k := range pm {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Validate parameters
		for _, k := range keys {
			if pm[k] == "" {
				return "", "", fmt.Errorf("empty value for parameter %s", k)
			}
		}

		params = append(params, pm)
	}

	jsonBytes, err := json.Marshal(params)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal parameters: %v", err)
	}

	// Build query string
	query := url.Values{}
	query.Set("batchOrders", string(jsonBytes))
	query.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	queryString := query.Encode()
	if queryString == "" {
		return "", "", fmt.Errorf("empty query string")
	}

	// Generate HMAC SHA256 signature
	if len(t.prikey) == 0 {
		return "", "", fmt.Errorf("empty private key")
	}
	h := hmac.New(sha256.New, []byte(t.prikey))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	return queryString, signature, nil
}
