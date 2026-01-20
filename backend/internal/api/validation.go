package api

import (
	"github.com/go-playground/validator/v10"
)

// Validate is the validator instance used for request validation.
var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// ValidateStruct validates a struct using the configured validator.
func ValidateStruct(s interface{}) error {
	return Validate.Struct(s)
}
