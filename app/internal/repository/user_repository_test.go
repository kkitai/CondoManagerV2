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

func createTestUser(t *testing.T, repo *repository.UserRepository, email string) *domain.User {
	t.Helper()
	hash := "testhash"
	dept := "Engineering"
	now := time.Now()
	u := &domain.User{
		Email:        email,
		PasswordHash: &hash,
		Name:         "Test User",
		Role:         domain.RoleGeneral,
		Department:   &dept,
		Status:       domain.StatusInvited,
		InvitedAt:    &now,
	}
	created, err := repo.Create(context.Background(), u)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return created
}

func cleanupUsers(t *testing.T, pool interface{ Exec(context.Context, string, ...any) (interface{ RowsAffected() int64 }, error) }) {
	t.Helper()
}

func TestUserRepository_Create(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("create-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, repo, email)

	if u.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
	if u.Email != email {
		t.Errorf("email = %q, want %q", u.Email, email)
	}
	if u.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("findbyid-%d@test.com", time.Now().UnixNano())
	created := createTestUser(t, repo, email)

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Email != email {
		t.Errorf("email = %q, want %q", found.Email, email)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.FindByID(context.Background(), 999999)
	if err == nil {
		t.Error("expected error for missing user")
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("findbyemail-%d@test.com", time.Now().UnixNano())
	createTestUser(t, repo, email)

	found, err := repo.FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatalf("FindByEmail: %v", err)
	}
	if found.Email != email {
		t.Errorf("email = %q, want %q", found.Email, email)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.FindByEmail(context.Background(), "nonexistent@test.com")
	if err == nil {
		t.Error("expected error for missing email")
	}
}

func TestUserRepository_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	email1 := fmt.Sprintf("list1-%s@test.com", suffix)
	email2 := fmt.Sprintf("list2-%s@test.com", suffix)
	createTestUser(t, repo, email1)
	createTestUser(t, repo, email2)

	users, total, err := repo.List(context.Background(), domain.UserListParams{
		Search:  suffix,
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 2 {
		t.Errorf("total = %d, want >= 2", total)
	}
	_ = users
}

func TestUserRepository_List_WithFilters(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	users, total, err := repo.List(context.Background(), domain.UserListParams{
		Status:  "active",
		Role:    "admin",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List with filters: %v", err)
	}
	_ = users
	_ = total
}

func TestUserRepository_List_WithSort(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	users, _, err := repo.List(context.Background(), domain.UserListParams{
		SortColumn: "name",
		SortOrder:  "ASC",
		Page:       1,
		PerPage:    5,
	})
	if err != nil {
		t.Fatalf("List with sort: %v", err)
	}
	_ = users
}

func TestUserRepository_Update(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("update-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, repo, email)

	u.Name = "Updated Name"
	u.Role = domain.RoleAdmin

	updated, err := repo.Update(context.Background(), u)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Errorf("Name = %q, want 'Updated Name'", updated.Name)
	}
}

func TestUserRepository_UpdateStatus(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("status-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, repo, email)

	if err := repo.UpdateStatus(context.Background(), u.ID, "active"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), u.ID)
	if found.Status != domain.StatusActive {
		t.Errorf("status = %q, want 'active'", found.Status)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("pass-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, repo, email)

	if err := repo.UpdatePassword(context.Background(), u.ID, "newhash"); err != nil {
		t.Fatalf("UpdatePassword: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), u.ID)
	if found.PasswordHash == nil || *found.PasswordHash != "newhash" {
		t.Errorf("password hash not updated")
	}
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	email := fmt.Sprintf("login-%d@test.com", time.Now().UnixNano())
	u := createTestUser(t, repo, email)

	now := time.Now().UTC().Truncate(time.Second)
	if err := repo.UpdateLastLogin(context.Background(), u.ID, now); err != nil {
		t.Fatalf("UpdateLastLogin: %v", err)
	}
}

func TestUserRepository_GetStats(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewUserRepository(pool)

	stats, err := repo.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.Total < 0 {
		t.Error("stats.Total should be >= 0")
	}
}
