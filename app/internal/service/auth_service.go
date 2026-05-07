package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
)

const (
	SessionCookieName = "session_token"
	sessionTokenBytes = 32
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user account is disabled")
	ErrSessionNotFound    = errors.New("session not found")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	sessionTTL  time.Duration
}

func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	sessionTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionTTL:  sessionTTL,
	}
}

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateToken() (string, error) {
	b := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func (s *AuthService) Login(ctx context.Context, email, password string, r *http.Request) (string, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if user.Status == domain.StatusDisabled {
		return "", nil, ErrUserDisabled
	}

	if user.PasswordHash == nil || !CheckPassword(*user.PasswordHash, password) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return "", nil, err
	}

	ipAddr := r.RemoteAddr
	ua := r.UserAgent()
	session := &domain.Session{
		UserID:    user.ID,
		TokenHash: repository.HashToken(token),
		ExpiresAt: time.Now().Add(s.sessionTTL),
		IPAddress: &ipAddr,
		UserAgent: &ua,
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return "", nil, err
	}

	now := time.Now()
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionRepo.Delete(ctx, token)
}

func (s *AuthService) GetUserByToken(ctx context.Context, token string) (*domain.User, error) {
	session, err := s.sessionRepo.FindByToken(ctx, token)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	if user.Status == domain.StatusDisabled {
		return nil, ErrUserDisabled
	}

	return user, nil
}
