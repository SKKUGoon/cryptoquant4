//go:build init
// +build init

package main

import (
	"context"
	"log"

	"cryptoquant.com/m/engine"
)

// Synchronize the redis database with the api information.
// Redis database is volatile, initialized at startup. Redis is included in docker-compose.yaml
func main() {
	ctx := context.Background()

	as := engine.NewAccountSource(ctx)
	if err := as.OnInit(); err != nil {
		log.Fatalf("failed to init account source: %v", err)
	}
}
