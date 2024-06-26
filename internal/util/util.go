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

func IsValidDateRange(start, end string) (bool, error) {
	startDate, err := time.Parse(time.DateOnly, start)
	if err != nil {
		return false, err
	}
	endDate, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return false, err
	}
	if startDate.UnixMilli() > endDate.UnixMilli() {
		return false, nil
	}
	return true, nil
}

func IsInArray(players pq.StringArray, playerID string) bool {
	for _, player := range players {
		if playerID == player {
			return true
		}
	}
	return false
}
