package util

import (
	"time"

	"github.com/lib/pq"
)

// TODO: Add Error if Age is < 0
func CalculateAge(submitted, comparison string) (int, error) {
	submittedDate, err := time.Parse(time.DateOnly, submitted)
	if err != nil {
		return 0, err
	}
	comparisonDate, err := time.Parse(time.DateOnly, comparison)
	if err != nil {
		return 0, err
	}
	yearsDiff := comparisonDate.Year() - submittedDate.Year()
	if comparisonDate.Month() < submittedDate.Month() ||
		(comparisonDate.Month() == submittedDate.Month() && comparisonDate.Day() < submittedDate.Day()) {
		yearsDiff--
	}
	return yearsDiff, nil
}

func IsInArray(players pq.StringArray, playerID string) bool {
	for _, player := range players {
		if playerID == player {
			return true
		}
	}
	return false
}
