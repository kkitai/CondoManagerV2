package mailer_test

import (
	"context"
	"testing"

	"github.com/kkitai/CondoManagerV2/app/internal/config"
	"github.com/kkitai/CondoManagerV2/app/internal/mailer"
)

func TestSMTPMailer_NoSMTPHost_PrintsToStdout(t *testing.T) {
	m := mailer.New(config.SMTPConfig{})
	err := m.SendInvitation(context.Background(), "user@example.com", "Test User", "http://example.com/invite/token")
	if err != nil {
		t.Errorf("expected nil error when SMTP host is empty, got: %v", err)
	}
}

func TestSMTPMailer_New(t *testing.T) {
	cfg := config.SMTPConfig{
		Host: "smtp.example.com",
		Port: "587",
		User: "user",
		From: "noreply@example.com",
	}
	m := mailer.New(cfg)
	if m == nil {
		t.Error("expected non-nil mailer")
	}
}
