package util

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Leagueify/api/internal/database"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func AuthRequired(f func(echo.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		apikey := c.Request().Header.Get("apiKey")
		if !VerifyToken(apikey) {
			return c.JSON(http.StatusUnauthorized,
				map[string]string{
					"status": "unauthorized",
				},
			)
		}
		db, err := database.Connect()
		if err != nil {
			return c.JSON(http.StatusBadGateway,
				map[string]string{
					"status": "bad gateway",
				},
			)
		}
		defer db.Close()
		result, err := db.Exec(`
			SELECT apikey FROM accounts where apikey = $1 AND is_active = true
		`, apikey[:len(apikey)-1])
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
		return f(c)
	}
}

func ComparePasswords(providedPassword, storedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
	return err == nil
}

func HashPassword(providedPassword *string) error {
	if len(*providedPassword) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(*providedPassword) > 64 {
		return errors.New("password must be at most 64 characters")
	}
	if !strings.ContainsAny(*providedPassword, "1234567890") {
		return errors.New("missing numeric character")
	}
	if !strings.ContainsAny(*providedPassword, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return errors.New("missing uppercase character")
	}
	if !strings.ContainsAny(*providedPassword, "abcdefghijklmnopqrstuvwxyz") {
		return errors.New("missing lowercase character")
	}
	if !strings.ContainsAny(*providedPassword, "~`!@#$%^&*()_-{[]},.") {
		return errors.New("missing special character")
	}
	password := []byte(*providedPassword)
	hashedPassword, err := bcrypt.GenerateFromPassword(password, 12)
	if err != nil {
		return err
	}
	*providedPassword = string(hashedPassword)
	return nil
}
