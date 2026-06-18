package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
)

type agentUsecase struct {
	repo connectors.AgentRepository
}

func NewAgentUsecase(repo connectors.AgentRepository) connectors.AgentUsecase {
	return &agentUsecase{repo: repo}
}

func (u *agentUsecase) ListByPlatform(ctx context.Context, platform string) ([]string, error) {
	if platform == "" {
		return []string{}, nil
	}
	return u.repo.ListNamesByPlatform(ctx, platform)
}
