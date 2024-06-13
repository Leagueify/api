package postgres

import (
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
)

func (p Postgres) CreatePositions(positions model.PositionCreation) error {
	tx, err := p.DB.Begin()
	if err != nil {
		sentry.CaptureException(err)
		return err
	}
	defer tx.Rollback()
	for _, position := range positions.Positions {
		_, err := tx.Exec(`
			INSERT INTO positions (id, name)
			VALUES ($1, $2)
		`, util.SignedToken(6), position)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (p Postgres) GetAllPositions() ([]model.Position, error) {
	positions := []model.Position{}

	rows, err := p.DB.Query(`SELECT * FROM positions`)
	if err != nil {
		return positions, err
	}
	defer rows.Close()
	for rows.Next() {
		var position model.Position
		if err := rows.Scan(
			&position.ID,
			&position.Name,
		); err != nil {
			return positions, err
		}
		position.ID = util.ReturnSignedToken(position.ID)
		positions = append(positions, position)
	}

	return positions, nil
}

func (p Postgres) GetTotalPositions() (int, error) {
	var totalPositions int

	row := p.DB.QueryRow(`SELECT COUNT(*) FROM positions`)
	if err := row.Scan(&totalPositions); err != nil {
		return 0, err
	}
	return totalPositions, nil
}
