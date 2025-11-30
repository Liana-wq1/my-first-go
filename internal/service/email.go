package service

import (
	"fmt"
	"net/smtp"
)

type EmailSender struct {
	From     string // email, с которого отправляем
	Password string // пароль
	Host     string // SMTP-хост
	Port     string // порт (587)
}

// Send отправляет простое текстовое письмо
func (e *EmailSender) Send(to, subject, body string) error {
	auth := smtp.PlainAuth("", e.From, e.Password, e.Host)

	msg := "From: " + e.From + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	addr := e.Host + ":" + e.Port

	return smtp.SendMail(addr, auth, e.From, []string{to}, []byte(msg))
}

// helper: создаём sender из настроек
func NewEmailSender(from, password, host, port string) *EmailSender {
	return &EmailSender{
		From:     from,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

// Сборка письма с ссылкой
func BuildConfirmEmailBody(link string) string {
	return fmt.Sprintf(
		"Здравствуйте!\n\nПерейдите по ссылке, чтобы подтвердить вашу регистрацию:\n%s\n\nЕсли вы не регистрировались — просто игнорируйте это письмо.",
		link,
	)
}
