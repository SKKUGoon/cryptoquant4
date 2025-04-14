package engine

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Account should be single source of truth.
// Account struct will be communicating with redis cache
// If there's an update
// 1. Save the available fund
// 2. Save the reserved fund
// 3. Save each position and its exchange information
// 4. Save the premium information

type AccountSource struct {
	Redis *redis.Client
	ctx   context.Context

	// AvailableFund: Total amount of money (USDT/KRW) that is free to be reserved by containers.
	// This money is shared by all containers.
	AvailableFund Fund
	// Amount already committed by a specific container for upcoming or in-progress trades.
	// This money is per container.
	ReservedFund Fund

	Position       map[string]string
	WalletSnapshot map[string]float64 // KRW, XRP, USDT, ... at startup.
}

type Fund struct {
	USDT float64
	KRW  float64
}

func (a *AccountSource) keyReservedFund(exchange string) string {
	return fmt.Sprintf("reserved_fund:%s", exchange)
}

func (a *AccountSource) SetAvailableFund(exchange string, amount float64) error {
	key := fmt.Sprintf("available_fund:%s", exchange)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

func (a *AccountSource) GetAvailableFund(exchange string) (float64, error) {
	key := fmt.Sprintf("available_fund:%s", exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (a *AccountSource) SetReservedFund(exchange string, amount float64, ttl time.Duration) error {
	key := a.keyReservedFund(exchange)
	return a.Redis.Set(a.ctx, key, amount, ttl).Err()
}

func (a *AccountSource) GetReservedFund(exchange string) (float64, error) {
	key := a.keyReservedFund(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}
