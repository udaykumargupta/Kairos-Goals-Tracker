package kairos

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

// SendPasswordReset emails the reset link. Content matches the Spring EmailServiceImpl
// so the message users receive is unchanged.
//
// Like the Java version this never fails the request: when SMTP isn't configured we
// log the link (handy in development), and send errors are logged and swallowed so
// the endpoint still answers with its neutral "if that email is registered" message.
func (a *App) SendPasswordReset(toEmail, resetLink string) {
	if a.Cfg.MailHost == "" || a.Cfg.MailUsername == "" {
		log.Printf("[DEV] Password reset for %s (email not configured) -> %s", toEmail, resetLink)
		return
	}

	from := a.Cfg.MailFrom
	if from == "" {
		from = a.Cfg.MailUsername
	}
	body := strings.Join([]string{
		"Hi,",
		"",
		"We received a request to set a new password for your Kairos account.",
		"Click the link below to choose a new password (it expires in 1 hour):",
		"",
		resetLink,
		"",
		"If you didn't request this, you can safely ignore this email.",
		"",
		"— Kairos",
	}, "\r\n")

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", a.Cfg.MailFromName, from),
		fmt.Sprintf("To: %s", toEmail),
		"Subject: Reset your Kairos password",
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"MIME-Version: 1.0",
		`Content-Type: text/plain; charset="UTF-8"`,
		"Content-Transfer-Encoding: 8bit",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", a.Cfg.MailHost, a.Cfg.MailPort)
	auth := smtp.PlainAuth("", a.Cfg.MailUsername, a.Cfg.MailPassword, a.Cfg.MailHost)
	// SendMail upgrades to TLS via STARTTLS when the server advertises it (Gmail on 587).
	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(msg)); err != nil {
		log.Printf("failed to send password-reset email to %s: %v", toEmail, err)
	}
}
