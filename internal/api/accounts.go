package api

import (
	"net/http"
	"time"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func (api *API) Accounts(e *echo.Group) {
	e.POST("/accounts", api.createAccount)
	e.POST("/accounts/:id/verify", api.verifyAccount)
}

func (api *API) createAccount(c echo.Context) (err error) {
	account := model.AccountCreation{}
	// Bind payload to account model
	if err := c.Bind(&account); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "invalid json payload",
			},
		)
	}
	// Validate payload against model
	if err := c.Validate(account); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Calculate Age
	today := time.Now().Format(time.DateOnly)
	age, err := util.CalculateAge(account.DateOfBirth, today)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": err.Error(),
			},
		)
	}
	if age < 18 {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "must be 18 or older to create an account",
			},
		)
	}
	// Set account.ID overriding provided ID
	account.ID = util.SignedToken(8)
	// Hash Password
	if err := util.HashPassword(&account.Password); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": err.Error(),
			},
		)
	}
	// Insert account into database
	_, err = api.DB.Exec(`
		INSERT INTO accounts (
			id, first_name, last_name, email, password,
			phone, date_of_birth, coach, volunteer, is_active
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`,
		account.ID[:len(account.ID)-1], account.FirstName,
		account.LastName, account.Email, account.Password,
		account.Phone, account.DateOfBirth, account.Coach,
		account.Volunteer, true,
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Successful Account Creation
	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}

func (api *API) verifyAccount(c echo.Context) (err error) {
	accountID := c.Param("id")
	// Verify Account ID
	if !util.VerifyToken(accountID) {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Update Account
	accountToken := util.SignedToken(10)
	result, err := api.DB.Exec(`
		UPDATE accounts SET is_active = true, token = $1 WHERE id = $2 AND is_active = false
	`, accountToken, accountID[:len(accountID)-1])
	if err != nil {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Generate JWT
	accountJWT, err := util.GenerateJWT(accountID, accountToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Return JWT
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
			"token":  accountJWT,
		},
	)
}
