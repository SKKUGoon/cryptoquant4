package account

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
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
	// Mutex for thread-safe operations
	Mu sync.Mutex

	Redis *redis.Client
	ctx   context.Context

	// API Keys
	binancePubkey string
	binancePrikey string

	upbitPubkey string
	upbitPrikey string

	// Wallet
	upbitWalletSnapshot   map[string]float64 // KRW, XRP, USDT, ... at startup.
	binanceWalletSnapshot map[string]float64 // KRW, XRP, USDT, ... at startup.

	UpbitPrincipalCurrency   string // KRW
	BinancePrincipalCurrency string // USDT
}

type Fund struct {
	// AvailableFund: Total amount of money (USDT/KRW) that is free to be reserved by containers.
	// This money is shared by all containers.
	AvailableFund float64

	// Amount already committed by a specific container for upcoming or in-progress trades.
	// This money is per container.
	ReservedFund float64
}

func NewAccountSource(ctx context.Context) *AccountSource {
	redisAddr := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisAddr, redisPort),
		Password: redisPassword,
		DB:       0,
	})
	// Check Redis connectivity
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Redis connected: %s", pong)

	// Environment variables
	binancePubkey := os.Getenv("BINANCE_API_KEY")
	binancePrikey := os.Getenv("BINANCE_SECRET_KEY")

	upbitPubkey := os.Getenv("UPBIT_API_KEY")
	upbitPrikey := os.Getenv("UPBIT_SECRET_KEY")

	upbitPrincipalCurrency := os.Getenv("UPBIT_PRINCIPAL_CURRENCY")
	binancePrincipalCurrency := os.Getenv("BINANCE_PRINCIPAL_CURRENCY")

	return &AccountSource{
		Redis: rdb,
		ctx:   ctx,

		binancePubkey: binancePubkey,
		binancePrikey: binancePrikey,

		upbitPubkey: upbitPubkey,
		upbitPrikey: upbitPrikey,

		UpbitPrincipalCurrency:   upbitPrincipalCurrency,
		BinancePrincipalCurrency: binancePrincipalCurrency,
	}
}

func (a *AccountSource) SetPrincipalCurrency(exchange, currency string) {
	switch exchange {
	case "upbit":
		a.UpbitPrincipalCurrency = currency
	case "binance":
		a.BinancePrincipalCurrency = currency
	default:
		log.Fatalf("Invalid exchange: %s", exchange)
	}
}

func (a *AccountSource) OnInit() error {
	// Sync upbit
	if err := a.syncRedisFromUpbit(); err != nil {
		return err
	}
	log.Println("Account: Upbit -> Redis. Fund synced")

	// Sync binance
	if err := a.syncRedisFromBinance(); err != nil {
		return err
	}
	log.Println("Account: Binance -> Redis. Fund synced")

	return nil
}

func (a *AccountSource) UpdateRedis() error {
	// Save to redis
	log.Println("Update Redis from Upbit")
	err := a.syncRedisFromUpbit()
	if err != nil {
		return fmt.Errorf("failed to sync Upbit: %v", err)
	}
	log.Println("Update Redis from Binance")
	err = a.syncRedisFromBinance()
	if err != nil {
		return fmt.Errorf("failed to sync Binance: %v", err)
	}

	// Update internal fund
	if err := a.SyncWalletSnapshotUpbit(); err != nil {
		return fmt.Errorf("failed to sync Upbit wallet snapshot: %v", err)
	}
	if err := a.SyncWalletSnapshotBinance(); err != nil {
		return fmt.Errorf("failed to sync Binance wallet snapshot: %v", err)
	}

	return nil
}

func (a *AccountSource) Sync() error {
	// Wallet sync
	if err := a.SyncWalletSnapshotUpbit(); err != nil {
		return fmt.Errorf("failed to sync Upbit wallet snapshot: %v", err)
	}
	if err := a.SyncWalletSnapshotBinance(); err != nil {
		return fmt.Errorf("failed to sync Binance wallet snapshot: %v", err)
	}
	return nil
}

func (a *AccountSource) Run() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			a.UpdateRedis()
		}
	}()
}
