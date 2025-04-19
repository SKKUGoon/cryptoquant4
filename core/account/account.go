package account

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

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

	// API Keys
	binancePubkey string
	binancePrikey string

	upbitPubkey string
	upbitPrikey string

	BinanceFund Fund
	UpbitFund   Fund

	upbitWalletSnapshot   map[string]float64 // KRW, XRP, USDT, ... at startup.
	binanceWalletSnapshot map[string]float64 // KRW, XRP, USDT, ... at startup.
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

	binancePubkey := os.Getenv("BINANCE_API_KEY")
	binancePrikey := os.Getenv("BINANCE_SECRET_KEY")

	upbitPubkey := os.Getenv("UPBIT_API_KEY")
	upbitPrikey := os.Getenv("UPBIT_SECRET_KEY")

	return &AccountSource{
		Redis: rdb,
		ctx:   ctx,

		binancePubkey: binancePubkey,
		binancePrikey: binancePrikey,

		upbitPubkey: upbitPubkey,
		upbitPrikey: upbitPrikey,
	}
}

func NewAccountSourceSync(ctx context.Context) *AccountSource {
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

	binancePubkey := os.Getenv("BINANCE_API_KEY")
	binancePrikey := os.Getenv("BINANCE_SECRET_KEY")

	upbitPubkey := os.Getenv("UPBIT_API_KEY")
	upbitPrikey := os.Getenv("UPBIT_SECRET_KEY")

	as := &AccountSource{
		Redis: rdb,
		ctx:   ctx,

		binancePubkey: binancePubkey,
		binancePrikey: binancePrikey,

		upbitPubkey: upbitPubkey,
		upbitPrikey: upbitPrikey,
	}

	if err := as.SyncAvailableFundUpbit(); err != nil {
		log.Fatalf("Failed to sync Upbit available fund: %v", err)
	}
	if err := as.SyncAvailableFundBinance(); err != nil {
		log.Fatalf("Failed to sync Binance available fund: %v", err)
	}
	if err := as.SyncReservedFundUpbit(); err != nil {
		log.Fatalf("Failed to sync Upbit reserved fund: %v", err)
	}
	if err := as.SyncReservedFundBinance(); err != nil {
		log.Fatalf("Failed to sync Binance reserved fund: %v", err)
	}
	if err := as.SyncWalletSnapshotUpbit(); err != nil {
		log.Fatalf("Failed to sync Upbit wallet snapshot: %v", err)
	}
	if err := as.SyncWalletSnapshotBinance(); err != nil {
		log.Fatalf("Failed to sync Binance wallet snapshot: %v", err)
	}

	return as
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

func (a *AccountSource) Update() error {
	// Save to redis
	log.Println("Account: Upbit -> Redis. Fund synced")
	a.syncRedisFromUpbit()
	log.Println("Account: Binance -> Redis. Fund synced")
	a.syncRedisFromBinance()

	// Update internal fund
	if err := a.SyncAvailableFundUpbit(); err != nil {
		return fmt.Errorf("failed to sync Upbit available fund: %v", err)
	}
	if err := a.SyncAvailableFundBinance(); err != nil {
		return fmt.Errorf("failed to sync Binance available fund: %v", err)
	}
	if err := a.SyncReservedFundUpbit(); err != nil {
		return fmt.Errorf("failed to sync Upbit reserved fund: %v", err)
	}
	if err := a.SyncReservedFundBinance(); err != nil {
		return fmt.Errorf("failed to sync Binance reserved fund: %v", err)
	}
	if err := a.SyncWalletSnapshotUpbit(); err != nil {
		return fmt.Errorf("failed to sync Upbit wallet snapshot: %v", err)
	}
	if err := a.SyncWalletSnapshotBinance(); err != nil {
		return fmt.Errorf("failed to sync Binance wallet snapshot: %v", err)
	}

	return nil
}

// Exchange -> Redis
func (a *AccountSource) syncRedisReservedFund(exchange string, amount float64) error {
	key := a.keyReservedFund(exchange)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

// Redis -> Go
func (a *AccountSource) syncGoReservedFund(exchange string) (float64, error) {
	key := a.keyReservedFund(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

// Redis -> UpbitFund.ReservedFund
func (a *AccountSource) SyncReservedFundUpbit() error {
	reserve, err := a.syncGoReservedFund("upbit")
	if err != nil {
		return err
	}

	a.UpbitFund.ReservedFund = reserve
	return nil
}

// Redis -> BinanceFund.ReservedFund
func (a *AccountSource) SyncReservedFundBinance() error {
	reserve, err := a.syncGoReservedFund("binance")
	if err != nil {
		return err
	}

	a.BinanceFund.ReservedFund = reserve
	return nil
}

// Exchange -> Redis
func (a *AccountSource) setRedisAvailableFund(exchange string, amount float64) error {
	key := a.keyAvailableFund(exchange)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

// Redis -> Go
func (a *AccountSource) syncGoAvailableFund(exchange string) (float64, error) {
	key := a.keyAvailableFund(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

// Redis -> UpbitFund.AvailableFund
func (a *AccountSource) SyncAvailableFundUpbit() error {
	fund, err := a.syncGoAvailableFund("upbit")
	if err != nil {
		return err
	}

	a.UpbitFund.AvailableFund = fund
	return nil
}

// Redis -> BinanceFund.AvailableFund
func (a *AccountSource) SyncAvailableFundBinance() error {
	fund, err := a.syncGoAvailableFund("binance")
	if err != nil {
		return err
	}

	a.BinanceFund.AvailableFund = fund
	return nil
}

func (a *AccountSource) SetPosition(exchange, currency string, amount float64) error {
	key := a.keyPosition(exchange, currency)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

func (a *AccountSource) GetPosition(exchange, currency string) (float64, error) {
	key := a.keyPosition(exchange, currency)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (a *AccountSource) SetWalletSnapshot(exchange string, snapshot map[string]float64) error {
	key := a.keyWalletSnapshot(exchange)
	json, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return a.Redis.Set(a.ctx, key, json, 0).Err()
}

func (a *AccountSource) GetWalletSnapshot(exchange string) (map[string]float64, error) {
	key := a.keyWalletSnapshot(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var snapshot map[string]float64
	if err := json.Unmarshal([]byte(val), &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (a *AccountSource) SyncWalletSnapshotUpbit() error {
	snapshot, err := a.GetWalletSnapshot("upbit")
	if err != nil {
		return err
	}

	a.upbitWalletSnapshot = snapshot
	return nil
}

func (a *AccountSource) SyncWalletSnapshotBinance() error {
	snapshot, err := a.GetWalletSnapshot("binance")
	if err != nil {
		return err
	}

	a.binanceWalletSnapshot = snapshot
	return nil
}
