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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateSeason(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// create mock db
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected creating mock DB", err)
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
			Description:        "Invalid request json",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing Required Fields",
			RequestBody:        `{}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Name StartDate EndDate RegistrationOpens RegistrationCloses\]"`,
		},
		{
			Description:        "Missing Required Field: Name",
			RequestBody:        `{"startDate":"2024-03-01","endDate":"2024-05-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Name\]"`,
		},
		{
			Description:        "Missing Required Field: StartDate",
			RequestBody:        `{"name":"Test Season","endDate":"2024-05-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[StartDate\]"`,
		},
		{
			Description:        "Missing Required Field: EndDate",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[EndDate\]"`,
		},
		{
			Description:        "Missing Required Field: RegistrationOpens",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-05-01","registrationCloses":"2024-03-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[RegistrationOpens\]"`,
		},
		{
			Description:        "Missing Required Field: RegistrationCloses",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-05-01","registrationOpens":"2024-01-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[RegistrationCloses\]"`,
		},
		{
			Description:        "Invalid Date Ranges: Season Dates",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-01-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[StartDate-EndDate\]"`,
		},
		{
			Description:        "Invalid Date Ranges: Registration Dates",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-05-01","registrationOpens":"2024-01-01","registrationCloses":"2023-12-31"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[RegistrationOpens-RegistrationCloses\]"`,
		},
		{
			Description:        "Invalid Date Ranges",
			RequestBody:        `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-01-01","registrationOpens":"2024-01-01","registrationCloses":"2023-12-31"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[StartDate-EndDate RegistrationOpens-RegistrationCloses\]"`,
		},
		{
			Description: "Valid Request",
			RequestBody: `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-05-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO seasons (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Duplicate Season Name",
			RequestBody: `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-05-01","registrationOpens":"2024-01-01","registrationCloses":"2024-03-01"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO seasons (.+) VALUES (.+)$").WillReturnError(&pq.Error{Code: "23505", Constraint: "seasons_name_key"})
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"name already in use"`,
		},
	}
	for _, test := range testCases {
		// use mock if set
		if test.Mock != nil {
			test.Mock(mock)
		}
		// echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := API{DB: db}
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// perform request
		if assert.NoError(t, api.createSeason(c)) {
			// assert status code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// validate request body
			match, err := regexp.MatchString(test.ExpectedContent, rec.Body.String())
			assert.NoError(t, err)
			assert.True(t, match, fmt.Sprintf("%v: Expected %v, but received %v",
				test.Description, test.ExpectedContent, rec.Body.String(),
			))
		}
		// assert all expectations where met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestListSeasons(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// create mock db
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected creating mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description: "Return Seasons",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name FROM seasons").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "2024-2025"))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "No Rows Returned",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name FROM seasons").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			ExpectedStatusCode: http.StatusNotFound,
		},
	}
	for _, test := range testCases {
		// utilize mock db if required
		if test.Mock != nil {
			test.Mock(mock)
		}
		// initialize echo
		e := echo.New()
		api := API{DB: db}
		req := httptest.NewRequest(http.MethodGet, "/api/seasons", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// perform request
		if assert.NoError(t, api.listSeasons(c)) {
			// assert status code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// validate response body
			assert.NoError(t, err)
		}
		// assert all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
