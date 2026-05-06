package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templateDir string
}

func NewRenderer(templateDir string) *Renderer {
	return &Renderer{templateDir: templateDir}
}

func (r *Renderer) HTML(w http.ResponseWriter, status int, tmplName string, data any) {
	pattern := filepath.Join(r.templateDir, "layout", "*.html")
	tmplPath := filepath.Join(r.templateDir, tmplName)

	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}
	tmpl, err = tmpl.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "template execute error", http.StatusInternalServerError)
	}
}

func (r *Renderer) Partial(w http.ResponseWriter, status int, tmplName string, data any) {
	tmplPath := filepath.Join(r.templateDir, tmplName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "template execute error", http.StatusInternalServerError)
	}
}

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, message string) {
	if isHTMXRequest(w, nil) {
		http.Error(w, message, status)
		return
	}
	JSON(w, status, map[string]string{"error": message})
}

func isHTMXRequest(_ http.ResponseWriter, r *http.Request) bool {
	if r == nil {
		return false
	}
	return r.Header.Get("HX-Request") == "true"
}
