package mail

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type MailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachments []string,
	) error
}

const (
	SMTPHost        = "smtp.gmail.com"
	SMTPServerAddr  = "smtp.gmail.com:587"
)

type GmailSender struct {
	name        string
	senderEmail string
	password    string
}

func NewGmailSender(name, senderEmail, password string) MailSender {
	return &GmailSender{
		name:        name,
		senderEmail: senderEmail,
		password:    password,
	}

}

func (sender *GmailSender) SendEmail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachments []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", sender.name, sender.senderEmail)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc

	for _, attachment := range attachments {
		if _, err := e.AttachFile(attachment); err != nil {
			return err
		}
	}

	auth := smtp.PlainAuth("", sender.senderEmail, sender.password, SMTPHost)
	return e.Send(SMTPServerAddr, auth)
}
