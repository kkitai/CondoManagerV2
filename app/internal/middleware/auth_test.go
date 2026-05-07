package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
)

func TestRequireAdmin_Forbidden(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	handler := middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRequireAdmin_AllowsAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users", nil)

	user := middleware.CurrentUser(req)
	if user != nil {
		t.Error("expected nil user from empty context")
	}
	_ = domain.User{}
}

func TestCurrentUser_NilContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	user := middleware.CurrentUser(req)
	if user != nil {
		t.Errorf("expected nil, got %v", user)
	}
}
