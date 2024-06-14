package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Leagueify/api/internal/database/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmailConfig(t *testing.T) {
	t.Parallel()
	// create mock db
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock DB: '%s'", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		RequestBody        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
		ExpectedContent    string
	}{
		{
			Description:        "Invalid JSON",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing email",
			RequestBody:        `{"smtpHost":"smtp.gmail.com","smtpPort":465,"smtpUser":"TestAccount","smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Email\]"`,
		},
		{
			Description:        "Missing smtpHost",
			RequestBody:        `{"email":"test@leagueify.org","smtpPort":465,"smtpUser":"TestAccount","smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[SMTPHost\]"`,
		},
		{
			Description:        "Missing smtpPort",
			RequestBody:        `{"email":"test@leagueify.org","smtpHost":"smtp.gmail.com","smtpUser":"TestAccount","smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[SMTPPort\]"`,
		},
		{
			Description:        "Missing smtpUser",
			RequestBody:        `{"email":"test@leagueify.org","smtpHost":"smtp.gmail.com","smtpPort":465,"smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[SMTPUser\]"`,
		},
		{
			Description:        "Missing smtpPass",
			RequestBody:        `{"email":"test@leagueify.org","smtpHost":"smtp.gmail.com","smtpPort":465,"smtpUser":"TestAccount"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[SMTPPass\]"`,
		},
		{
			Description:        "Missing Valid Payload",
			RequestBody:        `{}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Email SMTPHost SMTPPort SMTPUser SMTPPass\]"`,
		},
		{
			Description:        "Invalid Email",
			RequestBody:        `{"email":"test@leagueify","smtpHost":"smtp.gmail.com","smtpPort":465,"smtpUser":"TestAccount","smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid email"`,
		},
		{
			Description:        "Invalid SMTP Host",
			RequestBody:        `{"email":"test@leagueify.com","smtpHost":"ci-notfound.leagueify.org","smtpPort":465,"smtpUser":"TestAccount","smtpPass":"+#3Game!"}`,
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedContent:    `"detail":"host not found"`,
		},
	}
	for _, test := range testCases {
		if test.Mock != nil {
			test.Mock(mock)
		}
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := API{DB: db}
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/email/config", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if assert.NoError(t, api.createConfig(c)) {
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			match, err := regexp.MatchString(test.ExpectedContent, rec.Body.String())
			assert.NoError(t, err)
			assert.True(t, match, fmt.Sprintf("%v: Expected %v but received %v",
				test.Description, test.ExpectedContent, rec.Body.String(),
			))

		}
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
