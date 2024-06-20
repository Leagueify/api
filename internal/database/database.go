package database

import (
	"database/sql"
	"fmt"

	"github.com/Leagueify/api/internal/config"
	"github.com/Leagueify/api/internal/database/postgres"
	"github.com/Leagueify/api/internal/model"
	"github.com/lib/pq"
)

type Database interface {
	// account functions
	ActivateAccount(accountID, apikey string) error
	CreateAccount(account model.AccountCreation) error
	GetAccountByAPIKey(apikey string) (model.Account, error)
	GetAccountByEmail(email string) (model.Account, error)
	GetTotalAccounts() (int, error)
	SetAPIKey(apikey, accountID string) error
	SetPlayerIDs(playerIDs *pq.StringArray, accountID string, tx *sql.Tx) error
	SetRegistrationCode(tx *sql.Tx, code, accountID string) error
	UnsetAPIKey(accountID string) error
	// email functions
	CreateEmailConfig(emailConfig model.EmailConfig) error
	GetTotalEmailConfigs() (int, error)
	// league functions
	CreateLeague(league model.LeagueCreation) error
	GetTotalLeagues() (int, error)
	// player function
	CreatePlayer(player model.Player, tx *sql.Tx) error
	DeletePlayer(playerID string, tx *sql.Tx) error
	GetPlayer(playerID string) (model.Player, error)
	RegisterPlayer(tx *sql.Tx, playerID string) error
	// position functions
	CreatePositions(positions model.PositionCreation) error
	GetAllPositions() ([]model.Position, error)
	GetTotalPositions() (int, error)
	// registration functions
	CreateRegistration(tx *sql.Tx, registration model.Registration) error
	GetRegistration(tx *sql.Tx, registrationID string) (pq.StringArray, error)
	SetRegistration(tx *sql.Tx, playerIDs pq.StringArray, registrationID string) error
	// sport functions
	GetSports() ([]model.Sport, error)
	GetSportByID(sportID string) (model.Sport, error)
	// database functions
	BeginTransaction() (*sql.Tx, error)
	InitializeDatabase() error
}

func GetDatabase() (Database, error) {
	cfg := config.LoadConfig()
	switch cfg.DB {
	case "postgres":
		db, err := postgres.Connect(cfg.DBConnStr)
		if err != nil {
			return nil, fmt.Errorf("ERROR: Database Connection Error '%s'", err)
		}
		return postgres.Postgres{
			DB: db,
		}, nil
	default:
		return nil, fmt.Errorf("ERROR: Unsupported Database '%s'", cfg.DB)
	}
}
