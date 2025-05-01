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

// func (t *TimeScale) InsertPremiumLog(logs []PremiumLog) error {
// 	if len(logs) == 0 {
// 		return nil
// 	}

// 	// Start a transaction
// 	tx, err := t.db.Begin()
// 	if err != nil {
// 		return fmt.Errorf("error starting transaction: %v", err)
// 	}

// 	// Prepare the statement
// 	stmt, err := tx.Prepare(`
// 		INSERT INTO cryptoquant.premium_logs (time, symbol, anchor_price, kimchi_best_bid, kimchi_best_ask, cefi_best_bid, cefi_best_ask)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 	`)
// 	if err != nil {
// 		tx.Rollback()
// 		return fmt.Errorf("error preparing statement: %v", err)
// 	}
// 	defer stmt.Close()

// 	// Execute the statement for each log
// 	for _, log := range logs {
// 		_, err := stmt.Exec(
// 			log.Timestamp,
// 			log.Symbol,
// 			log.AnchorPrice,
// 			log.KimchiBestBid,
// 			log.KimchiBestAsk,
// 			log.CefiBestBid,
// 			log.CefiBestAsk,
// 		)
// 		if err != nil {
// 			tx.Rollback()
// 			return fmt.Errorf("error executing statement: %v", err)
// 		}
// 	}

// 	// Commit the transaction
// 	if err := tx.Commit(); err != nil {
// 		return fmt.Errorf("error committing transaction: %v", err)
// 	}

// 	return nil
// }

func (t *TimeScale) InsertAccountSnapshot(logs []AccountSnapshot) error {
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
		INSERT INTO cryptoquant.account_snapshots (time, exchange, available, reserved, total, wallet_balance_usdt, wallet_balance_krw)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
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
			log.Exchange,
			log.Available,
			log.Reserved,
			log.Total,
			log.WalletBalanceUSDT,
			log.WalletBalanceKRW,
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

func (t *TimeScale) LogEmergencyShutdown(positionCleared bool, note string) error {
	stmt, err := t.db.Prepare(`
		INSERT INTO cryptoquant.emergency_shutdown_logs (time, position_clear_success, note)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), positionCleared, note)
	if err != nil {
		return fmt.Errorf("error executing statement: %v", err)
	}

	return nil
}
