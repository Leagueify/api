package util

import (
	"testing"
)

func TestCalculateAge(t *testing.T) {
	testCases := []struct {
		Description    string
		SubmittedAge   string
		ComparisonAge  string
		ExpectedResult int
	}{
		{
			Description:    "Positive Non-Zero Difference",
			SubmittedAge:   "1990-08-31",
			ComparisonAge:  "2011-08-31",
			ExpectedResult: 21,
		},
		{
			Description:    "Positive Zero Difference - Updated Month",
			SubmittedAge:   "2020-01-01",
			ComparisonAge:  "2020-02-01",
			ExpectedResult: 0,
		},
		{
			Description:    "Positive Zero Difference - Updated Day",
			SubmittedAge:   "2024-01-01",
			ComparisonAge:  "2024-01-02",
			ExpectedResult: 0,
		},
	}

	for _, test := range testCases {
		result, _ := CalculateAge(test.SubmittedAge, test.ComparisonAge)
		if result != test.ExpectedResult {
			t.Errorf(
				`%v: Expected %v but received %v. Comparing %v and %v.`,
				test.Description, test.ExpectedResult, result,
				test.SubmittedAge, test.ComparisonAge,
			)
		}
	}
}
