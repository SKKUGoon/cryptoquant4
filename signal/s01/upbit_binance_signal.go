package s01signal

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	config "cryptoquant.com/m/config"
	database "cryptoquant.com/m/data/database"
	"cryptoquant.com/m/gen/traderpb"
	signal "cryptoquant.com/m/signal"
	s01 "cryptoquant.com/m/strategy/s01"
)

type UpbitBinanceSignal struct {
	ctx context.Context

	// Utility structs
	traderMessenger *signal.TraderMessenger
	database        *database.Database
	timeScale       *database.TimeScale

	// Mutable Config.
	// - Gained from Database and API
	UpbitExchangeConfig   *config.UpbitSpotTradeConfig
	BinanceExchangeConfig *config.BinanceFutureTradeConfig
	EnterPremiumBoundary  float64
	ExitPremiumBoundary   float64

	// Immutable Config.
	// - Gained from environment variables

	// Trade Process
	pair        *s01.UpbitBinancePair
	inPosition  bool
	premiumChan chan [3]float64 // [EnterPremium, ExitPremium, AnchorPrice]
	dataLogChan chan database.PremiumLog
}

func NewUpbitBinanceSignal(ctx context.Context, pair *s01.UpbitBinancePair) *UpbitBinanceSignal {
	var symbol string
	var traderAddr string

	// Get information from environment variables
	if symbol = os.Getenv("SYMBOL"); symbol == "" {
		panic("environment variable `SYMBOL` is not set")
	}

	if traderAddr = os.Getenv("TRADER_ADDRESS"); traderAddr == "" {
		panic("environment variable `TRADER_ADDRESS` is not set")
	}

	// Generate utility structs
	traderMessenger := signal.NewTraderMessenger(traderAddr, ctx)
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}
	ts, err := database.ConnectTS()
	if err != nil {
		panic(err)
	}

	return &UpbitBinanceSignal{
		ctx:             ctx,
		traderMessenger: traderMessenger,
		database:        db,
		timeScale:       ts,
		pair:            pair,
	}
}

// Setters
func (s *UpbitBinanceSignal) SetPremiumChan(premiumChan chan [3]float64) {
	s.premiumChan = premiumChan
}

func (s *UpbitBinanceSignal) SetDataLogChan(dataLogChan chan database.PremiumLog) {
	s.dataLogChan = dataLogChan
}

// Update methods
func (s *UpbitBinanceSignal) UpdateUpbitExchangeConfig() {
	upbitConfig, err := config.NewUpbitSpotTradeConfig()
	if err != nil {
		log.Printf("Failed to create Upbit config: %v\n", err)
	}
	s.UpbitExchangeConfig = upbitConfig
}

func (s *UpbitBinanceSignal) UpdateBinanceExchangeConfig() {
	binanceConfig, err := config.NewBinanceFutureTradeConfig()
	if err != nil {
		log.Printf("Failed to create Binance config: %v\n", err)
	}
	s.BinanceExchangeConfig = binanceConfig
}

func (s *UpbitBinanceSignal) UpdateEnterPremiumBoundary() {
	if s.database == nil {
		log.Println("[WARNING] database not set — skipping")
		return
	}

	key := fmt.Sprintf("%v_enter_premium_boundary", strings.ToLower(s.pair.PairInfo.Symbol))

	enterPremiumBoundary, err := s.database.GetTradeMetadata(key, 0.9980)
	if err != nil {
		log.Printf("Failed to get enter premium boundary: %v", err)
		panic(err)
	}
	s.EnterPremiumBoundary = enterPremiumBoundary.(float64)
}

func (s *UpbitBinanceSignal) UpdateExitPremiumBoundary() {
	if s.database == nil {
		log.Println("[WARNING] database not set — skipping")
		return
	}

	key := fmt.Sprintf("%v_exit_premium_boundary", strings.ToLower(s.pair.PairInfo.Symbol))

	exitPremiumBoundary, err := s.database.GetTradeMetadata(key, 1.0035)
	if err != nil {
		log.Printf("Failed to get exit premium boundary: %v", err)
		panic(err)
	}
	s.ExitPremiumBoundary = exitPremiumBoundary.(float64)
}

