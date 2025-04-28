package strategybase

import "sync"

type Asset struct {
	Mu     sync.Mutex
	Symbol string

	PriceChan      chan float64
	BestBidPrcChan chan float64
	BestBidQtyChan chan float64
	BestAskPrcChan chan float64
	BestAskQtyChan chan float64
}

func NewAsset(symbol string) *Asset {
	return &Asset{
		Symbol: symbol,
	}
}

func (a *Asset) SetPriceChannel(ch chan float64) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.PriceChan = ch
}

func (a *Asset) SetBestBidPrcChannel(ch chan float64) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.BestBidPrcChan = ch
}

func (a *Asset) SetBestBidQtyChannel(ch chan float64) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.BestBidQtyChan = ch
}

func (a *Asset) SetBestAskPrcChannel(ch chan float64) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.BestAskPrcChan = ch
}

func (a *Asset) SetBestAskQtyChannel(ch chan float64) {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	a.BestAskQtyChan = ch
}

func (a *Asset) Close() {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	close(a.PriceChan)
	close(a.BestBidPrcChan)
	close(a.BestBidQtyChan)
	close(a.BestAskPrcChan)
	close(a.BestAskQtyChan)
}

type SubscribableAsset struct {
	Mu     sync.RWMutex
	Symbol string

	bestBidPrcSubs []chan float64
	bestBidQtySubs []chan float64
	bestAskPrcSubs []chan float64
	bestAskQtySubs []chan float64
}

func NewSubscribableAsset(symbol string) *SubscribableAsset {
	return &SubscribableAsset{
		Symbol:         symbol,
		bestBidPrcSubs: make([]chan float64, 0),
		bestBidQtySubs: make([]chan float64, 0),
		bestAskPrcSubs: make([]chan float64, 0),
		bestAskQtySubs: make([]chan float64, 0),
	}
}

func (a *SubscribableAsset) UpdateBestBidPrc(prc float64) {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	for _, ch := range a.bestBidPrcSubs {
		select {
		case ch <- prc:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestBidQty(qty float64) {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	for _, ch := range a.bestBidQtySubs {
		select {
		case ch <- qty:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestAskPrc(prc float64) {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	for _, ch := range a.bestAskPrcSubs {
		select {
		case ch <- prc:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestAskQty(qty float64) {
	a.Mu.RLock()
	defer a.Mu.RUnlock()
	for _, ch := range a.bestAskQtySubs {
		select {
		case ch <- qty:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}
