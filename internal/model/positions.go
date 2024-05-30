package model

type (
	PositionCreation struct {
		Positions []string `json:"positions"`
	}

	Position struct {
		ID   string
		Name string
	}
)
