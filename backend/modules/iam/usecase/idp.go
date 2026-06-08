package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

// Cipher encrypts/decrypts the SP private key at rest.
type Cipher interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type identityProviderUsecase struct {
	repo   connectors.IdentityProviderRepository
	cipher Cipher
}

func NewIdentityProviderUsecase(repo connectors.IdentityProviderRepository, cipher Cipher) connectors.IdentityProviderUsecase {
	return &identityProviderUsecase{repo: repo, cipher: cipher}
}

// isSSOEntitled reports whether the current deploy/tenant may configure & use
// SAML SSO.
//
// TODO(billing): SAML SSO is a PAID feature. Once the billing module exists,
// gate this on the tenant's plan entitlement (e.g. plan >= Enterprise),
// resolved from the billing usecase / the request's tenant context. Until then
// it is always allowed so the feature can be developed and demoed. Mirror this
// gate in the public ListActive and in the live flow (usecase/saml.go) so a
// lapsed plan also stops SSO logins, not just new config.
func (u *identityProviderUsecase) isSSOEntitled(_ context.Context) bool {
	return true
}

func validIDP(req dto.IdentityProviderRequest) bool {
	return strings.TrimSpace(req.Name) != "" &&
		strings.TrimSpace(req.ProviderType) != "" &&
		strings.TrimSpace(req.MetadataURL) != "" &&
		strings.TrimSpace(req.SpEntityID) != "" &&
		strings.TrimSpace(req.SpAcsURL) != "" &&
		strings.TrimSpace(req.SpCertificatePem) != ""
}

func (u *identityProviderUsecase) Create(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error) {
	if !u.isSSOEntitled(ctx) {
		return nil, domain.ErrSSONotEntitled
	}
	if req.ID != 0 {
		return nil, domain.ErrIDPIDForbidden
	}
	if !validIDP(req) {
		return nil, domain.ErrIDPInvalidInput
	}
	if strings.TrimSpace(req.SpPrivateKeyPem) == "" {
		return nil, domain.ErrIDPKeyRequired
	}
	encKey, err := u.cipher.Encrypt(req.SpPrivateKeyPem)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	c := &domain.IdentityProviderConfig{
		Name:             req.Name,
		ProviderType:     req.ProviderType,
		MetadataURL:      req.MetadataURL,
		SpEntityID:       req.SpEntityID,
		SpAcsURL:         req.SpAcsURL,
		SpPrivateKeyPem:  encKey,
		SpCertificatePem: req.SpCertificatePem,
		Active:           req.Active,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := u.repo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (u *identityProviderUsecase) Update(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error) {
	if !u.isSSOEntitled(ctx) {
		return nil, domain.ErrSSONotEntitled
	}
	if req.ID == 0 {
		return nil, domain.ErrIDPIDRequired
	}
	if !validIDP(req) {
		return nil, domain.ErrIDPInvalidInput
	}
	existing, err := u.repo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrIDPNotFound
	}
	// Keep the stored (encrypted) key unless a new one is supplied.
	encKey := existing.SpPrivateKeyPem
	if strings.TrimSpace(req.SpPrivateKeyPem) != "" {
		encKey, err = u.cipher.Encrypt(req.SpPrivateKeyPem)
		if err != nil {
			return nil, err
		}
	}
	existing.Name = req.Name
	existing.ProviderType = req.ProviderType
	existing.MetadataURL = req.MetadataURL
	existing.SpEntityID = req.SpEntityID
	existing.SpAcsURL = req.SpAcsURL
	existing.SpPrivateKeyPem = encKey
	existing.SpCertificatePem = req.SpCertificatePem
	existing.Active = req.Active
	existing.UpdatedAt = time.Now().UTC()
	if err := u.repo.Save(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (u *identityProviderUsecase) GetByID(ctx context.Context, id uint64) (*domain.IdentityProviderConfig, error) {
	c, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, domain.ErrIDPNotFound
	}
	return c, nil
}

func (u *identityProviderUsecase) List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *identityProviderUsecase) ListActive(ctx context.Context) ([]dto.IdentityProviderPublic, error) {
	items, err := u.repo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.IdentityProviderPublic, 0, len(items))
	for _, c := range items {
		out = append(out, dto.IdentityProviderPublic{
			ID:           c.ID,
			Name:         c.Name,
			ProviderType: c.ProviderType,
			LoginURL:     "/api/v1/sso/saml/" + c.Name + "/login",
		})
	}
	return out, nil
}

func (u *identityProviderUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
