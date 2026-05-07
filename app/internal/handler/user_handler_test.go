package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/handler"
	"github.com/kkitai/CondoManagerV2/app/internal/middleware"
	"github.com/kkitai/CondoManagerV2/app/internal/service"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

var errTest = errors.New("test error")

func setupUserTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dirs := []string{"layout", "users"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	base := `{{define "base"}}{{block "content" .}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "layout", "base.html"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}

	list := `{{define "content"}}{{range .Users}}<p>{{.Name}}</p>{{end}}{{end}}`
	form := `{{define "content"}}<form>{{if .Errors}}errors{{end}}</form>{{end}}`
	show := `{{define "content"}}<h1>{{.User.Name}}</h1>{{end}}`
	row := `<tr id="user-row-{{.ID}}"><td>{{.Name}}</td></tr>`
	table := `{{range .Users}}<tr>{{.Name}}</tr>{{end}}`
	stats := `{{define "_stats.html"}}{{.Total}}{{end}}`

	files := map[string]string{
		"list.html":   list,
		"form.html":   form,
		"show.html":   show,
		"_row.html":   row,
		"_table.html": table,
		"_stats.html": stats,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, "users", name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func withAdminUser(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKeyForTest(), makeAdminUser())
	return r.WithContext(ctx)
}

func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestUserHandler_List_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{
		listUsers: []*domain.User{makeAdminUser()},
		listTotal: 1,
		stats:     &domain.UserStats{Total: 1, Active: 1},
	}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = withAdminUser(req)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_List_HTMX(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{
		listUsers: []*domain.User{},
		stats:     &domain.UserStats{},
	}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_List_ServiceError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{listErr: errTest}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_List_StatsError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{
		listUsers: []*domain.User{},
		statsErr:  errTest,
	}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_New(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	h := handler.NewUserHandler(rend, &mockUserService{}, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/new", nil)
	req = withAdminUser(req)
	rr := httptest.NewRecorder()
	h.New(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_Create_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	created := &domain.User{ID: 2, Name: "New User", Email: "new@example.com", Role: domain.RoleGeneral}
	userSvc := &mockUserService{createUser: created}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{
		"email": {"new@example.com"},
		"name":  {"New User"},
		"role":  {"general"},
	}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withAdminUser(req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestUserHandler_Create_ValidationError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{createErr: validator.Errors{"email": "必須項目です"}}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"email": {""}, "name": {"Test"}, "role": {"general"}}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withAdminUser(req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
}

func TestUserHandler_Create_EmailExists(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{createErr: service.ErrEmailAlreadyExists}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"email": {"dup@example.com"}, "name": {"Test"}, "role": {"general"}}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withAdminUser(req)
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
}

func TestUserHandler_Create_InternalError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{createErr: errTest}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"email": {"test@example.com"}, "name": {"Test"}, "role": {"general"}}
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_Show_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getUser: makeAdminUser()}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	req = withAdminUser(req)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_Show_NotFound(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getErr: service.ErrUserNotFound}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/99", nil)
	req = withChiURLParam(req, "id", "99")
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestUserHandler_Show_InvalidID(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	h := handler.NewUserHandler(rend, &mockUserService{}, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	req = withChiURLParam(req, "id", "abc")
	rr := httptest.NewRecorder()
	h.Show(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}

func TestUserHandler_Edit_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getUser: makeAdminUser()}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/1/edit", nil)
	req = withAdminUser(req)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.Edit(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_Edit_NotFound(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getErr: service.ErrUserNotFound}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/99/edit", nil)
	req = withChiURLParam(req, "id", "99")
	rr := httptest.NewRecorder()
	h.Edit(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rr.Code)
	}
}

func TestUserHandler_Update_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{updateUser: makeAdminUser()}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{
		"email": {"admin@example.com"},
		"name":  {"Admin"},
		"role":  {"admin"},
	}
	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withAdminUser(req)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestUserHandler_Update_ValidationError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{
		updateErr: validator.Errors{"email": "必須項目です"},
		getUser:   makeAdminUser(),
	}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"email": {""}, "name": {"Test"}, "role": {"admin"}}
	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withAdminUser(req)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
}

func TestUserHandler_Update_InternalError(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{updateErr: errTest}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"email": {"test@example.com"}, "name": {"Test"}, "role": {"general"}}
	req := httptest.NewRequest(http.MethodPut, "/users/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_UpdateStatus_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getUser: makeAdminUser()}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"status": {"disabled"}}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestUserHandler_UpdateStatus_HTMX(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{getUser: makeAdminUser()}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"status": {"active"}}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_UpdateStatus_Error(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{statusErr: errTest}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	form := url.Values{"status": {"disabled"}}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.UpdateStatus(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_SendInvitation_Success(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	h := handler.NewUserHandler(rend, &mockUserService{}, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodPost, "/users/1/invite", nil)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.SendInvitation(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
}

func TestUserHandler_SendInvitation_HTMX(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	h := handler.NewUserHandler(rend, &mockUserService{}, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodPost, "/users/1/invite", nil)
	req.Header.Set("HX-Request", "true")
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.SendInvitation(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestUserHandler_SendInvitation_Error(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	h := handler.NewUserHandler(rend, &mockUserService{}, &mockInvitationService{sendErr: errTest})

	req := httptest.NewRequest(http.MethodPost, "/users/1/invite", nil)
	req = withChiURLParam(req, "id", "1")
	rr := httptest.NewRecorder()
	h.SendInvitation(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestUserHandler_Export(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{
		listUsers: []*domain.User{makeAdminUser()},
		listTotal: 1,
	}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/export", nil)
	rr := httptest.NewRecorder()
	h.Export(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, "csv") {
		t.Errorf("Content-Type = %q, expected csv", ct)
	}
}

func TestUserHandler_Export_Error(t *testing.T) {
	dir := setupUserTemplateDir(t)
	rend := handler.NewRenderer(dir)
	userSvc := &mockUserService{listErr: errTest}
	h := handler.NewUserHandler(rend, userSvc, &mockInvitationService{})

	req := httptest.NewRequest(http.MethodGet, "/users/export", nil)
	rr := httptest.NewRecorder()
	h.Export(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}
