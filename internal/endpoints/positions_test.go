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
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreatePosition(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating mock DB", err)
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
			Description:        "Valid Request Body - No Positions",
			RequestBody:        `{"positions": []}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"status":"bad request"`,
		},
		{
			Description: "Valid Request Body - Single Position",
			RequestBody: `{"positions": ["skater"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM positions").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO positions (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Valid Request Body - Multiple Position",
			RequestBody: `{"positions": ["skater", "goalie"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM positions").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO positions (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO positions (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(2, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Valid Request Body - Duplicate Position",
			RequestBody: `{"positions": ["skater", "skater"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM positions").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO positions (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO positions (.+) VALUES (.+)$").WillReturnError(&pq.Error{Code: "23505", Constraint: "positions_name_key"})
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"name already in use"`,
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
		req := httptest.NewRequest(http.MethodPost, "/api/positions", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.createPosition(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// Validate Response Body
			match, err := regexp.MatchString(test.ExpectedContent, rec.Body.String())
			assert.True(t, match, fmt.Sprintf("%v: Expected %v but received %v",
				test.Description, test.ExpectedContent, rec.Body.String(),
			))
			assert.NoError(t, err)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestListPositions(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description: "Return Positions",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "No Rows Returned",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
			},
			ExpectedStatusCode: http.StatusNotFound,
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
		api := API{DB: db}
		req := httptest.NewRequest(http.MethodGet, "/api/positions", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.listPositions(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// Validate Response Body
			assert.NoError(t, err)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
