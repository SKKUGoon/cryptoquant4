package strategybase

import (
	"context"
	"log"
	"sync"
)

type SubscribableAsset struct {
	Mu     sync.RWMutex
	ctx    context.Context
	Symbol string

	// Streams
	orderbookChan chan [2][2]float64 // [[best bid, best bid qty], [best ask, best ask qty]]
	tradeChan     chan [2]float64

	// Subscribers
	BestBidPrcSubs map[string]chan float64
	BestBidQtySubs map[string]chan float64

	TradePrcSubs map[string]chan float64
	TradeQtySubs map[string]chan float64

	BestAskPrcSubs map[string]chan float64
	BestAskQtySubs map[string]chan float64
}

func NewSubscribableAsset(ctx context.Context, symbol string) *SubscribableAsset {
	return &SubscribableAsset{
		ctx:    ctx,
		Symbol: symbol,

		// Subscribers
		BestBidPrcSubs: make(map[string]chan float64),
		BestBidQtySubs: make(map[string]chan float64),
		BestAskPrcSubs: make(map[string]chan float64),
		BestAskQtySubs: make(map[string]chan float64),
		TradePrcSubs:   make(map[string]chan float64),
		TradeQtySubs:   make(map[string]chan float64),
	}
}

func (a *SubscribableAsset) Check() {
	log.Printf("\nSubscribers Table for %s:\n"+
		"+-----------------+----------+\n"+
		"| Channel         | Count    |\n"+
		"+-----------------+----------+\n"+
		"| BestBidPrc      | %8d |\n"+
		"| BestBidQty      | %8d |\n"+
		"| BestAskPrc      | %8d |\n"+
		"| BestAskQty      | %8d |\n"+
		"| TradePrc        | %8d |\n"+
		"| TradeQty        | %8d |\n"+
		"+-----------------+----------+",
		a.Symbol,
		len(a.BestBidPrcSubs),
		len(a.BestBidQtySubs),
		len(a.BestAskPrcSubs),
		len(a.BestAskQtySubs),
		len(a.TradePrcSubs),
		len(a.TradeQtySubs))
}

func (a *SubscribableAsset) SetOrderbookChan(ch chan [2][2]float64) {
	a.orderbookChan = ch
}

func (a *SubscribableAsset) SetTradeChan(ch chan [2]float64) {
	a.tradeChan = ch
}

func (a *SubscribableAsset) UpdateBestBidPrc(prc float64) {
	for _, ch := range a.BestBidPrcSubs {
		select {
		case ch <- prc:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestBidQty(qty float64) {
	for _, ch := range a.BestBidQtySubs {
		select {
		case ch <- qty:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestAskPrc(prc float64) {
	for _, ch := range a.BestAskPrcSubs {
		select {
		case ch <- prc:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateBestAskQty(qty float64) {
	for _, ch := range a.BestAskQtySubs {
		select {
		case ch <- qty:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateTradePrc(prc float64) {
	for _, ch := range a.TradePrcSubs {
		select {
		case ch <- prc:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) UpdateTradeQty(qty float64) {
	for _, ch := range a.TradeQtySubs {
		select {
		case ch <- qty:
		default:
			// Subscriber's channel is full or slow, we skip
		}
	}
}

func (a *SubscribableAsset) Listen() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case v := <-a.orderbookChan:
			a.Mu.Lock()
			a.UpdateBestBidPrc(v[0][0])
			a.UpdateBestBidQty(v[0][1])
			a.UpdateBestAskPrc(v[1][0])
			a.UpdateBestAskQty(v[1][1])
			a.Mu.Unlock()
		case v := <-a.tradeChan:
			a.Mu.Lock()
			a.UpdateTradePrc(v[0])
			a.UpdateTradeQty(v[1])
			a.Mu.Unlock()
		}
	}
}
