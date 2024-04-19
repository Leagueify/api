package util

import (
	"testing"
)

func TestSignedToken(t *testing.T) {
	testCases := []struct {
		Description     string
		SubmittedLength int
		ExpectedLength  int
	}{
		{
			Description:     "Valid Submitted Length",
			SubmittedLength: 2,
			ExpectedLength:  2,
		},
		{
			Description:     "Zero Submitted Length",
			SubmittedLength: 0,
			ExpectedLength:  0,
		},
		{
			Description:     "Invalid Submitted Positive Length",
			SubmittedLength: 1,
			ExpectedLength:  0,
		},
		{
			Description:     "Invalid Submitted Negative Length",
			SubmittedLength: -1,
			ExpectedLength:  0,
		},
	}

	for _, test := range testCases {
		result := SignedToken(test.SubmittedLength)
		if len(result) != test.ExpectedLength {
			t.Errorf(
				`%v: Expected a length of %v but recieved %v.`,
				test.Description, test.ExpectedLength, len(result),
			)
		}
	}
}

func TestUnsignedToken(t *testing.T) {
	testCases := []struct {
		Description    string
		Length         int
		ExpectedLength int
	}{
		{
			Description:    "Valid Submitted Length",
			Length:         1,
			ExpectedLength: 1,
		},
		{
			Description:    "Zero Submitted Length",
			Length:         0,
			ExpectedLength: 0,
		},
		{
			Description:    "Invalid Submitted Length",
			Length:         -1,
			ExpectedLength: 0,
		},
	}

	for _, test := range testCases {
		result := UnsignedToken(test.Length)
		if len(result) != test.ExpectedLength {
			t.Errorf(
				`%v: Expected a length of %v but received %v.`,
				test.Description, test.ExpectedLength, len(result),
			)
		}
	}
}

func TestVerifyToken(t *testing.T) {
	testCases := []struct {
		Description    string
		Token          string
		ExpectedResult bool
	}{
		{
			Description:    "Valid Token",
			Token:          "KJV1XK3",
			ExpectedResult: true,
		},
		{
			Description:    "Invalid Token",
			Token:          "A1B2C3D4",
			ExpectedResult: false,
		},
		{
			Description:    "Empty Token",
			Token:          "",
			ExpectedResult: false,
		},
	}

	for _, test := range testCases {
		result := VerifyToken(test.Token)
		if result != test.ExpectedResult {
			t.Errorf(
				`%v: Expected %v but received %v with value "%v".`,
				test.Description, test.ExpectedResult, result, test.Token,
			)
		}
	}
}

func TestVerifySignedToken(t *testing.T) {
	testCases := []struct {
		Description    string
		Length         int
		ExpectedResult bool
	}{
		{
			Description:    "Valid Submitted Length",
			Length:         2,
			ExpectedResult: true,
		},
		{
			Description:    "Zero Submitted Length",
			Length:         0,
			ExpectedResult: false,
		},
		{
			Description:    "Invalid Submitted Negative Length",
			Length:         -1,
			ExpectedResult: false,
		},
		{
			Description:    "Invalid Submitted Negative Length",
			Length:         -1,
			ExpectedResult: false,
		},
	}

	for _, test := range testCases {
		signedToken := SignedToken(test.Length)
		result := VerifyToken(signedToken)
		if result != test.ExpectedResult {
			t.Errorf(
				`%v: Expected %v but received %v with a submitted length of "%v".`,
				test.Description, test.ExpectedResult, result, test.Length,
			)
		}
	}
}
