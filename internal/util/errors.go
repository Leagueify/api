package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)

func HandleError(err error) string {
	switch errType := err.(type) {
	case *pq.Error:
		return postgresErrors(errType)
	case validator.ValidationErrors:
		return validationErrors(errType)
	default:
		return fmt.Sprintf("unknown error: %v :: %v", reflect.TypeOf(err), err.Error())
	}
}

func postgresErrors(err *pq.Error) string {
	if err.Code.Name() == "unique_violation" {
		key := strings.Split(err.Constraint, "_")[1]
		return fmt.Sprintf("%v already in use", key)
	}
	return err.Code.Name()
}

func validationErrors(validationErrors validator.ValidationErrors) string {
	var missingFields []string
	for _, err := range validationErrors {
		if err.Tag() == "required" {
			missingFields = append(missingFields, err.Field())
		}
		if err.Tag() == "e164" {
			return "phone must use the E.164 international standard"
		}
		if err.Tag() == "email" {
			return "invalid email"
		}
	}
	if len(missingFields) != 0 {
		return fmt.Sprintf("missing required field(s): %v", missingFields)
	}
	return fmt.Sprintf("validation error: %v", validationErrors)
}
