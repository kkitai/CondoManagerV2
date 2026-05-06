package csvexport

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"
)

type Exporter struct {
	w *csv.Writer
}

func New(w http.ResponseWriter, filename string) *Exporter {
	if filename == "" {
		filename = fmt.Sprintf("export_%s.csv", time.Now().Format("20060102_150405"))
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	// BOM for Excel UTF-8 compatibility
	fmt.Fprintf(w, "\xEF\xBB\xBF")
	return &Exporter{w: csv.NewWriter(w)}
}

func (e *Exporter) WriteHeader(columns []string) error {
	return e.w.Write(columns)
}

func (e *Exporter) WriteRow(row []string) error {
	return e.w.Write(row)
}

func (e *Exporter) Flush() {
	e.w.Flush()
}
