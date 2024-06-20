package api

import (
	"net/http"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

func (api *API) Players(e *echo.Group) {
	e.GET("/players", api.requiresAuth(api.getPlayers))
	e.POST("/players", api.requiresAuth(api.createPlayer))
	e.DELETE("/players/:id", api.requiresAuth(api.deletePlayer))
	e.GET("/players/:id", api.requiresAuth(api.getPlayer))
	e.POST("/players/register", api.requiresAuth(api.registerPlayer))
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
	// Verify players length
	if len(payload.Players) < 1 {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "payload contains no players",
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

func (api *API) deletePlayer(c echo.Context) error {
	playerID := c.Param("id")
	if !util.VerifyToken(playerID) {
		return c.JSON(http.StatusNoContent, nil)
	}
	// Remove checksum from playerID
	playerID = playerID[:len(playerID)-1]
	// Get account players
	players := api.Account.Players
	// Delete playerID if in players
	for playerIndex, player := range players {
		if player == playerID {
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
			// Delete Player Record
			if _, err := tx.Exec(`
				DELETE FROM players WHERE id = $1
			`, playerID); err != nil {
				return c.JSON(http.StatusBadRequest,
					map[string]string{
						"status": "bad request",
						"detail": util.HandleError(err),
					},
				)
			}
			// Remove playerID from account Players
			players = append(players[:playerIndex], players[playerIndex+1:]...)
			if _, err := tx.Exec(`
				UPDATE accounts SET player_ids = $1 WHERE id = $2
			`, players, api.Account.ID); err != nil {
				return c.JSON(http.StatusBadRequest,
					map[string]string{
						"status": "bad request",
						"detail": util.HandleError(err),
					},
				)

			}
			// Commit Transaction
			if err := tx.Commit(); err != nil {
				return c.JSON(http.StatusBadRequest,
					map[string]string{
						"status": "bad request",
						"detail": util.HandleError(err),
					},
				)
			}
		}
	}
	return c.JSON(http.StatusNoContent, nil)
}

func (api *API) getPlayer(c echo.Context) error {
	playerID := c.Param("id")
	if !util.VerifyToken(playerID) {
		return c.JSON(http.StatusNotFound,
			map[string]string{
				"status": "not found",
			},
		)
	}
	// Remove checksum from playerID
	playerID = playerID[:len(playerID)-1]
	// Get account players
	players := api.Account.Players
	var playerInfo model.Player
	for _, player := range players {
		if player == playerID {
			if err := api.DB.QueryRow(`
				SELECT * FROM players WHERE id = $1
			`, playerID).Scan(
				&playerInfo.ID,
				&playerInfo.FirstName,
				&playerInfo.LastName,
				&playerInfo.DateOfBirth,
				&playerInfo.Position,
				&playerInfo.Team,
				&playerInfo.Division,
				&playerInfo.IsRegistered,
			); err != nil {
				return c.JSON(http.StatusBadRequest,
					map[string]string{
						"status": "bad request",
						"detail": util.HandleError(err),
					},
				)
			}
			playerInfo.ID = util.ReturnSignedToken(playerInfo.ID)
			return c.JSON(http.StatusOK, playerInfo)
		}
	}
	return c.JSON(http.StatusNotFound,
		map[string]string{
			"status": "not found",
		},
	)
}

func (api *API) getPlayers(c echo.Context) error {
	players := api.Account.Players
	if len(players) == 0 {
		return c.JSON(http.StatusNotFound,
			map[string]string{
				"status": "not found",
			},
		)
	}
	for index, player := range players {
		players[index] = util.ReturnSignedToken(player)
	}
	return c.JSON(http.StatusOK,
		map[string]pq.StringArray{
			"players": players,
		},
	)
}

func (api *API) registerPlayer(c echo.Context) error {
	payload := model.PlayerRegistration{}
	// Bind payload to model
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "invalid json payload",
			},
		)
	}
	// Verify payload against model
	if err := c.Validate(payload); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Verify players length
	if len(payload.Players) < 1 {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": "payload contains no players",
			},
		)
	}
	// Generate Players to register
	var registerPlayers pq.StringArray
	// Begin Transaction
	tx, err := api.DB.Begin()
	defer tx.Rollback()
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			map[string]string{
				"status": "internal server error",
				"detail": util.HandleError(err),
			},
		)
	}
	// Generate Registration Code
	updateRegistration := false
	registrationCode := util.SignedToken(10)
	if api.Account.RegistrationCode != "" {
		updateRegistration = true
		registrationCode = util.ReturnSignedToken(api.Account.RegistrationCode)
	}
	storedRegistrationCode := registrationCode[:len(registrationCode)-1]
	if _, err := tx.Exec(`
		UPDATE accounts SET registration_code = $1 WHERE id = $2
	`, storedRegistrationCode, api.Account.ID); err != nil {
		return c.JSON(http.StatusBadRequest,
			map[string]string{
				"status": "bad request",
				"detail": util.HandleError(err),
			},
		)
	}
	// Add players to registration
	for _, player := range payload.Players {
		if !util.VerifyToken(player) {
			return c.JSON(http.StatusNotFound,
				map[string]string{
					"status": "not found",
				},
			)
		}
		// Update Player ID
		player = player[:len(player)-1]
		// Validate player in Account
		if !util.IsInArray(api.Account.Players, player) {
			return c.JSON(http.StatusNotFound,
				map[string]string{
					"status": "not found",
				},
			)
		}
		// Add Player to registerPlayers array
		registerPlayers = append(registerPlayers, player)
		if _, err := tx.Exec(`
			UPDATE players SET is_registered = true WHERE id = $1
		`, player); err != nil {
			return c.JSON(http.StatusInternalServerError,
				map[string]string{
					"status": "internal server error",
					"detail": util.HandleError(err),
				},
			)
		}
	}
	if updateRegistration {
		var registeredPlayers pq.StringArray
		if err := tx.QueryRow(`
				SELECT player_ids FROM registrations WHERE id = $1
			`, storedRegistrationCode).Scan(&registeredPlayers); err != nil {
			return c.JSON(http.StatusBadRequest,
				map[string]string{
					"status": "bad request",
					"detail": util.HandleError(err),
				},
			)
		}
		for _, player := range registerPlayers {
			if !util.IsInArray(registeredPlayers, player) {
				registeredPlayers = append(registeredPlayers, player)
			}
		}
		if _, err := tx.Exec(`
			UPDATE registrations SET player_ids = $1 WHERE id = $2
		`, registeredPlayers, storedRegistrationCode); err != nil {
			return c.JSON(http.StatusInternalServerError,
				map[string]string{
					"status": "internal server error",
					"detail": util.HandleError(err),
				},
			)
		}
	}
	if !updateRegistration {
		if _, err := tx.Exec(`
			INSERT INTO registrations (
				id, player_ids, amount_due, amount_paid
			)
			VALUES (
				$1, $2, $3, $4
			)`,
			storedRegistrationCode, registerPlayers, 0, 0,
		); err != nil {
			return c.JSON(http.StatusInternalServerError,
				map[string]string{
					"status": "internal server error",
					"detail": util.HandleError(err),
				},
			)
		}
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
