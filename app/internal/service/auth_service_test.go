package service_test

import (
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/service"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("expected non-empty hash")
	}

	if hash == password {
		t.Fatal("hash should not equal password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "mysecretpassword"
	hash, _ := service.HashPassword(password)

	if !service.CheckPassword(hash, password) {
		t.Fatal("expected password to match hash")
	}

	if service.CheckPassword(hash, "wrongpassword") {
		t.Fatal("expected wrong password to not match")
	}
}

func TestCheckPassword_EmptyHash(t *testing.T) {
	if service.CheckPassword("", "somepassword") {
		t.Fatal("empty hash should not match any password")
	}
}
