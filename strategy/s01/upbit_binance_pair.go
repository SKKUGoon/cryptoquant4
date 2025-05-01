package kimchiarb

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cryptoquant.com/m/data/database"
	strategybase "cryptoquant.com/m/strategy/base"
)

// UpbitBinancePair represents a trading pair between Upbit and Binance exchanges.
// It handles price data streams and calculates premiums between the two markets.
type UpbitBinancePair struct {
	*PairInfo
	*PairAssetChannels // Embedded Channels

	Mu  sync.Mutex
	ctx context.Context

	// Premiums represents the price differences between markets
	EnterPremium float64 // Premium threshold to enter a position
	ExitPremium  float64 // Premium threshold to exit a position

	// Exchange rate - Anchor price (e.g. USDT/KRW rate)
	anchorPrice float64

	// Best bid/ask prices and quantities from both exchanges
	upbitBestBid      float64
	upbitBestBidQty   float64
	upbitBestAsk      float64
	upbitBestAskQty   float64
	binanceBestBid    float64
	binanceBestBidQty float64
	binanceBestAsk    float64
	binanceBestAskQty float64

	// Channel for transferring premium data to strategy
	// Format: [EnterPremium, ExitPremium, AnchorPrice]
	premiumChan chan [3]float64
}

// NewUpbitBinancePair creates a new pair instance with initialized channels
func NewUpbitBinancePair(ctx context.Context, targetSymbol string, anchorSymbol string) *UpbitBinancePair {
	channels := NewPairAssetChannels()

	return &UpbitBinancePair{
		ctx:               ctx,
		PairAssetChannels: channels,
		PairInfo: &PairInfo{
			Symbol: targetSymbol,
			PairId: fmt.Sprintf("%s_%s_%s_%s", "UPBIT", "BINANCE", strings.ToUpper(anchorSymbol), strings.ToUpper(targetSymbol)),
		}, // Used for logging purposes
	}
}

func (p *UpbitBinancePair) GetExchange1Orderbook() [2][2]float64 {
	return [2][2]float64{
		{p.upbitBestBid, p.upbitBestBidQty},
		{p.upbitBestAsk, p.upbitBestAskQty},
	}
}

func (p *UpbitBinancePair) GetExchange2Orderbook() [2][2]float64 {
	return [2][2]float64{
		{p.binanceBestBid, p.binanceBestBidQty},
		{p.binanceBestAsk, p.binanceBestAskQty},
	}
}

// SetPremiumChan sets the channel used to send premium updates
func (p *UpbitBinancePair) SetPremiumChan(ch chan [3]float64) {
	p.premiumChan = ch
}

// SubscribeKoreanAsset subscribes to Korean market (Upbit) price feeds
func (p *UpbitBinancePair) SubscribeKoreanAsset(asset *strategybase.SubscribableAsset) {
	asset.BestBidPrcSubs[p.PairId] = p.KoreanAssetBestBidPrc
	asset.BestBidQtySubs[p.PairId] = p.KoreanAssetBestBidQty
	asset.BestAskPrcSubs[p.PairId] = p.KoreanAssetBestAskPrc
	asset.BestAskQtySubs[p.PairId] = p.KoreanAssetBestAskQty
}

// SubscribeForeignAsset subscribes to foreign market (Binance) price feeds
func (p *UpbitBinancePair) SubscribeForeignAsset(asset *strategybase.SubscribableAsset) {
	asset.BestBidPrcSubs[p.PairId] = p.ForeignAssetBestBidPrc
	asset.BestBidQtySubs[p.PairId] = p.ForeignAssetBestBidQty
	asset.BestAskPrcSubs[p.PairId] = p.ForeignAssetBestAskPrc
	asset.BestAskQtySubs[p.PairId] = p.ForeignAssetBestAskQty
}

// SubscribeAnchorPrice subscribes to exchange rate updates
func (p *UpbitBinancePair) SubscribeAnchorPrice(asset *strategybase.SubscribableAsset) {
	asset.TradePrcSubs[p.PairId] = p.AnchorPrice
}

// UnsubscribeKoreanAsset removes subscriptions from Korean market feeds
func (p *UpbitBinancePair) UnsubscribeKoreanAsset(asset *strategybase.SubscribableAsset) {
	delete(asset.BestBidPrcSubs, p.PairId)
	delete(asset.BestBidQtySubs, p.PairId)
	delete(asset.BestAskPrcSubs, p.PairId)
	delete(asset.BestAskQtySubs, p.PairId)
}

// UnsubscribeForeignAsset removes subscriptions from foreign market feeds
func (p *UpbitBinancePair) UnsubscribeForeignAsset(asset *strategybase.SubscribableAsset) {
	delete(asset.BestBidPrcSubs, p.PairId)
	delete(asset.BestBidQtySubs, p.PairId)
	delete(asset.BestAskPrcSubs, p.PairId)
	delete(asset.BestAskQtySubs, p.PairId)
}

