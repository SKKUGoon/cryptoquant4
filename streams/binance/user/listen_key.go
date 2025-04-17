package binanceuser

import (
	"encoding/json"
	"net/http"
	"os"

	binancews "cryptoquant.com/m/internal/binance/ws"
)

const listenKeyURL = "https://fapi.binance.com/fapi/v1/listenKey"

// CreateListenKey creates a listen key for user data stream
func CreateListenKey() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", listenKeyURL, nil)
	if err != nil {
		return "", err
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	req.Header.Add("X-MBX-APIKEY", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	var result binancews.ListenKeyResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.ListenKey, nil
}

func KeepAliveListenKey(listenKey string) error {
	client := &http.Client{}
	req, err := http.NewRequest("PUT", listenKeyURL, nil)
	if err != nil {
		return err
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	req.Header.Add("X-MBX-APIKEY", apiKey)
	req.Header.Add("Content-Type", "application/json")

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func CloseListenKey(listenKey string) error {
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", listenKeyURL, nil)
	if err != nil {
		return err
	}

	apiKey := os.Getenv("BINANCE_API_KEY")
	req.Header.Add("X-MBX-APIKEY", apiKey)

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}
