package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Leagueify/api/internal/database/postgres"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestListSports(t *testing.T) {
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
		Mock               func(mock sqlmock.Sqlmock)
		ExpectedStatusCode int
	}{
		{
			Description: "Return Sports",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM sports").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "Football").AddRow("2", "Hockey"))
			},
			ExpectedStatusCode: http.StatusOK,
		},
		{
			Description: "No Rows Returned",
			Mock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT \\* FROM sports").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
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
		req := httptest.NewRequest(http.MethodPost, "/api/sports", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Perform Request
		if assert.NoError(t, api.listSports(c)) {
			// Assert Status Code
			assert.Equal(t, test.ExpectedStatusCode, rec.Code)
			// Validate Response Body
			assert.NoError(t, err)
		}
		// Assert All Expectations Met
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}
