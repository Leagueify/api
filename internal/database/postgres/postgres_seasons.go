package postgres

import (
	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
)

func (p Postgres) CreateSeason(season model.Season) error {
	if _, err := p.DB.Exec(`
		INSERT INTO seasons (
			id, name, start_date, end_date, registration_opens,
			registration_closes
		)
		VALUES (
			$1, $2, $3, $4, $5, $6
		)`,
		season.ID[:len(season.ID)-1], season.Name, season.StartDate,
		season.EndDate, season.RegistrationOpens,
		season.RegistrationCloses,
	); err != nil {
		return err
	}
	return nil
}

func (p Postgres) ListSeasons() ([]model.SeasonList, error) {
	seasons := []model.SeasonList{}

	rows, err := p.DB.Query(`SELECT id, name FROM seasons`)
	if err != nil {
		return seasons, err
	}
	defer rows.Close()
	for rows.Next() {
		var season model.SeasonList
		if err := rows.Scan(
			&season.ID,
			&season.Name,
		); err != nil {
			return seasons, err
		}
		season.ID = util.ReturnSignedToken(season.ID)
		seasons = append(seasons, season)
	}

	return seasons, nil
}
