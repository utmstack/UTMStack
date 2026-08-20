package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/appconfig/dto"
)

type Usecase interface {
	List(ctx context.Context) ([]dto.ConfigResponse, error)
	Get(ctx context.Context, key string) (*dto.ConfigResponse, error)
	Update(ctx context.Context, actor, key string, input dto.UpsertRequest) (*dto.ConfigResponse, error)
	CheckMail(ctx context.Context, configs []domain.MailConfig) error
	GetDateFormat(ctx context.Context) dto.DateFormatResponse
}

type BrandingUsecase interface {
	Get(ctx context.Context) (*dto.BrandingResponse, error)
	Update(ctx context.Context, actor string, req dto.BrandingRequest) (*dto.BrandingResponse, error)
	Seed(ctx context.Context, req dto.BrandingRequest) (*dto.BrandingResponse, error)
	// SetAsset returns the updated branding plus the URL previously stored in
	// `slot` (empty if none). Callers use the previous URL to garbage-collect
	// the now-unreferenced file.
	SetAsset(ctx context.Context, actor, slot, url string) (resp *dto.BrandingResponse, previousURL string, err error)
	IsBrandingAssetReferenced(ctx context.Context, url string) (bool, error)
	GetPublic(ctx context.Context) (*dto.BrandingPublic, error)
	BrandNameProvider
}

type BrandNameProvider interface {
	ProductName(ctx context.Context) string
}

type Store interface {
	GetString(ctx context.Context, key string) (string, bool, error)
	SetString(ctx context.Context, key, value string, opts SetOpts) error
}

type SetOpts struct {
	IsSecret    bool
	Label       string
	Description string
}
