package kimchiarb

import (
	"sync"
)

type Asset struct {
	mu sync.Mutex

	Symbol string

	priceChan chan float64
	bestBid   chan float64
	bestAsk   chan float64

	pricePrecision    int
	quantityPrecision int
}

func NewAsset(symbol string) *Asset {
	return &Asset{
		Symbol: symbol,
	}
}

func (a *Asset) SetPriceChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.priceChan = ch
}

func (a *Asset) SetBestBidChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestBid = ch
}

func (a *Asset) SetBestAskChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestAsk = ch
}

func (a *Asset) SetPricePrecision(precision int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pricePrecision = precision
}

func (a *Asset) GetPricePrecision() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.pricePrecision
}

func (a *Asset) SetQuantityPrecision(precision int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.quantityPrecision = precision
}

func (a *Asset) GetQuantityPrecision() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.quantityPrecision
}

func (a *Asset) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	close(a.priceChan)
	close(a.bestBid)
	close(a.bestAsk)
}
