package postgres

import "github.com/Leagueify/api/internal/model"

func (p Postgres) CreateLeague(league model.LeagueCreation) error {
	if _, err := p.DB.Exec(`
		INSERT INTO leagues (id, name, sport_id, master_admin)
		VALUES ($1, $2, $3, $4)
		`, league.ID[:len(league.ID)-1], league.Name, league.SportID,
		league.MasterAdmin,
	); err != nil {
		return err
	}
	return nil
}

func (p Postgres) GetTotalLeagues() (int, error) {
	var totalLeagues int

	row := p.DB.QueryRow(`SELECT COUNT(*) FROM leagues`)
	if err := row.Scan(&totalLeagues); err != nil {
		return 0, err
	}
	return totalLeagues, nil
}
