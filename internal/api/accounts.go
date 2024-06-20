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
	if err := auth.HashPassword(&account.Password); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": err.Error(),
			},
		)
	}
	// Check for existing accounts
	var numAccounts int
	row := api.DB.QueryRow(
		`SELECT COUNT(*) FROM accounts`,
	)
	if err := row.Scan(&numAccounts); err != nil {
		return c.JSON(http.StatusBadGateway,
			map[string]string{
				"status": "bad gateway",
				"detail": util.HandleError(err),
			},
		)
	}
	// Default is_admin false
	is_admin := false
	if numAccounts < 1 {
		is_admin = true
	}
	// Insert account into database
	_, err = api.DB.Exec(`
		INSERT INTO accounts (
			id, first_name, last_name, email, password,
			phone, date_of_birth, registration_code, player_ids,
			coach, volunteer, apikey, is_active, is_admin
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)`,
		account.ID[:len(account.ID)-1], account.FirstName,
		account.LastName, account.Email, account.Password,
		account.Phone, account.DateOfBirth, "", "{}", account.Coach,
		account.Volunteer, "", true, is_admin,
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

func (api *API) loginAccount(c echo.Context) error {
	credentials := &model.AccountLogin{}
	account := &model.Account{}
	if err := c.Bind(&credentials); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	if err := api.DB.QueryRow(`
		SELECT password, is_active FROM accounts WHERE email = $1 AND is_active = true
	`, credentials.Email).Scan(
		&account.Password,
		&account.IsActive,
	); err != nil {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	if !auth.ComparePasswords(credentials.Password, account.Password) {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Generate API Key
	accountAPIKey := util.SignedToken(64)
	result, err := api.DB.Exec(`
		UPDATE accounts SET apikey = $1 WHERE email = $2
	`, accountAPIKey[:len(accountAPIKey)-1], credentials.Email)
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
	// Return API Key
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
			"apikey": accountAPIKey,
		},
	)
}

func (api *API) logoutAccount(c echo.Context) error {
	account := &model.Account{}
	if err := api.DB.QueryRow(`
		SELECT email, apikey FROM accounts WHERE apikey = $1
	`, &account.APIKey).Scan(
		&account.Email,
		&account.APIKey,
	); err != nil {
		return c.JSON(http.StatusUnauthorized,
			map[string]string{
				"status": "unauthorized",
			},
		)
	}
	// Remove API Key
	result, err := api.DB.Exec(`
		UPDATE accounts SET apikey = '' WHERE email = $1
	`, &account.Email)
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
	return c.JSON(http.StatusOK, "{}")
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
	// Generate API Key
	accountAPIKey := util.SignedToken(64)
	result, err := api.DB.Exec(`
		UPDATE accounts SET is_active = true, apikey = $1 WHERE id = $2
	`, accountAPIKey[:len(accountAPIKey)-1], accountID[:len(accountID)-1])
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
	// Return API Key
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
			"apikey": accountAPIKey,
		},
	)
}
