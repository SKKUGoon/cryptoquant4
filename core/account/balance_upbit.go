package account

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func (a *AccountSource) syncRedisFromUpbit() error {
	var balance float64
	var reserved float64
	var account upbitrest.Accounts
	var snapshot = make(map[string]float64)
	const urlBase = "https://api.upbit.com/v1/accounts"

	// Create authorization token
	nonce := uuid.NewString()
	claims := jwt.MapClaims{
		"access_key": a.upbitPubkey,
		"nonce":      nonce,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(a.upbitPrikey))
	if err != nil {
		return err
	}
	authToken := "Bearer " + signedToken

	// Send request

	req, err := http.NewRequest("GET", urlBase, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &account); err != nil {
		return err
	}

	// Save to redis

	for _, acc := range account {
		currency := acc.Currency

		if balance, err = strconv.ParseFloat(acc.Balance, 64); err != nil {
			return err
		}
		if reserved, err = strconv.ParseFloat(acc.Locked, 64); err != nil {
			return err
		}

		// Store individual currency balance
		if err := a.SetPosition("upbit", currency, balance); err != nil {
			return err
		}
		fmt.Printf("Account: Upbit %s balance: %f, reserved: %f\n", currency, balance, reserved)

		// Store in snapshot map
		snapshot[currency] = balance

		// Save KRW available to internal fund
		if currency == "KRW" {
			a.UpbitFund.AvailableFund = balance
			if err := a.setRedisAvailableFund("upbit", balance-reserved); err != nil {
				return err
			}
			if err := a.syncRedisReservedFund("upbit", reserved); err != nil {
				return err
			}
		}
	}

	// Save snapshot to redis
	if err := a.SetWalletSnapshot("upbit", snapshot); err != nil {
		return err
	}

	return nil
}
