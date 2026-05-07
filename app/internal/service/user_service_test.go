package service_test

import (
	"context"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

func TestUserService_Create_Validation(t *testing.T) {
	svc := service.NewUserService(nil)

	tests := []struct {
		name    string
		input   service.CreateUserInput
		wantErr bool
		errField string
	}{
		{
			name:     "empty email",
			input:    service.CreateUserInput{Name: "Test", Role: "general"},
			wantErr:  true,
			errField: "email",
		},
		{
			name:     "invalid email",
			input:    service.CreateUserInput{Email: "notanemail", Name: "Test", Role: "general"},
			wantErr:  true,
			errField: "email",
		},
		{
			name:     "empty name",
			input:    service.CreateUserInput{Email: "test@example.com", Role: "general"},
			wantErr:  true,
			errField: "name",
		},
		{
			name:     "invalid role",
			input:    service.CreateUserInput{Email: "test@example.com", Name: "Test", Role: "superuser"},
			wantErr:  true,
			errField: "role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				ve, ok := err.(validator.Errors)
				if !ok {
					t.Fatalf("expected validator.Errors, got %T: %v", err, err)
				}
				if _, exists := ve[tt.errField]; !exists {
					t.Errorf("expected error for field %q, got errors: %v", tt.errField, ve)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestUserService_UpdateStatus_Validation(t *testing.T) {
	svc := service.NewUserService(nil)

	err := svc.UpdateStatus(context.Background(), 1, "invalid_status")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}

	ve, ok := err.(validator.Errors)
	if !ok {
		t.Fatalf("expected validator.Errors, got %T", err)
	}
	if _, exists := ve["status"]; !exists {
		t.Errorf("expected error for field 'status', got: %v", ve)
	}
}
