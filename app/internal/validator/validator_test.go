package validator_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

func TestRequired(t *testing.T) {
	v := validator.New()
	v.Required("name", "")
	if v.Valid() {
		t.Error("expected invalid")
	}
	if _, ok := v.Errors()["name"]; !ok {
		t.Error("expected error for name field")
	}
}

func TestRequiredValid(t *testing.T) {
	v := validator.New()
	v.Required("name", "John")
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestMaxLength(t *testing.T) {
	v := validator.New()
	v.MaxLength("name", "あいうえおかきくけこ", 5)
	if v.Valid() {
		t.Error("expected invalid")
	}
}

func TestMinLength(t *testing.T) {
	v := validator.New()
	v.MinLength("password", "ab", 8)
	if v.Valid() {
		t.Error("expected invalid for too-short value")
	}
}

func TestMinLengthValid(t *testing.T) {
	v := validator.New()
	v.MinLength("password", "abcdefgh", 8)
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"invalid-email", false},
		{"", true},
	}
	for _, tc := range tests {
		v := validator.New()
		v.Email("email", tc.email)
		if v.Valid() != tc.valid {
			t.Errorf("email=%q: valid=%v, want %v", tc.email, v.Valid(), tc.valid)
		}
	}
}

func TestOneOf(t *testing.T) {
	v := validator.New()
	v.OneOf("role", "superuser", []string{"admin", "general"})
	if v.Valid() {
		t.Error("expected invalid for unknown role")
	}
}

func TestOneOfValid(t *testing.T) {
	v := validator.New()
	v.OneOf("role", "admin", []string{"admin", "general"})
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestErrorsError(t *testing.T) {
	e := validator.Errors{"field": "必須項目です"}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
	if !e.HasErrors() {
		t.Error("expected HasErrors() to be true")
	}
}
