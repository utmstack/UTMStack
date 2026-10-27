package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

type RetentionRequest struct {
	Dataset  string `json:"dataset" binding:"required"`
	KeepDays int    `json:"keepDays"`
	ColdDays int    `json:"coldDays"`
}

func (r RetentionRequest) ToDomain() domain.Retention {
	return domain.Retention{
		Dataset:  domain.Dataset(r.Dataset),
		KeepDays: r.KeepDays,
		ColdDays: r.ColdDays,
	}
}

type RetentionResponse struct {
	Dataset  string `json:"dataset"`
	KeepDays int    `json:"keepDays"`
	ColdDays int    `json:"coldDays"`
	Tiered   bool   `json:"tiered"`
}

func FromRetention(r domain.Retention) RetentionResponse {
	return RetentionResponse{
		Dataset:  string(r.Dataset),
		KeepDays: r.KeepDays,
		ColdDays: r.ColdDays,
		Tiered:   r.Tiered(),
	}
}

func FromRetentions(rs []domain.Retention) []RetentionResponse {
	out := make([]RetentionResponse, 0, len(rs))
	for _, r := range rs {
		out = append(out, FromRetention(r))
	}
	return out
}

type UsageResponse struct {
	Dataset   string     `json:"dataset"`
	Documents int64      `json:"documents"`
	Bytes     int64      `json:"bytes"`
	Oldest    *time.Time `json:"oldest,omitempty"`
	Newest    *time.Time `json:"newest,omitempty"`
}

func FromUsage(us []domain.Usage) []UsageResponse {
	out := make([]UsageResponse, 0, len(us))
	for _, u := range us {
		r := UsageResponse{
			Dataset:   string(u.Dataset),
			Documents: u.Documents,
			Bytes:     u.Bytes,
		}
		// An empty dataset has no span, and 1970 is not an answer.
		if u.Documents > 0 {
			oldest, newest := u.Oldest, u.Newest
			r.Oldest, r.Newest = &oldest, &newest
		}
		out = append(out, r)
	}
	return out
}

type HealthResponse struct {
	Status      string  `json:"status"`
	DiskUsedPct float64 `json:"diskUsedPct"`
	Message     string  `json:"message,omitempty"`
}

func FromHealth(h domain.Health) HealthResponse {
	return HealthResponse{Status: h.Status, DiskUsedPct: h.DiskUsedPct, Message: h.Message}
}

// ColdTargetRequest carries the secret in and never comes back out. Kind
// defaults to s3; a local target takes no other field, and an s3 target with useRole takes no keys. For azure, accessKey is the account name and secretKey its key.
type ColdTargetRequest struct {
	Kind       string `json:"kind,omitempty"`
	Endpoint   string `json:"endpoint"`
	Container  string `json:"container,omitempty"`
	UseRole    bool   `json:"useRole,omitempty"`
	AccessKey  string `json:"accessKey"`
	SecretKey  string `json:"secretKey"`
	CacheBytes int64  `json:"cacheBytes,omitempty"`
}

func (o ColdTargetRequest) ToDomain() domain.ColdTarget {
	return domain.ColdTarget{
		Kind:       domain.ColdKind(o.Kind),
		Container:  o.Container,
		UseRole:    o.UseRole,
		Endpoint:   o.Endpoint,
		AccessKey:  o.AccessKey,
		SecretKey:  o.SecretKey,
		CacheBytes: o.CacheBytes,
	}
}

type TieringResponse struct {
	Configured bool   `json:"configured"`
	Ready      bool   `json:"ready"`
	Kind       string `json:"kind,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	Policy     string `json:"policy,omitempty"`
}

func FromTiering(t domain.Tiering) TieringResponse {
	return TieringResponse{
		Configured: t.Configured,
		Ready:      t.Ready,
		Kind:       string(t.Kind),
		Endpoint:   t.Endpoint,
		Policy:     t.Policy,
	}
}
