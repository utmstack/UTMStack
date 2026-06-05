package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

type portsUsecase struct {
	repo connectors.PortsRepository
}

func NewPortsUsecase(repo connectors.PortsRepository) connectors.PortsUsecase {
	return &portsUsecase{repo: repo}
}

func (u *portsUsecase) Create(ctx context.Context, p *domain.UtmPorts) (*domain.UtmPorts, error) {
	if err := u.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (u *portsUsecase) Update(ctx context.Context, p *domain.UtmPorts) (*domain.UtmPorts, error) {
	if err := u.repo.Save(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (u *portsUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmPorts, error) {
	p, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (u *portsUsecase) ListByCriteria(ctx context.Context, c domain.PortsCriteria, p domain.Pagination) ([]domain.UtmPorts, int64, error) {
	return u.repo.ListByCriteria(ctx, c, p)
}

func (u *portsUsecase) CountByCriteria(ctx context.Context, c domain.PortsCriteria) (int64, error) {
	return u.repo.CountByCriteria(ctx, c)
}

func (u *portsUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
