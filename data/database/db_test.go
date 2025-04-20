package database_test

import (
	"testing"

	"cryptoquant.com/m/data/database"
	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
)

const ENV_LOC = "../../.env.local"

func TestGetTradeMetadataString(t *testing.T) {
	if err := godotenv.Load(ENV_LOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_STRING", nil)
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, "BTCUSDT")
}

func TestGetTradeMetadataInt(t *testing.T) {
	if err := godotenv.Load(ENV_LOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_INT", nil)
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, 2)
}

func TestGetTradeMetadataFloat(t *testing.T) {
	if err := godotenv.Load(ENV_LOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_FLOAT64", nil)
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, 0.001)
}

func TestGetTradeMetadataBool(t *testing.T) {
	if err := godotenv.Load(ENV_LOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_BOOL", nil)
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, true)
}

func TestGetTradeMetadataStringArray(t *testing.T) {
	if err := godotenv.Load(ENV_LOC); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	db, err := database.ConnectDB()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	metadata, err := db.GetTradeMetadata("TEST_STRING_ARRAY", nil)
	if err != nil {
		t.Fatalf("Failed to get trade metadata: %v", err)
	}

	assert.Equal(t, metadata, []string{"a", "bc", "def"})
}
