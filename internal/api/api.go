package api

import (
	"database/sql"

	"github.com/Leagueify/api/internal/database"
	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type API struct {
	DB        *sql.DB
	Validator *validator.Validate
}

func (api *API) Validate(i interface{}) error {
	if err := api.Validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func Routes(e *echo.Echo) {
	db, err := database.Connect()
	if err != nil {
		sentry.CaptureException(err)
	}
	api := &API{DB: db}
	e.Validator = &API{Validator: validator.New()}
	// Create API Group
	routes := e.Group("/api")
	// Register API Routes
	api.Accounts(routes)
}
