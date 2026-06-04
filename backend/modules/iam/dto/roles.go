package dto

type PermissionResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
}

type RoleResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description,omitempty"`
}

type RoleDetailResponse struct {
	RoleResponse
	Permissions []PermissionResponse `json:"permissions"`
}
