package api

import (
	"net/http"
	"time"

	"github.com/Leagueify/api/internal/auth"
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func (api *API) Accounts(e *echo.Group) {
	e.POST("/accounts", api.createAccount)
	e.POST("/accounts/:id/verify", api.verifyAccount)
	e.POST("/accounts/login", api.loginAccount)
	e.POST("/accounts/logout", api.requiresAuth(api.logoutAccount))
}

func (api *API) createAccount(c echo.Context) (err error) {
	account := model.AccountCreation{}
	// Bind payload to account model
	if err := c.Bind(&account); err != nil {
		sentry.CaptureException(err)
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}
	// Validate payload against model
	if err := c.Validate(account); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Calculate Age
	today := time.Now().Format(time.DateOnly)
	age, err := util.CalculateAge(account.DateOfBirth, today)
	if err != nil {
		// TODO: Update to use util.HandleError
		return util.SendStatus(http.StatusBadRequest, c, err.Error())
	}
	if age < 18 {
		return util.SendStatus(http.StatusBadRequest, c, "must be 18 or older to create an account")
	}
	// Set account.ID overriding provided ID
	account.ID = util.SignedToken(8)
	// Hash Password
	if err := auth.HashPassword(&account.Password); err != nil {
		sentry.CaptureException(err)
		// TODO: Update to use util.HandleError
		return util.SendStatus(http.StatusBadRequest, c, err.Error())
	}
	// Check for existing accounts
	totalAccounts, err := api.DB.GetTotalAccounts()
	if err != nil {
		return util.SendStatus(http.StatusBadGateway, c, util.HandleError(err))
	}
	// Default is_admin false
	account.IsAdmin = false
	if totalAccounts < 1 {
		account.IsAdmin = true
	}
	// Insert account into database
	if err := api.DB.CreateAccount(account); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Successful Account Creation
	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}

func (api *API) loginAccount(c echo.Context) error {
	credentials := &model.AccountLogin{}
	if err := c.Bind(&credentials); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "")
	}
	account, err := api.DB.GetAccountByEmail(credentials.Email)
	if err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	if !auth.ComparePasswords(credentials.Password, account.Password) {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	// Generate API Key
	apikey := util.SignedToken(64)
	if err := api.DB.SetAPIKey(apikey, account.ID); err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	// Return API Key
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
			"apikey": apikey,
		},
	)
}

func (api *API) logoutAccount(c echo.Context) error {
	if err := api.DB.SetAPIKey(" ", api.Account.ID); err != nil {
		return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
	}

	return c.JSON(http.StatusOK, "{}")
}

func (api *API) verifyAccount(c echo.Context) (err error) {
	accountID := c.Param("id")
	// Verify Account ID
	if !util.VerifyToken(accountID) {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	// Generate API Key
	apikey := util.SignedToken(64)
	if err := api.DB.ActivateAccount(accountID, apikey); err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}
	// Return API Key
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
			"apikey": apikey,
		},
	)
}