// UnsubscribeAnchorPrice removes subscription from exchange rate feed
func (p *UpbitBinancePair) UnsubscribeAnchorPrice(asset *strategybase.SubscribableAsset) {
	delete(asset.TradePrcSubs, p.PairId)
}

// GeneratePremiumLog creates a log entry of current premium state
func (p *UpbitBinancePair) GeneratePremiumLog() (database.PremiumLog, error) {
	if p.upbitBestAsk == 0 || p.upbitBestBid == 0 || p.binanceBestBid == 0 || p.binanceBestAsk == 0 || p.anchorPrice == 0 {
		return database.PremiumLog{}, fmt.Errorf("missing data")
	}
	return database.PremiumLog{
		Timestamp:    time.Now(),
		Symbol:       p.Symbol,
		AnchorPrice:  p.anchorPrice,
		EnterPremium: p.EnterPremium,
		ExitPremium:  p.ExitPremium,
	}, nil
}

// calculatePremiumEnter calculates the premium for entering a position
// Premium = (Upbit Ask) / (Binance Bid * Exchange Rate)
func (p *UpbitBinancePair) calculatePremiumEnter() {
	if p.upbitBestAsk == 0 || p.binanceBestBid == 0 || p.anchorPrice == 0 {
		return
	}

	cefiKimchify := p.binanceBestBid * p.anchorPrice
	kimchiKimchify := p.upbitBestAsk
	premium := kimchiKimchify / cefiKimchify

	p.EnterPremium = premium
}

// calculatePremiumExit calculates the premium for exiting a position
// Premium = (Upbit Bid) / (Binance Ask * Exchange Rate)
func (p *UpbitBinancePair) calculatePremiumExit() {
	if p.upbitBestBid == 0 || p.binanceBestAsk == 0 || p.anchorPrice == 0 {
		return
	}

	cefiKimchify := p.binanceBestAsk * p.anchorPrice
	kimchiKimchify := p.upbitBestBid
	premium := kimchiKimchify / cefiKimchify

	p.ExitPremium = premium
}

// Run starts the main event loop that processes incoming price updates
func (p *UpbitBinancePair) Run() {
	discarded := 0
	logTicker := time.NewTicker(5 * time.Second)
	defer logTicker.Stop()
	for {
		select {
		case <-p.ctx.Done():
			return

		case <-logTicker.C:
			if discarded > 0 {
				log.Printf("Discarded %d messages", discarded)
				discarded = 0
			}

		case v := <-p.KoreanAssetBestBidPrc:
			p.Mu.Lock()
			p.upbitBestBid = v
			p.Mu.Unlock()

		case v := <-p.KoreanAssetBestBidQty:
			p.Mu.Lock()
			p.upbitBestBidQty = v
			p.Mu.Unlock()

		case v := <-p.KoreanAssetBestAskPrc:
			p.Mu.Lock()
			p.upbitBestAsk = v
			p.Mu.Unlock()

		case v := <-p.KoreanAssetBestAskQty:
			p.Mu.Lock()
			p.upbitBestAskQty = v
			p.Mu.Unlock()

		case v := <-p.AnchorPrice:
			p.Mu.Lock()
			p.anchorPrice = v
			p.Mu.Unlock()

		case v := <-p.ForeignAssetBestBidPrc:
			p.Mu.Lock()
			p.binanceBestBid = v
			p.Mu.Unlock()

		case v := <-p.ForeignAssetBestBidQty:
			p.Mu.Lock()
			p.binanceBestBidQty = v
			p.Mu.Unlock()

		case v := <-p.ForeignAssetBestAskPrc:
			p.Mu.Lock()
			p.binanceBestAsk = v
			p.Mu.Unlock()

		case v := <-p.ForeignAssetBestAskQty:
			p.Mu.Lock()
			p.binanceBestAskQty = v
			p.Mu.Unlock()
		}

		// Check for correct data input
		if p.upbitBestAsk == 0 || p.upbitBestBid == 0 || p.binanceBestBid == 0 || p.binanceBestAsk == 0 || p.anchorPrice == 0 {
			continue
		}

		// Calculate premiums if ready
		p.calculatePremiumEnter()
		p.calculatePremiumExit()

		select {
		case p.premiumChan <- [3]float64{p.EnterPremium, p.ExitPremium, p.anchorPrice}:
		default:
			discarded++
		}
	}
}

// Close cleans up resources by closing channels and resetting state
func (p *UpbitBinancePair) Close() {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	// Close internal channels
	close(p.KoreanAssetBestBidPrc)
	close(p.KoreanAssetBestBidQty)
	close(p.KoreanAssetBestAskPrc)
	close(p.KoreanAssetBestAskQty)

	close(p.ForeignAssetBestBidPrc)
	close(p.ForeignAssetBestBidQty)
	close(p.ForeignAssetBestAskPrc)
	close(p.ForeignAssetBestAskQty)

	close(p.AnchorPrice)

	// (premiumChan is passed externally, maybe don't close here — or optionally close if you own it.)

	// Clear premium calculation numbers
	p.EnterPremium = 0
	p.ExitPremium = 0
}
