package model

import (
	"github.com/lib/pq"
)

type (
	Account struct {
		ID               string
		FirstName        string
		LastName         string
		Email            string
		Password         string
		Phone            string
		DateOfBirth      string
		RegistrationCode string
		Players          pq.StringArray
		Coach            bool
		Volunteer        bool
		APIKey           string
		IsActive         bool
		IsAdmin          bool
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
		IsActive    bool
		IsAdmin     bool
	}

	AccountLogin struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
)
