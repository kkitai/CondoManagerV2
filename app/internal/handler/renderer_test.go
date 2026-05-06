package handler_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/handler"
)

func setupTemplateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	layoutDir := filepath.Join(dir, "layout")
	if err := os.MkdirAll(layoutDir, 0o755); err != nil {
		t.Fatalf("mkdir layout: %v", err)
	}

	baseTmpl := `{{define "base"}}<!DOCTYPE html><html><body>{{template "content" .}}</body></html>{{end}}`
	if err := os.WriteFile(filepath.Join(layoutDir, "base.html"), []byte(baseTmpl), 0o644); err != nil {
		t.Fatalf("write base.html: %v", err)
	}

	pageTmpl := `{{define "content"}}<h1>{{.Title}}</h1>{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "page.html"), []byte(pageTmpl), 0o644); err != nil {
		t.Fatalf("write page.html: %v", err)
	}

	partialTmpl := `<p>{{.Message}}</p>`
	if err := os.WriteFile(filepath.Join(dir, "partial.html"), []byte(partialTmpl), 0o644); err != nil {
		t.Fatalf("write partial.html: %v", err)
	}

	return dir
}

func TestRendererHTML(t *testing.T) {
	dir := setupTemplateDir(t)
	r := handler.NewRenderer(dir)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = req

	r.HTML(rr, http.StatusOK, "page.html", map[string]string{"Title": "Test Page"})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestRendererHTMLMissingTemplate(t *testing.T) {
	r := handler.NewRenderer("/nonexistent/dir")

	rr := httptest.NewRecorder()
	r.HTML(rr, http.StatusOK, "missing.html", nil)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestRendererPartial(t *testing.T) {
	dir := setupTemplateDir(t)
	r := handler.NewRenderer(dir)

	rr := httptest.NewRecorder()
	r.Partial(rr, http.StatusOK, "partial.html", map[string]string{"Message": "hello"})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestRendererPartialMissingTemplate(t *testing.T) {
	r := handler.NewRenderer("/nonexistent/dir")

	rr := httptest.NewRecorder()
	r.Partial(rr, http.StatusOK, "missing.html", nil)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}
