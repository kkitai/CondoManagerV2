package service_test

import (
	"context"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

func TestPropertyService_Create(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)
	svc := service.NewPropertyService(repo)

	t.Run("valid input", func(t *testing.T) {
		p, err := svc.Create(context.Background(), service.CreatePropertyInput{
			Name:    "Test Property",
			Address: "Tokyo, Japan",
			Status:  "active",
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if p.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := svc.Create(context.Background(), service.CreatePropertyInput{
			Address: "Tokyo, Japan",
			Status:  "active",
		})
		if err == nil {
			t.Error("expected validation error for missing name")
		}
	})

	t.Run("missing address", func(t *testing.T) {
		_, err := svc.Create(context.Background(), service.CreatePropertyInput{
			Name:   "Test",
			Status: "active",
		})
		if err == nil {
			t.Error("expected validation error for missing address")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		_, err := svc.Create(context.Background(), service.CreatePropertyInput{
			Name:    "Test",
			Address: "Tokyo",
			Status:  "unknown",
		})
		if err == nil {
			t.Error("expected validation error for invalid status")
		}
	})

	t.Run("with area and unit_count", func(t *testing.T) {
		p, err := svc.Create(context.Background(), service.CreatePropertyInput{
			Name:      "Property With Numbers",
			Address:   "Osaka, Japan",
			Status:    "active",
			Area:      "2500.5",
			UnitCount: "50",
		})
		if err != nil {
			t.Fatalf("Create with numbers: %v", err)
		}
		if p.Area == nil || *p.Area != 2500.5 {
			t.Errorf("area = %v, want 2500.5", p.Area)
		}
		if p.UnitCount == nil || *p.UnitCount != 50 {
			t.Errorf("unit_count = %v, want 50", p.UnitCount)
		}
	})
}

func TestPropertyService_Update(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)
	svc := service.NewPropertyService(repo)

	created, err := svc.Create(context.Background(), service.CreatePropertyInput{
		Name:    "Original Name",
		Address: "Original Address",
		Status:  "active",
	})
	if err != nil {
		t.Fatalf("setup Create: %v", err)
	}

	t.Run("valid update", func(t *testing.T) {
		updated, err := svc.Update(context.Background(), created.ID, service.UpdatePropertyInput{
			Name:    "Updated Name",
			Address: "Updated Address",
			Status:  "inactive",
		})
		if err != nil {
			t.Fatalf("Update: %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Errorf("name = %q, want 'Updated Name'", updated.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Update(context.Background(), 999999999, service.UpdatePropertyInput{
			Name:    "X",
			Address: "X",
			Status:  "active",
		})
		if err != service.ErrPropertyNotFound {
			t.Errorf("expected ErrPropertyNotFound, got %v", err)
		}
	})
}

func TestPropertyService_Delete(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)
	svc := service.NewPropertyService(repo)

	created, err := svc.Create(context.Background(), service.CreatePropertyInput{
		Name:    "To Delete",
		Address: "Somewhere",
		Status:  "active",
	})
	if err != nil {
		t.Fatalf("setup Create: %v", err)
	}

	if err := svc.Delete(context.Background(), created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = svc.GetByID(context.Background(), created.ID)
	if err != service.ErrPropertyNotFound {
		t.Errorf("expected ErrPropertyNotFound after delete, got %v", err)
	}
}

func TestPropertyService_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)
	svc := service.NewPropertyService(repo)

	props, total, err := svc.List(context.Background(), domain.PropertyListParams{Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = props
	_ = total
}

func TestPropertyService_GetStats(t *testing.T) {
	pool := testutil.NewTestDB(t)
	repo := repository.NewPropertyRepository(pool)
	svc := service.NewPropertyService(repo)

	created, err := svc.Create(context.Background(), service.CreatePropertyInput{
		Name:    "Stats Property",
		Address: "Tokyo",
		Status:  "active",
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	stats, err := svc.GetStats(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.TotalClaims != 0 {
		t.Errorf("expected 0 total claims, got %d", stats.TotalClaims)
	}
}
