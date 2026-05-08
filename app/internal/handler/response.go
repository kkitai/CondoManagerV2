package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var templateFuncs = template.FuncMap{
	"deref": func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	},
	"string": func(v any) string {
		return fmt.Sprintf("%s", v)
	},
	"formatTime": func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("2006-01-02 15:04")
	},
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"add":      func(a, b int) int { return a + b },
	"subtract": func(a, b int) int { return a - b },
	"not":      func(b bool) bool { return !b },
	"derefFloat64": func(f *float64) float64 {
		if f == nil {
			return 0
		}
		return *f
	},
	"mulFloat64": func(a, b float64) float64 { return a * b },
	"derefInt": func(i *int) int {
		if i == nil {
			return 0
		}
		return *i
	},
}

type Renderer struct {
	templateDir string
}

func NewRenderer(templateDir string) *Renderer {
	return &Renderer{templateDir: templateDir}
}

func (r *Renderer) HTML(w http.ResponseWriter, status int, tmplName string, data any) {
	pattern := filepath.Join(r.templateDir, "layout", "*.html")
	tmplPath := filepath.Join(r.templateDir, tmplName)

	tmpl, err := template.New("").Funcs(templateFuncs).ParseGlob(pattern)
	if err != nil {
		http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err = tmpl.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "template execute error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (r *Renderer) Partial(w http.ResponseWriter, status int, tmplName string, data any) {
	tmplPath := filepath.Join(r.templateDir, tmplName)
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "template execute error: "+err.Error(), http.StatusInternalServerError)
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
