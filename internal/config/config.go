package config

import (
	"os"
)

type configuration struct {
	DBConnStr string
	JWTSecret string
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
	// JWT Secret
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.JWTSecret = jwtSecret
	}
	// Sentry DSN
	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		c.SentryDsn = sentryDSN
	}
}
