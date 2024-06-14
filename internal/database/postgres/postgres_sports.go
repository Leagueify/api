package postgres

import (
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
)

// sport functions
func (p Postgres) GetSports() ([]model.Sport, error) {
	sports := []model.Sport{}

	rows, err := p.DB.Query(`SELECT * FROM sports`)
	if err != nil {
		return sports, err
	}
	defer rows.Close()
	for rows.Next() {
		var sport model.Sport
		if err := rows.Scan(
			&sport.ID,
			&sport.Name,
		); err != nil {
			return sports, err
		}
		sport.ID = util.ReturnSignedToken(sport.ID)
		sports = append(sports, sport)
	}

	return sports, nil
}

func (p Postgres) GetSportByID(sportID string) (model.Sport, error) {
	var sport model.Sport

	if err := p.DB.QueryRow(`
		SELECT * FROM sports WHERE id = $1
	`, sportID[:len(sportID)-1]).Scan(
		&sport.ID,
		&sport.Name,
	); err != nil {
		return sport, err
	}
	return sport, nil
}
