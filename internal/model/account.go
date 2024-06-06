package model

type (
	Account struct {
		ID          string
		FirstName   string
		LastName    string
		Email       string
		Password    string
		Phone       string
		DateOfBirth string
		Coach       bool
		Volunteer   bool
		APIKey      string
		IsActive    bool
		IsAdmin     bool
	}

	AccountCreation struct {
		ID          string
		FirstName   string `json:"firstName" validate:"required"`
		LastName    string `json:"lastName" validate:"required"`
		Email       string `json:"email" validate:"required,email"`
		Password    string `json:"password" validate:"required"`
		Phone       string `json:"phone" validate:"required,e164"`
		DateOfBirth string `json:"dateOfBirth" validate:"required"`
		Coach       bool   `json:"coach"`
		Volunteer   bool   `json:"volunteer"`
	}

	AccountLogin struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
)
