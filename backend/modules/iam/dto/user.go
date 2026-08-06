package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
)

type PageInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type CreateUserRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	Name      string   `json:"name,omitempty"`
	LangKey   string   `json:"lang_key,omitempty"`
	RoleNames []string `json:"role_names,omitempty"`
}

type UpdateUserRequest struct {
	Email   string             `json:"email,omitempty" binding:"omitempty,email"`
	Name    string             `json:"name,omitempty"`
	LangKey string             `json:"lang_key,omitempty"`
	Status  *domain.UserStatus `json:"status,omitempty"`
}

type AssignRolesRequest struct {
	RoleNames []string `json:"role_names"`
}

type UserDetailResponse struct {
	UserResponse
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Roles     []RoleDigest `json:"roles,omitempty"`
}

type RoleDigest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type UserListItem struct {
	UserResponse
	Roles []RoleDigest `json:"roles,omitempty"`
}

type UserListResponse struct {
	Data     []UserListItem `json:"data"`
	PageInfo PageInfo       `json:"page_info"`
}

type ListUsersQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Search   string `form:"search"`
}
