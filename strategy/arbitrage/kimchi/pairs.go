package kimchiarb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cryptoquant.com/m/data/database"
)

type KimchiPremium struct {
	mu          sync.Mutex
	KimchiAsset *Asset
	AnchorAsset *Asset // USDTKRW
	CefiAsset   *Asset // Foreign Binance

	Premium         float64
	PremiumEnterPos float64
	PremiumExitPos  float64
	KimchiPrice     float64
	AnchorPrice     float64
	CefiPrice       float64

	// For more accurate premium calculation
	CefiBestBid      float64
	CefiBestBidQty   float64 // Calculate how much market can take
	CefiBestAsk      float64
	CefiBestAskQty   float64 // Calculate how much market can take
	KimchiBestBid    float64
	KimchiBestBidQty float64 // Calculate how much market can take
	KimchiBestAsk    float64
	KimchiBestAskQty float64 // Calculate how much market can take
}

func (p *KimchiPremium) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case anchorPrice := <-p.AnchorAsset.priceChan:
			p.mu.Lock()
			p.AnchorPrice = anchorPrice
			p.calculatePremium()
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		// Cefi
		case cefiPrice := <-p.CefiAsset.priceChan:
			p.mu.Lock()
			p.CefiPrice = cefiPrice
			p.calculatePremium()
			p.mu.Unlock()
		case cefibbp := <-p.CefiAsset.bestBidPrcChan:
			p.mu.Lock()
			p.CefiBestBid = cefibbp
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case cefibap := <-p.CefiAsset.bestAskPrcChan:
			p.mu.Lock()
			p.CefiBestAsk = cefibap
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case cefibbq := <-p.CefiAsset.bestBidQtyChan:
			p.mu.Lock()
			p.CefiBestBidQty = cefibbq
			p.mu.Unlock()
		case cefibaq := <-p.CefiAsset.bestAskQtyChan:
			p.mu.Lock()
			p.CefiBestAskQty = cefibaq
			p.mu.Unlock()
		// Kimchi
		case kimchiPrice := <-p.KimchiAsset.priceChan:
			p.mu.Lock()
			p.KimchiPrice = kimchiPrice
			p.calculatePremium()
			p.mu.Unlock()
		case kimchiBestBid := <-p.KimchiAsset.bestBidPrcChan:
			p.mu.Lock()
			p.KimchiBestBid = kimchiBestBid
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case kimchiBestAsk := <-p.KimchiAsset.bestAskPrcChan:
			p.mu.Lock()
			p.KimchiBestAsk = kimchiBestAsk
			p.calculatePremiumEnterPos()
			p.mu.Unlock()
		case kimchiBestBidQty := <-p.KimchiAsset.bestBidQtyChan:
			p.mu.Lock()
			p.KimchiBestBidQty = kimchiBestBidQty
			p.mu.Unlock()
		case kimchiBestAskQty := <-p.KimchiAsset.bestAskQtyChan:
			p.mu.Lock()
			p.KimchiBestAskQty = kimchiBestAskQty
			p.mu.Unlock()
		}
	}
}

func (p *KimchiPremium) Status() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.KimchiPrice == 0 || p.AnchorPrice == 0 || p.CefiPrice == 0 {
		return false, "Waiting for data..."
	}
	return true, fmt.Sprintf(
		"Premium: %v, AccPremium: %v, KimchiPrice: %v, AnchorPrice: %v, CefiPrice: %v, KimchiBestBid: %v, KimchiBestAsk: %v, CefiBestBid: %v, CefiBestAsk: %v",
		p.Premium, p.PremiumEnterPos, p.KimchiPrice, p.AnchorPrice, p.CefiPrice, p.KimchiBestBid, p.KimchiBestAsk, p.CefiBestBid, p.CefiBestAsk,
	)
}

func (p *KimchiPremium) CheckEnter(enter float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.PremiumEnterPos < enter
}

func (p *KimchiPremium) CheckExit(exit float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.PremiumExitPos > exit
}

func (p *KimchiPremium) ToPremiumLog() database.PremiumLog {
	p.mu.Lock()
	defer p.mu.Unlock()
	return database.PremiumLog{
		Timestamp:       time.Now(),
		Symbol:          p.KimchiAsset.Symbol,
		Premium:         p.Premium,
		PremiumEnterPos: p.PremiumEnterPos,
		PremiumExitPos:  p.PremiumExitPos,
		KimchiPrice:     p.KimchiPrice,
		AnchorPrice:     p.AnchorPrice,
		CefiPrice:       p.CefiPrice,
		KimchiBestBid:   p.KimchiBestBid,
		KimchiBestAsk:   p.KimchiBestAsk,
		CefiBestBid:     p.CefiBestBid,
		CefiBestAsk:     p.CefiBestAsk,
	}
}

func (p *KimchiPremium) calculatePremium() {
	if p.KimchiPrice == 0 || p.AnchorPrice == 0 || p.CefiPrice == 0 {
		return
	}

	cefiKimchify := p.CefiPrice * p.AnchorPrice
	kimchiKimchify := p.KimchiPrice
	premium := kimchiKimchify / cefiKimchify

	p.Premium = premium
}

func (p *KimchiPremium) calculatePremiumEnterPos() {
	if p.KimchiBestAsk == 0 || p.CefiBestBid == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.CefiBestBid * p.AnchorPrice
	kimchiKimchify := p.KimchiBestAsk
	premium := kimchiKimchify / cefiKimchify
	p.PremiumEnterPos = premium
}

func (p *KimchiPremium) calculatePremiumExitPos() {
	if p.KimchiBestBid == 0 || p.CefiBestAsk == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.CefiBestAsk * p.AnchorPrice
	kimchiKimchify := p.KimchiBestBid
	premium := kimchiKimchify / cefiKimchify
	p.PremiumExitPos = premium
}

func (p *KimchiPremium) GetPremium() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Premium
}
