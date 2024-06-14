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
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreatePlayer(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
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
			Description:        "Invalid Payload",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing Players",
			RequestBody:        `{}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Players\]"`,
		},
		{
			Description:        "No Players in Payload",
			RequestBody:        `{"players":[]}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"payload contains no players"`,
		},
		{
			Description: "Single Player Missing FirstName",
			RequestBody: `{"players":[{"lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[FirstName\]"`,
		},
		{
			Description: "Single Player Missing LastName",
			RequestBody: `{"players":[{"firstName":"Leagueify","dateOfBirth":"2016-12-10","position":"goalie"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[LastName\]"`,
		},
		{
			Description: "Single Player Missing DateOfBirth",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","position":"goalie"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[DateOfBirth\]"`,
		},
		{
			Description: "Single Player Missing Position",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Position\]"`,
		},
		{
			Description: "Single Player Invalid Position",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"skate"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid position"`,
		},
		{
			Description: "Create Single Player",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Second Player Missing FirstName",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"lastName":"Test","dateOfBirth":"2019-02-14","position":"skater"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[FirstName\]"`,
		},
		{
			Description: "Second Player Missing LastName",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"firstName":"Second","dateOfBirth":"2019-02-14","position":"skater"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[LastName\]"`,
		},
		{
			Description: "Second Player Missing DateOfBirth",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"firstName":"Second","lastName":"Test","position":"skater"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[DateOfBirth\]"`,
		},
		{
			Description: "Second Player Missing Position",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"firstName":"Second","lastName":"Test","dateOfBirth":"2019-02-14"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Position\]"`,
		},
		{
			Description: "Second Player Invalid Position",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"firstName":"Second","lastName":"Test","dateOfBirth":"2019-02-14","position":"skate"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid position"`,
		},
		{
			Description: "Create Multiple Players",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"firstName":"Second","lastName":"Test","dateOfBirth":"2019-02-14","position":"skater"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM positions").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "skater").AddRow("2", "goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
	}
	// Execute Test Cases
	for _, test := range testCases {
		if test.Mock != nil {
			test.Mock(mock)
		}
		// Initialize Echo and the Echo validator
		e := echo.New()
		account := model.Account{
			ID: "123ABC",
		}
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: account}
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/players", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.createPlayer(c)) {
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

