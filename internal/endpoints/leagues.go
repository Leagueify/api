package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/labstack/echo/v4"
)

func (api *API) Leagues(e *echo.Group) {
	e.POST("/leagues", api.requiresAdmin(api.createLeague))
}

func (api *API) createLeague(c echo.Context) error {
	league := model.LeagueCreation{}
	// bind payload to league model
	if err := c.Bind(&league); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}

	// validate payload against model
	if err := c.Validate(league); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	// check for existing league
	leagues, err := api.DB.GetTotalLeagues()
	if err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}
	if leagues > 0 {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	// validate sportID
	if _, err := api.DB.GetSportByID(league.SportID); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "invalid SportID")
	}

	// Set league.ID overriding provided ID
	league.ID = util.SignedToken(6)
	league.MasterAdmin = api.Account.ID
	// Insert league into database
	if err := api.DB.CreateLeague(league); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Successful League Creation
	return c.JSON(http.StatusCreated,
		map[string]string{
			"message": "successful",
		},
	)
}
