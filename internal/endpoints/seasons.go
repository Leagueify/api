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
	e.GET("/seasons/:id", api.getSeason)
	e.PATCH("/seasons/:id", api.requiresAdmin(api.updateSeason))
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

func (api *API) getSeason(c echo.Context) error {
	seasonID := c.Param("id")
	if !util.VerifyToken(seasonID) {
		return util.SendStatus(http.StatusNotFound, c, "")
	}
	// search for season
	season, err := api.DB.GetSeason(seasonID)
	if err != nil {
		return util.SendStatus(http.StatusNotFound, c, "")
	}
	return c.JSON(http.StatusOK, season)
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

func (api *API) updateSeason(c echo.Context) error {
	seasonID := c.Param("id")
	if !util.VerifyToken(seasonID) {
		return util.SendStatus(http.StatusNotFound, c, "")
	}

	// bind payload
	payload := model.SeasonUpdate{}
	if err := c.Bind(&payload); err != nil {
		return util.SendStatus(
			http.StatusBadRequest, c, "invalid json payload",
		)
	}

	// search for season
	season, err := api.DB.GetSeason(seasonID)
	if err != nil {
		return util.SendStatus(http.StatusNotFound, c, "")
	}

	// TODO: update this block of code to remove the stack of if statements
	if payload.Name != "" {
		season.Name = payload.Name
	}
	if payload.StartDate != "" {
		season.StartDate = payload.StartDate
	}
	if payload.EndDate != "" {
		season.EndDate = payload.EndDate
	}
	if payload.RegistrationOpens != "" {
		season.RegistrationOpens = payload.RegistrationOpens
	}
	if payload.RegistrationCloses != "" {
		season.RegistrationCloses = payload.RegistrationCloses
	}

	// validate dates
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

	// store updates within database
	if err := api.DB.UpdateSeason(season); err != nil {
		return util.SendStatus(
			http.StatusBadRequest, c, util.HandleError(err),
		)
	}

	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
		},
	)
}
