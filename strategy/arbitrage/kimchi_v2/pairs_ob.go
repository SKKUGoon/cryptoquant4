package kimchiarbv2

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

// At real time we cannot use the premium.
// We only enter at best bid and best ask.
// Calculated EnterPremium and ExitPremium
type UpbitBinancePair struct {
	mu          sync.Mutex
	KimchiAsset *AssetBest
	AnchorAsset *AssetBest // USDTKRW
	CefiAsset   *AssetBest // Foreign Binance

	EnterPremium float64
	ExitPremium  float64
	KimchiPrice  float64
	AnchorPrice  float64
	CefiPrice    float64

	BinanceBestBid    float64
	BinanceBestBidQty float64
	BinanceBestAsk    float64
	BinanceBestAskQty float64

	UpbitBestBid    float64
	UpbitBestBidQty float64
	UpbitBestAsk    float64
	UpbitBestAskQty float64

	Premiums chan [2]float64 // [EnterPremium, ExitPremium]
}

// Run is a function that runs the strategy.
// It listens to the price channels and calculates the premium.
// Use less `case` to reduce the number of select cases. - Increase performance.
func (p *UpbitBinancePair) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case anchorPrice := <-p.AnchorAsset.priceChan:
			p.mu.Lock()
			p.AnchorPrice = anchorPrice
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		// Cefi
		case cefibbp := <-p.CefiAsset.bestBidPrcChan:
			p.mu.Lock()
			p.BinanceBestBid = cefibbp
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case cefibap := <-p.CefiAsset.bestAskPrcChan:
			p.mu.Lock()
			p.BinanceBestAsk = cefibap
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case cefibbq := <-p.CefiAsset.bestBidQtyChan:
			p.mu.Lock()
			p.BinanceBestBidQty = cefibbq
			p.mu.Unlock()
		case cefibaq := <-p.CefiAsset.bestAskQtyChan:
			p.mu.Lock()
			p.BinanceBestAskQty = cefibaq
			p.mu.Unlock()
		// Kimchi
		case kimchiBestBid := <-p.KimchiAsset.bestBidPrcChan:
			p.mu.Lock()
			p.UpbitBestBid = kimchiBestBid
			p.calculatePremiumEnterPos()
			p.calculatePremiumExitPos()
			p.mu.Unlock()
		case kimchiBestAsk := <-p.KimchiAsset.bestAskPrcChan:
			p.mu.Lock()
			p.UpbitBestAsk = kimchiBestAsk
			p.calculatePremiumEnterPos()
			p.mu.Unlock()
		case kimchiBestBidQty := <-p.KimchiAsset.bestBidQtyChan:
			p.mu.Lock()
			p.UpbitBestBidQty = kimchiBestBidQty
			p.mu.Unlock()
		case kimchiBestAskQty := <-p.KimchiAsset.bestAskQtyChan:
			p.mu.Lock()
			p.UpbitBestAskQty = kimchiBestAskQty
			p.mu.Unlock()
		}
	}
}

func (p *UpbitBinancePair) Status() (bool, string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.KimchiPrice == 0 || p.AnchorPrice == 0 || p.CefiPrice == 0 {
		return false, "Waiting for data..."
	}
	return true, fmt.Sprintf(
		"EnterP: %v, ExitP: %v, KimchiBB: %v, KimchiBA: %v, CefiBB: %v, CefiBA: %v",
		p.EnterPremium, p.ExitPremium, p.UpbitBestBid, p.UpbitBestAsk, p.BinanceBestBid, p.BinanceBestAsk,
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
	longEnter := math.Min(p.UpbitBestAskQty*p.UpbitBestAsk, longFund)       // In KRW
	shortEnter := math.Min(p.BinanceBestBidQty*p.BinanceBestBid, shortFund) // In USDT

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
		Quantity:     decimal.NewFromFloat(enterFundUSDT / p.BinanceBestBid),
	}
	return upbitOrderSheet, binanceOrderSheet, nil
}

// CreatePremiumShortOrders sends empty order sheets.
// It is filled by the engine and account source.
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
		Timestamp:     time.Now(),
		Symbol:        p.KimchiAsset.Symbol,
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
