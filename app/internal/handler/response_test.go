package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/handler"
)

func TestJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	handler.JSON(rr, http.StatusOK, map[string]string{"key": "value"})

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q", ct)
	}
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("body[key] = %q, want value", body["key"])
	}
}

func TestJSONCreated(t *testing.T) {
	rr := httptest.NewRecorder()
	handler.JSON(rr, http.StatusCreated, map[string]int{"id": 1})

	if rr.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", rr.Code)
	}
}

func TestErrorNonHTMX(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = req
	handler.Error(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rr.Code)
	}
}
