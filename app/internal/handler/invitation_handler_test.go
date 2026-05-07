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

func setupInvitationTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dirs := []string{"layout", "invitation"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	base := `{{define "base"}}{{block "content" .}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "layout", "base.html"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}

	accept := `{{define "content"}}{{if .Error}}<p class="error">{{.Error}}</p>{{end}}{{if .Token}}<form>ok</form>{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "invitation", "accept.html"), []byte(accept), 0o644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func makeValidToken(userID int64) *domain.InvitationToken {
	return &domain.InvitationToken{
		ID:        1,
		UserID:    userID,
		ExpiresAt: time.Now().Add(72 * time.Hour),
	}
}

func TestInvitationHandler_ShowAcceptForm_Valid(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{
		validateToken: makeValidToken(1),
		validateUser:  &domain.User{ID: 1, Name: "Test User", Email: "test@example.com"},
	}
	h := handler.NewInvitationHandler(rend, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/invite/validtoken", nil)
	req = withChiURLParam(req, "token", "validtoken")
	rr := httptest.NewRecorder()
	h.ShowAcceptForm(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestInvitationHandler_ShowAcceptForm_InvalidToken(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{validateErr: service.ErrTokenInvalid}
	h := handler.NewInvitationHandler(rend, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/invite/badtoken", nil)
	req = withChiURLParam(req, "token", "badtoken")
	rr := httptest.NewRecorder()
	h.ShowAcceptForm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "無効") {
		t.Errorf("expected invalid message in body, got: %s", body)
	}
}

func TestInvitationHandler_ShowAcceptForm_ExpiredToken(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{validateErr: service.ErrTokenExpired}
	h := handler.NewInvitationHandler(rend, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/invite/expiredtoken", nil)
	req = withChiURLParam(req, "token", "expiredtoken")
	rr := httptest.NewRecorder()
	h.ShowAcceptForm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "有効期限") {
		t.Errorf("expected expiry message in body, got: %s", body)
	}
}

func TestInvitationHandler_ShowAcceptForm_UsedToken(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{validateErr: service.ErrTokenUsed}
	h := handler.NewInvitationHandler(rend, invSvc)

	req := httptest.NewRequest(http.MethodGet, "/invite/usedtoken", nil)
	req = withChiURLParam(req, "token", "usedtoken")
	rr := httptest.NewRecorder()
	h.ShowAcceptForm(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestInvitationHandler_AcceptInvitation_Success(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{}
	h := handler.NewInvitationHandler(rend, invSvc)

	form := url.Values{
		"password":         {"newpassword123"},
		"password_confirm": {"newpassword123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/invite/validtoken", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "token", "validtoken")
	rr := httptest.NewRecorder()
	h.AcceptInvitation(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); !strings.Contains(loc, "invited=1") {
		t.Errorf("Location = %q, expected to contain invited=1", loc)
	}
}

func TestInvitationHandler_AcceptInvitation_PasswordMismatch(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{
		validateToken: makeValidToken(1),
		validateUser:  &domain.User{ID: 1, Name: "Test"},
	}
	h := handler.NewInvitationHandler(rend, invSvc)

	form := url.Values{
		"password":         {"password123"},
		"password_confirm": {"different123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/invite/validtoken", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "token", "validtoken")
	rr := httptest.NewRecorder()
	h.AcceptInvitation(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "一致") {
		t.Errorf("expected mismatch message in body, got: %s", body)
	}
}

func TestInvitationHandler_AcceptInvitation_AcceptError(t *testing.T) {
	dir := setupInvitationTemplateDir(t)
	rend := handler.NewRenderer(dir)
	invSvc := &mockInvitationService{acceptErr: service.ErrTokenExpired}
	h := handler.NewInvitationHandler(rend, invSvc)

	form := url.Values{
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/invite/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "token", "token")
	rr := httptest.NewRecorder()
	h.AcceptInvitation(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}
