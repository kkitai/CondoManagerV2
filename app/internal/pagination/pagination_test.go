package pagination_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/pagination"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		wantPage    int
		wantPerPage int
	}{
		{"defaults", "", pagination.DefaultPage, pagination.DefaultPerPage},
		{"valid values", "page=3&per_page=50", 3, 50},
		{"page below 1", "page=0", pagination.DefaultPage, pagination.DefaultPerPage},
		{"per_page above max", "per_page=200", pagination.DefaultPage, pagination.MaxPerPage},
		{"invalid strings", "page=abc&per_page=xyz", pagination.DefaultPage, pagination.DefaultPerPage},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tc.query, nil)
			p := pagination.Parse(req)
			if p.Page != tc.wantPage {
				t.Errorf("Page = %d, want %d", p.Page, tc.wantPage)
			}
			if p.PerPage != tc.wantPerPage {
				t.Errorf("PerPage = %d, want %d", p.PerPage, tc.wantPerPage)
			}
		})
	}
}

func TestOffsetLimit(t *testing.T) {
	p := pagination.Params{Page: 3, PerPage: 20}
	if got := p.Offset(); got != 40 {
		t.Errorf("Offset = %d, want 40", got)
	}
	if got := p.Limit(); got != 20 {
		t.Errorf("Limit = %d, want 20", got)
	}
}

func TestSetTotal(t *testing.T) {
	p := pagination.Params{Page: 1, PerPage: 20}
	p.SetTotal(55)
	if p.Total != 55 {
		t.Errorf("Total = %d, want 55", p.Total)
	}
	if p.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", p.TotalPages)
	}
}
