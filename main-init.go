//go:build init && !trader && !server
// +build init,!trader,!server

package main

import (
	"context"
	"log"

	account "cryptoquant.com/m/core/account"
)

// Synchronize the redis database with the api information.
// Redis database is volatile, initialized at startup. Redis is included in docker-compose.yaml
func main() {
	ctx := context.Background()

	as := account.NewAccountSource(ctx)
	if err := as.OnInit(); err != nil {
		log.Fatalf("failed to init account source: %v", err)
	}
}
