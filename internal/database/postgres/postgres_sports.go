package postgres

import (
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/getsentry/sentry-go"
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
			sentry.CaptureException(err)
			return sports, err
		}
		sport.ID = util.ReturnSignedToken(sport.ID)
		sports = append(sports, sport)
	}

	return sports, nil
}
