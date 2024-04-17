package database

import (
	"database/sql"

	"github.com/Leagueify/api/internal/config"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
)

func init() {
	initAccounts()
}

func Connect() (*sql.DB, error) {
	cfg := config.LoadConfig()
	db, err := sql.Open("postgres", cfg.DBConnStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func initAccounts() {
	db, err := Connect()
	if err != nil {
		sentry.CaptureException(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			date_of_birth TEXT NOT NULL,
			coach BOOLEAN DEFAULT false,
			volunteer BOOLEAN DEFAULT false,
			token TEXT,
			is_active BOOLEAN DEFAULT false
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
	}
}
