package usecase

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/connectors"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

type BootstrapService struct {
	repo     connectors.IndexPatternRepository
	registry *domain.IndexPatternRegistry
	ismRepo  connectors.ISMRepository // nil in Slice 1; wired in Slice 2
}

func NewBootstrapService(repo connectors.IndexPatternRepository, registry *domain.IndexPatternRegistry) *BootstrapService {
	return &BootstrapService{repo: repo, registry: registry}
}

func (b *BootstrapService) SetISMRepo(ismRepo connectors.ISMRepository) {
	b.ismRepo = ismRepo
}

func (b *BootstrapService) InitPatterns(ctx context.Context) error {
	patterns, err := b.repo.FindAll(ctx)
	if err != nil {
		return err
	}
	if len(patterns) == 0 {
		return errors.New("no index patterns configured")
	}

	byID := make(map[int64]string, len(patterns))
	for _, p := range patterns {
		byID[p.ID] = p.Pattern
	}

	b.registry.Logs = byID[1]
	b.registry.Alerts = byID[2]
	b.registry.LogsWindows = byID[8]

	if b.registry.Logs == "" {
		logger.Warn("indexpattern: system pattern id=1 (LOGS) not found in database")
	}
	if b.registry.Alerts == "" {
		logger.Warn("indexpattern: system pattern id=2 (ALERTS) not found in database")
	}
	if b.registry.LogsWindows == "" {
		logger.Warn("indexpattern: system pattern id=8 (LOGS_WINDOWS) not found in database")
	}

	logger.Info("indexpattern: registry populated — LOGS=" + b.registry.Logs +
		" ALERTS=" + b.registry.Alerts +
		" LOGS_WINDOWS=" + b.registry.LogsWindows)
	return nil
}

func (b *BootstrapService) InitPolicy(ctx context.Context) error {
	if b.ismRepo == nil {
		logger.Info("indexpattern: ISM repository not configured — skipping policy bootstrap")
		return nil
	}

	if err := b.createPolicy(ctx); err != nil {
		return err
	}
	return b.registerSnapshotRepo(ctx)
}

func (b *BootstrapService) createPolicy(ctx context.Context) error {
	existing, err := b.ismRepo.GetPolicy(ctx)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	policy := buildDefaultPolicy(b.registry)
	if err := b.ismRepo.CreatePolicy(ctx, policy); err != nil {
		return err
	}
	logger.Info("indexpattern: ISM policy created")

	logsPattern := b.registry.Get(domain.SysPatternLogs)
	alertsPattern := b.registry.Get(domain.SysPatternAlerts)

	if logsPattern != "" {
		if err := b.ismRepo.AddPolicyToIndex(ctx, logsPattern); err != nil {
			return err
		}
	}
	if alertsPattern != "" {
		if err := b.ismRepo.AddPolicyToIndex(ctx, alertsPattern); err != nil {
			return err
		}
	}
	return nil
}

func (b *BootstrapService) registerSnapshotRepo(ctx context.Context) error {
	exists, err := b.ismRepo.GetSnapshotRepo(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	logger.Info("indexpattern: registering snapshot repository")
	return b.ismRepo.RegisterSnapshotRepo(ctx)
}

func buildDefaultPolicy(registry *domain.IndexPatternRegistry) domain.IndexPolicy {
	logsPattern := registry.Get(domain.SysPatternLogs)
	alertsPattern := registry.Get(domain.SysPatternAlerts)

	if logsPattern == "" {
		logsPattern = "v11-log-*" // fallback: post-Liquibase 20241227001 value
	}
	if alertsPattern == "" {
		alertsPattern = "v11-alert-*" // fallback: post-Liquibase 20241227001 value
	}

	policy := domain.Policy{
		PolicyID:     domain.PolicyID,
		Description:  "UtmStack main index lifecycle",
		DefaultState: domain.StateIngest,
		IsmTemplate: []domain.IsmTemplate{
			{
				IndexPatterns: []string{logsPattern, alertsPattern},
				Priority:      1,
			},
		},
		States: []domain.State{
			// 1. ingest → open (24h)
			{
				Name:    domain.StateIngest,
				Actions: []domain.Action{},
				Transitions: []domain.Transition{
					{
						StateName:  domain.StateOpen,
						Conditions: &domain.TransitionCondition{MinIndexAge: "24h"},
					},
				},
			},
			// 2. open → delete (30d)
			{
				Name:    domain.StateOpen,
				Actions: []domain.Action{},
				Transitions: []domain.Transition{
					{
						StateName:  domain.StateDelete,
						Conditions: &domain.TransitionCondition{MinIndexAge: "30d"},
					},
				},
			},
			// 3. backup → snapshot, then unconditional → open
			{
				Name: domain.StateBackup,
				Actions: []domain.Action{
					{Snapshot: &domain.ActionSnapshot{
						Repository: domain.SnapshotRepoName,
						Snapshot:   domain.SnapshotName,
					}},
				},
				Transitions: []domain.Transition{
					{
						StateName:  domain.StateOpen,
						Conditions: nil, // omitempty — no condition on this transition
					},
				},
			},
			// 4. delete — terminates lifecycle
			{
				Name: domain.StateDelete,
				Actions: []domain.Action{
					{Delete: &domain.ActionDelete{}},
				},
				Transitions: []domain.Transition{},
			},
			// 5. safe_delete — snapshot then delete
			{
				Name: domain.StateSafeDelete,
				Actions: []domain.Action{
					{Snapshot: &domain.ActionSnapshot{
						Repository: domain.SnapshotRepoName,
						Snapshot:   domain.SnapshotName,
					}},
					{Delete: &domain.ActionDelete{}},
				},
				Transitions: []domain.Transition{},
			},
		},
	}

	return domain.IndexPolicy{Policy: &policy}
}
