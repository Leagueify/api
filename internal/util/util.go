package util

import (
	"time"
)

func CalculateAge(submitted, comparison string) (int, error) {
	submittedDate, err := time.Parse(time.DateOnly, submitted)
	if err != nil {
		return 0, err
	}
	currentDate := time.Now()
	yearsDiff := currentDate.Year() - submittedDate.Year()
	if currentDate.Month() < submittedDate.Month() ||
		(currentDate.Month() == submittedDate.Month() && currentDate.Day() < submittedDate.Day()) {
		yearsDiff--
	}
	return yearsDiff, nil
}
