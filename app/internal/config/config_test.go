package config_test

import (
	"os"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/config"
)

func TestLoad(t *testing.T) {
	os.Setenv("SESSION_SECRET", "test-secret")
	defer os.Unsetenv("SESSION_SECRET")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Server.Port, "8080")
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("SSLMode = %q, want %q", cfg.Database.SSLMode, "disable")
	}
	if cfg.App.SessionSecret != "test-secret" {
		t.Errorf("SessionSecret = %q", cfg.App.SessionSecret)
	}
}

func TestLoadMissingSecret(t *testing.T) {
	os.Unsetenv("SESSION_SECRET")
	_, err := config.Load()
	if err == nil {
		t.Error("expected error when SESSION_SECRET is missing")
	}
}

func TestDSN(t *testing.T) {
	d := config.DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "secret",
		DBName:   "mydb",
		SSLMode:  "disable",
	}
	dsn := d.DSN()
	if dsn == "" {
		t.Error("expected non-empty DSN")
	}
}
