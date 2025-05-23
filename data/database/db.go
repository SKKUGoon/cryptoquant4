package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func ConnectDB() (*Database, error) {
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%s",
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("PG_NAME"),
		os.Getenv("PG_HOST"),
		os.Getenv("PG_PORT"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) GetTradeMetadata(key string, defaultValue any) (any, error) {
	query := `SELECT value, value_type 
		FROM cryptoquant.trading_metadata
		WHERE key = $1`

	rows, err := d.db.Query(query, key)
	if err != nil {
		return nil, fmt.Errorf("error querying data: %v", err)
	}
	defer rows.Close()

	var value string
	var valueType string

	for rows.Next() {
		if err := rows.Scan(&value, &valueType); err != nil {
			log.Printf("error scanning row: %v\n", err)
			continue
		}

		switch valueType {
		case "int":
			return strconv.Atoi(value)
		case "float64":
			return strconv.ParseFloat(value, 64)
		case "bool":
			return strconv.ParseBool(value)
		case "string":
			return value, nil
		case "[]string":
			var out []string
			err := json.Unmarshal([]byte(value), &out)
			return out, err
		case "[]int":
			var out []int
			err := json.Unmarshal([]byte(value), &out)
			return out, err
		case "[]float64":
			var out []float64
			err := json.Unmarshal([]byte(value), &out)
			return out, err
		case "[]bool":
			var out []bool
			err := json.Unmarshal([]byte(value), &out)
			return out, err
		default:
			return nil, fmt.Errorf("unsupported value_type: %s", valueType)
		}
	}
	log.Printf("key not found: %s. Using default value: %v.", key, defaultValue)
	return defaultValue, nil
}

// ExportTradeLog exports trade log data
func (d *Database) InsertStrategyKimchiOrderLog(logs []KimchiOrderLog) error {
	if len(logs) == 0 {
		return nil
	}

	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO cryptoquant.strategy_kimchi_order_logs (pair_id, order_time, execution_time, pair_side, exchange, side, order_price, executed_price, anchor_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	for _, log := range logs {
		_, err := stmt.Exec(
			log.PairID,
			log.OrderTime,
			log.ExecutionTime,
			log.PairSide,
			log.Exchange,
			log.Side,
			log.OrderPrice,
			log.ExecutedPrice,
			log.AnchorPrice,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing statement: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}
