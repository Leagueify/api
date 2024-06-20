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
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}
	// Validate payload against model
	if err := c.Validate(payload); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Verify players length
	if len(payload.Players) < 1 {
		return util.SendStatus(http.StatusBadRequest, c, "payload contains no players")
	}
	// retrieve league positions
	leaguePositions, err := api.DB.GetAllPositions()
	if err != nil {
		return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
	}
	// check for existing players
	playerIDs := api.Account.Players
	// begin transaction
	tx, err := api.DB.BeginTransaction()
	if err != nil {
		return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
	}
	defer tx.Rollback()
	for _, player := range payload.Players {
		// validate players
		if err := c.Validate(player); err != nil {
			return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
		}
		// validate player position in positions
		var validPosition = false
		for _, position := range leaguePositions {
			if position.Name == player.Position {
				validPosition = true
				break
			}
		}
		if !validPosition {
			return util.SendStatus(http.StatusBadRequest, c, "invalid position")
		}
		player.ID = util.SignedToken(10)
		playerIDs = append(playerIDs, player.ID[:len(player.ID)-1])
		if err := api.DB.CreatePlayer(player, tx); err != nil {
			return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
		}
	}
	if err := api.DB.SetPlayerIDs(&playerIDs, api.Account.ID, tx); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	if err := tx.Commit(); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	return c.JSON(http.StatusCreated,
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
			tx, err := api.DB.BeginTransaction()
			if err != nil {
				return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
			}
			// delete player record
			if err := api.DB.DeletePlayer(playerID, tx); err != nil {
				return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
			}
			// Remove playerID from account Players
			players = append(players[:playerIndex], players[playerIndex+1:]...)
			if err := api.DB.SetPlayerIDs(&players, api.Account.ID, tx); err != nil {
				return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
			}
			// Commit Transaction
			if err := tx.Commit(); err != nil {
				return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
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
	for _, player := range players {
		if player == playerID {
			playerInfo, err := api.DB.GetPlayer(player)
			if err != nil {
				return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
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
		return util.SendStatus(http.StatusNotFound, c, "")
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
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}
	// Verify payload against model
	if err := c.Validate(payload); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Verify players length
	if len(payload.Players) < 1 {
		return util.SendStatus(http.StatusBadRequest, c, "payload contains no players")
	}
	// Generate Players to register
	var registerPlayers pq.StringArray
	// Begin Transaction
	tx, err := api.DB.BeginTransaction()
	if err != nil {
		return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
	}
	defer tx.Rollback()
	// Generate Registration Code
	updateRegistration := false
	registrationCode := util.SignedToken(10)
	if api.Account.RegistrationCode != "" {
		updateRegistration = true
		registrationCode = util.ReturnSignedToken(api.Account.RegistrationCode)
	}
	storedRegistrationCode := registrationCode[:len(registrationCode)-1]
	if err := api.DB.SetRegistrationCode(tx, storedRegistrationCode, api.Account.ID); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	// Add players to registration
	for _, player := range payload.Players {
		if !util.VerifyToken(player) {
			return util.SendStatus(http.StatusNotFound, c, "")
		}
		// Update Player ID
		player = player[:len(player)-1]
		// Validate player in Account
		if !util.IsInArray(api.Account.Players, player) {
			return util.SendStatus(http.StatusNotFound, c, "")
		}
		// Add Player to registerPlayers array
		registerPlayers = append(registerPlayers, player)
		if err := api.DB.RegisterPlayer(tx, player); err != nil {
			return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
		}
	}
	if updateRegistration {
		registeredPlayers, err := api.DB.GetRegistration(tx, registrationCode)
		if err != nil {
			return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
		}

		for _, player := range registerPlayers {
			if !util.IsInArray(registeredPlayers, player) {
				registeredPlayers = append(registeredPlayers, player)
			}
		}

		if err := api.DB.SetRegistration(tx, registeredPlayers, registrationCode); err != nil {
			return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
		}
	}
	if !updateRegistration {
		registration := model.Registration{
			ID:         storedRegistrationCode,
			PlayerIDs:  registerPlayers,
			AmountDue:  0,
			AmountPaid: 0,
		}
		if err := api.DB.CreateRegistration(tx, registration); err != nil {
			return util.SendStatus(http.StatusInternalServerError, c, util.HandleError(err))
		}
	}
	if err := tx.Commit(); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	return c.JSON(http.StatusOK,
		map[string]string{
			"status": "successful",
		},
	)
}
