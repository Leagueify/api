package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/database"
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type API struct {
	Account   model.Account
	DB        database.Database
	Validator *validator.Validate
}

func (api *API) requiresAdmin(f func(echo.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		var err error

		apikey := c.Request().Header.Get("apiKey")
		if !util.VerifyToken(apikey) {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}

		api.Account, err = api.DB.GetAccountByAPIKey(apikey)
		if err != nil {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}
		if !api.Account.IsActive {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}
		if !api.Account.IsAdmin {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}

		return f(c)
	}
}

func (api *API) requiresAuth(f func(echo.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		var err error

		apikey := c.Request().Header.Get("apiKey")
		if !util.VerifyToken(apikey) {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}

		api.Account, err = api.DB.GetAccountByAPIKey(apikey)
		if err != nil {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}
		if !api.Account.IsActive {
			return util.SendStatus(http.StatusUnauthorized, c, "")
		}

		return f(c)
	}
}

func (api *API) Validate(i interface{}) error {
	if err := api.Validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func Routes(e *echo.Echo) {
	db, err := database.GetDatabase()
	if err != nil {
		sentry.CaptureException(err)
	}
	api := &API{DB: db}
	e.Validator = &API{Validator: validator.New()}
	// Create API Group
	routes := e.Group("/api")
	// Register API Routes
	api.Accounts(routes)
	api.Email(routes)
	api.Leagues(routes)
	api.Players(routes)
	api.Positions(routes)
	api.Sports(routes)
}
