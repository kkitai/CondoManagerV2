package util_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/util"
)

func TestValidatorRequired(t *testing.T) {
	v := util.NewValidator()
	v.Required("name", "")
	if v.Valid() {
		t.Error("expected invalid")
	}
	if _, ok := v.Errors()["name"]; !ok {
		t.Error("expected error for name field")
	}
}

func TestValidatorRequiredValid(t *testing.T) {
	v := util.NewValidator()
	v.Required("name", "John")
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestValidatorMaxLength(t *testing.T) {
	v := util.NewValidator()
	v.MaxLength("name", "あいうえおかきくけこ", 5)
	if v.Valid() {
		t.Error("expected invalid")
	}
}

func TestValidatorEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"invalid-email", false},
		{"", true},
	}
	for _, tc := range tests {
		v := util.NewValidator()
		v.Email("email", tc.email)
		if v.Valid() != tc.valid {
			t.Errorf("email=%q: valid=%v, want %v", tc.email, v.Valid(), tc.valid)
		}
	}
}

func TestValidatorOneOf(t *testing.T) {
	v := util.NewValidator()
	v.OneOf("role", "superuser", []string{"admin", "general"})
	if v.Valid() {
		t.Error("expected invalid for unknown role")
	}
}

func TestValidatorOneOfValid(t *testing.T) {
	v := util.NewValidator()
	v.OneOf("role", "admin", []string{"admin", "general"})
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestValidationErrorsError(t *testing.T) {
	e := util.ValidationErrors{"field": "必須項目です"}
	if e.Error() == "" {
		t.Error("expected non-empty error message")
	}
	if !e.HasErrors() {
		t.Error("expected HasErrors() to be true")
	}
}
