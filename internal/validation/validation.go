package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var (
	// ErrInvalidEmail is returned when an email address is invalid
	ErrInvalidEmail = errors.New("invalid email address")
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	// ErrPasswordTooWeak is returned when password doesn't meet complexity requirements
	ErrPasswordTooWeak = errors.New("password must contain at least one uppercase letter, one lowercase letter, and one number")
	// ErrInvalidValue is returned for invalid numeric values
	ErrInvalidValue = errors.New("invalid value")
)

// Email regex pattern - RFC 5322 simplified
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidEmail
	}
	if len(email) > 254 { // RFC 5321
		return ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if len(password) > 72 { // bcrypt limit
		return errors.New("password must be at most 72 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber {
		return ErrPasswordTooWeak
	}

	return nil
}

// ValidatePositiveInt checks if an integer is positive
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%w: %s must be positive", ErrInvalidValue, fieldName)
	}
	return nil
}

// ValidateNonNegativeInt checks if an integer is non-negative
func ValidateNonNegativeInt(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%w: %s must be non-negative", ErrInvalidValue, fieldName)
	}
	return nil
}

// ValidatePositiveFloat checks if a float is positive
func ValidatePositiveFloat(value float64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%w: %s must be positive", ErrInvalidValue, fieldName)
	}
	return nil
}

// ValidateIntRange checks if an integer is within a range
func ValidateIntRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%w: %s must be between %d and %d", ErrInvalidValue, fieldName, min, max)
	}
	return nil
}

// ValidateFloatRange checks if a float is within a range
func ValidateFloatRange(value, min, max float64, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%w: %s must be between %.2f and %.2f", ErrInvalidValue, fieldName, min, max)
	}
	return nil
}

// ValidateStringLength checks if a string length is within bounds
func ValidateStringLength(value string, minLen, maxLen int, fieldName string) error {
	length := len(value)
	if length < minLen || length > maxLen {
		return fmt.Errorf("%w: %s must be between %d and %d characters", ErrInvalidValue, fieldName, minLen, maxLen)
	}
	return nil
}