func TestDeletePlayer(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		ID                 string
		Account            model.Account
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid Player ID",
			ID:                 "ABD1234",
			Account:            model.Account{ID: "123ABC"},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description:        "Valid Player ID no Player ID in Account",
			ID:                 "QP4RD39CEF",
			Account:            model.Account{ID: "123ABC"},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description:        "Valid Player ID not in Account",
			ID:                 "49QRBF09YA",
			Account:            model.Account{ID: "123ABC", Players: pq.StringArray{"QP4RD39CE"}},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description: "Valid Player ID in Account",
			ID:          "49QRBF09YA",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"49QRBF09Y"}},
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM players WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description: "Valid Player ID in Account with Multiple Player IDs",
			ID:          "49QRBF09YA",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"12345ABCD", "49QRBF09Y"}},
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("DELETE FROM players WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
	}
	// Execute Test Cases
	for _, test := range testCases {
		if test.Mock != nil {
			test.Mock(mock)
		}
		// Initialize Echo and the Echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: test.Account}
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/players/%s", test.ID), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(test.ID)
		// Perform Request
		if assert.NoError(t, api.deletePlayer(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestGetPlayer(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		ID                 string
		Account            model.Account
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid Player ID",
			ID:                 "ABD1234",
			Account:            model.Account{ID: "123ABC"},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description:        "Valid Player ID no Player ID in Account",
			ID:                 "QP4RD39CEF",
			Account:            model.Account{ID: "123ABC"},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description:        "Valid Player ID not in Account",
			ID:                 "49QRBF09YA",
			Account:            model.Account{ID: "123ABC", Players: pq.StringArray{"QP4RD39CE"}},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Valid Player ID in Account",
			ID:          "49QRBF09YA",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"49QRBF09Y"}},
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM players WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "date_of_birth", "position", "team", "division", "is_registered"}).AddRow("49QRBF09YA", "Leagueify", "Test", "1990-08-31", "goalie", "", "", false))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Valid Player ID in Account with Multiple Player IDs",
			ID:          "49QRBF09YA",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"12345ABCD", "49QRBF09Y"}},
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM players WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "date_of_birth", "position", "team", "division", "is_registered"}).AddRow("49QRBF09YA", "Leagueify", "Test", "1990-08-31", "goalie", "", "", false))
			},
			ExpectedStatusCode: http.StatusOK,
		},
	}
	// Execute Test Cases
	for _, test := range testCases {
		if test.Mock != nil {
			test.Mock(mock)
		}
		// Initialize Echo and the Echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: test.Account}
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/players/%s", test.ID), nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(test.ID)
		// Perform Request
		if assert.NoError(t, api.getPlayer(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}

}

func TestGetPlayers(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		Account            model.Account
		ExpectedStatusCode int
	}{
		{
			Description:        "No Results",
			Account:            model.Account{ID: "123ABC"},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description:        "Result Found",
			Account:            model.Account{ID: "123ABC", Players: pq.StringArray{"QP4RD39CEF"}},
			ExpectedStatusCode: http.StatusOK,
		},
	}
	// Execute Test Cases
	for _, test := range testCases {
		// Initialize Echo and the Echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: test.Account}
		req := httptest.NewRequest(http.MethodGet, "/api/players", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.getPlayers(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestRegisterPlayer(t *testing.T) {
	// run test in parallel
	t.Parallel()
	// Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		Account            model.Account
		RequestBody        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedContent    string
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid Payload",
			Account:            model.Account{ID: "123ABC"},
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing Players",
			Account:            model.Account{ID: "123ABC"},
			RequestBody:        `{}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Players\]"`,
		},
		{
			Description:        "No Players in payload",
			Account:            model.Account{ID: "123ABC"},
			RequestBody:        `{"players":[]}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"payload contains no players"`,
		},
		{
			Description: "Invalid Player ID",
			Account:     model.Account{ID: "123ABC"},
			RequestBody: `{"players":["ABD123"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedContent:    `"status":"not found"`,
		},
		{
			Description: "Valid Player ID no Player ID in Account",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{}},
			RequestBody: `{"players":["W4SBH35WV8"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedContent:    `"status":"not found"`,
		},
		{
			Description: "Valid Player ID not in Account",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"49QRBF09Y"}},
			RequestBody: `{"players":["DW74MSY5XQ"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectRollback()
			},
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedContent:    `"status":"not found"`,
		},
		{
			Description: "Valid Player ID in Account",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"DW74MSY5X"}},
			RequestBody: `{"players":["DW74MSY5XQ"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE players SET is_registered = true WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO registrations (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Valid Player ID in Account with Multiple Player IDs",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"W4SBH35WV", "DW74MSY5X"}},
			RequestBody: `{"players":["DW74MSY5XQ"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE players SET is_registered = true WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO registrations (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Valid Player ID add to Existing Registration",
			Account:     model.Account{ID: "123ABC", Players: pq.StringArray{"W4SBH35WV", "DW74MSY5X"}, RegistrationCode: "123ABC"},
			RequestBody: `{"players":["DW74MSY5XQ"]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE accounts SET registration_code = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE players SET is_registered = true WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT player_ids FROM registrations WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{'W4SBH35WV'}"))
				mock.ExpectExec("UPDATE registrations SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedContent:    `"status":"successful"`,
		},
		// TODO: Add more tests
	}
	// Execute Test Cases
	for _, test := range testCases {
		if test.Mock != nil {
			test.Mock(mock)
		}
		// Initialize Echo and the Echo validator
		e := echo.New()
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: test.Account}
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/players/register", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.registerPlayer(c)) {
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
