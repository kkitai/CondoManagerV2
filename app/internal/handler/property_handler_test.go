package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/domain"
	"github.com/kkitai/CondoManagerV2/app/internal/handler"
	"github.com/kkitai/CondoManagerV2/app/internal/validator"
)

func setupPropertyTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dirs := []string{"layout", "properties"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	base := `{{define "base"}}{{block "content" .}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "layout", "base.html"), []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}

	tmpl := func(name, body string) {
		p := filepath.Join(dir, "properties", name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	tmpl("list.html", `{{define "content"}}<ul>{{range .Properties}}<li>{{.Name}}</li>{{end}}</ul>{{end}}`)
	tmpl("_table.html", `<ul>{{range .Properties}}<li>{{.Name}}</li>{{end}}</ul>`)
	tmpl("form.html", `{{define "content"}}<form>{{if .Errors}}errors{{end}}</form>{{end}}`)
	tmpl("detail.html", `{{define "content"}}<h1>{{.Property.Name}}</h1>{{end}}`)
	tmpl("_stats.html", `<p>stats</p>`)

	return dir
}

func makeTestProperty(id int64, name string) *domain.Property {
	return &domain.Property{
		ID:      id,
		Name:    name,
		Address: "Tokyo, Japan",
		Status:  domain.PropertyStatusActive,
	}
}

func makeTestPropertyStats() *domain.PropertyStats {
	return &domain.PropertyStats{
		TotalClaims:     5,
		OpenClaims:      2,
		CompletedClaims: 3,
	}
}

func TestPropertyHandler_List(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	props := []*domain.Property{
		makeTestProperty(1, "Property A"),
		makeTestProperty(2, "Property B"),
	}
	svc := &mockPropertyService{listProps: props, listTotal: 2}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties", nil)
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestPropertyHandler_List_Error(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{listErr: errTest}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties", nil)
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.List(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestPropertyHandler_List_HTMX(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{listProps: []*domain.Property{makeTestProperty(1, "P")}}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties", nil)
	r.Header.Set("HX-Request", "true")
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestPropertyHandler_New(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties/new", nil)
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.New(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestPropertyHandler_Create_Success(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	created := makeTestProperty(10, "New Property")
	svc := &mockPropertyService{createProp: created}
	h := handler.NewPropertyHandler(renderer, svc)

	form := url.Values{
		"name":    {"New Property"},
		"address": {"Tokyo"},
		"status":  {"active"},
	}
	r := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", w.Code)
	}
}

func TestPropertyHandler_Create_ValidationError(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{createErr: validator.Errors{"name": "必須項目です"}}
	h := handler.NewPropertyHandler(renderer, svc)

	form := url.Values{"name": {""}, "status": {"active"}}
	r := httptest.NewRequest(http.MethodPost, "/properties", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = withAdminUser(r)
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

func TestPropertyHandler_Show(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	prop := makeTestProperty(1, "Property A")
	stats := makeTestPropertyStats()
	svc := &mockPropertyService{getProp: prop, stats: stats}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties/1", nil)
	r = withAdminUser(r)
	r = withChiURLParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.Show(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestPropertyHandler_Show_NotFound(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{getErr: errTest}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodGet, "/properties/999", nil)
	r = withAdminUser(r)
	r = withChiURLParam(r, "id", "999")
	w := httptest.NewRecorder()

	h.Show(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestPropertyHandler_Delete_Success(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodDelete, "/properties/1", nil)
	r = withAdminUser(r)
	r = withChiURLParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", w.Code)
	}
}

func TestPropertyHandler_Delete_Error(t *testing.T) {
	dir := setupPropertyTemplateDir(t)
	renderer := handler.NewRenderer(dir)
	svc := &mockPropertyService{deleteErr: errTest}
	h := handler.NewPropertyHandler(renderer, svc)

	r := httptest.NewRequest(http.MethodDelete, "/properties/1", nil)
	r = withAdminUser(r)
	r = withChiURLParam(r, "id", "1")
	w := httptest.NewRecorder()

	h.Delete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}
