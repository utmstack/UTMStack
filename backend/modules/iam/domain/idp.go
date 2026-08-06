package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ProviderType string

const (
	ProviderSAML ProviderType = "saml"
	ProviderOIDC ProviderType = "oidc"
	ProviderLDAP ProviderType = "ldap"
)

func (p ProviderType) Valid() bool {
	switch p {
	case ProviderSAML, ProviderOIDC, ProviderLDAP:
		return true
	}
	return false
}

func (p ProviderType) Redirecting() bool {
	return p == ProviderSAML || p == ProviderOIDC
}

type IdentityProviderConfig struct {
	ID               uuid.UUID      `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID         uuid.UUID      `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:ux_idp_tenant_name,priority:1" json:"-"`
	Name             string         `gorm:"column:name;size:64;not null;uniqueIndex:ux_idp_tenant_name,priority:2" json:"name"`
	ProviderType     ProviderType   `gorm:"column:provider_type;size:16;not null;index" json:"providerType"`
	Active           bool           `gorm:"column:active;not null;default:true" json:"active"`
	Settings         datatypes.JSON `gorm:"column:settings;type:jsonb" json:"settings"`
	JITProvisioning  bool           `gorm:"column:jit_provisioning;not null;default:false" json:"jitProvisioning"`
	DefaultRoleID    *uuid.UUID     `gorm:"column:default_role_id;type:uuid" json:"defaultRoleId,omitempty"`
	GroupsAttribute  string         `gorm:"column:groups_attribute;size:128" json:"groupsAttribute,omitempty"`
	SyncRolesOnLogin bool           `gorm:"column:sync_roles_on_login;not null;default:false" json:"syncRolesOnLogin"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null" json:"updatedAt"`
}

func (IdentityProviderConfig) TableName() string { return "identity_provider" }

type IdentityProviderGroupMapping struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	IdentityProviderID uuid.UUID `gorm:"column:identity_provider_id;type:uuid;not null;index;uniqueIndex:ux_idp_group,priority:1" json:"identityProviderId"`
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index" json:"-"`
	Group              string    `gorm:"column:group_name;size:255;not null;uniqueIndex:ux_idp_group,priority:2" json:"group"`
	RoleID             uuid.UUID `gorm:"column:role_id;type:uuid;not null" json:"roleId"`
}

func (IdentityProviderGroupMapping) TableName() string { return "identity_provider_group" }

type SAMLSettings struct {
	MetadataURL      string `json:"metadataUrl"`
	SpEntityID       string `json:"spEntityId"`
	SpACSURL         string `json:"spAcsUrl"`
	SpPrivateKeyPem  string `json:"spPrivateKeyPem,omitempty"`
	SpCertificatePem string `json:"spCertificatePem"`
}

type OIDCSettings struct {
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret,omitempty"`
	RedirectURL  string   `json:"redirectUrl"`
	Scopes       []string `json:"scopes,omitempty"`
}

type LDAPSettings struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	StartTLS       bool   `json:"startTls"`
	InsecureTLS    bool   `json:"insecureTls,omitempty"`
	BindDN         string `json:"bindDn"`
	BindPassword   string `json:"bindPassword,omitempty"`
	BaseDN         string `json:"baseDn"`
	UserFilter     string `json:"userFilter"`
	EmailAttribute string `json:"emailAttribute"`
	NameAttribute  string `json:"nameAttribute"`
	GroupAttribute string `json:"groupAttribute"`
}

type FederatedIdentity struct {
	Email  string
	Name   string
	Groups []string
}
