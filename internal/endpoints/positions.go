package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/labstack/echo/v4"
)

func (api *API) Positions(e *echo.Group) {
	e.GET("/positions", api.requiresAuth(api.listPositions))
	e.POST("/positions", api.requiresAdmin(api.createPosition))
}

func (api *API) createPosition(c echo.Context) (err error) {
	positions := &model.PositionCreation{}
	// Bind payload to positions model
	if err := c.Bind(&positions); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}
	if len(positions.Positions) <= 0 {
		return util.SendStatus(http.StatusBadRequest, c, "")
	}

	// Check for existing positions
	totalPositions, err := api.DB.GetTotalPositions()
	if err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}
	if totalPositions > 0 {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	// Create Positions
	if err := api.DB.CreatePositions(*positions); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}

func (api *API) listPositions(c echo.Context) (err error) {
	positions, err := api.DB.GetAllPositions()
	if err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "")
	}

	if len(positions) < 1 {
		return util.SendStatus(http.StatusNotFound, c, "")
	}

	return c.JSON(http.StatusOK, positions)
}
