package model

type (
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
)
