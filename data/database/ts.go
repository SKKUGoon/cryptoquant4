package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

type TimeScale struct {
	db *sql.DB
}

type PremiumLog struct {
	Timestamp       time.Time
	Symbol          string
	Premium         float64
	PremiumEnterPos float64
	PremiumExitPos  float64
	KimchiPrice     float64
	AnchorPrice     float64
	CefiPrice       float64
	// Best Bid and Ask
	KimchiBestBid float64
	KimchiBestAsk float64
	CefiBestBid   float64
	CefiBestAsk   float64
}

func ConnectTS() (*TimeScale, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("TS_NAME"),
		os.Getenv("TS_HOST"),
		os.Getenv("TS_PORT"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &TimeScale{db: db}, nil
}

func (t *TimeScale) Close() error {
	return t.db.Close()
}

func (t *TimeScale) InsertPremiumLog(logs []PremiumLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Start a transaction
	tx, err := t.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	// Prepare the statement
	stmt, err := tx.Prepare(`
		INSERT INTO premium_logs (time, symbol, premium, kimchi_price, anchor_price, cefi_price, kimchi_best_bid, kimchi_best_ask, anchor_best_bid, anchor_best_ask, cefi_best_bid, cefi_best_ask)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	// Execute the statement for each log
	for _, log := range logs {
		_, err := stmt.Exec(
			log.Timestamp,
			log.Symbol,
			log.Premium,
			log.KimchiPrice,
			log.AnchorPrice,
			log.CefiPrice,
			log.KimchiBestBid,
			log.KimchiBestAsk,
			nil,
			nil,
			log.CefiBestBid,
			log.CefiBestAsk,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing statement: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
