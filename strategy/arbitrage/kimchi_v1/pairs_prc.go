package kimchiarb

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"cryptoquant.com/m/data/database"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/shopspring/decimal"
)

type UpbitBinancePair struct {
	mu          sync.Mutex
	KimchiAsset *Asset
	AnchorAsset *Asset // USDTKRW
	CefiAsset   *Asset // Foreign Binance

	Premium      float64
	EnterPremium float64
	ExitPremium  float64
	KimchiPrice  float64
	AnchorPrice  float64
	CefiPrice    float64

	// For more accurate premium calculation
	CefiBestBid      float64
	CefiBestBidQty   float64 // Calculate how much market can take
	CefiBestAsk      float64
	CefiBestAskQty   float64 // Calculate how much market can take
	KimchiBestBid    float64
	KimchiBestBidQty float64 // Calculate how much market can take
	KimchiBestAsk    float64
	KimchiBestAskQty float64 // Calculate how much market can take

	PremiumChan chan [3]float64 // [EnterPremium, ExitPremium]
}

func (p *UpbitBinancePair) Run(ctx context.Context) {
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

		// Check for correct data input
		if p.KimchiPrice == 0 || p.KimchiBestBid == 0 || p.KimchiBestAsk == 0 ||
			p.CefiPrice == 0 || p.CefiBestBid == 0 || p.CefiBestAsk == 0 ||
			p.AnchorPrice == 0 {
			continue
		}

		p.PremiumChan <- [3]float64{p.EnterPremium, p.ExitPremium, p.AnchorPrice}
	}
}

func (p *UpbitBinancePair) Status() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.KimchiPrice == 0 || p.AnchorPrice == 0 || p.CefiPrice == 0 {
		return false, "Waiting for data..."
	}
	return true, fmt.Sprintf(
		"Premium: %v, AccPremium: %v, KimchiPrice: %v, AnchorPrice: %v, CefiPrice: %v, KimchiBestBid: %v, KimchiBestAsk: %v, CefiBestBid: %v, CefiBestAsk: %v",
		p.Premium, p.EnterPremium, p.KimchiPrice, p.AnchorPrice, p.CefiPrice, p.KimchiBestBid, p.KimchiBestAsk, p.CefiBestBid, p.CefiBestAsk,
	)
}

func (p *UpbitBinancePair) CheckEnter(enter float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.EnterPremium < enter
}

func (p *UpbitBinancePair) CreatePremiumLongOrders(longFund, shortFund float64) (upbitrest.OrderSheet, binancerest.OrderSheet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate the maximum amount
	// Long and Short should equal in fund - to gurantee perfect delta hedge
	// Use only 80% of the minimum of the best bid and best ask quantity.
	// For example:
	// - If upbit's best ask's quantity is 1000
	//      binance's best bid's quantity is 900
	//   Then, we can only use 900 * 0.8 = 720
	longEnter := math.Min(p.KimchiBestAskQty*p.KimchiBestAsk, longFund) // In KRW
	shortEnter := math.Min(p.CefiBestBidQty*p.CefiBestBid, shortFund)   // In USDT

	enterFundUSDT := math.Min(shortEnter, longEnter*p.AnchorPrice) * 0.8 // In USDT
	enterFundKRW := enterFundUSDT / p.AnchorPrice                        // In KRW. Upbit requires Price * Quantity in Price

	// Upbit is always long - market order
	upbitOrderSheet := upbitrest.OrderSheet{
		Symbol:  p.KimchiAsset.Symbol,
		Side:    "bid",
		Price:   strconv.FormatFloat(enterFundKRW, 'f', -1, 64),
		OrdType: "price", // Market order
	}

	// Binance is always short - market order
	binanceOrderSheet := binancerest.OrderSheet{
		Symbol:       p.CefiAsset.Symbol,
		Side:         "SELL",
		PositionSide: "BOTH", // for One way mode.
		Type:         "MARKET",
		Quantity:     decimal.NewFromFloat(enterFundUSDT / p.CefiBestBid),
	}
	return upbitOrderSheet, binanceOrderSheet, nil
}

func (p *UpbitBinancePair) CreatePremiumShortOrders(symbol string) (upbitrest.OrderSheet, binancerest.OrderSheet, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	upbitOrderSheet := upbitrest.OrderSheet{}
	binanceOrderSheet := binancerest.OrderSheet{}
	return upbitOrderSheet, binanceOrderSheet, nil
}

func (p *UpbitBinancePair) CheckExit(exit float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ExitPremium > exit
}

func (p *UpbitBinancePair) ToPremiumLog() database.PremiumLog {
	p.mu.Lock()
	defer p.mu.Unlock()
	return database.PremiumLog{
		Timestamp:       time.Now(),
		Symbol:          p.KimchiAsset.Symbol,
		Premium:         p.Premium,
		PremiumEnterPos: p.EnterPremium,
		PremiumExitPos:  p.ExitPremium,
		KimchiPrice:     p.KimchiPrice,
		AnchorPrice:     p.AnchorPrice,
		CefiPrice:       p.CefiPrice,
		KimchiBestBid:   p.KimchiBestBid,
		KimchiBestAsk:   p.KimchiBestAsk,
		CefiBestBid:     p.CefiBestBid,
		CefiBestAsk:     p.CefiBestAsk,
	}
}

func (p *UpbitBinancePair) calculatePremium() {
	if p.KimchiPrice == 0 || p.AnchorPrice == 0 || p.CefiPrice == 0 {
		return
	}

	cefiKimchify := p.CefiPrice * p.AnchorPrice
	kimchiKimchify := p.KimchiPrice
	premium := kimchiKimchify / cefiKimchify

	p.Premium = premium
}

func (p *UpbitBinancePair) calculatePremiumEnterPos() {
	if p.KimchiBestAsk == 0 || p.CefiBestBid == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.CefiBestBid * p.AnchorPrice
	kimchiKimchify := p.KimchiBestAsk
	premium := kimchiKimchify / cefiKimchify
	p.EnterPremium = premium
}

func (p *UpbitBinancePair) calculatePremiumExitPos() {
	if p.KimchiBestBid == 0 || p.CefiBestAsk == 0 || p.AnchorPrice == 0 {
		return
	}

	cefiKimchify := p.CefiBestAsk * p.AnchorPrice
	kimchiKimchify := p.KimchiBestBid
	premium := kimchiKimchify / cefiKimchify
	p.ExitPremium = premium
}

func (p *UpbitBinancePair) GetPremium() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.Premium
}
