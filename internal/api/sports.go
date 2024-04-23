package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/auth"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

type sport struct {
	ID   string
	Name string
}

func (api *API) Sports(e *echo.Group) {
	e.GET("/sports", auth.AuthRequired(api.listSports))
}

func (api *API) listSports(c echo.Context) (err error) {
	sports := []sport{}
	rows, err := api.DB.Query(
		`SELECT * FROM sports`,
	)
	if err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadGateway,
			map[string]string{
				"status": "bad gateway",
			},
		)
	}
	defer rows.Close()
	for rows.Next() {
		var sport sport
		if err := rows.Scan(
			&sport.ID,
			&sport.Name,
		); err != nil {
			sentry.CaptureException(err)
			return c.JSON(http.StatusBadGateway,
				map[string]string{
					"status": "bad gateway",
				},
			)
		}
		sports = append(sports, sport)
	}
	if len(sports) < 1 {
		sentry.CaptureMessage("No rows returned in Sports table")
		return c.JSON(http.StatusNotFound,
			map[string]string{
				"status": "not found",
			},
		)
	}
	return c.JSON(http.StatusOK, sports)
}
