package binancetrade

import (
	"log"
	"os"
	"time"

	binancerest "cryptoquant.com/m/internal/binance/rest"
)

type Trader struct {
	// API Key
	pubkey string // Binance API Key
	prikey string // Binance API Secret

	// Order rate limit
	rateLimit   int
	currentRate int

	// Channels to receive orders from strategy
	orders     chan binancerest.OrderSheet
	inPosition bool
}

func NewTrader() *Trader {
	pubkey := os.Getenv("BINANCE_API_KEY")
	prikey := os.Getenv("BINANCE_SECRET_KEY")

	if pubkey == "" || prikey == "" {
		log.Fatalf("[trader] BINANCE_API_KEY or BINANCE_SECRET_KEY is not set")
	}

	return &Trader{
		pubkey:     pubkey,
		prikey:     prikey,
		inPosition: false,
	}
}

func (t *Trader) SetTestPubKey(pubkey string) {
	t.pubkey = pubkey
}

func (t *Trader) SetTestPriKey(prikey string) {
	t.prikey = prikey
}

func (t *Trader) SetOrderChannel(ch chan binancerest.OrderSheet) {
	t.orders = ch
}

func (t *Trader) UpdateRateLimit(rateLimit int) {
	t.rateLimit = rateLimit
}

func (t *Trader) UpdateCurrentRate(currentRate int) {
	t.currentRate = currentRate
}

func (t *Trader) checkRateLimit(weight int) {
	if t.rateLimit-t.currentRate < weight {
		// Wait until the start of next minute
		now := time.Now()
		nextMinute := now.Add(time.Minute).Truncate(time.Minute)
		time.Sleep(nextMinute.Sub(now))
		t.currentRate = 0 // Reset rate limit counter for new minute
	}
}
