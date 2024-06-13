package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/util"
	"github.com/labstack/echo/v4"
)

func (api *API) Sports(e *echo.Group) {
	e.GET("/sports", api.requiresAuth(api.listSports))
}

func (api *API) listSports(c echo.Context) (err error) {
	sports, err := api.DB.GetSports()
	if err != nil {
		return util.SendStatus(http.StatusInternalServerError, c, "")
	}
	if len(sports) < 1 {
		return util.SendStatus(http.StatusNotFound, c, "")
	}
	return c.JSON(http.StatusOK, sports)
}
