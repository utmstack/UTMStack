package dto

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

// GroupMapping is one line of "this directory group grants this role here".
type GroupMapping struct {
	Group  string    `json:"group"`
	RoleID uuid.UUID `json:"roleId"`
}

type IdentityProviderRequest struct {
	ID               uuid.UUID       `json:"id"`
	Name             string          `json:"name"`
	ProviderType     string          `json:"providerType"`
	Active           bool            `json:"active"`
	Settings         json.RawMessage `json:"settings"`
	JITProvisioning  bool            `json:"jitProvisioning"`
	DefaultRoleID    *uuid.UUID      `json:"defaultRoleId,omitempty"`
	GroupsAttribute  string          `json:"groupsAttribute,omitempty"`
	SyncRolesOnLogin bool            `json:"syncRolesOnLogin"`
	GroupMappings    []GroupMapping  `json:"groupMappings,omitempty"`
}

type IdentityProviderFilter struct {
	Name         string `form:"name"`
	ProviderType string `form:"providerType"`
	Active       *bool  `form:"active"`
	database.Params
}

type IdentityProviderPublic struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	ProviderType string    `json:"providerType"`
	LoginURL     string    `json:"loginUrl"`
}
