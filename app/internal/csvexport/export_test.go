package csvexport_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/csvexport"
)

func TestExporter(t *testing.T) {
	rr := httptest.NewRecorder()
	exp := csvexport.New(rr, "test.csv")

	if err := exp.WriteHeader([]string{"id", "name", "email"}); err != nil {
		t.Fatalf("WriteHeader error: %v", err)
	}
	if err := exp.WriteRow([]string{"1", "山田太郎", "yamada@example.com"}); err != nil {
		t.Fatalf("WriteRow error: %v", err)
	}
	exp.Flush()

	res := rr.Result()
	if ct := res.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/csv") {
		t.Errorf("Content-Type = %q, want text/csv", ct)
	}
	if cd := res.Header.Get("Content-Disposition"); !strings.Contains(cd, "test.csv") {
		t.Errorf("Content-Disposition = %q", cd)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "id") || !strings.Contains(body, "山田太郎") {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestExporterDefaultFilename(t *testing.T) {
	rr := httptest.NewRecorder()
	exp := csvexport.New(rr, "")
	exp.Flush()

	cd := rr.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "export_") {
		t.Errorf("Content-Disposition = %q, expected default filename", cd)
	}
}
