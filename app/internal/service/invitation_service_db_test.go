package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

func TestInvitationService_SendAndAccept(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	mailer := &mockMailer{}
	svc := service.NewInvitationService(userRepo, invRepo, mailer, "http://localhost:8080")

	email := fmt.Sprintf("inv-send-%d@test.com", time.Now().UnixNano())
	now := time.Now()
	u := &domain.User{
		Email:     email,
		Name:      "Invite Test",
		Role:      domain.RoleGeneral,
		Status:    domain.StatusInvited,
		InvitedAt: &now,
	}
	created, err := userRepo.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := svc.SendInvitation(context.Background(), created.ID); err != nil {
		t.Fatalf("SendInvitation: %v", err)
	}
	if !mailer.called {
		t.Error("expected mailer to be called")
	}
	if mailer.lastURL == "" {
		t.Error("expected non-empty invite URL")
	}

	// extract token from URL
	url := mailer.lastURL
	// URL format: http://localhost:8080/invite/{token}
	token := url[len("http://localhost:8080/invite/"):]

	// validate token
	inv, user, err := svc.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if inv == nil {
		t.Fatal("expected non-nil invitation")
	}
	if user.Email != email {
		t.Errorf("user email = %q, want %q", user.Email, email)
	}

	// accept invitation
	if err := svc.AcceptInvitation(context.Background(), token, "newpassword123"); err != nil {
		t.Fatalf("AcceptInvitation: %v", err)
	}

	// token should now be used
	_, _, err = svc.ValidateToken(context.Background(), token)
	if err != service.ErrTokenUsed {
		t.Errorf("expected ErrTokenUsed after accept, got %v", err)
	}
}

func TestInvitationService_ValidateToken_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	svc := service.NewInvitationService(userRepo, invRepo, &mockMailer{}, "http://localhost")

	_, _, err := svc.ValidateToken(context.Background(), "nonexistent-token")
	if err != service.ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestInvitationService_SendInvitation_UserNotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	svc := service.NewInvitationService(userRepo, invRepo, &mockMailer{}, "http://localhost")

	err := svc.SendInvitation(context.Background(), 999999)
	if err != service.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestInvitationService_ValidateToken_Expired(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	svc := service.NewInvitationService(userRepo, invRepo, &mockMailer{}, "http://localhost")

	email := fmt.Sprintf("expired-inv-%d@test.com", time.Now().UnixNano())
	now := time.Now()
	u := &domain.User{
		Email:     email,
		Name:      "Expired",
		Role:      domain.RoleGeneral,
		Status:    domain.StatusInvited,
		InvitedAt: &now,
	}
	created, _ := userRepo.Create(context.Background(), u)

	// create expired token directly in repository
	tokenHash := repository.HashToken(fmt.Sprintf("expired-%d", time.Now().UnixNano()))
	expiredAt := time.Now().Add(-1 * time.Hour)
	if _, err := invRepo.Create(context.Background(), created.ID, tokenHash, expiredAt); err != nil {
		t.Fatalf("create expired token: %v", err)
	}

	// find the raw token by using the hash-reversed lookup won't work
	// instead test with a token that produces a known hash
	rawToken := fmt.Sprintf("rawexpired-%d", time.Now().UnixNano())
	rawHash := repository.HashToken(rawToken)
	if _, err := invRepo.Create(context.Background(), created.ID, rawHash, time.Now().Add(-1*time.Hour)); err != nil {
		t.Fatalf("create expired raw token: %v", err)
	}

	_, _, err := svc.ValidateToken(context.Background(), rawToken)
	if err != service.ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}
