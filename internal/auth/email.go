package auth

import (
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPEmailSender struct {
	config SMTPConfig
}

func NewSMTPEmailSender(config SMTPConfig) *SMTPEmailSender {
	return &SMTPEmailSender{config: config}
}

func (s *SMTPEmailSender) SendVerificationEmail(email, code string) error {
	subject := "이메일 인증 코드"
	body := fmt.Sprintf(`
안녕하세요,

LottoSmash 회원가입을 위한 인증 코드입니다.

인증 코드: %s

이 코드는 10분간 유효합니다.

감사합니다.
LottoSmash 팀
`, code)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailSender) SendPasswordResetEmail(email, code string) error {
	subject := "비밀번호 재설정 코드"
	body := fmt.Sprintf(`
안녕하세요,

비밀번호 재설정을 위한 인증 코드입니다.

인증 코드: %s

이 코드는 10분간 유효합니다.

본인이 요청하지 않았다면 이 이메일을 무시해주세요.

감사합니다.
LottoSmash 팀
`, code)

	return s.sendEmail(email, subject, body)
}

func (s *SMTPEmailSender) sendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		to, subject, body))

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{to}, msg)
}

// NoopEmailSender 개발/테스트용 이메일 발송기 (실제로 발송하지 않음)
type NoopEmailSender struct{}

func NewNoopEmailSender() *NoopEmailSender {
	return &NoopEmailSender{}
}

func (n *NoopEmailSender) SendVerificationEmail(email, code string) error {
	fmt.Printf("[DEV] Verification email to %s: code=%s\n", email, code)
	return nil
}

func (n *NoopEmailSender) SendPasswordResetEmail(email, code string) error {
	fmt.Printf("[DEV] Password reset email to %s: code=%s\n", email, code)
	return nil
}
