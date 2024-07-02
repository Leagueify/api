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

func (p Postgres) GetSeason(seasonID string) (model.Season, error) {
	season := model.Season{}

	if err := p.DB.QueryRow(`
		SELECT * FROM seasons WHERE id = $1
	`, seasonID[:len(seasonID)-1]).Scan(
		&season.ID,
		&season.Name,
		&season.StartDate,
		&season.EndDate,
		&season.RegistrationOpens,
		&season.RegistrationCloses,
	); err != nil {
		return season, err
	}
	season.ID = util.ReturnSignedToken(season.ID)

	return season, nil
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

func (p Postgres) UpdateSeason(season model.Season) error {
	if _, err := p.DB.Exec(`
		UPDATE seasons
		SET name = $1, start_date = $2, end_date = $3,
			registration_opens = $4, registration_closes = $5
		WHERE id = $6
	`,
		season.Name, season.StartDate, season.EndDate,
		season.RegistrationOpens, season.RegistrationCloses,
		season.ID[:len(season.ID)-1],
	); err != nil {
		return err
	}
	return nil
}
