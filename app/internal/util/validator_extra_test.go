package util_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/util"
)

func TestValidatorMinLength(t *testing.T) {
	v := util.NewValidator()
	v.MinLength("password", "ab", 8)
	if v.Valid() {
		t.Error("expected invalid for too-short value")
	}
}

func TestValidatorMinLengthValid(t *testing.T) {
	v := util.NewValidator()
	v.MinLength("password", "abcdefgh", 8)
	if !v.Valid() {
		t.Error("expected valid")
	}
}

func TestFilterBuilderAdd(t *testing.T) {
	fb := util.NewFilterBuilder()
	fb.Add("status = $1", "active")
	clause := fb.WhereClause()
	if clause == "" {
		t.Error("expected non-empty where clause")
	}
	if len(fb.Args()) != 1 {
		t.Errorf("Args len = %d, want 1", len(fb.Args()))
	}
}

func TestFilterBuilderNextIndex(t *testing.T) {
	fb := util.NewFilterBuilder()
	if fb.NextIndex() != 1 {
		t.Errorf("NextIndex = %d, want 1", fb.NextIndex())
	}
	fb.AddEqual("status", "active")
	if fb.NextIndex() != 2 {
		t.Errorf("NextIndex = %d, want 2", fb.NextIndex())
	}
}
