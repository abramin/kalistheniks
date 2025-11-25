package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with dots", "first.last@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"empty email", "", true},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"missing local part", "@example.com", true},
		{"invalid characters", "user name@example.com", true},
		{"no TLD", "user@example", true},
		{"double @", "user@@example.com", true},
		{"too long", strings.Repeat("a", 250) + "@example.com", true},
		{"whitespace only", "   ", true},
		{"whitespace trimmed", "  user@example.com  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidEmail)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		password string
		wantErr error
	}{
		{"valid password", "Password123", nil},
		{"valid with special chars", "P@ssw0rd!", nil},
		{"too short", "Pass1", ErrPasswordTooShort},
		{"no uppercase", "password123", ErrPasswordTooWeak},
		{"no lowercase", "PASSWORD123", ErrPasswordTooWeak},
		{"no number", "PasswordABC", ErrPasswordTooWeak},
		{"empty", "", ErrPasswordTooShort},
		{"exactly 8 chars valid", "Passw0rd", nil},
		{"only uppercase and numbers", "PASSWORD123", ErrPasswordTooWeak},
		{"only lowercase and numbers", "password123", ErrPasswordTooWeak},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePasswordTooLong(t *testing.T) {
	// Test bcrypt 72 character limit
	longPassword := strings.Repeat("A", 73) + "a1"
	err := ValidatePassword(longPassword)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at most 72")
}

func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		field   string
		wantErr bool
	}{
		{"positive value", 10, "reps", false},
		{"zero value", 0, "reps", true},
		{"negative value", -5, "reps", true},
		{"large positive", 1000, "reps", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
				assert.Contains(t, err.Error(), tt.field)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNonNegativeInt(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		field   string
		wantErr bool
	}{
		{"positive value", 10, "set_index", false},
		{"zero value", 0, "set_index", false},
		{"negative value", -1, "set_index", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNonNegativeInt(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePositiveFloat(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		field   string
		wantErr bool
	}{
		{"positive value", 20.5, "weight", false},
		{"zero value", 0, "weight", true},
		{"negative value", -5.5, "weight", true},
		{"small positive", 0.1, "weight", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveFloat(tt.value, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateIntRange(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		min     int
		max     int
		field   string
		wantErr bool
	}{
		{"within range", 5, 1, 10, "rpe", false},
		{"at minimum", 1, 1, 10, "rpe", false},
		{"at maximum", 10, 1, 10, "rpe", false},
		{"below minimum", 0, 1, 10, "rpe", true},
		{"above maximum", 11, 1, 10, "rpe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntRange(tt.value, tt.min, tt.max, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFloatRange(t *testing.T) {
	tests := []struct {
		name    string
		value   float64
		min     float64
		max     float64
		field   string
		wantErr bool
	}{
		{"within range", 50.5, 0, 500, "weight", false},
		{"at minimum", 0, 0, 500, "weight", false},
		{"at maximum", 500, 0, 500, "weight", false},
		{"below minimum", -0.1, 0, 500, "weight", true},
		{"above maximum", 500.1, 0, 500, "weight", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFloatRange(tt.value, tt.min, tt.max, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		minLen  int
		maxLen  int
		field   string
		wantErr bool
	}{
		{"within range", "hello", 1, 10, "notes", false},
		{"at minimum", "a", 1, 10, "notes", false},
		{"at maximum", "1234567890", 1, 10, "notes", false},
		{"too short", "", 1, 10, "notes", true},
		{"too long", "12345678901", 1, 10, "notes", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.value, tt.minLen, tt.maxLen, tt.field)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidValue)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
