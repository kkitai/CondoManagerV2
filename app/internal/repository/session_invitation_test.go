package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

func TestSessionRepository_CreateAndFind(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)

	email := fmt.Sprintf("session-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	token := "test-session-token-unique-123"
	hash := repository.HashToken(token)
	ip := "127.0.0.1"
	ua := "test-agent"

	session := &domain.Session{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: &ip,
		UserAgent: &ua,
	}

	created, err := sessionRepo.Create(context.Background(), session)
	if err != nil {
		t.Fatalf("Create session: %v", err)
	}
	if created.ID == 0 {
		t.Error("expected non-zero session ID")
	}

	found, err := sessionRepo.FindByToken(context.Background(), token)
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if found.UserID != u.ID {
		t.Errorf("UserID = %d, want %d", found.UserID, u.ID)
	}
}

func TestSessionRepository_FindByToken_Expired(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)

	email := fmt.Sprintf("expsession-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	token := fmt.Sprintf("expired-token-%d", time.Now().UnixNano())
	hash := repository.HashToken(token)

	session := &domain.Session{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // already expired
	}
	if _, err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Create expired session: %v", err)
	}

	_, err := sessionRepo.FindByToken(context.Background(), token)
	if err == nil {
		t.Error("expected error for expired session")
	}
}

func TestSessionRepository_Delete(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)

	email := fmt.Sprintf("delsession-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	token := fmt.Sprintf("del-token-%d", time.Now().UnixNano())
	hash := repository.HashToken(token)
	session := &domain.Session{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if _, err := sessionRepo.Create(context.Background(), session); err != nil {
		t.Fatalf("Create session: %v", err)
	}

	if err := sessionRepo.Delete(context.Background(), token); err != nil {
		t.Fatalf("Delete session: %v", err)
	}

	_, err := sessionRepo.FindByToken(context.Background(), token)
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestSessionRepository_DeleteExpired(t *testing.T) {
	pool := testutil.NewTestDB(t)
	sessionRepo := repository.NewSessionRepository(pool)

	if err := sessionRepo.DeleteExpired(context.Background()); err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
}

func TestSessionRepository_DeleteByUserID(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)

	email := fmt.Sprintf("deluidsession-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	for i := 0; i < 2; i++ {
		token := fmt.Sprintf("uid-token-%d-%d", time.Now().UnixNano(), i)
		hash := repository.HashToken(token)
		session := &domain.Session{
			UserID:    u.ID,
			TokenHash: hash,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if _, err := sessionRepo.Create(context.Background(), session); err != nil {
			t.Fatalf("Create session %d: %v", i, err)
		}
	}

	if err := sessionRepo.DeleteByUserID(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteByUserID: %v", err)
	}
}

func TestInvitationRepository_CreateAndFind(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)

	email := fmt.Sprintf("invitation-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	token := fmt.Sprintf("inv-token-%d", time.Now().UnixNano())
	tokenHash := repository.HashToken(token)
	expiresAt := time.Now().Add(72 * time.Hour)

	inv, err := invRepo.Create(context.Background(), u.ID, tokenHash, expiresAt)
	if err != nil {
		t.Fatalf("Create invitation: %v", err)
	}
	if inv.ID == 0 {
		t.Error("expected non-zero invitation ID")
	}

	found, err := invRepo.FindByTokenHash(context.Background(), tokenHash)
	if err != nil {
		t.Fatalf("FindByTokenHash: %v", err)
	}
	if found.UserID != u.ID {
		t.Errorf("UserID = %d, want %d", found.UserID, u.ID)
	}
}

func TestInvitationRepository_FindByTokenHash_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	invRepo := repository.NewInvitationRepository(pool)

	_, err := invRepo.FindByTokenHash(context.Background(), "nonexistent-hash")
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestInvitationRepository_MarkAsUsed(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)

	email := fmt.Sprintf("markused-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, userRepo, email)

	tokenHash := repository.HashToken(fmt.Sprintf("markused-%d", time.Now().UnixNano()))
	inv, err := invRepo.Create(context.Background(), u.ID, tokenHash, time.Now().Add(72*time.Hour))
	if err != nil {
		t.Fatalf("Create invitation: %v", err)
	}

	if err := invRepo.MarkAsUsed(context.Background(), inv.ID); err != nil {
		t.Fatalf("MarkAsUsed: %v", err)
	}

	found, _ := invRepo.FindByTokenHash(context.Background(), tokenHash)
	if found.UsedAt == nil {
		t.Error("expected used_at to be set")
	}
}
