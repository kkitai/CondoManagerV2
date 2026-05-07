package domain_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
)

func TestUser_IsAdmin(t *testing.T) {
	admin := &domain.User{Role: domain.RoleAdmin}
	general := &domain.User{Role: domain.RoleGeneral}

	if !admin.IsAdmin() {
		t.Error("admin user should be admin")
	}
	if general.IsAdmin() {
		t.Error("general user should not be admin")
	}
}

func TestUser_IsActive(t *testing.T) {
	active := &domain.User{Status: domain.StatusActive}
	invited := &domain.User{Status: domain.StatusInvited}
	disabled := &domain.User{Status: domain.StatusDisabled}

	if !active.IsActive() {
		t.Error("active user should be active")
	}
	if invited.IsActive() {
		t.Error("invited user should not be active")
	}
	if disabled.IsActive() {
		t.Error("disabled user should not be active")
	}
}

func TestUserRole_Values(t *testing.T) {
	if string(domain.RoleAdmin) != "admin" {
		t.Errorf("RoleAdmin = %q, want 'admin'", domain.RoleAdmin)
	}
	if string(domain.RoleGeneral) != "general" {
		t.Errorf("RoleGeneral = %q, want 'general'", domain.RoleGeneral)
	}
}

func TestUserStatus_Values(t *testing.T) {
	if string(domain.StatusActive) != "active" {
		t.Errorf("StatusActive = %q, want 'active'", domain.StatusActive)
	}
	if string(domain.StatusInvited) != "invited" {
		t.Errorf("StatusInvited = %q, want 'invited'", domain.StatusInvited)
	}
	if string(domain.StatusDisabled) != "disabled" {
		t.Errorf("StatusDisabled = %q, want 'disabled'", domain.StatusDisabled)
	}
}
