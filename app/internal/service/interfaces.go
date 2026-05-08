package service

import (
	"context"
	"net/http"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
)

type AuthServicer interface {
	Login(ctx context.Context, email, password string, r *http.Request) (string, *domain.User, error)
	Logout(ctx context.Context, token string) error
	GetUserByToken(ctx context.Context, token string) (*domain.User, error)
}

type UserServicer interface {
	Create(ctx context.Context, in CreateUserInput) (*domain.User, error)
	Update(ctx context.Context, id int64, in UpdateUserInput) (*domain.User, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error)
	GetStats(ctx context.Context) (*domain.UserStats, error)
}

type InvitationServicer interface {
	SendInvitation(ctx context.Context, userID int64) error
	ValidateToken(ctx context.Context, token string) (*domain.InvitationToken, *domain.User, error)
	AcceptInvitation(ctx context.Context, token, password string) error
}

type PropertyServicer interface {
	Create(ctx context.Context, in CreatePropertyInput) (*domain.Property, error)
	Update(ctx context.Context, id int64, in UpdatePropertyInput) (*domain.Property, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.Property, error)
	List(ctx context.Context, params domain.PropertyListParams) ([]*domain.Property, int64, error)
	GetStats(ctx context.Context, propertyID int64) (*domain.PropertyStats, error)
}
