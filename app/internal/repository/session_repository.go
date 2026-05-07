package repository

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
)

type SessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	const q = `
		INSERT INTO sessions (user_id, token_hash, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.IPAddress,
		session.UserAgent,
	).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (*domain.Session, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, ip_address, user_agent, created_at
		FROM sessions
		WHERE token_hash = $1 AND expires_at > NOW()`

	s := &domain.Session{}
	err := r.db.QueryRow(ctx, q, HashToken(token)).Scan(
		&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt,
		&s.IPAddress, &s.UserAgent, &s.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find session by token: %w", err)
	}
	return s, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM sessions WHERE token_hash = $1`,
		HashToken(token),
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at <= $1`, time.Now())
	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete sessions by user: %w", err)
	}
	return nil
}
