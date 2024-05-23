package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func (api *API) Positions(e *echo.Group) {
	e.GET("/positions", api.AuthRequired(api.listPositions))
	e.POST("/positions", api.AuthRequired(api.createPosition))
}

func (api *API) createPosition(c echo.Context) (err error) {
	positions := &model.PositionCreation{}
	// Bind payload to positions model
	if err := c.Bind(&positions); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if len(positions.Positions) <= 0 {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
			},
		)
	}
	// Check for existing positions
	var existingPositions model.Position
	if err := api.DB.QueryRow(`SELECT id FROM positions`).Scan(
		&existingPositions.ID,
	); err == nil {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Create Positions
	for _, position := range positions.Positions {
		_, err := api.DB.Exec(`
			INSERT INTO positions (
				id, name
			)
			VALUES ($1, $2)
		`,
			util.UnsignedToken(6), position,
		)
		if err != nil {
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
					"detail": util.HandleError(err),
				},
			)
		}
	}
	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}

func (api *API) listPositions(c echo.Context) (err error) {
	positions := []model.Position{}
	rows, err := api.DB.Query(
		`SELECT * FROM positions`,
	)
	if err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
			},
		)
	}
	defer rows.Close()
	for rows.Next() {
		var position model.Position
		if err := rows.Scan(
			&position.ID,
			&position.Name,
		); err != nil {
			sentry.CaptureException(err)
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
				},
			)
		}
		positions = append(positions, position)
	}
	if len(positions) < 1 {
		sentry.CaptureMessage("No rows returned in Positions table")
		return c.JSON(http.StatusNotFound,
			map[string]string{
				"status": "not found",
			},
		)
	}
	return c.JSON(http.StatusOK, positions)
}
