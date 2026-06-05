package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
	"github.com/utmstack/utmstack/backend/modules/network_scan/dto"
)

type probeUsecase struct {
	client connectors.ProbeClient
}

func NewProbeUsecase(client connectors.ProbeClient) connectors.ProbeUsecase {
	return &probeUsecase{client: client}
}

func (u *probeUsecase) Scan(ctx context.Context, req dto.ProbeScanRequest) ([]domain.NetworkScan, error) {
	return u.client.Scan(ctx, req.ProbeID, req.Interface, req.NetworkRange)
}

func (u *probeUsecase) Ping(ctx context.Context, probeID uint64) error {
	return u.client.Ping(ctx, probeID)
}

func (u *probeUsecase) CheckInterface(ctx context.Context, probeID uint64, iface string) (bool, error) {
	return u.client.CheckInterface(ctx, probeID, iface)
}
