package binanceaccount

import (
	"log"
	"strconv"
	"sync"
	"time"

	binancews "cryptoquant.com/m/internal/binance/ws"
	binanceuser "cryptoquant.com/m/streams/binance/user"
)

type Account struct {
	AvailableFund float64
	QuotingAsset  string

	listenKey       string
	accountUpdateCh chan binancews.AccountUpdateEvent

	mu sync.Mutex
}

func NewAccount() *Account {
	listenKey, err := binanceuser.CreateListenKey()
	if err != nil {
		log.Fatalf("Failed to create listen key: %v", err)
	}

	return &Account{
		listenKey: listenKey,
	}
}

func (a *Account) SetAccountUpdateCh(ch chan binancews.AccountUpdateEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.accountUpdateCh = ch
}

func (a *Account) SetQuotingAsset(asset string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.QuotingAsset = asset
}

func (a *Account) GetAvailableFund() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.AvailableFund
}

func (a *Account) GetListenKey() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.listenKey
}

func (a *Account) Status() {
	a.mu.Lock()
	defer a.mu.Unlock()

	log.Printf("[account] status: %v", a.AvailableFund)
}

func (a *Account) Run(done chan struct{}) {
	// Keep the listen key alive
	renewTicker := time.NewTicker(time.Minute * 55)
	statusTicker := time.NewTicker(time.Second * 10)
	defer renewTicker.Stop()
	defer statusTicker.Stop()

	for {
		select {
		case <-statusTicker.C:
			a.Status()
		case <-renewTicker.C:
			// Send `PUT` request to keep the listen key alive
			binanceuser.KeepAliveListenKey(a.listenKey)
		case update := <-a.accountUpdateCh:
			for _, balance := range update.AccountUpdateData.Balances {
				if balance.Asset == a.QuotingAsset {
					balanceFloat, err := strconv.ParseFloat(balance.WalletBalance, 64)
					if err != nil {
						log.Printf("Failed to parse balance: %v", err)
						continue
					}
					a.AvailableFund = balanceFloat
				}
			}
		case <-done:
			// Send `DELETE` request to close the listen key
			binanceuser.CloseListenKey(a.listenKey)
			return
		}
	}
}
