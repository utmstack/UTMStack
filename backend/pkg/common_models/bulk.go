package common_models

// BulkTenantSelector picks which tenants a platform-admin bulk op targets.
// If AllTenants is true, TenantIDs is ignored and the handler enumerates
// every ACTIVE tenant via the tenant usecase.
type BulkTenantSelector struct {
	TenantIDs  []string `json:"tenantIds"`
	AllTenants bool     `json:"allTenants"`
}

// BulkFailure records one per-tenant error so callers can retry just the
// tenants that failed instead of the whole batch.
type BulkFailure struct {
	TenantID string `json:"tenantId"`
	Error    string `json:"error"`
}

// BulkResult is what every platform-admin bulk endpoint returns. Partial
// success is expected — system-owned guards and per-tenant validation errors
// land in Failed while the rest of the loop continues.
type BulkResult struct {
	Succeeded []string      `json:"succeeded"`
	Failed    []BulkFailure `json:"failed"`
}

// Append records an outcome for one tenant. Nil err means success.
func (r *BulkResult) Append(tenantID string, err error) {
	if err != nil {
		r.Failed = append(r.Failed, BulkFailure{TenantID: tenantID, Error: err.Error()})
		return
	}
	r.Succeeded = append(r.Succeeded, tenantID)
}
