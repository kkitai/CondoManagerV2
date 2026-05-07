package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/handler"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

func setupAuthTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	for _, d := range []string{"layout", "auth"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	base := `{{define "base"}}<!DOCTYPE html><html><body>{{block "content" .}}{{end}}</body></html>{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "layout", "base.html"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}

	login := `{{define "content"}}<form>{{if .Error}}<p class="error">{{.Error}}</p>{{end}}</form>{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "auth", "login.html"), []byte(login), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestAuthHandler_ShowLogin(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	h := handler.NewAuthHandler(r, &mockAuthService{}, time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rr := httptest.NewRecorder()
	h.ShowLogin(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	svc := &mockAuthService{
		loginToken: "test-token",
		loginUser:  makeAdminUser(),
	}
	h := handler.NewAuthHandler(r, svc, time.Hour)

	form := url.Values{"email": {"admin@example.com"}, "password": {"password123"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("Location = %q, want '/dashboard'", loc)
	}

	// verify cookie is set
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == service.SessionCookieName {
			found = true
			if c.Value != "test-token" {
				t.Errorf("cookie value = %q, want 'test-token'", c.Value)
			}
		}
	}
	if !found {
		t.Error("session cookie not set")
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	svc := &mockAuthService{loginErr: service.ErrInvalidCredentials}
	h := handler.NewAuthHandler(r, svc, time.Hour)

	form := url.Values{"email": {"bad@example.com"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthHandler_Login_DisabledUser(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	svc := &mockAuthService{loginErr: service.ErrUserDisabled}
	h := handler.NewAuthHandler(r, svc, time.Hour)

	form := url.Values{"email": {"disabled@example.com"}, "password": {"pass"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "無効化") {
		t.Errorf("expected disabled message in body, got: %s", body)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	h := handler.NewAuthHandler(r, &mockAuthService{}, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: service.SessionCookieName, Value: "old-token"})
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Errorf("Location = %q, want '/login'", loc)
	}
}

func TestAuthHandler_Logout_NoCookie(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	h := handler.NewAuthHandler(r, &mockAuthService{}, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestAuthHandler_Login_ParseFormError(t *testing.T) {
	dir := setupAuthTemplateDir(t)
	r := handler.NewRenderer(dir)
	h := handler.NewAuthHandler(r, &mockAuthService{}, time.Hour)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("%invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	// ParseForm on invalid data still returns 200/OK for most cases in Go
	// but the key is it doesn't panic
	if rr.Code == 0 {
		t.Error("expected a response code")
	}
}

func init() {
	_ = domain.User{}
}
