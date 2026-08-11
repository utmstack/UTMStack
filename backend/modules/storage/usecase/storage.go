package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/storage/connectors"
	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

type service struct {
	store  connectors.StoreRepository
	config connectors.ConfigRepository
}

func New(store connectors.StoreRepository, config connectors.ConfigRepository) connectors.Usecase {
	return &service{store: store, config: config}
}

func (s *service) Retentions(ctx context.Context) ([]domain.Retention, error) {
	out := make([]domain.Retention, 0, len(domain.Datasets()))
	for _, d := range domain.Datasets() {
		r, err := s.store.Retention(ctx, d)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *service) SetRetention(ctx context.Context, want domain.Retention) (domain.Retention, error) {
	if err := want.Validate(); err != nil {
		return domain.Retention{}, err
	}

	if want.Tiered() {
		ready, err := s.store.ColdStorageReady(ctx)
		if err != nil {
			return domain.Retention{}, err
		}
		if !ready {
			return domain.Retention{}, domain.ErrTieringRequired
		}
		if err := s.store.AdoptTiering(ctx, want); err != nil {
			return domain.Retention{}, err
		}
	} else if err := s.store.SetRetention(ctx, want); err != nil {
		return domain.Retention{}, err
	}

	return s.store.Retention(ctx, want.Dataset)
}

func (s *service) Usage(ctx context.Context) ([]domain.Usage, error) { return s.store.Usage(ctx) }

func (s *service) Health(ctx context.Context) (domain.Health, error) { return s.store.Health(ctx) }

func (s *service) Tiering(ctx context.Context) (domain.Tiering, error) {
	t, err := s.config.Read()
	if err != nil {
		return domain.Tiering{}, err
	}

	ready, err := s.store.ColdStorageReady(ctx)
	if err != nil {
		return domain.Tiering{}, err
	}
	t.Ready = ready
	if ready {
		t.Policy = domain.PolicyName
	}
	return t, nil
}

func (s *service) EnableTiering(ctx context.Context, o domain.ObjectStore) (domain.Tiering, error) {
	o = o.Normalized()
	if err := o.Validate(); err != nil {
		return domain.Tiering{}, err
	}

	current, err := s.config.Read()
	if err != nil {
		return domain.Tiering{}, err
	}
	if current.Configured && current.Endpoint != o.Endpoint {
		moved, err := s.anyDatasetTiered(ctx)
		if err != nil {
			return domain.Tiering{}, err
		}
		if moved {
			return domain.Tiering{}, domain.ErrEndpointLocked
		}
	}

	undo, err := s.config.Write(o)
	if err != nil {
		return domain.Tiering{}, err
	}

	// The store checks the bucket while reading the file — it writes a probe
	// object — so a wrong endpoint or key fails here rather than a day later.
	// The file has to come back out: left behind, it is what the server would
	// load on its next start.
	if err := s.store.ReloadConfig(ctx); err != nil {
		if undo != nil {
			_ = undo()
			_ = s.store.ReloadConfig(ctx)
		}
		return domain.Tiering{}, fmt.Errorf("%w: %s", domain.ErrColdRefused, err.Error())
	}

	// Not ready after a reload that succeeded is a deployment problem rather than a
	// wrong bucket, so the configuration stays: the operator fixes the mount and
	// does not have to type it again.
	ready, err := s.waitForColdStorage(ctx)
	if err != nil {
		return domain.Tiering{}, err
	}
	if !ready {
		return domain.Tiering{}, domain.ErrColdNotReady
	}

	return domain.Tiering{
		Configured: true,
		Ready:      true,
		Endpoint:   o.Endpoint,
		Policy:     domain.PolicyName,
	}, nil
}

func (s *service) anyDatasetTiered(ctx context.Context) (bool, error) {
	for _, d := range domain.Datasets() {
		r, err := s.store.Retention(ctx, d)
		if err != nil {
			return false, err
		}
		if r.Tiered() {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) waitForColdStorage(ctx context.Context) (bool, error) {
	var lastErr error
	for i := 0; i < 5; i++ {
		ready, err := s.store.ColdStorageReady(ctx)
		if err == nil && ready {
			return true, nil
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}
	return false, lastErr
}
