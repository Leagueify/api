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
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateLeague(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ERROR: '%s' was not expected when creating mock DB", err)
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
			Description:        "Invalid Request Body",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing Name",
			RequestBody:        `{"sportID":"4"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Name\]"`,
		},
		{
			Description:        "Missing SportID",
			RequestBody:        `{"name":"Leagueify Sporting League"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[SportID\]"`,
		},
		{
			Description:        "Inalid Request Body - Min Name Violation",
			RequestBody:        `{"name":"LE","sportID":"4"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"'Name' must have a minimum length of '3' characters"`,
		},
		{
			Description: "Invalid SportID",
			RequestBody: `{"name":"Leagueify Sporting League","sportID":"65"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM leagues").WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT \\* FROM sports WHERE id = (.+)").WillReturnRows(mock.NewRows([]string{"id", "name"}))
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid SportID"`,
		},
		{
			Description: "Valid Request Body",
			RequestBody: `{"name":"Leagueify Sporting League","sportID":"65"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM leagues").WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectQuery("SELECT \\* FROM sports WHERE id = (.+)").WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow("65", "hockey"))
				mock.ExpectExec("INSERT INTO leagues (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusCreated,
		},
	}
	// Execute Test Cases
	for _, test := range testCases {
		// Determine if tests should have mock DB
		if test.Mock != nil {
			test.Mock(mock)
		}
		// Initialize Echo and the Echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := API{DB: db}
		api.Account = model.Account{}
		api.Account.ID = util.SignedToken(8)
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/leagues", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.createLeague(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// Validate Response Body
			match, err := regexp.MatchString(test.ExpectedContent, rec.Body.String())
			assert.NoError(t, err)
			assert.True(t, match, fmt.Sprintf("%v: Expected %v but received %v",
				test.Description, test.ExpectedContent, rec.Body.String(),
			))
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
