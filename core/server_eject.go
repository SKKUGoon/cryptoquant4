package core

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/shopspring/decimal"
)

// In case of emergency
// 1. Stop all trading
// 2. Close all positions
// 3. Log the emergency

func (s *Server) KimchiPremiumEject() {
	s.Account.Mu.Lock()
	defer s.Account.Mu.Unlock()

	log.Println("[EMERGENCY] Ejecting trading server")

	// Update redis and wallet information
	err := s.Account.UpdateRedis()
	if err != nil {
		log.Printf("Failed to update redis: %v", err)
	}
	err = s.Account.Sync()
	if err != nil {
		log.Printf("Failed to sync: %v", err)
	}

	upbitWallet := s.Account.GetUpbitWalletSnapshot()
	binanceWallet := s.Account.GetBinanceWalletSnapshot()

	// Close all positions - upbit
	walletCleared := true
	note := []string{}
	for asset, amount := range upbitWallet {
		if asset == "KRW" {
			continue
		}
		if amount > 0 {
			log.Printf("Closing position: %s %f", asset, amount)

			orderSheet := upbitrest.OrderSheet{
				Symbol:  "KRW-" + asset,
				Side:    "ask",
				Volume:  strconv.FormatFloat(amount, 'f', -1, 64),
				OrdType: "market",
			}

			_, err := s.UpbitTrader.SendOrder(orderSheet)
			if err != nil {
				log.Printf("Failed to send close %v: %v", asset, err)
				walletCleared = false
				note = append(note, fmt.Sprintf("Failed to send close %v: %v", asset, err))
			}
		}
	}

	// Close all positions - binance
	for asset, amount := range binanceWallet {
		if asset == "USDT" {
			continue
		}
		if amount > 0 {
			log.Printf("Closing position: %s %f", asset, amount)

			orderSheet := &binancerest.OrderSheet{
				Symbol:     asset,
				Side:       "BUY",
				ReduceOnly: "true",
				Type:       "MARKET",
				Quantity:   decimal.NewFromFloat(amount * 1.0005), // Buffer of 0.05% ensure all position closed (Reduce only)
			}

			_, err := s.BinanceTrader.SendSingleOrder(orderSheet)
			if err != nil {
				log.Printf("Failed to send close %v: %v", asset, err)
				walletCleared = false
				note = append(note, fmt.Sprintf("Failed to send close %v: %v", asset, err))
			}
		}
	}

	// Log the emergency
	s.TimeScale.LogEmergencyShutdown(walletCleared, strings.Join(note, "\n"))

	// Crash the trading program
	os.Exit(1)
}
