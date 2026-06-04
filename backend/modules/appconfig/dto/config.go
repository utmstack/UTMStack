package dto

import "time"

type ConfigResponse struct {
	Key         string    `json:"key"`
	Value       string    `json:"value,omitempty"`
	IsSecret    bool      `json:"is_secret"`
	IsSet       bool      `json:"is_set"`
	Label       string    `json:"label,omitempty"`
	Description string    `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by,omitempty"`
}

type UpsertRequest struct {
	Value       string `json:"value"`
	IsSecret    *bool  `json:"is_secret,omitempty"`
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
}

type CheckMailRequest struct {
	Host     string   `json:"host" binding:"required"`
	Port     string   `json:"port" binding:"required"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	Orgname  string   `json:"orgname"`
	SmtpAuth string   `json:"smtp_auth"`
	BaseUrl  string   `json:"base_url"`
	To       []string `json:"to" binding:"required,min=1"`
}

type CheckMailResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
