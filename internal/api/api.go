package api

import (
	"database/sql"
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
	Account   *model.Account
	DB        *sql.DB
	Validator *validator.Validate
}

func (api *API) AuthRequired(f func(echo.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		account := model.Account{}
		apikey := c.Request().Header.Get("apiKey")
		if !util.VerifyToken(apikey) {
			return c.JSON(http.StatusUnauthorized,
				map[string]string{
					"status": "unauthorized",
				},
			)
		}
		err := api.DB.QueryRow(`
			SELECT * FROM accounts where apikey = $1 AND is_active = true
		`, apikey[:len(apikey)-1]).Scan(
			&account.ID,
			&account.FirstName,
			&account.LastName,
			&account.Email,
			&account.Password,
			&account.Phone,
			&account.DateOfBirth,
			&account.Coach,
			&account.Volunteer,
			&account.APIKey,
			&account.IsActive,
		)
		api.Account = &account
		if err != nil {
			return c.JSON(http.StatusUnauthorized,
				map[string]string{
					"status": "unauthorized",
				},
			)
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
	api.Leagues(routes)
	api.Positions(routes)
	api.Sports(routes)
}
