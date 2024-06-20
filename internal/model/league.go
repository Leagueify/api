package model

type LeagueCreation struct {
	ID          string
	Name        string `json:"name" validate:"required,min=3"`
	SportID     string `json:"sportID" validate:"required"`
	MasterAdmin string
}
