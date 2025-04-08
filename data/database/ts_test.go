package database_test

import (
	"testing"

	"cryptoquant.com/m/data/database"
	"github.com/joho/godotenv"
)

const TSENVLOC = "../../.env"

func TestConnectTS(t *testing.T) {
	if err := godotenv.Load(TSENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	ts, err := database.ConnectTS()
	if err != nil {
		t.Fatalf("Failed to connect to TimeScale: %v", err)
	}

	defer ts.Close()
}
