package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, department, job_title,
		       status, invited_at, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role,
		&u.Department, &u.JobTitle, &u.Status, &u.InvitedAt,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		SELECT id, email, password_hash, name, role, department, job_title,
		       status, invited_at, last_login_at, created_at, updated_at
		FROM users WHERE email = $1`

	u := &domain.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role,
		&u.Department, &u.JobTitle, &u.Status, &u.InvitedAt,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (r *UserRepository) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error) {
	fb := queryparam.NewFilterBuilder()

	if params.Search != "" {
		idx := fb.NextIndex()
		fb.Add(
			fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d OR department ILIKE $%d)", idx, idx, idx),
			"%"+params.Search+"%",
		)
	}
	fb.AddEqual("status", params.Status)
	fb.AddEqual("role", params.Role)
	fb.AddLike("department", params.Department)
	fb.AddLike("job_title", params.JobTitle)
	fb.AddDateRange("created_at", params.CreatedFrom, params.CreatedTo)
	fb.AddDateRange("last_login_at", params.LastLoginFrom, params.LastLoginTo)

	where := fb.WhereClause()
	args := fb.Args()

	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM users %s`, where)
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	sortCol := params.SortColumn
	if sortCol == "" {
		sortCol = "created_at"
	}
	sortOrder := strings.ToUpper(params.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	allowedCols := map[string]bool{
		"name": true, "email": true, "created_at": true,
		"last_login_at": true, "status": true, "role": true,
	}
	if !allowedCols[sortCol] {
		sortCol = "created_at"
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	perPage := params.PerPage
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	limitIdx := fb.NextIndex()
	offsetIdx := limitIdx + 1
	listQ := fmt.Sprintf(`
		SELECT id, email, password_hash, name, role, department, job_title,
		       status, invited_at, last_login_at, created_at, updated_at
		FROM users %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d`,
		where, sortCol, sortOrder, limitIdx, offsetIdx,
	)
	listArgs := append(args, perPage, offset)

	rows, err := r.db.Query(ctx, listQ, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role,
			&u.Department, &u.JobTitle, &u.Status, &u.InvitedAt,
			&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `
		INSERT INTO users (email, password_hash, name, role, department, job_title, status, invited_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, q,
		u.Email, u.PasswordHash, u.Name, u.Role,
		u.Department, u.JobTitle, u.Status, u.InvitedAt,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) (*domain.User, error) {
	const q = `
		UPDATE users
		SET email = $1, name = $2, role = $3, department = $4, job_title = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRow(ctx, q,
		u.Email, u.Name, u.Role, u.Department, u.JobTitle, u.ID,
	).Scan(&u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update user status: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, id,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int64, t time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = $1, updated_at = NOW() WHERE id = $2`,
		t, id,
	)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

func (r *UserRepository) GetStats(ctx context.Context) (*domain.UserStats, error) {
	const q = `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE status = 'active')   AS active,
			COUNT(*) FILTER (WHERE status = 'invited')  AS invited,
			COUNT(*) FILTER (WHERE status = 'disabled') AS disabled,
			COUNT(*) FILTER (WHERE role = 'admin')      AS admin,
			COUNT(*) FILTER (WHERE role = 'general')    AS general
		FROM users`

	s := &domain.UserStats{}
	err := r.db.QueryRow(ctx, q).Scan(
		&s.Total, &s.Active, &s.Invited, &s.Disabled, &s.Admin, &s.General,
	)
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}
	return s, nil
}
