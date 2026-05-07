package mailer

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/kkitai/CondoManagerV2/app/internal/config"
)

type SMTPMailer struct {
	cfg config.SMTPConfig
}

func New(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendInvitation(_ context.Context, toEmail, toName, inviteURL string) error {
	if m.cfg.Host == "" {
		fmt.Printf("[mailer] invitation to %s <%s>: %s\n", toName, toEmail, inviteURL)
		return nil
	}

	subject := "CondoManager への招待"
	body := fmt.Sprintf(`%s さん、

CondoManager にご招待しました。
下記のリンクからパスワードを設定してアカウントを有効化してください。

%s

このリンクは72時間有効です。

CondoManager 管理チーム
`, toName, inviteURL)

	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		toEmail, m.cfg.From, subject, body)

	addr := m.cfg.Host + ":" + m.cfg.Port
	auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)

	if err := smtp.SendMail(addr, auth, m.cfg.From, []string{toEmail}, []byte(msg)); err != nil {
		return fmt.Errorf("send invitation email: %w", err)
	}
	return nil
}
