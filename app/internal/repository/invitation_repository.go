package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
)

type InvitationRepository struct {
	db *pgxpool.Pool
}

func NewInvitationRepository(db *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{db: db}
}

func (r *InvitationRepository) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*domain.InvitationToken, error) {
	const q = `
		INSERT INTO invitation_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	t := &domain.InvitationToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	err := r.db.QueryRow(ctx, q, userID, tokenHash, expiresAt).Scan(&t.ID, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create invitation token: %w", err)
	}
	return t, nil
}

func (r *InvitationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.InvitationToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM invitation_tokens
		WHERE token_hash = $1`

	t := &domain.InvitationToken{}
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find invitation token: %w", err)
	}
	return t, nil
}

func (r *InvitationRepository) MarkAsUsed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE invitation_tokens SET used_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark invitation token as used: %w", err)
	}
	return nil
}
