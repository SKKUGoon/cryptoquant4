package kimchiarb

import (
	"context"
	"fmt"
)

type KimchiPremium struct {
	KimchiAsset *Asset
	AnchorAsset *Asset // USDTKRW
	CefiAsset   *Asset // Foreign Binance

	Premium     float64
	kimchiPrice float64
	anchorPrice float64
	cefiPrice   float64
}

func (p *KimchiPremium) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case anchorPrice := <-p.AnchorAsset.priceChan:
			p.anchorPrice = anchorPrice
			p.calculatePremium()
		case cefiPrice := <-p.CefiAsset.priceChan:
			p.cefiPrice = cefiPrice
			p.calculatePremium()
		case kimchiPrice := <-p.KimchiAsset.priceChan:
			p.kimchiPrice = kimchiPrice
			p.calculatePremium()
		}
	}
}

func (p *KimchiPremium) Status() string {
	return fmt.Sprintf("Premium: %v, KimchiPrice: %v, AnchorPrice: %v, CefiPrice: %v", p.Premium, p.kimchiPrice, p.anchorPrice, p.cefiPrice)
}

func (p *KimchiPremium) calculatePremium() {
	if p.kimchiPrice == 0 || p.anchorPrice == 0 || p.cefiPrice == 0 {
		return
	}

	cefiKimchify := p.cefiPrice * p.anchorPrice
	kimchiKimchify := p.kimchiPrice
	premium := kimchiKimchify / cefiKimchify

	p.Premium = premium
}
