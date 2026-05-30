package mail

import (
	"testing"

	"github.com/diatbbin/QuotaFlow/util"
	"github.com/stretchr/testify/require"
)

func TestGmailSender_SendEmail(t *testing.T) {
	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddr, config.EmailPassword)

	subject := "Test Email from QuotaFlow"
	content := `
	<h1>Hello From QuotaFlow</h1>
	<p>This is a test email</p>
	`
	to := []string{"diatbbin1@gmail.com"}
	attachments := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachments)
	require.NoError(t, err)
}