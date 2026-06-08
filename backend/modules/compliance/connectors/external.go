package connectors

import "context"

type MailSender interface {
	SendComplianceReport(ctx context.Context, toEmail, subject string, pdfData []byte) error
}
