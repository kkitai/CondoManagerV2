package domain

import "time"

type UserRole string
type UserStatus string

const (
	RoleAdmin   UserRole = "admin"
	RoleGeneral UserRole = "general"
)

const (
	StatusActive   UserStatus = "active"
	StatusInvited  UserStatus = "invited"
	StatusDisabled UserStatus = "disabled"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash *string
	Name         string
	Role         UserRole
	Department   *string
	JobTitle     *string
	Status       UserStatus
	InvitedAt    *time.Time
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

type UserStats struct {
	Total    int64
	Active   int64
	Invited  int64
	Disabled int64
	Admin    int64
	General  int64
}

type UserListParams struct {
	Search     string
	Status     string
	Role       string
	Department string
	JobTitle   string
	CreatedFrom time.Time
	CreatedTo   time.Time
	LastLoginFrom time.Time
	LastLoginTo   time.Time
	Page       int
	PerPage    int
	SortColumn string
	SortOrder  string
}

type Session struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	IPAddress *string
	UserAgent *string
	CreatedAt time.Time
}

type InvitationToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}
