package dto

import "time"

type PageInfo struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// CreateUserRequest carries the fields an admin provides when inviting a user.
// No password: the account is created deactivated and the user sets their own
// password through the invitation email link (matches the legacy Java flow).
type CreateUserRequest struct {
	Login     string   `json:"login" binding:"required,min=3,max=50"`
	Email     string   `json:"email" binding:"required,email"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	LangKey   string   `json:"lang_key,omitempty"`
	RoleNames []string `json:"role_names,omitempty"`
}

type UpdateUserRequest struct {
	Email     string `json:"email,omitempty" binding:"omitempty,email"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	LangKey   string `json:"lang_key,omitempty"`
	Activated *bool  `json:"activated,omitempty"`
}

type AssignRolesRequest struct {
	RoleNames []string `json:"role_names"`
}

type UserDetailResponse struct {
	UserResponse
	DefaultPassword  bool         `json:"default_password"`
	CreatedBy        string       `json:"created_by,omitempty"`
	CreatedDate      *time.Time   `json:"created_date,omitempty"`
	LastModifiedBy   string       `json:"last_modified_by,omitempty"`
	LastModifiedDate *time.Time   `json:"last_modified_date,omitempty"`
	Roles            []RoleDigest `json:"roles,omitempty"`
}

type RoleDigest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type UserListResponse struct {
	Data     []UserResponse `json:"data"`
	PageInfo PageInfo       `json:"page_info"`
}

type ListUsersQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Search   string `form:"search"`
}
