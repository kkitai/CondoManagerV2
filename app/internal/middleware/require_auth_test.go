package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

type mockAuthSvc struct {
	user *domain.User
	err  error
}

func (m *mockAuthSvc) Login(_ context.Context, _, _ string, _ *http.Request) (string, *domain.User, error) {
	return "", nil, nil
}
func (m *mockAuthSvc) Logout(_ context.Context, _ string) error { return nil }
func (m *mockAuthSvc) GetUserByToken(_ context.Context, _ string) (*domain.User, error) {
	return m.user, m.err
}

func TestRequireAuth_NoCookie(t *testing.T) {
	svc := &mockAuthSvc{}
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want /login", loc)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	svc := &mockAuthSvc{err: service.ErrSessionNotFound}
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: service.SessionCookieName, Value: "invalid-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect, got %d", rr.Code)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	user := &domain.User{ID: 1, Role: domain.RoleAdmin, Status: domain.StatusActive}
	svc := &mockAuthSvc{user: user}
	handler := middleware.RequireAuth(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := middleware.CurrentUser(r)
		if u == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: service.SessionCookieName, Value: "valid-token"})
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequireAdmin_WithAdminUser(t *testing.T) {
	user := &domain.User{Role: domain.RoleAdmin, Status: domain.StatusActive}
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKeyForTest(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d", rr.Code)
	}
}

func TestRequireAdmin_WithGeneralUser(t *testing.T) {
	user := &domain.User{Role: domain.RoleGeneral, Status: domain.StatusActive}
	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), middleware.UserContextKeyForTest(), user)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 for general user, got %d", rr.Code)
	}
}
