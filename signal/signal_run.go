package signal

import (
	"log"
	"time"

	traderpb "cryptoquant.com/m/gen/traderpb"
	binancews "cryptoquant.com/m/internal/binance/ws"
	upbitws "cryptoquant.com/m/internal/upbit/ws"
	kimchiarbv1 "cryptoquant.com/m/strategy/arbitrage/kimchi_v1"
	binancemarket "cryptoquant.com/m/streams/binance/market"
	upbitmarket "cryptoquant.com/m/streams/upbit/market"
)

// StartAssetPair initializes the asset objects for the trading pairs and sets up their channels.
// It creates:
// 1. KimchiAsset: Represents the Kimchi trading pair
// 2. ForexAsset: Represents the Cefi trading pair
// 3. AnchorAsset: Represents the anchor trading pair
func (e *SignalContext) StartAssetPair() {
	log.Printf("Starting asset for %v", e.UpbitAssetSymbol)
	kimchiAsset := kimchiarbv1.NewAsset(e.UpbitAssetSymbol)
	forexAsset := kimchiarbv1.NewAsset(e.BinanceAssetSymbol)
	anchorAsset := kimchiarbv1.NewAsset(e.AnchorAssetSymbol)

	kimchiAsset.SetPriceChannel(e.upbitTradeChan)
	forexAsset.SetPriceChannel(e.binanceTradeChan)
	anchorAsset.SetPriceChannel(e.anchorTradeChan)

	kimchiAsset.SetBestBidPrcChannel(e.upbitBestBidPrcChan)
	forexAsset.SetBestBidPrcChannel(e.binanceBestBidPrcChan)
	anchorAsset.SetBestBidPrcChannel(e.anchorBestBidPrcChan)

	kimchiAsset.SetBestAskPrcChannel(e.upbitBestAskPrcChan)
	forexAsset.SetBestAskPrcChannel(e.binanceBestAskPrcChan)
	anchorAsset.SetBestAskPrcChannel(e.anchorBestAskPrcChan)

	kimchiAsset.SetBestBidQtyChannel(e.upbitBestBidQtyChan)
	forexAsset.SetBestBidQtyChannel(e.binanceBestBidQtyChan)
	anchorAsset.SetBestBidQtyChannel(e.anchorBestBidQtyChan)

	kimchiAsset.SetBestAskQtyChannel(e.upbitBestAskQtyChan)
	forexAsset.SetBestAskQtyChannel(e.binanceBestAskQtyChan)
	anchorAsset.SetBestAskQtyChannel(e.anchorBestAskQtyChan)

	e.UpbitBinancePairs = &kimchiarbv1.UpbitBinancePair{
		AnchorAsset: anchorAsset,
		CefiAsset:   forexAsset,
		KimchiAsset: kimchiAsset,
		PremiumChan: e.premiumChan,
	}

	log.Printf("Asset Pairs(%v, %v, %v) initialized", e.UpbitAssetSymbol, e.BinanceAssetSymbol, e.AnchorAssetSymbol)
}

// StartAssetStreams starts the streams for the asset pairs.
// It creates total of 5 streams:
// 1. KimchiStream: Represents the Kimchi trading pair (Upbit, 2 streams: price and book)
// 2. CefiStream: Represents the Cefi trading pair (Binance, 2 streams: price and book)
// 3. AnchorStream: Represents the anchor trading pair (Upbit, 1 stream: price)
func (e *SignalContext) StartAssetStreams() {
	log.Println("Engine started")

	// Start stream - Kimchi - Upbit
	log.Printf("Starting stream for %v", e.UpbitAssetSymbol)
	kimchi1 := upbitmarket.NewPriceHandler(e.upbitTradeChan)
	kimchi2 := upbitmarket.NewBestBidPrcHandler(e.upbitBestBidPrcChan)
	kimchi3 := upbitmarket.NewBestBidQtyHandler(e.upbitBestBidQtyChan)
	kimchi4 := upbitmarket.NewBestAskPrcHandler(e.upbitBestAskPrcChan)
	kimchi5 := upbitmarket.NewBestAskQtyHandler(e.upbitBestAskQtyChan)
	kimchiHandlerPrice := []func(upbitws.SpotTrade) error{kimchi1}
	kimchiHandlerBook := []func(upbitws.SpotOrderbook) error{kimchi2, kimchi3, kimchi4, kimchi5}
	go upbitmarket.SubscribeTrade(e.ctx, e.UpbitAssetSymbol, kimchiHandlerPrice)
	go upbitmarket.SubscribeBook(e.ctx, e.UpbitAssetSymbol, kimchiHandlerBook)

	// Start stream - Binance
	log.Printf("Starting stream for %v", e.BinanceAssetSymbol)
	cefi1 := binancemarket.NewPriceHandler(e.binanceTradeChan)
	cefi2 := binancemarket.NewBestBidPrcHandler(e.binanceBestBidPrcChan)
	cefi3 := binancemarket.NewBestBidQtyHandler(e.binanceBestBidQtyChan)
	cefi4 := binancemarket.NewBestAskPrcHandler(e.binanceBestAskPrcChan)
	cefi5 := binancemarket.NewBestAskQtyHandler(e.binanceBestAskQtyChan)
	cefiHandlerPrice := []func(binancews.FutureAggTrade) error{cefi1}
	cefiHandlerBook := []func(binancews.FutureBookTicker) error{cefi2, cefi3, cefi4, cefi5}
	go binancemarket.SubscribeAggtrade(e.ctx, e.BinanceAssetSymbol, cefiHandlerPrice)
	go binancemarket.SubscribeBook(e.ctx, e.BinanceAssetSymbol, cefiHandlerBook)

	// Start stream - Anchor - Upbit
	log.Printf("Starting stream for %v", e.AnchorAssetSymbol)
	anchor1 := upbitmarket.NewPriceHandler(e.anchorTradeChan)
	h3s := []func(upbitws.SpotTrade) error{anchor1}
	go upbitmarket.SubscribeTrade(e.ctx, e.AnchorAssetSymbol, h3s)
}

