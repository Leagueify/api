package util

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Leagueify/api/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func ComparePasswords(providedPassword, storedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
	return err == nil
}

func GenerateJWT(accountID, accountToken string) (string, error) {
	cfg := config.LoadConfig()
	jwtIssueTime := time.Now()
	jwtKey := fmt.Sprintf("%s.%s", accountToken, cfg.JWTSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"acc": accountID,
		"exp": jwt.NewNumericDate(jwtIssueTime.Add(time.Hour * 24 * 7)),
		"iat": jwt.NewNumericDate(jwtIssueTime),
		"jti": accountToken,
		"nbf": jwt.NewNumericDate(jwtIssueTime),
	})
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
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
