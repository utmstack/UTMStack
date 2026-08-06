package dto

import "github.com/google/uuid"

type PermissionResponse struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	System      bool      `json:"system"`
}

type RoleDetailResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}

type RoleUpsertRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=50"`
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}
