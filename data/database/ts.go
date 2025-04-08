package database

import (
	"database/sql"
	"fmt"
	"os"
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

func (t *TimeScale) InsertPremiumLog() {}
