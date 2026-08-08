package connectors

import "context"

type MailSender interface {
	SendComplianceReport(ctx context.Context, to, cc []string, subject string, pdfData []byte) error
}
