package postgres

import (
	"database/sql"

	"github.com/Leagueify/api/internal/model"
	"github.com/lib/pq"
)

func (p Postgres) CreateRegistration(tx *sql.Tx, registration model.Registration) error {
	if _, err := tx.Exec(`
		INSERT INTO registrations (
			id, player_ids, amount_due, amount_paid
		)
		VALUES (
			$1, $2, $3, $4
		)`,
		registration.ID, registration.PlayerIDs, registration.AmountDue,
		registration.AmountPaid,
	); err != nil {
		return err
	}
	return nil
}

func (p Postgres) GetRegistration(tx *sql.Tx, registrationID string) (pq.StringArray, error) {
	var registeredPlayers pq.StringArray

	if err := tx.QueryRow(`
		SELECT player_ids FROM registrations WHERE id = $1
	`, registrationID).Scan(&registeredPlayers); err != nil {
		return registeredPlayers, err
	}
	return registeredPlayers, nil
}

func (p Postgres) SetRegistration(tx *sql.Tx, playerIDs pq.StringArray, registrationID string) error {
	if _, err := tx.Exec(`
		UPDATE registrations SET player_ids = $1 WHERE id = $2
	`); err != nil {
		return err
	}
	return nil
}
