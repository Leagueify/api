package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

func (api *API) Players(e *echo.Group) {
	e.POST("/players", api.requiresAuth(api.createPlayer))
}

func (api *API) createPlayer(c echo.Context) error {
	payload := model.PlayerCreation{}
	// Bind payload to player model
	if err := c.Bind(&payload); err != nil {
		sentry.CaptureException(err)
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "invalid json payload",
			},
		)
	}
	// Validate payload against model
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Retrieve league positions
	var leaguePositions []string
	rows, err := api.DB.Query(`
		SELECT name FROM positions
	`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			map[string]string{
				"status": "internal server error",
				"detail": util.HandleError(err),
			},
		)
	}
	defer rows.Close()
	for rows.Next() {
		var leaguePosition string
		err := rows.Scan(
			&leaguePosition,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError,
				map[string]string{
					"status": "internal server error",
					"detail": util.HandleError(err),
				},
			)
		}
		leaguePositions = append(leaguePositions, leaguePosition)
	}
	// Check for existing players
	playerIDs := api.Account.Players
	// Begin Transaction
	tx, err := api.DB.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			map[string]string{
				"status": "internal server error",
				"detail": util.HandleError(err),
			},
		)
	}
	defer tx.Rollback()
	for _, player := range payload.Players {
		// Validate Players
		if err := c.Validate(player); err != nil {
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
					"detail": util.HandleError(err),
				},
			)
		}
		// Validate player position in positions
		var validPosition = false
		for _, position := range leaguePositions {
			if position == player.Position {
				validPosition = true
				break
			}
		}
		if !validPosition {
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
					"detail": "invalid position",
				},
			)
		}
		player.ID = util.SignedToken(10)
		playerIDs = append(playerIDs, player.ID[:len(player.ID)-1])
		// Add Player to players table
		if _, err := tx.Exec(`
			INSERT INTO players (
				id, first_name, last_name, date_of_birth, position,
				team, division, is_registered
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8
			)`,
			player.ID[:len(player.ID)-1], player.FirstName, player.LastName,
			player.DateOfBirth, player.Position, "", "", false,
		); err != nil {
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
					"detail": util.HandleError(err),
				},
			)
		}
	}
	if _, err := tx.Exec(`
		UPDATE accounts SET player_ids = $1 WHERE id = $2
	`, playerIDs, api.Account.ID); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
		},
	)
}
