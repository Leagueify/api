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

func TestGetSeason(t *testing.T) {
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
		ID                 string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid Season ID",
			ID:                 "ABC1234",
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Valid Season ID no Season with ID",
			ID:          "BJ7Q4NVRNQ",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationsCloses"}))
			},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Valid Season ID Found",
			ID:          "BJ7Q4NVRNQ",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
			},
			ExpectedStatusCode: http.StatusOK,
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
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/seasons/%s", test.ID), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(test.ID)

		// perform request
		if assert.NoError(t, api.getSeason(c)) {
			// assert status code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// validate response body
			assert.NoError(t, err)
		}
		// assert all expectations were met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestUpdateSeason(t *testing.T) {
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
		ID                 string
		RequestBody        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
		ExpectedContent    string
	}{
		{
			Description:        "Invalid JSON Payload",
			ID:                 "1RFWY1T1B~",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Invalid Season ID",
			ID:                 "1RFWY1T1B1",
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Valid Season ID no Season with ID",
			ID:          "BJ7Q4NVRNQ",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationsCloses"}))
			},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Invalid Date Ranges: Season Dates",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"endDate":"2024-01-01"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[StartDate-EndDate\]"`,
		},
		{
			Description: "Invalid Date Ranges: Registration Dates",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"registrationCloses":"2023-12-31"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[RegistrationOpens-RegistrationCloses\]"`,
		},
		{
			Description: "Invalid Date Ranges",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"name":"Test Season","startDate":"2024-03-01","endDate":"2024-01-01","registrationOpens":"2024-01-01","registrationCloses":"2023-12-31"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"incorrect date range\(s\): \[StartDate-EndDate RegistrationOpens-RegistrationCloses\]"`,
		},
		{
			Description: "Update Name",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"name":"Test Season"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Update StartDate",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"start_date":"2024-03-02"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Update EndDate",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"end_date":"2024-05-02"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Update RegistrationOpens",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"registration_opens":"2024-01-02"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Update RegistrationCloses",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"registration_closes":"2024-03-02"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Update Entire Season",
			ID:          "BJ7Q4NVRNQ",
			RequestBody: `{"name":"Test Season","start_date":"2024-03-02","end_date":"2024-05-02","registration_opens":"2024-01-02","registration_closes":"2024-03-02"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM seasons WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "startDate", "endDate", "registrationOpens", "registrationCloses"}).AddRow("BJ7Q4NVRNQ", "2024-2025", "2024-03-01", "2024-05-01", "2024-01-01", "2024-03-01"))
				mock.ExpectExec("UPDATE seasons SET name = (.+), start_date = (.+), end_date = (.+), registration_opens = (.+), registration_closes = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
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
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/seasons/%s", test.ID), bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(test.ID)
		// perform request
		if assert.NoError(t, api.updateSeason(c)) {
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
