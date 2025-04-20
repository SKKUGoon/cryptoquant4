package account

import "strconv"

/* Position sync */

// UpdateRedisPosition updates Redis with position information
func (a *AccountSource) UpdateRedisPosition(exchange, currency string, amount float64) error {
	key := a.keyPosition(exchange, currency)
	return a.Redis.Set(a.ctx, key, amount, 0).Err()
}

// SyncFromRedisPosition syncs position from Redis to Go struct
func (a *AccountSource) SyncFromRedisPosition(exchange, currency string) (float64, error) {
	key := a.keyPosition(exchange, currency)
	val, err := a.Redis.Get(a.ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}
