package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	// "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
	// Create Mock DB
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ERROR: '%s' was not expected when creating mock DB", err)
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
			Description: "Valid Request Body",
			RequestBody: `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO accounts (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"token":".+\..+\..+"`,
		},
		{
			Description:        "Invalid JSON Payload",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid json payload"`,
		},
		{
			Description:        "Missing FirstName",
			RequestBody:        `{"lastName":"Tests","email":"test@leagueify.org","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[FirstName\]"`,
		},
		{
			Description:        "Missing LastName",
			RequestBody:        `{"firstName":"Leagueify","email":"test@leagueify.org","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[LastName\]"`,
		},
		{
			Description:        "Missing Email",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Email\]"`,
		},
		{
			Description:        "Missing Password",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Password\]"`,
		},
		{
			Description:        "Missing DateOfBirth",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test123!","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[DateOfBirth\]"`,
		},
		{
			Description:        "Missing Phone",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test123!","dateOfBirth":"1990-08-31"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[Phone\]"`,
		},
		{
			Description:        "Missing Valid Payload",
			RequestBody:        `{}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing required field\(s\): \[FirstName LastName Email Password Phone DateOfBirth\]"`,
		},
		{
			Description:        "Invalid Password Too Short",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"password must be at least 8 characters"`,
		},
		{
			Description:        "Invalid Password Too Long",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test123!901234567890123456789012345678901234567890123456789012345","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"password must be at most 64 characters"`,
		},
		{
			Description:        "Invalid Password Missing Numeric Character",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Testing!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing numeric character"`,
		},
		{
			Description:        "Invalid Password Missing Uppercase Character",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing uppercase character"`,
		},
		{
			Description:        "Invalid Password Missing Lowercase Character",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"TEST123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing lowercase character"`,
		},
		{
			Description:        "Invalid Password Missing Special Character",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test1234","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"missing special character"`,
		},
		{
			Description:        "Invalid Email",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"invalid email"`,
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
		reqBody := []byte(test.RequestBody)
		req := httptest.NewRequest(http.MethodPost, "/api/accounts", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.createAccount(c)) {
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
