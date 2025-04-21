package core_test

import (
	"context"
	"testing"

	"cryptoquant.com/m/core"
	"github.com/joho/godotenv"
)

const ENV_PATH = "../.env.local"

func TestKimchiPremiumEject(t *testing.T) {
	if err := godotenv.Load(ENV_PATH); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()
	server, err := core.NewTraderServer(ctx)
	if err != nil {
		t.Fatalf("Failed to create trader server: %v", err)
	}

	server.KimchiPremiumEject()
}
