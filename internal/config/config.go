package config

import (
	"os"
)

type Configuration struct {
	DBConnStr string
	SentryDsn string
}

func (c *Configuration) LoadFromEnv() {
	// Database Connection String
	if dbConnStr := os.Getenv("DB_CONN_STR"); dbConnStr != "" {
		c.DBConnStr = dbConnStr
	}
	// Sentry DSN
	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		c.SentryDsn = sentryDSN
	}
}
