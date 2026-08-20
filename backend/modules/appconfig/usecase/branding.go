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

var ErrWhiteLabelNotEntitled = errors.New("white-labeling requires an enterprise license")
var ErrUnknownAssetSlot = errors.New("unknown branding asset slot")

const (
	defaultProductName = "UTMStack"
	brandingConfigKey  = "branding"
	systemLanguageKey  = "utmstack.system.language"
	defaultLanguage    = "en"
)

const (
	AssetLogo        = "logo"
	AssetLogoDark    = "logoDark"
	AssetFavicon     = "favicon"
	AssetReportLogo  = "reportLogo"
	AssetReportCover = "reportCover"
)

type brandingService struct {
	repo     connectors.Repository
	entitled func() bool
}

func NewBranding(repo connectors.Repository) *brandingService {
	return &brandingService{repo: repo}
}

func (s *brandingService) SetEntitlement(fn func() bool) {
	s.entitled = fn
}

func (s *brandingService) isWhiteLabelEntitled(_ context.Context) bool {
	if s.entitled == nil {
		return true
	}
	return s.entitled()
}

// read loads the stored branding (JSON in the `branding` config row), falling
// back to defaults when unset or unparseable. Uses GetOwn so a tenant without
// its own row gets defaults instead of inheriting the master tenant's brand.
func (s *brandingService) read(ctx context.Context) (dto.BrandingResponse, error) {
	resp := dto.BrandingResponse{ProductName: defaultProductName}
	row, err := s.repo.GetOwn(ctx, brandingConfigKey)
	if err != nil {
		return resp, err
	}
	if row == nil || strings.TrimSpace(row.Value) == "" {
		return resp, nil
	}
	var stored dto.BrandingResponse
	if err := json.Unmarshal([]byte(row.Value), &stored); err != nil {
		return resp, nil // corrupt value → defaults
	}
	if strings.TrimSpace(stored.ProductName) == "" {
		stored.ProductName = defaultProductName
	}
	return stored, nil
}

func (s *brandingService) save(ctx context.Context, actor string, resp *dto.BrandingResponse) error {
	now := time.Now()
	resp.UpdatedAt = &now
	resp.UpdatedBy = actor
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	row, err := s.repo.GetByKey(ctx, brandingConfigKey)
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("branding config row %q is not seeded", brandingConfigKey)
	}
	row.Value = string(data)
	row.UpdatedAt = &now
	row.UpdatedBy = actor
	return s.repo.Save(ctx, row)
}

// Get returns the full branding (admin view), falling back to defaults.
func (s *brandingService) Get(ctx context.Context) (*dto.BrandingResponse, error) {
	resp, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func fromRequest(req dto.BrandingRequest) dto.BrandingResponse {
	return dto.BrandingResponse{
		Enabled:        req.Enabled,
		ProductName:    req.ProductName,
		AccentColor:    req.AccentColor,
		LogoURL:        req.LogoURL,
		LogoDarkURL:    req.LogoDarkURL,
		FaviconURL:     req.FaviconURL,
		ReportLogoURL:  req.ReportLogoURL,
		ReportCoverURL: req.ReportCoverURL,
	}
}

func (s *brandingService) Update(ctx context.Context, actor string, req dto.BrandingRequest) (*dto.BrandingResponse, error) {
	if req.Enabled && !s.isWhiteLabelEntitled(ctx) {
		return nil, ErrWhiteLabelNotEntitled
	}
	resp := fromRequest(req)
	if err := s.save(ctx, actor, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *brandingService) Seed(ctx context.Context, req dto.BrandingRequest) (*dto.BrandingResponse, error) {
	resp := fromRequest(req)
	if err := s.save(ctx, "installer", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *brandingService) SetAsset(ctx context.Context, actor, slot, url string) (*dto.BrandingResponse, string, error) {
	cur, err := s.read(ctx)
	if err != nil {
		return nil, "", err
	}
	var previous string
	switch slot {
	case AssetLogo:
		previous, cur.LogoURL = cur.LogoURL, url
	case AssetLogoDark:
		previous, cur.LogoDarkURL = cur.LogoDarkURL, url
	case AssetFavicon:
		previous, cur.FaviconURL = cur.FaviconURL, url
	case AssetReportLogo:
		previous, cur.ReportLogoURL = cur.ReportLogoURL, url
	case AssetReportCover:
		previous, cur.ReportCoverURL = cur.ReportCoverURL, url
	default:
		return nil, "", ErrUnknownAssetSlot
	}
	if err := s.save(ctx, actor, &cur); err != nil {
		return nil, "", err
	}
	return &cur, previous, nil
}

// IsBrandingAssetReferenced reports whether any tenant's branding row still
// mentions `url`. Callers use this after replacing an asset to decide whether
// the file on disk can be removed.
func (s *brandingService) IsBrandingAssetReferenced(ctx context.Context, url string) (bool, error) {
	if strings.TrimSpace(url) == "" {
		return false, nil
	}
	n, err := s.repo.CountValueContains(ctx, brandingConfigKey, url)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// GetPublic returns the effective branding for the (unauthenticated) login page.
// White-labeling only takes effect when it is both enabled AND entitled;
// otherwise the default UTMStack brand is returned.
func (s *brandingService) GetPublic(ctx context.Context) (*dto.BrandingPublic, error) {
	b, err := s.read(ctx)
	if err != nil {
		return nil, err
	}

	lang := s.language(ctx)
	if !b.Enabled || !s.isWhiteLabelEntitled(ctx) {
		return &dto.BrandingPublic{ProductName: defaultProductName, Language: lang}, nil
	}
	return &dto.BrandingPublic{
		Enabled:     b.Enabled,
		ProductName: b.ProductName,
		AccentColor: b.AccentColor,
		LogoURL:     b.LogoURL,
		LogoDarkURL: b.LogoDarkURL,
		FaviconURL:  b.FaviconURL,
		Language:    lang,
	}, nil
}

// language returns the configured platform-default UI language, falling back to
// English when unset. Read straight off the config row (not a secret).
func (s *brandingService) language(ctx context.Context) string {
	row, err := s.repo.GetByKey(ctx, systemLanguageKey)
	if err != nil || row == nil || strings.TrimSpace(row.Value) == "" {
		return defaultLanguage
	}
	return strings.TrimSpace(row.Value)
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
