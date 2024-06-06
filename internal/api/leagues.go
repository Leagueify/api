package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func (api *API) Leagues(e *echo.Group) {
	e.POST("/leagues", api.requiresAdmin(api.createLeague))
}

func (api *API) createLeague(c echo.Context) error {
	league := model.LeagueCreation{}
	// Bind payload to league model
	if err := c.Bind(&league); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "invalid json payload",
			},
		)
	}
	// Validate payload against model
	if err := c.Validate(league); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Check for Existing League
	var existingLeague model.LeagueCreation
	if err := api.DB.QueryRow(`SELECT id FROM leagues`).Scan(
		&existingLeague.ID,
	); err == nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Set league.ID overriding provided ID
	league.ID = util.SignedToken(6)
	league.MasterAdmin = api.Account.ID
	// Insert league into database
	_, err := api.DB.Exec(`
		INSERT INTO leagues (
			id, name, sport_id, master_admin
		)
		VALUES (
			$1, $2, $3, $4
		)`,
		league.ID[:len(league.ID)-1], league.Name,
		league.SportID, league.MasterAdmin,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Successful League Creation
	return c.JSON(http.StatusCreated,
		map[string]string{
			"message": "successful",
			"detail":  league.ID,
		},
	)
}
