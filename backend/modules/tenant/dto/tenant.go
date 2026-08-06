package dto

import (
	"bytes"
	"encoding/json"

	"github.com/utmstack/utmstack/backend/modules/tenant/domain"
)

type CreateRequest struct {
	Name       string `json:"name"   binding:"required"`
	Domain     string `json:"domain" binding:"required"`
	AdminEmail string `json:"adminEmail" binding:"required,email"`
}

type UpdateRequest struct {
	Name   string              `json:"name,omitempty"`
	Domain string              `json:"domain,omitempty"`
	Status domain.TenantStatus `json:"status,omitempty"`
	// Raw, because the cap is a tri-state and a *int only carries two: a JSON
	// null and an absent field both decode to nil. Read it with AILimit.
	MaxAIRequests json.RawMessage `json:"maxAIRequests,omitempty"`
}

// AILimit reads the cap the caller sent: absent leaves the stored one alone,
// null removes it, and a number sets it. Without the middle case a cap could be
// handed out but never taken back.
func (r UpdateRequest) AILimit() (limit *int, present bool, err error) {
	raw := bytes.TrimSpace(r.MaxAIRequests)
	if len(raw) == 0 {
		return nil, false, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return nil, true, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, false, domain.ErrLimitInvalid
	}
	return &n, true, nil
}

type Filter struct {
	Name   string              `form:"name"`
	Domain string              `form:"domain"`
	Status domain.TenantStatus `form:"status"`
	Page   int                 `form:"page"`
	Size   int                 `form:"size"`
}

type SupportAccessRequest struct {
	SupportAccess domain.SupportAccess `json:"supportAccess" binding:"required"`
}

// SupportAccessResponse is what a tenant may read about itself. Deliberately
// only the grant: the rest of the row is the platform's business, not theirs.
type SupportAccessResponse struct {
	SupportAccess domain.SupportAccess `json:"supportAccess"`
}
