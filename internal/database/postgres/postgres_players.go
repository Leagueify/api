package postgres

import (
	"database/sql"

	"github.com/Leagueify/api/internal/model"
)

func (p Postgres) CreatePlayer(player model.Player, tx *sql.Tx) error {
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
		return err
	}
	return nil
}

func (p Postgres) DeletePlayer(playerID string, tx *sql.Tx) error {
	if _, err := tx.Exec(`
		DELETE FROM players WHERE id = $1
	`, playerID); err != nil {
		return err
	}
	return nil
}

func (p Postgres) GetPlayer(playerID string) (model.Player, error) {
	var player model.Player

	if err := p.DB.QueryRow(`
		SELECT * FROM players WHERE id = $1
	`, playerID).Scan(
		&player.ID,
		&player.FirstName,
		&player.LastName,
		&player.DateOfBirth,
		&player.Position,
		&player.Team,
		&player.Division,
		&player.IsRegistered,
	); err != nil {
		return player, err
	}

	return player, nil
}

func (p Postgres) RegisterPlayer(tx *sql.Tx, playerID string) error {
	if _, err := tx.Exec(`
		UPDATE players SET is_registered = true WHERE id = $1
	`, playerID); err != nil {
		return err
	}
	return nil
}
