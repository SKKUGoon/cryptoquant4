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

	Premium     float64
	KimchiPrice float64
	AnchorPrice float64
	CefiPrice   float64
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
			p.mu.Unlock()
		case cefiPrice := <-p.CefiAsset.priceChan:
			p.mu.Lock()
			p.CefiPrice = cefiPrice
			p.calculatePremium()
			p.mu.Unlock()
		case kimchiPrice := <-p.KimchiAsset.priceChan:
			p.mu.Lock()
			p.KimchiPrice = kimchiPrice
			p.calculatePremium()
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
	return true, fmt.Sprintf("Premium: %v, KimchiPrice: %v, AnchorPrice: %v, CefiPrice: %v", p.Premium, p.KimchiPrice, p.AnchorPrice, p.CefiPrice)
}

func (p *KimchiPremium) ToPremiumLog() database.PremiumLog {
	p.mu.Lock()
	defer p.mu.Unlock()
	return database.PremiumLog{
		Timestamp:   time.Now(),
		Symbol:      p.KimchiAsset.Symbol,
		Premium:     p.Premium,
		KimchiPrice: p.KimchiPrice,
		AnchorPrice: p.AnchorPrice,
		CefiPrice:   p.CefiPrice,
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

func (p *KimchiPremium) GetPremium() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Premium
}
