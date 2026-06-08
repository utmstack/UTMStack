package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

// IdentityProviderRequest is the create/update body for a SAML IdP config.
// SpPrivateKeyPem is plaintext on input (encrypted by the usecase); on update it
// may be empty to keep the stored key.
type IdentityProviderRequest struct {
	ID               uint64 `json:"id"`
	Name             string `json:"name"`
	ProviderType     string `json:"providerType"`
	MetadataURL      string `json:"metadataUrl"`
	SpEntityID       string `json:"spEntityId"`
	SpAcsURL         string `json:"spAcsUrl"`
	SpPrivateKeyPem  string `json:"spPrivateKeyPem"`
	SpCertificatePem string `json:"spCertificatePem"`
	Active           bool   `json:"active"`
}

// IdentityProviderFilter is the list filter for IdP configs.
type IdentityProviderFilter struct {
	Name         string `form:"name"`
	ProviderType string `form:"providerType"`
	Active       *bool  `form:"active"`
	database.Params
}

// IdentityProviderPublic is the unauthenticated login-page view of an active
// IdP: just enough to render a "Login with …" button.
type IdentityProviderPublic struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"providerType"`
	LoginURL     string `json:"loginUrl"`
}
