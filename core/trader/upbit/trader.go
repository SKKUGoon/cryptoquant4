package upbittrade

import (
	"log"
	"os"
	"time"

	upbitrest "cryptoquant.com/m/internal/upbit/rest"
)

type Trader struct {
	pubkey string // Upbit API Key
	prikey string // Upbit API Secret

	// Order rate limit
	rateLimit   int
	currentRate int

	// Channels to receive orders from strategy
	orders     chan upbitrest.OrderSheet
	inPosition bool
}

func NewTrader() *Trader {
	pubkey := os.Getenv("UPBIT_API_KEY")
	prikey := os.Getenv("UPBIT_SECRET_KEY")

	if pubkey == "" || prikey == "" {
		log.Fatalf("[trader] UPBIT_API_KEY or UPBIT_SECRET_KEY is not set")
	}

	return &Trader{
		pubkey:     pubkey,
		prikey:     prikey,
		inPosition: false,
	}
}

func (t *Trader) SetOrderChannel(ch chan upbitrest.OrderSheet) {
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
	}
}

func (t *Trader) Run() {

}
