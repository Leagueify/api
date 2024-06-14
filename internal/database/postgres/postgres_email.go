package postgres

import "github.com/Leagueify/api/internal/model"

func (p Postgres) CreateEmailConfig(emailConfig model.EmailConfig) error {
	if _, err := p.DB.Exec(`
		INSERT INTO email (
			id, email, smtp_host, smtp_port, smtp_user, smtp_pass,
			is_active, has_error
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		`, emailConfig.ID[:len(emailConfig.ID)-1], emailConfig.Email,
		emailConfig.SMTPHost, emailConfig.SMTPPort, emailConfig.SMTPUser,
		emailConfig.SMTPPass, emailConfig.IsEnabled, emailConfig.HasError,
	); err != nil {
		return err
	}
	return nil
}

func (p Postgres) GetTotalEmailConfigs() (int, error) {
	var totalEmailConfigs int

	row := p.DB.QueryRow(`SELECT COUNT(*) FROM email`)
	if err := row.Scan(&totalEmailConfigs); err != nil {
		return 0, err
	}

	return totalEmailConfigs, nil
}
