package util

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"
)

type CSVExporter struct {
	w *csv.Writer
}

func NewCSVExporter(w http.ResponseWriter, filename string) *CSVExporter {
	if filename == "" {
		filename = fmt.Sprintf("export_%s.csv", time.Now().Format("20060102_150405"))
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	// BOM for Excel UTF-8 compatibility
	fmt.Fprintf(w, "\xEF\xBB\xBF")
	return &CSVExporter{w: csv.NewWriter(w)}
}

func (e *CSVExporter) WriteHeader(columns []string) error {
	return e.w.Write(columns)
}

func (e *CSVExporter) WriteRow(row []string) error {
	return e.w.Write(row)
}

func (e *CSVExporter) Flush() {
	e.w.Flush()
}
