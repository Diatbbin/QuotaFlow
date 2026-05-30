package mail

import "fmt"

func TransferNotificationSubject() string {
	return "QuotaFlow: You received a token transfer"
}

func TransferNotificationContent(senderEmail string, tokens int64) string {
	return fmt.Sprintf(`
		<h1>Token transfer received</h1>
		<p>You received <strong>%d</strong> tokens from <strong>%s</strong>.</p>
		<p>— QuotaFlow</p>
	`, tokens, senderEmail)
}
