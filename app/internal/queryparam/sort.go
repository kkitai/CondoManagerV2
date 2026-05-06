package queryparam

import (
	"fmt"
	"net/http"
	"strings"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type SortParams struct {
	Column string
	Order  SortOrder
}

func ParseSort(r *http.Request, allowedColumns []string, defaultColumn string) SortParams {
	column := r.URL.Query().Get("sort")
	order := strings.ToUpper(r.URL.Query().Get("order"))

	if !isAllowed(column, allowedColumns) {
		column = defaultColumn
	}

	if order != string(SortAsc) && order != string(SortDesc) {
		order = string(SortDesc)
	}

	return SortParams{
		Column: column,
		Order:  SortOrder(order),
	}
}

func (s SortParams) SQL() string {
	if s.Column == "" {
		return ""
	}
	return fmt.Sprintf("%s %s", s.Column, s.Order)
}

func isAllowed(column string, allowed []string) bool {
	for _, a := range allowed {
		if a == column {
			return true
		}
	}
	return false
}
