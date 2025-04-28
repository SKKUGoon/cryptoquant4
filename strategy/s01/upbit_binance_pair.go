package kimchiarb

import (
	"context"
	"sync"
	"time"

	"cryptoquant.com/m/data/database"
	strategybase "cryptoquant.com/m/strategy/base"
)

type UpbitBinancePair struct {
	Mu           sync.Mutex
	UpbitAsset   *strategybase.Asset
	AnchorAsset  *strategybase.Asset // USDTKRW
	BinanceAsset *strategybase.Asset // Foreign Binance

	Premium      float64
	EnterPremium float64
	ExitPremium  float64
	AnchorPrice  float64

	// For more accurate premium calculation
	BinanceBestBid    float64
	BinanceBestBidQty float64 // Calculate how much market can take

	BinanceBestAsk    float64
	BinanceBestAskQty float64 // Calculate how much market can take

	UpbitBestBid    float64
	UpbitBestBidQty float64 // Calculate how much market can take

	UpbitBestAsk    float64
	UpbitBestAskQty float64 // Calculate how much market can take

	PremiumChan chan [3]float64 // [EnterPremium, ExitPremium]
}

func (p *UpbitBinancePair) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cefibbp := <-p.BinanceAsset.BestBidPrcChan:
			p.Mu.Lock()
			p.BinanceBestBid = cefibbp
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.Mu.Unlock()
		case cefibap := <-p.BinanceAsset.BestAskPrcChan:
			p.Mu.Lock()
			p.BinanceBestAsk = cefibap
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.Mu.Unlock()
		case cefibbq := <-p.BinanceAsset.BestBidQtyChan:
			p.Mu.Lock()
			p.BinanceBestBidQty = cefibbq
			p.Mu.Unlock()
		case cefibaq := <-p.BinanceAsset.BestAskQtyChan:
			p.Mu.Lock()
			p.BinanceBestAskQty = cefibaq
			p.Mu.Unlock()
		// Kimchi
		case kimchiBestBid := <-p.UpbitAsset.BestBidPrcChan:
			p.Mu.Lock()
			p.UpbitBestBid = kimchiBestBid
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.Mu.Unlock()
		case kimchiBestAsk := <-p.UpbitAsset.BestAskPrcChan:
			p.Mu.Lock()
			p.UpbitBestAsk = kimchiBestAsk
			p.calculatePremiumEnterPos()
			p.Mu.Unlock()
		case kimchiBestBidQty := <-p.UpbitAsset.BestBidQtyChan:
			p.Mu.Lock()
			p.UpbitBestBidQty = kimchiBestBidQty
			p.Mu.Unlock()
		case kimchiBestAskQty := <-p.UpbitAsset.BestAskQtyChan:
			p.Mu.Lock()
			p.UpbitBestAskQty = kimchiBestAskQty
			p.Mu.Unlock()
		}

		// Check for correct data input
		if p.UpbitBestBid == 0 || p.UpbitBestAsk == 0 || p.BinanceBestBid == 0 || p.BinanceBestAsk == 0 || p.AnchorPrice == 0 {
			continue
		}

		p.PremiumChan <- [3]float64{p.EnterPremium, p.ExitPremium, p.AnchorPrice}
	}
}

func (p *UpbitBinancePair) ToPremiumLog() database.PremiumLog {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	return database.PremiumLog{
		Timestamp:     time.Now(),
		Symbol:        p.UpbitAsset.Symbol,
		AnchorPrice:   p.AnchorPrice,
		KimchiBestBid: p.UpbitBestBid,
		KimchiBestAsk: p.UpbitBestAsk,
		CefiBestBid:   p.BinanceBestBid,
		CefiBestAsk:   p.BinanceBestAsk,
	}
}

func (p *UpbitBinancePair) calculatePremiumEnterPos() {
	if p.UpbitBestAsk == 0 || p.BinanceBestBid == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.BinanceBestBid * p.AnchorPrice
	kimchiKimchify := p.UpbitBestAsk
	premium := kimchiKimchify / cefiKimchify
	p.EnterPremium = premium
}

func (p *UpbitBinancePair) calculatePremiumExitPos() {
	if p.UpbitBestBid == 0 || p.BinanceBestAsk == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.BinanceBestAsk * p.AnchorPrice
	kimchiKimchify := p.UpbitBestBid
	premium := kimchiKimchify / cefiKimchify
	p.ExitPremium = premium
}
