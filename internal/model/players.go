package model

type (
	Player struct {
		ID           string
		FirstName    string `json:"firstName" validate:"required"`
		LastName     string `json:"lastName" validate:"required"`
		DateOfBirth  string `json:"dateOfBirth" validate:"required"`
		Position     string `json:"position" validate:"required"`
		Team         string
		Division     string
		IsRegistered bool
	}

	PlayerCreation struct {
		Players []Player `json:"players" validate:"required"`
	}

	PlayerRegistration struct {
		Players []string `json:"players" validate:"required"`
	}
)
