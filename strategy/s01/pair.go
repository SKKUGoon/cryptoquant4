package kimchiarb

import (
	"context"

	"cryptoquant.com/m/data/database"
	strategybase "cryptoquant.com/m/strategy/base"
)

type PairInfo struct {
	Symbol          string // Symbol without quoting asset, Uppercase e.g. BTC-USDT => BTC
	PairId          string // <Exchange1>_<Exchange2>_<QuotingAsset>_<Asset> e.g. UPBIT_BINANCE_USDT_BTC
	Exchange1Symbol string // BTC-KRW
	Exchange2Symbol string // BTCUSDT
	AnchorSymbol    string // USDT
}

type PairAssetChannels struct {
	KoreanAssetBestBidPrc chan float64
	KoreanAssetBestBidQty chan float64
	KoreanAssetBestAskPrc chan float64
	KoreanAssetBestAskQty chan float64

	ForeignAssetBestBidPrc chan float64
	ForeignAssetBestBidQty chan float64
	ForeignAssetBestAskPrc chan float64
	ForeignAssetBestAskQty chan float64

	AnchorPrice chan float64 // e.g. USDTKRW rate
}

// Interface's name is
// <Exchange1><Exchange2>Pair. e.g.) UpbitBinancePair
// - Exchange1: Upbit
// - Exchange2: Binance
// - QuotingAsset: USDT
type Pair interface {
	Run(ctx context.Context) error

	// Subscribers
	SubscribeKoreanAsset(*strategybase.SubscribableAsset)
	UnsubscribeKoreanAsset(*strategybase.SubscribableAsset)

	SubscribeForeignAsset(*strategybase.SubscribableAsset)
	UnsubscribeForeignAsset(*strategybase.SubscribableAsset)

	SubscribeAnchorPrice(*strategybase.SubscribableAsset)
	UnsubscribeAnchorPrice(*strategybase.SubscribableAsset)

	// Logging methods for database
	GeneratePremiumLog() database.PremiumLog

	// Premium calculation for trading purpose
	calculatePremiumEnter()
	calculatePremiumExit()

	// Setters
	SetPremiumChan(chan [3]float64)

	// Getters
	GetExchange1Orderbook() [2][2]float64
	GetExchange2Orderbook() [2][2]float64

	Close()
}

func NewPairAssetChannels() *PairAssetChannels {
	return &PairAssetChannels{
		KoreanAssetBestBidPrc: make(chan float64, 10),
		KoreanAssetBestBidQty: make(chan float64, 10),
		KoreanAssetBestAskPrc: make(chan float64, 10),
		KoreanAssetBestAskQty: make(chan float64, 10),

		ForeignAssetBestBidPrc: make(chan float64, 10),
		ForeignAssetBestBidQty: make(chan float64, 10),
		ForeignAssetBestAskPrc: make(chan float64, 10),
		ForeignAssetBestAskQty: make(chan float64, 10),

		AnchorPrice: make(chan float64, 10),
	}
}
