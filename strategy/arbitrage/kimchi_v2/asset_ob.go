package kimchiarbv2

import (
	"sync"
	"time"
)

type AssetBest struct {
	mu sync.Mutex

	Symbol string

	priceChan      chan float64
	bestBidPrcChan chan float64
	bestBidQtyChan chan float64
	bestAskPrcChan chan float64
	bestAskQtyChan chan float64

	timeChan chan time.Time
}

func NewAssetBest(symbol string) *AssetBest {
	return &AssetBest{
		Symbol: symbol,
	}
}

func (a *AssetBest) SetPriceChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.priceChan = ch
}

func (a *AssetBest) SetBestBidPrcChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestBidPrcChan = ch
}

func (a *AssetBest) SetBestBidQtyChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestBidQtyChan = ch
}

func (a *AssetBest) SetBestAskPrcChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestAskPrcChan = ch
}

func (a *AssetBest) SetBestAskQtyChannel(ch chan float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.bestAskQtyChan = ch
}

func (a *AssetBest) SetTimeChannel(ch chan time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.timeChan = ch
}

func (a *AssetBest) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	close(a.priceChan)
	close(a.bestBidPrcChan)
	close(a.bestBidQtyChan)
	close(a.bestAskPrcChan)
	close(a.bestAskQtyChan)
	close(a.timeChan)
}
