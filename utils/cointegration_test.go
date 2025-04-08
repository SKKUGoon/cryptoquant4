package utils_test

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/joho/godotenv"
// )

// func TestCointegrationTest(t *testing.T) {
// 	if err := godotenv.Load("../../.env"); err != nil {
// 		t.Fatalf("Error loading .env file: %v", err)
// 	}

// 	db, err := data.ConnectDB()
// 	if err != nil {
// 		t.Fatalf("Failed to connect to database: %v", err)
// 	}
// 	defer db.Close()

// 	prices1, prices2, err := db.GetBacktestData()
// 	if err != nil {
// 		t.Fatalf("Failed to import Nimbus data: %v", err)
// 	}

// 	pvalue, err := calculation.CointegrationTest(prices1, prices2)
// 	if err != nil {
// 		t.Fatalf("Error calculating cointegration test: %v", err)
// 	}

// 	fmt.Println("Cointegration Test p-value:", pvalue)
// }
