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
