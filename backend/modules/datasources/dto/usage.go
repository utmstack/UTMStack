package dto

type UsageResponse struct {
	Count     int64 `json:"count"`
	Limit     int64 `json:"limit"`     // effective cap; 0 when unlimited
	Unlimited bool  `json:"unlimited"` // true for an unlimited enterprise license
	OverLimit bool  `json:"overLimit"` // count exceeds the cap (data beyond it is at risk)
}
