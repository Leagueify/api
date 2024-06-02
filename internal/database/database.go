package database

import (
	"database/sql"

	"github.com/Leagueify/api/internal/config"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
)

func init() {
	initAccounts()
	initLeagues()
	initPositions()
	initSports()
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
			apikey TEXT,
			is_active BOOLEAN DEFAULT false,
			is_admin BOOLEAN DEFAULT false
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
	}
}

func initLeagues() {
	db, err := Connect()
	if err != nil {
		sentry.CaptureException(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS leagues (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			sport_id INTEGER NOT NULL,
			master_admin TEXT NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
	}
}

func initPositions() {
	db, err := Connect()
	if err != nil {
		sentry.CaptureException(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS positions (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
	}
}

func initSports() {
	sports := []string{
		"baseball", "basketball", "football", "hockey", "quidditch",
		"rugby", "soccer", "softball", "volleyball",
	}
	db, err := Connect()
	if err != nil {
		sentry.CaptureException(err)
	}
	defer db.Close()
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sports (
			id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
	}
	for _, sport := range sports {
		if _, err = db.Exec(`
			INSERT INTO sports (id, name) VALUES (DEFAULT, $1)
		`, sport); err != nil {
			sentry.CaptureException(err)
		}
	}
	if err != nil {
		sentry.CaptureException(err)
	}
}
