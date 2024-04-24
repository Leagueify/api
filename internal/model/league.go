package model

type LeagueCreation struct {
	ID          string
	Name        string `json:"name" validate:"required,min=3"`
	SportID     int    `json:"sportID" validate:"required"`
	MasterAdmin string
}
