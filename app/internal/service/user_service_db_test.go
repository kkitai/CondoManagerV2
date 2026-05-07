package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

func TestUserService_Create_Success(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	email := fmt.Sprintf("svc-create-%d@test.com", time.Now().UnixNano())
	u, err := svc.Create(context.Background(), service.CreateUserInput{
		Email: email,
		Name:  "Service Test",
		Role:  "general",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	email := fmt.Sprintf("dup-%d@test.com", time.Now().UnixNano())
	_, err := svc.Create(context.Background(), service.CreateUserInput{
		Email: email, Name: "First", Role: "general",
	})
	if err != nil {
		t.Fatalf("first Create: %v", err)
	}

	_, err = svc.Create(context.Background(), service.CreateUserInput{
		Email: email, Name: "Second", Role: "general",
	})
	if err != service.ErrEmailAlreadyExists {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestUserService_Update_Success(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	email := fmt.Sprintf("svc-update-%d@test.com", time.Now().UnixNano())
	created, _ := svc.Create(context.Background(), service.CreateUserInput{
		Email: email, Name: "Original", Role: "general",
	})

	updated, err := svc.Update(context.Background(), created.ID, service.UpdateUserInput{
		Email: email,
		Name:  "Updated",
		Role:  "admin",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Name = %q, want 'Updated'", updated.Name)
	}
}

func TestUserService_Update_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	_, err := svc.Update(context.Background(), 999999, service.UpdateUserInput{
		Email: "test@example.com",
		Name:  "Test",
		Role:  "general",
	})
	if err != service.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	_, err := svc.GetByID(context.Background(), 999999)
	if err != service.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	users, total, err := svc.List(context.Background(), domain.UserListParams{
		Page: 1, PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = users
	_ = total
}

func TestUserService_GetStats(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats == nil {
		t.Error("expected non-nil stats")
	}
}

func TestUserService_UpdateStatus_ValidStatuses(t *testing.T) {
	pool := testutil.NewTestDB(t)
	userRepo := repository.NewUserRepository(pool)
	svc := service.NewUserService(userRepo)

	email := fmt.Sprintf("svc-status-%d@test.com", time.Now().UnixNano())
	created, _ := svc.Create(context.Background(), service.CreateUserInput{
		Email: email, Name: "Status Test", Role: "general",
	})

	// active status
	if err := svc.UpdateStatus(context.Background(), created.ID, "active"); err != nil {
		t.Fatalf("UpdateStatus active: %v", err)
	}

	// disabled status
	if err := svc.UpdateStatus(context.Background(), created.ID, "disabled"); err != nil {
		t.Fatalf("UpdateStatus disabled: %v", err)
	}
}
