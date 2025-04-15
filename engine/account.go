package engine

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
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
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
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
	as.SyncAvailableFundUpbit()
	as.SyncAvailableFundBinance()
	as.SyncReservedFundUpbit()
	as.SyncReservedFundBinance()
	as.SyncWalletSnapshotUpbit()
	as.SyncWalletSnapshotBinance()

	return as
}

func (a *AccountSource) OnInit() error {
	// Sync upbit
	if err := a.upbitFundSync(); err != nil {
		return err
	}
	log.Println("Account: Upbit -> Redis. Fund synced")

	// Sync binance
	if err := a.binanceFundSync(); err != nil {
		return err
	}
	log.Println("Account: Binance -> Redis. Fund synced")

	return nil
}

func (a *AccountSource) OnUpdate() error {
	log.Println("Account: Redis -> Upbit. Fund synced")
	log.Println("Account: Redis -> Binance. Fund synced")

	return nil
}

func (a *AccountSource) upbitFundSync() error {
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
	const urlBase = "https://api.upbit.com/v1/accounts"

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

	var accounts upbitrest.Accounts
	if err := json.Unmarshal(body, &accounts); err != nil {
		return err
	}

	// Save to redis
	snapshot := make(map[string]float64)
	for _, acc := range accounts {
		currency := acc.Currency
		var balance, reserved float64
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
			if err := a.SetAvailableFund("upbit", balance-reserved); err != nil {
				return err
			}
			if err := a.SetReservedFund("upbit", reserved); err != nil {
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

func (a *AccountSource) binanceFundSync() error {
	const urlBase = "https://fapi.binance.com/fapi/v3/account"

	params := map[string]int64{
		"timestamp": time.Now().UnixMilli(),
	}

	// Build query string
	query := url.Values{}
	for k, v := range params {
		query.Add(k, strconv.FormatInt(v, 10))
	}
	queryString := query.Encode()
	if queryString == "" {
		return fmt.Errorf("empty query string")
	}

	// Build HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(a.binancePrikey))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))

	fullQuery := queryString + "&signature=" + signature

	// Send request
	fullURL := urlBase + "?" + fullQuery
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-MBX-APIKEY", a.binancePubkey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var account binancerest.AccountInfo
	if err := json.Unmarshal(body, &account); err != nil {
		return err
	}

	// Save to redis - Assets (Quoting)
	snapshot := make(map[string]float64)
	for _, asset := range account.Assets {
		var total, available float64
		if total, err = strconv.ParseFloat(asset.WalletBalance, 64); err != nil {
			return err
		}
		if available, err = strconv.ParseFloat(asset.AvailableBalance, 64); err != nil {
			return err
		}

		// Store individual currency balance
		if err := a.SetPosition("binance", asset.Asset, total); err != nil {
			return err
		}
		fmt.Printf("Account: Binance %s balance: %f, available: %f\n", asset.Asset, total, available)

		snapshot[asset.Asset] = available

		if asset.Asset == "USDT" {
			a.BinanceFund.AvailableFund = available
			if err := a.SetAvailableFund("binance", available); err != nil {
				return err
			}
			if err := a.SetReservedFund("binance", total-available); err != nil {
				return err
			}
		}
	}

	// Save to redis - Positions (Quoting)
	for _, position := range account.Positions {
		var total, available float64
		if total, err = strconv.ParseFloat(position.PositionAmt, 64); err != nil {
			return err
		}
		available = total // TODO: For now, it's market order only.

		if err := a.SetPosition("binance", position.Symbol, total); err != nil {
			return err
		}

		snapshot[position.Symbol] = available
	}

	if err := a.SetWalletSnapshot("binance", snapshot); err != nil {
		return err
	}
	return nil
}

// ReservedFund: Total amount of money (USDT/KRW) that is reserved by containers.
func (a *AccountSource) keyReservedFund(exchange string) string {
	return fmt.Sprintf("reserved_fund:%s", exchange)
}

func (a *AccountSource) SetReservedFund(exchange string, amount float64) error {
	key := a.keyReservedFund(exchange)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

func (a *AccountSource) getReservedFund(exchange string) (float64, error) {
	key := a.keyReservedFund(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (a *AccountSource) SyncReservedFundUpbit() error {
	reserve, err := a.getReservedFund("upbit")
	if err != nil {
		return err
	}

	a.UpbitFund.ReservedFund = reserve
	return nil
}

func (a *AccountSource) SyncReservedFundBinance() error {
	reserve, err := a.getReservedFund("binance")
	if err != nil {
		return err
	}

	a.BinanceFund.ReservedFund = reserve
	return nil
}

// AvailableFund: Total amount of money (USDT/KRW) that is free to be reserved by containers.
func (a *AccountSource) keyAvailableFund(exchange string) string {
	return fmt.Sprintf("available_fund:%s", exchange)
}

func (a *AccountSource) SetAvailableFund(exchange string, amount float64) error {
	key := a.keyAvailableFund(exchange)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

func (a *AccountSource) getAvailableFund(exchange string) (float64, error) {
	key := a.keyAvailableFund(exchange)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (a *AccountSource) SyncAvailableFundUpbit() error {
	fund, err := a.getAvailableFund("upbit")
	if err != nil {
		return err
	}

	a.UpbitFund.AvailableFund = fund
	return nil
}

func (a *AccountSource) SyncAvailableFundBinance() error {
	fund, err := a.getAvailableFund("binance")
	if err != nil {
		return err
	}

	a.BinanceFund.AvailableFund = fund
	return nil
}

// Position: Total amount of money (USDT/KRW) that is reserved by containers.
func (a *AccountSource) keyPosition(exchange, currency string) string {
	return fmt.Sprintf("wallet:%s:%s", exchange, currency)
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

// WalletSnapshot: Snapshot of the wallet at startup.
func (a *AccountSource) keyWalletSnapshot(exchange string) string {
	return fmt.Sprintf("wallet_snapshot:%s", exchange)
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
