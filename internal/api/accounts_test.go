package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Leagueify/api/internal/auth"
	"github.com/Leagueify/api/internal/database/postgres"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccount(t *testing.T) {
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
			Description: "Valid Request Body",
			RequestBody: `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.org","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM accounts").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("INSERT INTO accounts (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusCreated,
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
		{
			Description: "Duplicate Email Address",
			RequestBody: `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.com","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM accounts").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("^INSERT INTO accounts (.+) VALUES (.+)$").WillReturnError(&pq.Error{Code: "23505", Constraint: "accounts_email_key"})
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"email already in use"`,
		},
		{
			Description: "Duplicate Phone",
			RequestBody: `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.com","password":"Test123!","dateOfBirth":"1990-08-31","phone":"+12085550000"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM accounts").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("^INSERT INTO accounts (.+) VALUES (.+)$").WillReturnError(&pq.Error{Code: "23505", Constraint: "accounts_phone_key"})
			},
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"phone already in use"`,
		},
		{
			Description:        "Invalid Phone Format",
			RequestBody:        `{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.com","password":"Test123!","dateOfBirth":"1990-08-31","phone":"(208) 555-0000"}`,
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"phone must use the E.164 international standard"`,
		},
		{
			Description: "Exactly 18 Account Creator",
			RequestBody: fmt.Sprintf(`{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.com","password":"Testu123!","dateOfBirth":"%v","phone":"+12085550000"}`, time.Now().AddDate(-18, 0, 0).Format(time.DateOnly)),
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM accounts").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				mock.ExpectExec("INSERT INTO accounts (.+) VALUES (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusCreated,
			ExpectedContent:    `"status":"successful"`,
		},
		{
			Description:        "Underage Account Creator",
			RequestBody:        fmt.Sprintf(`{"firstName":"Leagueify","lastName":"Tests","email":"test@leagueify.com","password":"Testu123!","dateOfBirth":"%v","phone":"+12085550000"}`, time.Now().AddDate(-18, 0, 1).Format(time.DateOnly)),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedContent:    `"detail":"must be 18 or older to create an account"`,
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

func TestLoginAccount(t *testing.T) {
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ERROR: '%s' was not expected when creating mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	// Setup Password
	validPassword := "Test123!"
	if err := auth.HashPassword(&validPassword); err != nil {
		t.Fatalf("ERROR: '%s' was not expected when hashing password", err)
	}
	testCases := []struct {
		Description        string
		RequestBody        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description:        "Invalid JSON Payload",
			RequestBody:        `{`,
			ExpectedStatusCode: http.StatusBadRequest,
		},
		{
			Description: "Valid Account Credentials",
			RequestBody: `{"email":"test@leagueify.org","password":"Test123!"}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM accounts WHERE email = (.+)$").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "phone", "date_of_birth", "registration_code", "players", "coach", "volunteer", "apikey", "is_active", "is_admin"}).AddRow("TEST1234", "Leagueify", "Test", "test@leagieuify.org", &validPassword, "+12085551234", "1990-08-31", "", pq.StringArray{}, false, false, "", true, false))
				mock.ExpectExec("UPDATE accounts SET apikey = (.+) WHERE id = (.+)$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "Valid Credentials Inactive Account",
			RequestBody: `{}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM accounts WHERE email = (.+)$").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "phone", "date_of_birth", "registration_code", "players", "coach", "volunteer", "apikey", "is_active", "is_admin"}).AddRow("TEST1234", "Leagueify", "Test", "test@leagieuify.org", &validPassword, "+12085551234", "1990-08-31", "", pq.StringArray{}, false, false, "", true, false))
			},
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Description: "Invalid Account Credentials - Incorrect Password",
			RequestBody: `{}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM accounts WHERE email = (.+)$").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "phone", "date_of_birth", "registration_code", "players", "coach", "volunteer", "apikey", "is_active", "is_admin"}).AddRow("TEST1234", "Leagueify", "Test", "test@leagieuify.org", &validPassword, "+12085551234", "1990-08-31", "", pq.StringArray{}, false, false, "", true, false))
			},
			ExpectedStatusCode: http.StatusUnauthorized,
		},
		{
			Description: "Invalid Account Credentials - Incorrect Email",
			RequestBody: `{}`,
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM accounts WHERE email = (.+)$").WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email", "password", "phone", "date_of_birth", "registration_code", "players", "coach", "volunteer", "apikey", "is_active", "is_admin"}).AddRow("TEST1234", "Leagueify", "Test", "test@leagieuify.org", &validPassword, "+12085551234", "1990-08-31", "", pq.StringArray{}, false, false, "", true, false))
			},
			ExpectedStatusCode: http.StatusUnauthorized,
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
		req := httptest.NewRequest(http.MethodPost, "/api/accounts/login", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.loginAccount(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestLogoutAccount(t *testing.T) {
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ERROR: '%s' was not expected when creating mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description: "Account Logout",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE accounts SET apikey = (.+) WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
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
		req := httptest.NewRequest(http.MethodPost, "/api/accounts/logout", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.logoutAccount(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestVerifyAccount(t *testing.T) {
	// Create Mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("ERROR: '%s' was not expected when creating mock DB", err)
	}
	db := postgres.Postgres{DB: mockDB}
	testCases := []struct {
		Description        string
		ID                 string
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
		ExpectedContent    string
	}{
		{
			Description: "Valid Account ID",
			ID:          "ERCXNX57",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("^UPDATE accounts SET apikey = (.+), is_active = true WHERE id = (.+) AND is_active = false$").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ExpectedStatusCode: http.StatusOK,
			ExpectedContent:    `"apikey":"(.+)"`,
		},
		{
			Description:        "Invalid Account ID",
			ID:                 "12345678",
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedContent:    `"status":"unauthorized"`,
		},
		{
			Description: "Account ID not in Database",
			ID:          "ERCXNX57",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE accounts SET apikey = (.+), is_active = true WHERE id = (.+) AND is_active = false$").WillReturnResult(sqlmock.NewResult(0, 0))
			},
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedContent:    `"status":"unauthorized"`,
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
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/accounts/%s/verify", test.ID), bytes.NewBuffer(nil))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(test.ID)
		// Perform Request
		if assert.NoError(t, api.verifyAccount(c)) {
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