// Run starts the arbitrage strategy.
// It creates a goroutine that runs the arbitrage strategy and tracks consecutive failures.
// If the strategy is not ready, it logs an error and panics. Otherwise, it logs a message and starts the strategy.
func (e *SignalContext) Run() {
	log.Println("Starting strategy")
	go e.UpbitBinancePairs.Run(e.ctx)

	// TEST Send test trade once after 60 seconds
	// TODO: Remove this after testing
	go func() {
		timer := time.NewTimer(60 * time.Second)
		defer timer.Stop()

		select {
		case <-e.ctx.Done():
			return
		case <-timer.C:
			log.Println("Sending one-time test trade after 60 seconds")
			_, err := e.traderMessenger.SubmitTrade(&traderpb.TradeRequest{
				OrderType: &traderpb.TradeRequest_PairOrder{
					PairOrder: &traderpb.PairOrderSheet{
						BaseSymbol: e.SignalID,
						UpbitOrder: &traderpb.ExchangeOrder{
							Symbol:    e.UpbitAssetSymbol,
							Side:      "buy",
							Amount:    1,
							AmountKey: traderpb.AmountKey_TOTAL_VALUE,
						},
						BinanceOrder: &traderpb.ExchangeOrder{
							Symbol:    e.BinanceAssetSymbol,
							Side:      "sell",
							Amount:    1,
							AmountKey: traderpb.AmountKey_QUANTITY,
						},
					},
				},
			})
			if err != nil {
				log.Printf("Failed to submit test trade: %v", err)
			}
			return // Exit goroutine after single execution
		}
	}()

	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return
			case premiums := <-e.premiumChan:
				enter := premiums[0]
				exit := premiums[1]

				switch true {
				case e.inPosition && exit > e.ExitPremiumBoundary:
					log.Printf("Exiting position: %v (standard: %v)", exit, e.ExitPremiumBoundary)

					// TODO: Send Exit Position order signal with protobuf
					_, err := e.traderMessenger.SubmitTrade(&traderpb.TradeRequest{
						OrderType: &traderpb.TradeRequest_PairOrder{
							PairOrder: &traderpb.PairOrderSheet{
								BaseSymbol: e.SignalID,
								UpbitOrder: &traderpb.ExchangeOrder{
									Symbol:    e.UpbitAssetSymbol,
									Side:      "sell",
									Amount:    1,
									AmountKey: traderpb.AmountKey_TOTAL_VALUE,
								},
								BinanceOrder: &traderpb.ExchangeOrder{
									Symbol:    e.BinanceAssetSymbol,
									Side:      "buy",
									Amount:    1,
									AmountKey: traderpb.AmountKey_QUANTITY,
								},
							},
						},
					})
					if err != nil {
						log.Printf("Failed to submit trade: %v", err)
					} else {
						e.ChangePositionStatus()
					}

				case !e.inPosition && enter < e.EnterPremiumBoundary:
					log.Printf("Entering position: %v (standard: %v)", enter, e.EnterPremiumBoundary)

					_, err := e.traderMessenger.SubmitTrade(&traderpb.TradeRequest{
						OrderType: &traderpb.TradeRequest_PairOrder{
							PairOrder: &traderpb.PairOrderSheet{
								BaseSymbol: e.SignalID,
								UpbitOrder: &traderpb.ExchangeOrder{
									Symbol:    e.UpbitAssetSymbol,
									Side:      "buy",
									Amount:    1,
									AmountKey: traderpb.AmountKey_TOTAL_VALUE,
								},
								BinanceOrder: &traderpb.ExchangeOrder{
									Symbol:    e.BinanceAssetSymbol,
									Side:      "sell",
									Amount:    1,
									AmountKey: traderpb.AmountKey_QUANTITY,
								},
							},
						},
					})
					if err != nil {
						log.Printf("Failed to submit trade: %v", err)
					} else {
						e.ChangePositionStatus()
					}
				}
				log.Printf("Premium: %v, %v", enter, exit)
			}
		}
	}()
}
