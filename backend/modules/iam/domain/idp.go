package domain

import (
	"errors"
	"time"
)

var (
	ErrIDPNotFound     = errors.New("identity provider not found")
	ErrIDPIDForbidden  = errors.New("id must be absent on create")
	ErrIDPIDRequired   = errors.New("id is required for update")
	ErrIDPInvalidInput = errors.New("name, providerType, metadataUrl, spEntityId, spAcsUrl and certificate are required")
	ErrIDPKeyRequired  = errors.New("sp private key is required on create")
)

type IdentityProviderConfig struct {
	ID               uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"column:name;size:255;not null;uniqueIndex:idx_idp_provider_name" json:"name"`
	ProviderType     string    `gorm:"column:provider_type;size:50;not null;uniqueIndex:idx_idp_provider_name;index" json:"providerType"`
	MetadataURL      string    `gorm:"column:metadata_url;size:512;not null" json:"metadataUrl"`
	SpEntityID       string    `gorm:"column:sp_entity_id;size:512;not null" json:"spEntityId"`
	SpAcsURL         string    `gorm:"column:sp_acs_url;size:512;not null" json:"spAcsUrl"`
	SpPrivateKeyPem  string    `gorm:"column:sp_private_key_pem" json:"-"`
	SpCertificatePem string    `gorm:"column:sp_certificate_pem" json:"spCertificatePem"`
	Active           bool      `gorm:"column:active;not null;default:true" json:"active"`
	CreatedAt        time.Time `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt        time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`
}

func (IdentityProviderConfig) TableName() string { return "utm_identity_provider_config" }
