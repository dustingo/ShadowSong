package notifier

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

var globalDB *gorm.DB

// SetDB sets the global DB reference used by EmailSender to load SmtpConfig.
// Must be called once at application startup after DB initialization.
func SetDB(db *gorm.DB) {
	globalDB = db
}

type EmailConfig struct {
	FromName string `json:"from_name"`
}

type EmailSender struct {
	config EmailConfig
	db     *gorm.DB
}

func NewEmailSender(config []byte) (Sender, error) {
	var ec EmailConfig
	if err := json.Unmarshal(config, &ec); err != nil {
		return nil, err
	}
	return &EmailSender{config: ec, db: globalDB}, nil
}

func (s *EmailSender) Send(title, content string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("SMTP 服务未配置: 数据库未初始化")
	}

	var smtpCfg models.SmtpConfig
	if err := s.db.Where("id = 1").First(&smtpCfg).Error; err != nil {
		return fmt.Errorf("SMTP 服务未配置: %w", err)
	}
	if !smtpCfg.Enabled {
		return fmt.Errorf("SMTP 服务未启用")
	}

	recipientsRaw, ok := data["recipients"]
	if !ok || recipientsRaw == nil {
		return fmt.Errorf("邮件收件人为空")
	}

	var recipients []string
	switch v := recipientsRaw.(type) {
	case []interface{}:
		for _, item := range v {
			if addr, ok := item.(string); ok && addr != "" {
				recipients = append(recipients, addr)
			}
		}
	case []string:
		recipients = v
	case string:
		if v != "" {
			recipients = []string{v}
		}
	}
	if len(recipients) == 0 {
		return fmt.Errorf("邮件收件人为空")
	}

	fromName := s.config.FromName
	if fromName == "" {
		fromName = smtpCfg.FromName
	}
	if fromName == "" {
		fromName = "告警系统"
	}

	from := smtpCfg.FromAddr
	subject := encodeRFC2047(title)
	message := buildHTMLMessage(from, fromName, recipients, subject, content)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)
	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	var client *smtp.Client
	var err error

	if smtpCfg.TLS {
		tlsConfig := &tls.Config{
			ServerName: smtpCfg.Host,
		}
		conn, dialErr := tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp",
			addr,
			tlsConfig,
		)
		if dialErr != nil {
			return fmt.Errorf("failed to connect SMTP server: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, smtpCfg.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP server: %w", err)
		}
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed for %s: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("SMTP QUIT failed: %w", err)
	}

	return nil
}

func encodeRFC2047(s string) string {
	for _, r := range s {
		if r > 127 {
			encoded := base64.StdEncoding.EncodeToString([]byte(s))
			return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
		}
	}
	return s
}

func buildHTMLMessage(from, fromName string, to []string, subject, htmlBody string) string {
	var buf strings.Builder

	fromAddr := mail.Address{Name: fromName, Address: from}
	buf.WriteString(fmt.Sprintf("From: %s\r\n", fromAddr.String()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)

	return buf.String()
}