package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
)

// ErrWhiteLabelNotEntitled is returned when a deploy/tenant tries to enable
// white-labeling without the required (paid) plan. Not reachable yet — see the
// billing TODO in isWhiteLabelEntitled.
var ErrWhiteLabelNotEntitled = errors.New("white-labeling requires a paid plan")

const (
	defaultProductName = "UTMStack"
	// brandingConfigKey is the single utm_configuration_parameter row that stores
	// the whole branding object as JSON. Reuses the existing config table instead
	// of a dedicated table — the row is seeded in 000001_init.up.sql.
	brandingConfigKey = "branding"
)

type brandingService struct {
	repo connectors.Repository
}

func NewBranding(repo connectors.Repository) *brandingService {
	return &brandingService{repo: repo}
}

// isWhiteLabelEntitled reports whether the current deploy/tenant may use
// white-labeling.
//
// TODO(billing): white-labeling is a PAID feature. Once the billing module
// exists, gate this on the tenant's plan entitlement (e.g. plan >= Enterprise),
// resolved from the billing usecase / the request's tenant context. Until then
// it is always allowed so the feature can be developed and demoed.
func (s *brandingService) isWhiteLabelEntitled(_ context.Context) bool {
	return true
}

// read loads the stored branding (JSON in the `branding` config row), falling
// back to defaults when unset or unparseable.
func (s *brandingService) read(ctx context.Context) (dto.BrandingResponse, error) {
	resp := dto.BrandingResponse{ProductName: defaultProductName}
	row, err := s.repo.GetByKey(ctx, brandingConfigKey)
	if err != nil {
		return resp, err
	}
	if row == nil || strings.TrimSpace(row.ConfParamValue) == "" {
		return resp, nil
	}
	var stored dto.BrandingResponse
	if err := json.Unmarshal([]byte(row.ConfParamValue), &stored); err != nil {
		return resp, nil // corrupt value → defaults
	}
	if strings.TrimSpace(stored.ProductName) == "" {
		stored.ProductName = defaultProductName
	}
	return stored, nil
}

// Get returns the full branding (admin view), falling back to defaults.
func (s *brandingService) Get(ctx context.Context) (*dto.BrandingResponse, error) {
	resp, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update persists the branding overrides as JSON on the `branding` config row.
// Enabling white-labeling requires entitlement (TODO: billing); editing while
// disabled is always allowed so the admin can prepare the brand before buying.
func (s *brandingService) Update(ctx context.Context, actor string, req dto.BrandingRequest) (*dto.BrandingResponse, error) {
	if req.Enabled && !s.isWhiteLabelEntitled(ctx) {
		return nil, ErrWhiteLabelNotEntitled
	}
	now := time.Now()
	resp := dto.BrandingResponse{
		Enabled:            req.Enabled,
		ProductName:        req.ProductName,
		LogoURL:            req.LogoURL,
		LogoDarkURL:        req.LogoDarkURL,
		FaviconURL:         req.FaviconURL,
		PrimaryColor:       req.PrimaryColor,
		AccentColor:        req.AccentColor,
		LoginBackgroundURL: req.LoginBackgroundURL,
		FooterText:         req.FooterText,
		SupportURL:         req.SupportURL,
		HidePoweredBy:      req.HidePoweredBy,
		ReportLogoURL:      req.ReportLogoURL,
		ReportCoverURL:     req.ReportCoverURL,
		UpdatedAt:          &now,
		UpdatedBy:          actor,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	row, err := s.repo.GetByKey(ctx, brandingConfigKey)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("branding config row %q is not seeded", brandingConfigKey)
	}
	row.ConfParamValue = string(data)
	row.ModificationTime = &now
	row.ModificationUser = actor
	if err := s.repo.Update(ctx, row); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPublic returns the effective branding for the (unauthenticated) login page.
// White-labeling only takes effect when it is both enabled AND entitled;
// otherwise the default UTMStack brand is returned.
func (s *brandingService) GetPublic(ctx context.Context) (*dto.BrandingPublic, error) {
	b, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	// TODO(billing): a lapsed plan must stop white-labeling from taking effect
	// even if the stored row says Enabled. isWhiteLabelEntitled covers that here.
	if !b.Enabled || !s.isWhiteLabelEntitled(ctx) {
		return &dto.BrandingPublic{ProductName: defaultProductName}, nil
	}
	return &dto.BrandingPublic{
		Enabled:            b.Enabled,
		ProductName:        b.ProductName,
		LogoURL:            b.LogoURL,
		LogoDarkURL:        b.LogoDarkURL,
		FaviconURL:         b.FaviconURL,
		PrimaryColor:       b.PrimaryColor,
		AccentColor:        b.AccentColor,
		LoginBackgroundURL: b.LoginBackgroundURL,
		FooterText:         b.FooterText,
		SupportURL:         b.SupportURL,
		HidePoweredBy:      b.HidePoweredBy,
	}, nil
}

// ProductName returns the effective product name for branded output: the
// configured brand when white-labeling is enabled & entitled (and a name is
// set), otherwise the default "UTMStack". Used by emails, the TOTP issuer and
// report subjects so a paying customer sees their company name instead of ours.
func (s *brandingService) ProductName(ctx context.Context) string {
	b, err := s.read(ctx)
	if err != nil || !b.Enabled || !s.isWhiteLabelEntitled(ctx) {
		return defaultProductName
	}
	if name := strings.TrimSpace(b.ProductName); name != "" {
		return name
	}
	return defaultProductName
}
