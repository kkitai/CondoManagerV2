package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

var (
	ErrEmailAlreadyExists = errors.New("email already in use")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserInput struct {
	Email      string
	Name       string
	Role       string
	Department string
	JobTitle   string
	Password   string
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	v := validator.New()
	v.Required("email", in.Email)
	v.Email("email", in.Email)
	v.Required("name", in.Name)
	v.MaxLength("name", in.Name, 100)
	v.OneOf("role", in.Role, []string{"admin", "general"})
	if !v.Valid() {
		return nil, v.Errors()
	}

	if _, err := s.userRepo.FindByEmail(ctx, in.Email); err == nil {
		return nil, ErrEmailAlreadyExists
	}

	u := &domain.User{
		Email:  in.Email,
		Name:   in.Name,
		Role:   domain.UserRole(in.Role),
		Status: domain.StatusInvited,
	}
	if in.Department != "" {
		u.Department = &in.Department
	}
	if in.JobTitle != "" {
		u.JobTitle = &in.JobTitle
	}

	now := time.Now()
	u.InvitedAt = &now

	if in.Password != "" {
		h, err := HashPassword(in.Password)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		u.PasswordHash = &h
		u.Status = domain.StatusActive
	}

	return s.userRepo.Create(ctx, u)
}

type UpdateUserInput struct {
	Email      string
	Name       string
	Role       string
	Department string
	JobTitle   string
}

func (s *UserService) Update(ctx context.Context, id int64, in UpdateUserInput) (*domain.User, error) {
	v := validator.New()
	v.Required("email", in.Email)
	v.Email("email", in.Email)
	v.Required("name", in.Name)
	v.MaxLength("name", in.Name, 100)
	v.OneOf("role", in.Role, []string{"admin", "general"})
	if !v.Valid() {
		return nil, v.Errors()
	}

	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	u.Email = in.Email
	u.Name = in.Name
	u.Role = domain.UserRole(in.Role)
	if in.Department != "" {
		u.Department = &in.Department
	} else {
		u.Department = nil
	}
	if in.JobTitle != "" {
		u.JobTitle = &in.JobTitle
	} else {
		u.JobTitle = nil
	}

	return s.userRepo.Update(ctx, u)
}

func (s *UserService) UpdateStatus(ctx context.Context, id int64, status string) error {
	v := validator.New()
	v.OneOf("status", status, []string{"active", "disabled"})
	if !v.Valid() {
		return v.Errors()
	}
	return s.userRepo.UpdateStatus(ctx, id, status)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *UserService) List(ctx context.Context, params domain.UserListParams) ([]*domain.User, int64, error) {
	return s.userRepo.List(ctx, params)
}

func (s *UserService) GetStats(ctx context.Context) (*domain.UserStats, error) {
	return s.userRepo.GetStats(ctx)
}
