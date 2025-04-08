package database_test

import (
	"fmt"
	"testing"

	"cryptoquant.com/m/data/database"
	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
)

const PGENVLOC = "../../.env"

func TestGetBacktestData(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	prices1, prices2, err := db.GetBacktestData()
	if err != nil {
		t.Fatalf("Failed to import Nimbus data: %v", err)
	}

	fmt.Println(prices1)
	fmt.Println(prices2)
}

func TestGetTradeMetadataString(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_STRING")
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, "BTCUSDT")
}

func TestGetTradeMetadataInt(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_INT")
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, 2)
}

func TestGetTradeMetadataFloat(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_FLOAT64")
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, 0.001)
}

func TestGetTradeMetadataBool(t *testing.T) {
	if err := godotenv.Load(PGENVLOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_BOOL")
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, true)
}

func TestGetTradeMetadataStringArray(t *testing.T) {
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_STRING_ARRAY")
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, []string{"a", "bc", "def"})
}
