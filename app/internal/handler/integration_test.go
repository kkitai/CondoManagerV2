package handler_test

// Integration tests: real DB (seed fixtures) + real templates.
// Skipped unless DB_HOST is set (same condition as repository integration tests).
//
// Purpose: catch template rendering errors (missing partials, broken expressions,
// type mismatches) that unit tests with stub templates cannot detect.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/handler"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/repository"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/testutil"
)

// withUser injects a domain.User into the request context (bypasses auth middleware).
func withUser(r *http.Request, u *domain.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKeyForTest(), u)
	return r.WithContext(ctx)
}

// seedAdmin returns the admin user inserted by the seed fixture.
func seedAdmin() *domain.User {
	return &domain.User{
		ID:    1,
		Name:  "田中 太郎",
		Email: "admin@example.com",
		Role:  domain.RoleAdmin,
	}
}

// assertOK fails the test if the response is not 200 or contains a template error.
func assertOK(t *testing.T, rr *httptest.ResponseRecorder, label string) {
	t.Helper()
	if rr.Code != http.StatusOK {
		t.Errorf("%s: status = %d, body = %s", label, rr.Code, rr.Body.String())
		return
	}
	if body := rr.Body.String(); strings.Contains(body, "template") && strings.Contains(body, "error") {
		t.Errorf("%s: response body contains template error: %s", label, body)
	}
}

func TestIntegration_UserHandler_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	userSvc := service.NewUserService(userRepo)
	invSvc := service.NewInvitationService(userRepo, invRepo, nil, "")
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewUserHandler(rend, userSvc, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.List(rr, req)

	assertOK(t, rr, "GET /users")
}

func TestIntegration_UserHandler_New(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	userRepo := repository.NewUserRepository(pool)
	invRepo := repository.NewInvitationRepository(pool)
	userSvc := service.NewUserService(userRepo)
	invSvc := service.NewInvitationService(userRepo, invRepo, nil, "")
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewUserHandler(rend, userSvc, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/users/new", nil)
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.New(rr, req)

	assertOK(t, rr, "GET /users/new")
}

func TestIntegration_PropertyHandler_List(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	propRepo := repository.NewPropertyRepository(pool)
	propSvc := service.NewPropertyService(propRepo)
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewPropertyHandler(rend, propSvc)

	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.List(rr, req)

	assertOK(t, rr, "GET /properties")
}

func TestIntegration_PropertyHandler_Show(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	propRepo := repository.NewPropertyRepository(pool)
	propSvc := service.NewPropertyService(propRepo)
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewPropertyHandler(rend, propSvc)

	req := httptest.NewRequest(http.MethodGet, "/properties/1", nil)
	req = withChiURLParam(req, "id", "1")
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	assertOK(t, rr, "GET /properties/1")
}

func TestIntegration_PropertyHandler_New(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	propRepo := repository.NewPropertyRepository(pool)
	propSvc := service.NewPropertyService(propRepo)
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewPropertyHandler(rend, propSvc)

	req := httptest.NewRequest(http.MethodGet, "/properties/new", nil)
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.New(rr, req)

	assertOK(t, rr, "GET /properties/new")
}

func TestIntegration_PropertyHandler_Edit(t *testing.T) {
	pool := testutil.NewTestDB(t)
	testutil.SeedTestDB(t, pool)

	propRepo := repository.NewPropertyRepository(pool)
	propSvc := service.NewPropertyService(propRepo)
	rend := handler.NewRenderer(testutil.TemplateDir())
	h := handler.NewPropertyHandler(rend, propSvc)

	req := httptest.NewRequest(http.MethodGet, "/properties/1/edit", nil)
	req = withChiURLParam(req, "id", "1")
	req = withUser(req, seedAdmin())
	rr := httptest.NewRecorder()
	h.Edit(rr, req)

	assertOK(t, rr, "GET /properties/1/edit")
}
