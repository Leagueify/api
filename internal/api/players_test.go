package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Leagueify/api/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreatePlayer(t *testing.T) {
	// Mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	defer db.Close()
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
			Description: "Single Player Missing FirstName",
			RequestBody: `{"players":[{"lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description: "Second Player Missing FirstName",
			RequestBody: `{"players":[{"firstName":"Leagueify","lastName":"Test","dateOfBirth":"2016-12-10","position":"goalie"},{"lastName":"Test","dateOfBirth":"2019-02-14","position":"skater"}]}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
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
				mock.ExpectQuery("SELECT name FROM positions").WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("skater").AddRow("goalie"))
				mock.ExpectBegin()
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("INSERT INTO players (.+) VALUES (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE accounts SET player_ids = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			ExpectedStatusCode: http.StatusOK,
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
		account := &model.Account{
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
	// Mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	defer db.Close()
	testCases := []struct {
		Description        string
		ID                 string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid Player ID",
			ID:                 "ABD1234",
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description: "Valid Player ID no Player ID in Account",
			ID:          "QP4RD39CEF",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{}"))
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description: "Valid Player ID not in Account",
			ID:          "49QRBF09YA",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{12345ABCDE}"))
			},
			ExpectedStatusCode: http.StatusNoContent,
		},
		{
			Description: "Valid Player ID in Account",
			ID:          "49QRBF09YA",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{49QRBF09YA}"))
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
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{12345ABCDE,49QRBF09YA}"))
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
		account := &model.Account{
			ID: "123ABC",
		}
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: account}
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

func TestGetPlayers(t *testing.T) {
	// Mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error: '%s' was not expected when creating the mock DB", err)
	}
	defer db.Close()
	testCases := []struct {
		Description        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description: "No Results",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{}"))
			},
			ExpectedStatusCode: http.StatusNotFound,
		},
		{
			Description: "Result Found",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT player_ids FROM accounts WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"player_ids"}).AddRow("{12345ABCDE}"))
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
		account := &model.Account{
			ID: "123ABC",
		}
		e.Validator = &API{Validator: validator.New()}
		api := &API{DB: db, Account: account}
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
