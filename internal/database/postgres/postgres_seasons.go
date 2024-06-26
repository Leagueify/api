package postgres

import (
	"github.com/Leagueify/api/internal/model"
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
