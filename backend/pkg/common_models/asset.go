package common_models

// AssetSensitivity is one asset's CIA rating, projected to the event processor
// so it can weigh an alert against what the target is worth.
//
// TenantID is what keeps the projection honest: the event processor matches an
// alert's tenant before it looks at any asset, so an asset that arrives without
// one is either never applied or applied to somebody else's alerts.
type AssetSensitivity struct {
	TenantID        string
	Name            string
	Hostnames       []string
	Ips             []string
	Confidentiality int
	Integrity       int
	Availability    int
}
