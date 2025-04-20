package account

import "encoding/json"

/* Wallet snapshot sync */

// UpdateRedisWalletSnapshot updates Redis with wallet snapshot
func (a *AccountSource) UpdateRedisWalletSnapshot(exchange string, snapshot map[string]float64) error {
	key := a.keyWalletSnapshot(exchange)
	json, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return a.Redis.Set(a.ctx, key, json, 0).Err()
}

// SyncFromRedisWalletSnapshot syncs wallet snapshot from Redis to Go struct
func (a *AccountSource) SyncFromRedisWalletSnapshot(exchange string) (map[string]float64, error) {
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

// SyncWalletSnapshotUpbit syncs Upbit wallet snapshot from Redis to Go struct
func (a *AccountSource) SyncWalletSnapshotUpbit() error {
	snapshot, err := a.SyncFromRedisWalletSnapshot("upbit")
	if err != nil {
		return err
	}

	a.upbitWalletSnapshot = snapshot
	return nil
}

func (a *AccountSource) GetUpbitWalletSnapshot() map[string]float64 {
	return a.upbitWalletSnapshot
}

// SyncWalletSnapshotBinance syncs Binance wallet snapshot from Redis to Go struct
func (a *AccountSource) SyncWalletSnapshotBinance() error {
	snapshot, err := a.SyncFromRedisWalletSnapshot("binance")
	if err != nil {
		return err
	}

	a.binanceWalletSnapshot = snapshot
	return nil
}

func (a *AccountSource) GetBinanceWalletSnapshot() map[string]float64 {
	return a.binanceWalletSnapshot
}
