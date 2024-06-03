package config

import (
	"os"
	"strconv"
	"strings"
)

type configuration struct {
	DB        string
	DBConnStr string
	Sentry    bool
	SentryDSN string
	SentryTSR float64
}

func LoadConfig() *configuration {
	config := &configuration{}
	config.setDefaults()
	config.loadFromEnv()
	return config
}

func (c *configuration) loadFromEnv() {
	// Database Configuration
	// Database Service
	if db := os.Getenv("DATABASE"); db != "" {
		c.DB = strings.TrimSpace(db)
	}
	// Database Connection String
	if dbConnStr := os.Getenv("DB_CONN_STR"); dbConnStr != "" {
		c.DBConnStr = strings.TrimSpace(dbConnStr)
	}
	// Sentry Configuration
	// Sentry
	if sentry := os.Getenv("SENTRY"); sentry != "" {
		b, err := strconv.ParseBool(strings.TrimSpace(sentry))
		if err != nil {
			panic("Invalid SENTRY Environment Variable")
		}
		c.Sentry = b
	}
	// Sentry DSN
	if sentryDSN := os.Getenv("SENTRY_DSN"); sentryDSN != "" {
		c.SentryDSN = strings.TrimSpace(sentryDSN)
	}
	// Sentry Trace Sample Rate
	if sentryTSR := os.Getenv("SENTRY_TSR"); sentryTSR != "" {
		f, err := strconv.ParseFloat(strings.TrimSpace(sentryTSR), 64)
		if err != nil {
			panic("Invalid SENTRY_TSR Environment Variable")
		}
		c.SentryTSR = f
	}
}

func (c *configuration) setDefaults() {
	// Database
	c.DB = "postgres"
	// Sentry
	c.Sentry = true
	c.SentryDSN = "https://e7e4580a95ed8183cdf475d2fc826255@o4504687817261056.ingest.us.sentry.io/4506582744956928"
	c.SentryTSR = 1.0
}
