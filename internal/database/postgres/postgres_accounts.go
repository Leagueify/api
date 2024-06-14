package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Leagueify/api/internal/model"
	"github.com/lib/pq"
)

func (p Postgres) ActivateAccount(accountID, apikey string) error {
	results, err := p.DB.Exec(`
		UPDATE accounts SET apikey = $1, is_active = true
		WHERE id = $2 AND is_active = false
	`, apikey[:len(apikey)-1], accountID[:len(accountID)-1])
	if err != nil {
		return err
	}

	rows, err := results.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("Account update failed")
	}

	return nil
}

func (p Postgres) CreateAccount(account model.AccountCreation) error {
	if _, err := p.DB.Exec(`
		INSERT INTO accounts (
			id, first_name, last_name, email, password, phone,
			date_of_birth, registration_code, player_ids, coach,
			volunteer, apikey, is_active, is_admin
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)`,
		account.ID[:len(account.ID)-1], account.FirstName,
		account.LastName, account.Email, account.Password,
		account.Phone, account.DateOfBirth, "", "{}", account.Coach,
		account.Volunteer, "", account.IsActive, account.IsAdmin,
	); err != nil {
		return err
	}

	return nil
}

func (p Postgres) GetAccountByAPIKey(apikey string) (model.Account, error) {
	account := model.Account{}

	if err := p.DB.QueryRow(`
		SELECT * FROM accounts WHERE apikey = $1
	`, apikey[:len(apikey)-1]).Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Email,
		&account.Password,
		&account.Phone,
		&account.DateOfBirth,
		&account.RegistrationCode,
		&account.Players,
		&account.Coach,
		&account.Volunteer,
		&account.APIKey,
		&account.IsActive,
		&account.IsAdmin,
	); err != nil {
		return account, err
	}

	return account, nil
}

func (p Postgres) GetAccountByEmail(email string) (model.Account, error) {
	account := model.Account{}

	if err := p.DB.QueryRow(`
		SELECT * FROM accounts WHERE email = $1
	`, email).Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Email,
		&account.Password,
		&account.Phone,
		&account.DateOfBirth,
		&account.RegistrationCode,
		&account.Players,
		&account.Coach,
		&account.Volunteer,
		&account.APIKey,
		&account.IsActive,
		&account.IsAdmin,
	); err != nil {
		return account, err
	}

	return account, nil
}

func (p Postgres) GetTotalAccounts() (int, error) {
	var totalAccounts int

	row := p.DB.QueryRow(`SELECT COUNT(*) FROM accounts`)
	if err := row.Scan(&totalAccounts); err != nil {
		return 0, err
	}

	return totalAccounts, nil
}

func (p Postgres) SetAPIKey(apikey, accountID string) error {
	if _, err := p.DB.Exec(`
		UPDATE accounts SET apikey = $1 WHERE id = $2
	`, apikey[:len(apikey)-1], accountID); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p Postgres) SetPlayerIDs(playerIDs *pq.StringArray, accountID string, tx *sql.Tx) error {
	if _, err := tx.Exec(`
		UPDATE accounts SET player_ids = $1 WHERE id = $2
	`, playerIDs, accountID); err != nil {
		return err
	}
	return nil
}

func (p Postgres) SetRegistrationCode(tx *sql.Tx, code, accountID string) error {
	if _, err := tx.Exec(`
		UPDATE accounts SET registration_code = $1 WHERE id = $2
	`, code, accountID); err != nil {
		return err
	}
	return nil
}

func (p Postgres) UnsetAPIKey(accountID string) error {
	if _, err := p.DB.Exec(`
		UPDATE accounts SET apikey = "" WHERE id = $1
	`, accountID); err != nil {
		return err
	}

	return nil
}
