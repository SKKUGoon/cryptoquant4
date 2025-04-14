//go:build init
// +build init

package main

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// Synchronize the redis database with the api information.
// Redis database is volatile, initialized at startup. Redis is included in docker-compose.yaml
func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	// Check Redis connectivity
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Redis connected: %s", pong)
}
