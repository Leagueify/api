package util

import (
	"errors"
	"testing"
)

func TestComparePasswords(t *testing.T) {
	testCases := []struct {
		Description      string
		ProvidedPassword string
		StoredPassword   string
		ExpectedResult   bool
	}{
		{
			Description:      "Matching Passwords",
			ProvidedPassword: "Testu123!",
			StoredPassword:   "$2a$12$ZkxvGZrKKGcertas0oGoieM/TkMIigF4kZYbahwYbPeau1fk.EnMO",
			ExpectedResult:   true,
		},
		{
			Description:      "Mismatch Password",
			ProvidedPassword: "Test123!",
			StoredPassword:   "$2a$12$ZkxvGZrKKGcertas0oGoieM/TkMIigF4kZYbahwYbPeau1fk.EnMO",
			ExpectedResult:   false,
		},
		{
			Description:      "Empty Provided Password",
			ProvidedPassword: "",
			StoredPassword:   "$2a$12$ZkxvGZrKKGcertas0oGoieM/TkMIigF4kZYbahwYbPeau1fk.EnMO",
			ExpectedResult:   false,
		},
		{
			Description:      "Empty Stored Password",
			ProvidedPassword: "Test123!",
			StoredPassword:   "",
			ExpectedResult:   false,
		},
		{
			Description:      "Empty Provided and Stored Passwords",
			ProvidedPassword: "",
			StoredPassword:   "",
			ExpectedResult:   false,
		},
	}

	for _, test := range testCases {
		result := ComparePasswords(test.ProvidedPassword, test.StoredPassword)
		if result != test.ExpectedResult {
			t.Errorf(
				"%v: Expected %v received %v for %v and %v",
				test.Description, test.ExpectedResult, result,
				test.ProvidedPassword, test.StoredPassword,
			)
		}
	}
}

func TestHashPassword(t *testing.T) {
	testCases := []struct {
		Description       string
		SubmittedPassword string
		Error             error
	}{
		{
			Description:       "Valid Password",
			SubmittedPassword: "Test123!",
			Error:             nil,
		},
		{
			Description:       "Invalid Password Too Short",
			SubmittedPassword: "Test12!",
			Error:             errors.New("password must be at least 8 characters"),
		},
		{
			Description:       "Invalid Password Too Ling",
			SubmittedPassword: "!Tt45678901234567890123456789012345678901234567890123456789012345",
			Error:             errors.New("password must be at most 64 characters"),
		},
		{
			Description:       "Invalid Password Missing Uppercase Character",
			SubmittedPassword: "testu123!",
			Error:             errors.New("missing uppercase character"),
		},
		{
			Description:       "Invalid Password Missing Lowercase Character",
			SubmittedPassword: "TEST123!",
			Error:             errors.New("missing lowercase character"),
		},
		{
			Description:       "Invalid Password Missing Numeric Character",
			SubmittedPassword: "TESTTEST!",
			Error:             errors.New("missing numeric character"),
		},
		{
			Description:       "Invalid Password Missing Special Character",
			SubmittedPassword: "Test1234",
			Error:             errors.New("missing special character"),
		},
	}

	for _, test := range testCases {
		err := HashPassword(&test.SubmittedPassword)
		if err != test.Error && err.Error() != test.Error.Error() {
			t.Errorf(
				`%v: Expected %v received %v with password %v`,
				test.Description, test.Error, err.Error(), test.SubmittedPassword,
			)
		}
	}
}
