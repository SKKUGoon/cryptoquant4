package binancetrade

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"time"

	"cryptoquant.com/m/utils"
)

// GenerateSignature creates a signed query string for Binance API requests.
// It takes a struct containing request parameters, ensures a timestamp is set if the field exists,
// converts the struct to URL parameters, sorts them alphabetically for consistency,
// and generates an HMAC SHA256 signature using the trader's private key.
// Returns the URL-encoded query string and hex-encoded signature, or an error if validation fails.
// The signature is required by Binance for authenticated API endpoints.
func (t *Trader) GenerateSignature(data any) (string, string, error) {
	// Enforce pointer to struct
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return "", "", fmt.Errorf("input must be a pointer to a struct")
	}
	v = v.Elem()

	// Set Timestamp if present
	if tsField := v.FieldByName("Timestamp"); tsField.IsValid() &&
		tsField.CanSet() && tsField.Kind() == reflect.Int64 && tsField.Int() == 0 {
		tsField.SetInt(time.Now().UnixMilli())
	}

	params := utils.StructToParamsMap(data)
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
