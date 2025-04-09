package kimchiarb

import (
	"sync"
)

type Asset struct {
	mu sync.Mutex

	Symbol string

	priceChan      chan float64
	bestBidPrcChan chan float64
	bestBidQtyChan chan float64
	bestAskPrcChan chan float64
	bestAskQtyChan chan float64

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

func (a *Asset) SetBestBidPrcChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestBidPrcChan = ch
}

func (a *Asset) SetBestBidQtyChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestBidQtyChan = ch
}

func (a *Asset) SetBestAskPrcChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestAskPrcChan = ch
}

func (a *Asset) SetBestAskQtyChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestAskQtyChan = ch
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
	close(a.bestBidPrcChan)
	close(a.bestBidQtyChan)
	close(a.bestAskPrcChan)
	close(a.bestAskQtyChan)
}
