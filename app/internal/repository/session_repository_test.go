package repository_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/repository"
)

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-value"
	h1 := repository.HashToken(token)
	h2 := repository.HashToken(token)

	if h1 != h2 {
		t.Error("HashToken should be deterministic")
	}

	if h1 == token {
		t.Error("hash should not equal input")
	}

	if len(h1) != 64 {
		t.Errorf("expected SHA-256 hex length 64, got %d", len(h1))
	}
}

func TestHashToken_DifferentInputs(t *testing.T) {
	h1 := repository.HashToken("token-a")
	h2 := repository.HashToken("token-b")

	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}
