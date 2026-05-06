package handler

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	dbStatus := "ok"

	if h.db != nil {
		if err := h.db.Ping(context.Background()); err != nil {
			dbStatus = "error"
			status = "degraded"
		}
	}

	code := http.StatusOK
	if status != "ok" {
		code = http.StatusServiceUnavailable
	}

	JSON(w, code, map[string]string{
		"status":   status,
		"database": dbStatus,
	})
}
