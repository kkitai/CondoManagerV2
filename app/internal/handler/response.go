package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

var buildingPalette = [][3]string{
	{"#CFE0EE", "#E9EFF4", "#3D4756"},
	{"#D8E4D2", "#EEF3EA", "#3F4E48"},
	{"#E5DDEE", "#F1ECF6", "#4A4456"},
	{"#EFE0CF", "#F7EFE3", "#594631"},
	{"#CDDBE6", "#E5ECF1", "#384857"},
	{"#E2D7CC", "#F1EAE0", "#503E2B"},
	{"#D7DCE8", "#ECEFF5", "#3B4255"},
	{"#D2E1DC", "#E6EFEC", "#36504A"},
	{"#E6D6D6", "#F2E7E7", "#5A3F3F"},
	{"#D5DEE0", "#E9EEEF", "#3B4949"},
}

func buildingThumb(id int64) template.HTML {
	p := buildingPalette[(id-1+int64(len(buildingPalette)))%int64(len(buildingPalette))]
	return template.HTML(fmt.Sprintf(`<svg viewBox="0 0 60 60" preserveAspectRatio="xMidYMid slice"><rect width="60" height="60" fill="%s"/><rect x="14" y="10" width="32" height="44" fill="%s" stroke="#9AA4B0" stroke-width=".5"/><g fill="%s"><rect x="18" y="14" width="6" height="4"/><rect x="27" y="14" width="6" height="4"/><rect x="36" y="14" width="6" height="4"/><rect x="18" y="22" width="6" height="4"/><rect x="27" y="22" width="6" height="4"/><rect x="36" y="22" width="6" height="4"/><rect x="18" y="30" width="6" height="4"/><rect x="27" y="30" width="6" height="4"/><rect x="36" y="30" width="6" height="4"/><rect x="18" y="38" width="6" height="4"/><rect x="27" y="38" width="6" height="4"/><rect x="36" y="38" width="6" height="4"/><rect x="27" y="46" width="6" height="8"/></g></svg>`, p[0], p[1], p[2]))
}

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
	"buildingThumb": buildingThumb,
	"formatFloat1": func(f float64) string { return fmt.Sprintf("%.1f", f) },
	"multiply":     func(a, b int) int { return a * b },
	"minInt": func(a, b int) int {
		if a < b {
			return a
		}
		return b
	},
	"int64ToInt": func(n int64) int { return int(n) },
	"iterate": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	},
	"percentInt": func(part, total int64) string {
		if total == 0 {
			return "0"
		}
		return fmt.Sprintf("%.0f", float64(part)/float64(total)*100)
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
	// load partials (_*.html) in the same directory as the template
	partialsPattern := filepath.Join(filepath.Dir(tmplPath), "_*.html")
	if matches, _ := filepath.Glob(partialsPattern); len(matches) > 0 {
		tmpl, err = tmpl.ParseGlob(partialsPattern)
		if err != nil {
			http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
			return
		}
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
