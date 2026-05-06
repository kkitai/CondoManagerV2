package util_test

import (
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/util"
)

func TestParseSort(t *testing.T) {
	allowed := []string{"name", "created_at", "updated_at"}

	tests := []struct {
		name       string
		query      string
		wantCol    string
		wantOrder  util.SortOrder
	}{
		{"default", "", "created_at", util.SortDesc},
		{"valid asc", "sort=name&order=ASC", "name", util.SortAsc},
		{"valid desc", "sort=created_at&order=DESC", "created_at", util.SortDesc},
		{"invalid column", "sort=injection;DROP", "created_at", util.SortDesc},
		{"invalid order defaults to DESC", "sort=name&order=INVALID", "name", util.SortDesc},
		{"case insensitive order", "sort=name&order=asc", "name", util.SortAsc},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tc.query, nil)
			s := util.ParseSort(req, allowed, "created_at")
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
	s := util.SortParams{Column: "name", Order: util.SortAsc}
	if got := s.SQL(); got != "name ASC" {
		t.Errorf("SQL = %q, want %q", got, "name ASC")
	}

	empty := util.SortParams{}
	if got := empty.SQL(); got != "" {
		t.Errorf("SQL = %q, want empty", got)
	}
}
