package dto

// TenantUsageDTO is one tenant's successfully-ingested volume for one UTC
// day, as reported to an internal caller (the installer's usage-report job,
// forwarding this to Customer Manager).
type TenantUsageDTO struct {
	TenantID   string `json:"tenant_id"`
	EventCount int64  `json:"event_count"`
	Bytes      int64  `json:"bytes"`
}
