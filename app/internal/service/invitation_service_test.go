package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

var _ = errors.New // ensure errors import used

type mockMailer struct {
	called   bool
	lastURL  string
	sendErr  error
}

func (m *mockMailer) SendInvitation(_ context.Context, _, _, inviteURL string) error {
	m.called = true
	m.lastURL = inviteURL
	return m.sendErr
}

func TestInvitationService_AcceptInvitation_PasswordTooShort(t *testing.T) {
	svc := service.NewInvitationService(nil, nil, nil, "http://localhost")
	err := svc.AcceptInvitation(context.Background(), "sometoken", "short")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestInvitationErrors_Identity(t *testing.T) {
	if !errors.Is(service.ErrTokenExpired, service.ErrTokenExpired) {
		t.Error("ErrTokenExpired should match itself")
	}
	if !errors.Is(service.ErrTokenUsed, service.ErrTokenUsed) {
		t.Error("ErrTokenUsed should match itself")
	}
	if !errors.Is(service.ErrTokenInvalid, service.ErrTokenInvalid) {
		t.Error("ErrTokenInvalid should match itself")
	}

	// distinct sentinel errors should not match each other
	if errors.Is(service.ErrTokenExpired, service.ErrTokenUsed) {
		t.Error("ErrTokenExpired should not match ErrTokenUsed")
	}
	if errors.Is(service.ErrTokenUsed, service.ErrTokenInvalid) {
		t.Error("ErrTokenUsed should not match ErrTokenInvalid")
	}
}
