package config

import (
	"os"
)

type configuration struct {
	DBConnStr string
	SentryDsn string
}

func LoadConfig() *configuration {
	config := &configuration{}
	config.loadFromEnv()
	return config
}

func (c *configuration) loadFromEnv() {
	// Database Connection String
	if dbConnStr := os.Getenv("DB_CONN_STR"); dbConnStr != "" {
		c.DBConnStr = dbConnStr
	}
	// Sentry DSN
	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		c.SentryDsn = sentryDSN
	}
}
