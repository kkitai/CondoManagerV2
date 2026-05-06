package util

import (
	"fmt"
	"strings"
	"time"
)

type FilterBuilder struct {
	conditions []string
	args       []any
	argIndex   int
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{argIndex: 1}
}

func (f *FilterBuilder) Add(condition string, value any) {
	f.conditions = append(f.conditions, condition)
	f.args = append(f.args, value)
	f.argIndex++
}

func (f *FilterBuilder) AddLike(column string, value string) {
	if value == "" {
		return
	}
	f.conditions = append(f.conditions, fmt.Sprintf("%s ILIKE $%d", column, f.argIndex))
	f.args = append(f.args, "%"+value+"%")
	f.argIndex++
}

func (f *FilterBuilder) AddEqual(column string, value any) {
	if value == nil || value == "" {
		return
	}
	f.conditions = append(f.conditions, fmt.Sprintf("%s = $%d", column, f.argIndex))
	f.args = append(f.args, value)
	f.argIndex++
}

func (f *FilterBuilder) AddDateRange(column string, from, to time.Time) {
	if !from.IsZero() {
		f.conditions = append(f.conditions, fmt.Sprintf("%s >= $%d", column, f.argIndex))
		f.args = append(f.args, from)
		f.argIndex++
	}
	if !to.IsZero() {
		f.conditions = append(f.conditions, fmt.Sprintf("%s <= $%d", column, f.argIndex))
		f.args = append(f.args, to)
		f.argIndex++
	}
}

func (f *FilterBuilder) WhereClause() string {
	if len(f.conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(f.conditions, " AND ")
}

func (f *FilterBuilder) Args() []any {
	return f.args
}

func (f *FilterBuilder) NextIndex() int {
	return f.argIndex
}
