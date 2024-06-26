package model

type Season struct {
	ID                 string
	Name               string `json:"name" validate:"required"`
	StartDate          string `json:"startDate" validate:"required"`
	EndDate            string `json:"endDate" validate:"required"`
	RegistrationOpens  string `json:"registrationOpens" validate:"required"`
	RegistrationCloses string `json:"registrationCloses" validate:"required"`
}
