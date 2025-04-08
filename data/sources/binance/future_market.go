package binancesource

import (
	"fmt"
)

type BinanceFutureMarketData struct {
	rateLimit   int
	currentRate int
}

func NewBinanceFutureMarketData() *BinanceFutureMarketData {
	return &BinanceFutureMarketData{
		rateLimit:   1000,
		currentRate: 0,
	}
}

func (bm *BinanceFutureMarketData) GetStatus() string {
	return fmt.Sprintf("RateLimit: %d | CurrentRate: %d", bm.rateLimit, bm.currentRate)
}

func (bm *BinanceFutureMarketData) UpdateRateLimit(rateLimit int) {
	bm.rateLimit = rateLimit
}

func (bm *BinanceFutureMarketData) UpdateCurrentRate(currentRate int) {
	bm.currentRate = currentRate
}
