package queryparam_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
)

func TestParseSort(t *testing.T) {
	allowed := []string{"name", "created_at", "updated_at"}

	tests := []struct {
		name      string
		query     string
		wantCol   string
		wantOrder queryparam.SortOrder
	}{
		{"default", "", "created_at", queryparam.SortDesc},
		{"valid asc", "sort=name&order=ASC", "name", queryparam.SortAsc},
		{"valid desc", "sort=created_at&order=DESC", "created_at", queryparam.SortDesc},
		{"invalid column", "sort=injection;DROP", "created_at", queryparam.SortDesc},
		{"invalid order defaults to DESC", "sort=name&order=INVALID", "name", queryparam.SortDesc},
		{"case insensitive order", "sort=name&order=asc", "name", queryparam.SortAsc},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tc.query, nil)
			s := queryparam.ParseSort(req, allowed, "created_at")
			if s.Column != tc.wantCol {
				t.Errorf("Column = %q, want %q", s.Column, tc.wantCol)
			}
			if s.Order != tc.wantOrder {
				t.Errorf("Order = %q, want %q", s.Order, tc.wantOrder)
			}
		})
	}
}

func TestSortSQL(t *testing.T) {
	s := queryparam.SortParams{Column: "name", Order: queryparam.SortAsc}
	if got := s.SQL(); got != "name ASC" {
		t.Errorf("SQL = %q, want %q", got, "name ASC")
	}

	empty := queryparam.SortParams{}
	if got := empty.SQL(); got != "" {
		t.Errorf("SQL = %q, want empty", got)
	}
}
