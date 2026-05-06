package config_test

import (
	"os"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/config"
)

func TestLoadWithCustomPort(t *testing.T) {
	os.Setenv("SESSION_SECRET", "secret")
	os.Setenv("PORT", "9090")
	os.Setenv("SESSION_MAX_AGE", "3600")
	os.Setenv("MAX_UPLOAD_SIZE", "5242880")
	os.Setenv("SERVER_READ_TIMEOUT", "30s")
	defer func() {
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("PORT")
		os.Unsetenv("SESSION_MAX_AGE")
		os.Unsetenv("MAX_UPLOAD_SIZE")
		os.Unsetenv("SERVER_READ_TIMEOUT")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Server.Port)
	}
	if cfg.App.SessionMaxAge != 3600 {
		t.Errorf("SessionMaxAge = %d, want 3600", cfg.App.SessionMaxAge)
	}
	if cfg.App.MaxUploadSize != 5242880 {
		t.Errorf("MaxUploadSize = %d, want 5242880", cfg.App.MaxUploadSize)
	}
}

func TestLoadInvalidTimeout(t *testing.T) {
	os.Setenv("SESSION_SECRET", "secret")
	os.Setenv("SERVER_READ_TIMEOUT", "not-a-duration")
	defer func() {
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("SERVER_READ_TIMEOUT")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	// Should fall back to default
	if cfg.Server.ReadTimeout == 0 {
		t.Error("expected non-zero ReadTimeout default")
	}
}

func TestLoadInvalidInt(t *testing.T) {
	os.Setenv("SESSION_SECRET", "secret")
	os.Setenv("SESSION_MAX_AGE", "not-a-number")
	os.Setenv("MAX_UPLOAD_SIZE", "not-a-number")
	defer func() {
		os.Unsetenv("SESSION_SECRET")
		os.Unsetenv("SESSION_MAX_AGE")
		os.Unsetenv("MAX_UPLOAD_SIZE")
	}()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	// Should fall back to defaults
	if cfg.App.SessionMaxAge == 0 {
		t.Error("expected non-zero SessionMaxAge default")
	}
}
