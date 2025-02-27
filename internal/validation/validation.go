package validation

import (
	"net/mail"
	"strings"
)

// ValidationError represents a field-specific validation error.
type ValidationError struct {
	Field   string
	Message string
}

// ValidateRequired checks if a field is not empty.
func ValidateRequired(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: "This field is required.",
		}
	}
	return nil
}

// ValidateEmail checks if a value is a valid email address.
func ValidateEmail(field, value string) *ValidationError {
	_, err := mail.ParseAddress(value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: "Invalid email format.",
		}
	}
	return nil
}

// ValidateMinLength checks if a string has a minimum length.
func ValidateMinLength(field, value string, minLength int) *ValidationError {
	if len(strings.TrimSpace(value)) < minLength {
		return &ValidationError{
			Field:   field,
			Message: "Must be at least " + string(minLength) + " characters long.",
		}
	}
	return nil
}

// ValidateForm aggregates validation errors into a map of field names to error messages.
func ValidateForm(validations ...*ValidationError) map[string]string {
	errors := make(map[string]string)
	for _, v := range validations {
		if v != nil {
			errors[v.Field] = v.Message
		}
	}
	return errors
}
