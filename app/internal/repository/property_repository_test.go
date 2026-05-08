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

func createTestProperty(t *testing.T, repo *repository.PropertyRepository, name string) *domain.Property {
	t.Helper()
	area := 1500.5
	units := 20
	mgmt := "Test Management Co"
	p := &domain.Property{
		Name:              name,
		Address:           "Tokyo, Japan",
		Area:              &area,
		UnitCount:         &units,
		Status:            domain.PropertyStatusActive,
		ManagementCompany: &mgmt,
	}
	created, err := repo.Create(context.Background(), p)
	if err != nil {
		t.Fatalf("create property: %v", err)
	}
	return created
}

func TestPropertyRepository_Create(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	name := fmt.Sprintf("prop-create-%d", time.Now().UnixNano())
	p := createTestProperty(t, repo, name)

	if p.ID == 0 {
		t.Error("expected non-zero ID after create")
	}
	if p.Name != name {
		t.Errorf("name = %q, want %q", p.Name, name)
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected non-zero created_at")
	}
}

func TestPropertyRepository_FindByID(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	name := fmt.Sprintf("prop-find-%d", time.Now().UnixNano())
	created := createTestProperty(t, repo, name)

	found, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Name != name {
		t.Errorf("name = %q, want %q", found.Name, name)
	}
	if found.Status != domain.PropertyStatusActive {
		t.Errorf("status = %q, want active", found.Status)
	}
}

func TestPropertyRepository_FindByID_NotFound(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	_, err := repo.FindByID(context.Background(), 999999999)
	if err == nil {
		t.Error("expected error for nonexistent property")
	}
}

func TestPropertyRepository_Update(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	name := fmt.Sprintf("prop-update-%d", time.Now().UnixNano())
	p := createTestProperty(t, repo, name)

	newName := name + "-updated"
	p.Name = newName
	p.Status = domain.PropertyStatusInactive

	updated, err := repo.Update(context.Background(), p)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != newName {
		t.Errorf("name = %q, want %q", updated.Name, newName)
	}

	found, err := repo.FindByID(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("FindByID after update: %v", err)
	}
	if found.Status != domain.PropertyStatusInactive {
		t.Errorf("status = %q, want inactive", found.Status)
	}
}

func TestPropertyRepository_Delete(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	name := fmt.Sprintf("prop-delete-%d", time.Now().UnixNano())
	p := createTestProperty(t, repo, name)

	if err := repo.Delete(context.Background(), p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(context.Background(), p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestPropertyRepository_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	prefix := fmt.Sprintf("listtest-%d-", time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		createTestProperty(t, repo, fmt.Sprintf("%s%d", prefix, i))
	}

	props, total, err := repo.List(context.Background(), domain.PropertyListParams{
		Search:  prefix,
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total < 3 {
		t.Errorf("total = %d, want >= 3", total)
	}
	if len(props) < 3 {
		t.Errorf("len(props) = %d, want >= 3", len(props))
	}
}

func TestPropertyRepository_List_StatusFilter(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	prefix := fmt.Sprintf("statusfilter-%d-", time.Now().UnixNano())
	p := createTestProperty(t, repo, prefix+"active")
	p2 := createTestProperty(t, repo, prefix+"inactive")
	p2.Status = domain.PropertyStatusInactive
	_, err := repo.Update(context.Background(), p2)
	if err != nil {
		t.Fatalf("update p2: %v", err)
	}
	_ = p

	active, _, err := repo.List(context.Background(), domain.PropertyListParams{
		Search:  prefix,
		Status:  "active",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	for _, prop := range active {
		if prop.Status != domain.PropertyStatusActive {
			t.Errorf("expected active status, got %q", prop.Status)
		}
	}
}

func TestPropertyRepository_List_Pagination(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	prefix := fmt.Sprintf("pagtest-%d-", time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		createTestProperty(t, repo, fmt.Sprintf("%s%d", prefix, i))
	}

	page1, total, err := repo.List(context.Background(), domain.PropertyListParams{
		Search:  prefix,
		Page:    1,
		PerPage: 2,
	})
	if err != nil {
		t.Fatalf("List page1: %v", err)
	}
	if total < 5 {
		t.Errorf("total = %d, want >= 5", total)
	}
	if len(page1) != 2 {
		t.Errorf("len(page1) = %d, want 2", len(page1))
	}

	page2, _, err := repo.List(context.Background(), domain.PropertyListParams{
		Search:  prefix,
		Page:    2,
		PerPage: 2,
	})
	if err != nil {
		t.Fatalf("List page2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("len(page2) = %d, want 2", len(page2))
	}
	if page1[0].ID == page2[0].ID {
		t.Error("page1 and page2 returned same first item")
	}
}

func TestPropertyRepository_GetStats(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)

	name := fmt.Sprintf("prop-stats-%d", time.Now().UnixNano())
	p := createTestProperty(t, repo, name)

	stats, err := repo.GetStats(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalClaims != 0 {
		t.Errorf("expected 0 claims for new property, got %d", stats.TotalClaims)
	}
}
