package service_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

func TestAuthService_Login_Success(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	// create a user with known password
	hash, _ := service.HashPassword("testpassword")
	email := fmt.Sprintf("authlogin-%d@test.com", time.Now().UnixNano())
	u := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         "Auth Test",
		Role:         domain.RoleGeneral,
		Status:       domain.StatusActive,
	}
	if _, err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	token, user, err := authSvc.Login(context.Background(), email, "testpassword", req)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	if user.Email != email {
		t.Errorf("user email = %q, want %q", user.Email, email)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	hash, _ := service.HashPassword("correctpassword")
	email := fmt.Sprintf("wrongpass-%d@test.com", time.Now().UnixNano())
	u := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         "Test",
		Role:         domain.RoleGeneral,
		Status:       domain.StatusActive,
	}
	if _, err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	_, _, err := authSvc.Login(context.Background(), email, "wrongpassword", req)
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_DisabledUser(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	hash, _ := service.HashPassword("password")
	email := fmt.Sprintf("disabled-%d@test.com", time.Now().UnixNano())
	u := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         "Disabled",
		Role:         domain.RoleGeneral,
		Status:       domain.StatusDisabled,
	}
	if _, err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	_, _, err := authSvc.Login(context.Background(), email, "password", req)
	if err != service.ErrUserDisabled {
		t.Errorf("expected ErrUserDisabled, got %v", err)
	}
}

func TestAuthService_Login_NonExistentUser(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	_, _, err := authSvc.Login(context.Background(), "nobody@test.com", "password", req)
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_NilPasswordHash(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	email := fmt.Sprintf("nohash-%d@test.com", time.Now().UnixNano())
	now := time.Now()
	u := &domain.User{
		Email:     email,
		Name:      "NoHash",
		Role:      domain.RoleGeneral,
		Status:    domain.StatusInvited,
		InvitedAt: &now,
	}
	if _, err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	_, _, err := authSvc.Login(context.Background(), email, "password", req)
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials for nil hash, got %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	hash, _ := service.HashPassword("pass")
	email := fmt.Sprintf("logout-%d@test.com", time.Now().UnixNano())
	u := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         "Logout",
		Role:         domain.RoleGeneral,
		Status:       domain.StatusActive,
	}
	if _, err := userRepo.Create(context.Background(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	token, _, err := authSvc.Login(context.Background(), email, "pass", req)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if err := authSvc.Logout(context.Background(), token); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// after logout, should not find user by token
	_, err = authSvc.GetUserByToken(context.Background(), token)
	if err != service.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after logout, got %v", err)
	}
}

func TestAuthService_GetUserByToken_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	authSvc := service.NewAuthService(userRepo, sessionRepo, 24*time.Hour)

	_, err := authSvc.GetUserByToken(context.Background(), "nonexistent-token")
	if err != service.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}
