package util

import (
	"math"
	"net/http"
	"strconv"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type PaginationParams struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
}

func ParsePagination(r *http.Request) PaginationParams {
	page := parseInt(r.URL.Query().Get("page"), DefaultPage)
	perPage := parseInt(r.URL.Query().Get("per_page"), DefaultPerPage)

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

func (p *PaginationParams) SetTotal(total int64) {
	p.Total = total
	if p.PerPage > 0 {
		p.TotalPages = int(math.Ceil(float64(total) / float64(p.PerPage)))
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p PaginationParams) Limit() int {
	return p.PerPage
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return v
}
