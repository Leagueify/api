package api

import (
	"fmt"
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/labstack/echo/v4"
)

func (api *API) Seasons(e *echo.Group) {
	e.POST("/seasons", api.requiresAdmin(api.createSeason))
	e.GET("/seasons", api.listSeasons)
}

func (api *API) createSeason(c echo.Context) error {
	season := model.Season{}
	// bind payload to model
	if err := c.Bind(&season); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}
	// validate payload against model
	if err := c.Validate(season); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	var dateErrors []string
	validDate, err := util.IsValidDateRange(
		season.StartDate, season.EndDate,
	)
	if err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	if !validDate {
		dateErrors = append(
			dateErrors,
			"StartDate-EndDate",
		)
	}

	validRegistrationDate, err := util.IsValidDateRange(
		season.RegistrationOpens, season.RegistrationCloses,
	)
	if err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	if !validRegistrationDate {
		dateErrors = append(
			dateErrors,
			"RegistrationOpens-RegistrationCloses",
		)
	}

	if len(dateErrors) != 0 {
		return util.SendStatus(
			http.StatusBadRequest, c,
			fmt.Sprintf(
				"incorrect date range(s): %v",
				dateErrors,
			),
		)
	}

	season.ID = util.SignedToken(10)
	if err := api.DB.CreateSeason(season); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}

func (api *API) listSeasons(c echo.Context) error {
	seasons, err := api.DB.ListSeasons()
	if err != nil {
		return util.SendStatus(
			http.StatusInternalServerError, c,
			util.HandleError(err),
		)
	}

	if len(seasons) == 0 {
		return util.SendStatus(http.StatusNotFound, c, "")
	}

	return c.JSON(http.StatusOK, seasons)
}
