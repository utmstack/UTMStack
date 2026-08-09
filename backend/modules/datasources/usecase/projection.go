package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const projectionDebounce = 2 * time.Second

type Notifier interface{ Notify() }

type projector interface {
	ProjectAssets(ctx context.Context) error
}
type AssetProjection struct {
	uc    projector
	dirty chan struct{}
}

func NewAssetProjection(uc projector) *AssetProjection {
	return &AssetProjection{uc: uc, dirty: make(chan struct{}, 1)}
}

func (p *AssetProjection) Notify() {
	select {
	case p.dirty <- struct{}{}:
	default:
	}
}

func (p *AssetProjection) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.dirty:
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(projectionDebounce):
		}

		if err := p.uc.ProjectAssets(ctx); err != nil {
			_ = catcher.Error("datasources: asset projection failed", err, nil)
		}
	}
}
