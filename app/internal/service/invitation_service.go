package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
)

const invitationTokenBytes = 32
const invitationTTL = 72 * time.Hour

var (
	ErrTokenExpired  = errors.New("invitation token has expired")
	ErrTokenUsed     = errors.New("invitation token has already been used")
	ErrTokenInvalid  = errors.New("invitation token is invalid")
)

type Mailer interface {
	SendInvitation(ctx context.Context, toEmail, toName, inviteURL string) error
}

type InvitationService struct {
	userRepo       *repository.UserRepository
	invitationRepo *repository.InvitationRepository
	mailer         Mailer
	appBaseURL     string
}

func NewInvitationService(
	userRepo *repository.UserRepository,
	invitationRepo *repository.InvitationRepository,
	mailer Mailer,
	appBaseURL string,
) *InvitationService {
	return &InvitationService{
		userRepo:       userRepo,
		invitationRepo: invitationRepo,
		mailer:         mailer,
		appBaseURL:     appBaseURL,
	}
}

func generateInvitationToken() (string, error) {
	b := make([]byte, invitationTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate invitation token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func hashInvitationToken(token string) string {
	return repository.HashToken(token)
}

func (s *InvitationService) SendInvitation(ctx context.Context, userID int64) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	token, err := generateInvitationToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(invitationTTL)
	if _, err := s.invitationRepo.Create(ctx, userID, hashInvitationToken(token), expiresAt); err != nil {
		return fmt.Errorf("create invitation token: %w", err)
	}

	now := time.Now()
	if err := s.userRepo.UpdateStatus(ctx, userID, string(domain.StatusInvited)); err != nil {
		return err
	}
	user.InvitedAt = &now
	user.Status = domain.StatusInvited

	inviteURL := fmt.Sprintf("%s/invite/%s", s.appBaseURL, token)
	return s.mailer.SendInvitation(ctx, user.Email, user.Name, inviteURL)
}

func (s *InvitationService) ValidateToken(ctx context.Context, token string) (*domain.InvitationToken, *domain.User, error) {
	inv, err := s.invitationRepo.FindByTokenHash(ctx, hashInvitationToken(token))
	if err != nil {
		return nil, nil, ErrTokenInvalid
	}

	if inv.UsedAt != nil {
		return nil, nil, ErrTokenUsed
	}

	if time.Now().After(inv.ExpiresAt) {
		return nil, nil, ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(ctx, inv.UserID)
	if err != nil {
		return nil, nil, ErrUserNotFound
	}

	return inv, user, nil
}

func (s *InvitationService) AcceptInvitation(ctx context.Context, token, password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	inv, _, err := s.ValidateToken(ctx, token)
	if err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	if err := s.userRepo.UpdatePassword(ctx, inv.UserID, hash); err != nil {
		return err
	}

	if err := s.userRepo.UpdateStatus(ctx, inv.UserID, string(domain.StatusActive)); err != nil {
		return err
	}

	return s.invitationRepo.MarkAsUsed(ctx, inv.ID)
}
