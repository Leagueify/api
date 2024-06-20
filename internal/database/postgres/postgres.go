package postgres

import (
	"database/sql"
	"fmt"

	"github.com/Leagueify/api/internal/config"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
)

type Postgres struct {
	DB *sql.DB
}

func init() {
	cfg := config.LoadConfig()
	database, err := Connect(cfg.DBConnStr)
	if err != nil {
		panic(fmt.Sprintf("Error initializing database '%s'", err))
	}
	db := Postgres{
		DB: database,
	}
	if err := db.InitializeDatabase(); err != nil {
		sentry.CaptureException(err)
	}
}

func Connect(DBConnStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DBConnStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (p Postgres) BeginTransaction() (*sql.Tx, error) {
	tx, err := p.DB.Begin()
	if err != nil {
		sentry.CaptureException(err)
		return nil, err
	}
	return tx, nil
}

func (p Postgres) InitializeDatabase() error {
	tx, err := p.DB.Begin()
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	defer tx.Rollback()

	// create accounts table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			phone TEXT NOT NULL UNIQUE,
			date_of_birth TEXT NOT NULL,
			registration_code TEXT NOT NULL,
			player_ids TEXT[] NOT NULL,
			coach BOOLEAN DEFAULT false,
			volunteer BOOLEAN DEFAULT false,
			apikey TEXT NOT NULL,
			is_active BOOLEAN DEFAULT false,
			is_admin BOOLEAN DEFAULT false
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// create leagues table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS leagues (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			sport_id INTEGER NOT NULL,
			master_admin TEXT NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// create players table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS players (
			id TEXT PRIMARY KEY,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			date_of_birth TEXT NOT NULL,
			position TEXT NOT NULL,
			team TEXT NOT NULL,
			division TEXT NOT NULL,
			is_registered BOOLEAN DEFAULT false
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// create positions table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS positions (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// create registrations table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS registrations (
			id TEXT PRIMARY KEY,
			player_ids TEXT[] NOT NULL,
			amount_due INTEGER NOT NULL,
			amount_paid INTEGER NOT NULL
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// create sports table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS sports (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// add sports to table
	sports := []string{
		"baseball", "basketball", "football", "hockey", "quidditch",
		"rugby", "soccer", "softball", "volleyball",
	}
	for _, sport := range sports {
		sportID := util.SignedToken(4)
		if _, err = tx.Exec(`
			INSERT INTO sports (id, name) VALUES ($1, $2)
		`, sportID[:len(sportID)-1], sport); err != nil {
			sentry.CaptureException(err)
			return err
		}
	}
	if err != nil {
		sentry.CaptureException(err)
		return err
	}

	// commit database initialization
	if err := tx.Commit(); err != nil {
		sentry.CaptureException(err)
		return err
	}
	return nil
}