// Check it before running
func (s *UpbitBinanceSignal) Check() {
	if s.UpbitExchangeConfig == nil || s.BinanceExchangeConfig == nil {
		log.Println("[WARNING] Upbit or Binance config not set — skipping")
		return
	}

	if !s.UpbitExchangeConfig.IsAvailableSymbol(s.pair.PairInfo.Exchange1Symbol) {
		log.Println("[WARNING] Upbit asset symbol not available — skipping")
		return
	}

	if !s.BinanceExchangeConfig.IsAvailableSymbol(s.pair.PairInfo.Exchange2Symbol) {
		log.Println("[WARNING] Binance asset symbol not available — skipping")
		return
	}
}

func (s *UpbitBinanceSignal) Run() {
	logTicker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-logTicker.C:
			s.pair.Mu.Lock()
			if dl, err := s.pair.GeneratePremiumLog(); err == nil {
				select {
				case s.dataLogChan <- dl:
				default:
					log.Println("[WARNING] dataLogChan buffer full — skipping")
				}
			} else {
				log.Printf("Failed to generate premium log: %v", err)
			}
			s.pair.Mu.Unlock()

		case premium := <-s.premiumChan:
			// Lock the pair to prevent race condition
			log.Println("premium:", premium)
			enterPremium := premium[0]
			exitPremium := premium[1]
			anchorPrice := premium[2]

			switch s.inPosition {
			case true:
				if exitPremium > s.ExitPremiumBoundary {
					s.pair.Mu.Lock()
					exchange1Orderbook := s.pair.GetExchange1Orderbook() // [best bid, best bid qty], [best ask, best ask qty]
					exchange2Orderbook := s.pair.GetExchange2Orderbook() // [best bid, best bid qty], [best ask, best ask qty]

					_, err := s.traderMessenger.SubmitTrade(&traderpb.TradeRequest{
						OrderType: &traderpb.TradeRequest_PairOrder{
							PairOrder: &traderpb.PairOrderSheet{
								BaseSymbol:    s.pair.PairInfo.Symbol,
								PairOrderType: traderpb.PairOrderType_PairOrderExit,
								ExchangeRate:  anchorPrice,
								// Upbit: Exchange 1
								UpbitOrder: &traderpb.ExchangeOrder{
									Symbol: s.pair.PairInfo.Exchange1Symbol,
									Side:   "sell",
									Price:  exchange1Orderbook[0][0],
									Amount: exchange1Orderbook[0][1],
								},
								// Binance: Exchange 2
								BinanceOrder: &traderpb.ExchangeOrder{
									Symbol: s.pair.PairInfo.Exchange2Symbol,
									Side:   "buy",
									Price:  exchange2Orderbook[1][0],
									Amount: exchange2Orderbook[1][1],
								},
							},
						},
					})
					if err != nil {
						log.Printf("Failed to submit trade: %v", err)
					} else {
						s.inPosition = false
					}
					s.pair.Mu.Unlock()
				}
			case false:
				if enterPremium < s.EnterPremiumBoundary {
					s.pair.Mu.Lock()
					exchange1Orderbook := s.pair.GetExchange1Orderbook() // [best bid, best bid qty], [best ask, best ask qty]
					exchange2Orderbook := s.pair.GetExchange2Orderbook() // [best bid, best bid qty], [best ask, best ask qty]

					_, err := s.traderMessenger.SubmitTrade(&traderpb.TradeRequest{
						OrderType: &traderpb.TradeRequest_PairOrder{
							PairOrder: &traderpb.PairOrderSheet{
								BaseSymbol:    s.pair.PairInfo.Symbol,
								PairOrderType: traderpb.PairOrderType_PairOrderEnter,
								ExchangeRate:  anchorPrice,
								// Upbit: Exchange 1
								UpbitOrder: &traderpb.ExchangeOrder{
									Symbol: s.pair.PairInfo.Exchange1Symbol,
									Side:   "buy",
									Price:  exchange1Orderbook[1][0],
									Amount: exchange1Orderbook[1][1],
								},
								// Binance: Exchange 2
								BinanceOrder: &traderpb.ExchangeOrder{
									Symbol: s.pair.PairInfo.Exchange2Symbol,
									Side:   "sell",
									Price:  exchange2Orderbook[0][0],
									Amount: exchange2Orderbook[0][1],
								},
							},
						},
					})
					if err != nil {
						log.Printf("Failed to submit trade: %v", err)
					} else {
						s.inPosition = true
					}
					s.pair.Mu.Unlock()
				}
			}

		case <-s.ctx.Done():
			return
		}
	}
}
