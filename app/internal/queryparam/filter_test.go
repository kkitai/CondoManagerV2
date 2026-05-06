package queryparam_test

import (
	"testing"
	"time"

	"github.com/kkitai/CondoManagerV2/app/internal/queryparam"
)

func TestFilterBuilderEmpty(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	if got := fb.WhereClause(); got != "" {
		t.Errorf("WhereClause = %q, want empty", got)
	}
	if got := len(fb.Args()); got != 0 {
		t.Errorf("Args len = %d, want 0", got)
	}
}

func TestFilterBuilderAddLike(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.AddLike("name", "test")
	clause := fb.WhereClause()
	if clause != "WHERE name ILIKE $1" {
		t.Errorf("WhereClause = %q", clause)
	}
	args := fb.Args()
	if len(args) != 1 || args[0] != "%test%" {
		t.Errorf("Args = %v", args)
	}
}

func TestFilterBuilderAddLikeEmpty(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.AddLike("name", "")
	if got := fb.WhereClause(); got != "" {
		t.Errorf("expected empty clause, got %q", got)
	}
}

func TestFilterBuilderAddEqual(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.AddEqual("status", "active")
	clause := fb.WhereClause()
	if clause != "WHERE status = $1" {
		t.Errorf("WhereClause = %q", clause)
	}
}

func TestFilterBuilderAddEqualEmpty(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.AddEqual("status", "")
	if got := fb.WhereClause(); got != "" {
		t.Errorf("expected empty clause, got %q", got)
	}
}

func TestFilterBuilderMultiple(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.AddLike("name", "foo")
	fb.AddEqual("status", "active")
	clause := fb.WhereClause()
	if clause != "WHERE name ILIKE $1 AND status = $2" {
		t.Errorf("WhereClause = %q", clause)
	}
}

func TestFilterBuilderDateRange(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	fb.AddDateRange("created_at", from, to)
	clause := fb.WhereClause()
	if clause != "WHERE created_at >= $1 AND created_at <= $2" {
		t.Errorf("WhereClause = %q", clause)
	}
	if len(fb.Args()) != 2 {
		t.Errorf("Args len = %d, want 2", len(fb.Args()))
	}
}

func TestFilterBuilderAdd(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	fb.Add("status = $1", "active")
	clause := fb.WhereClause()
	if clause == "" {
		t.Error("expected non-empty where clause")
	}
}

func TestFilterBuilderNextIndex(t *testing.T) {
	fb := queryparam.NewFilterBuilder()
	if fb.NextIndex() != 1 {
		t.Errorf("NextIndex = %d, want 1", fb.NextIndex())
	}
	fb.AddEqual("status", "active")
	if fb.NextIndex() != 2 {
		t.Errorf("NextIndex = %d, want 2", fb.NextIndex())
	}
}
