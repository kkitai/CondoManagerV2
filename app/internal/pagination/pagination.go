package pagination

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

type Params struct {
	Page       int
	PerPage    int
	Total      int64
	TotalPages int
}

func Parse(r *http.Request) Params {
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

	return Params{
		Page:    page,
		PerPage: perPage,
	}
}

func (p *Params) SetTotal(total int64) {
	p.Total = total
	if p.PerPage > 0 {
		p.TotalPages = int(math.Ceil(float64(total) / float64(p.PerPage)))
	}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

func (p Params) Limit() int {
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
