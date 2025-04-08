package binancesource

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

func (bm *BinanceFutureMarketData) GetKlineData(symbol string, interval string, limit int) (binancerest.KlineDataREST, error) {
	const weight = 5
	const urlBase = "https://fapi.binance.com/fapi/v1/klines"

	if bm.rateLimit-bm.currentRate < weight {
		// Wait until the start of next minute
		now := time.Now()
		nextMinute := now.Add(time.Minute).Truncate(time.Minute)
		time.Sleep(nextMinute.Sub(now))
		bm.currentRate = 0 // Reset rate limit counter for new minute
	}

	url := fmt.Sprintf("%s?symbol=%s&interval=%s&limit=%d", urlBase, symbol, interval, limit)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Update rate limit from response headers
	if weightStr := resp.Header.Get("X-MBX-USED-WEIGHT-1M"); weightStr != "" {
		var currentWeight int
		if _, err := fmt.Sscanf(weightStr, "%d", &currentWeight); err == nil {
			bm.UpdateCurrentRate(currentWeight)
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klineData binancerest.KlineDataREST
	if err := json.Unmarshal(body, &klineData); err != nil {
		log.Printf("Failed to unmarshal kline data: %v. Body: %s", err, string(body))
		return nil, err
	}

	return klineData, nil
}
