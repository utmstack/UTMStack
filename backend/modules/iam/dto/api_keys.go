package dto

import "time"

type APIKeyUpsertRequest struct {
	Name      string     `json:"name" binding:"required"`
	AllowedIP []string   `json:"allowed_ip"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type APIKeyAuthRequest struct {
	APIKey   string `json:"apiKey" binding:"required"`
	ClientIP string `json:"clientIp"`
}

type APIKeyResponse struct {
	ID          uint64     `json:"id"`
	Name        string     `json:"name"`
	AllowedIP   []string   `json:"allowed_ip"`
	CreatedAt   time.Time  `json:"created_at"`
	GeneratedAt time.Time  `json:"generated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

type ListAPIKeysQuery struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

type APIKeyListResponse struct {
	Data     []APIKeyResponse `json:"data"`
	PageInfo PageInfo         `json:"page_info"`
}
