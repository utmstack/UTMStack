package usecase

import (
	"context"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	"github.com/utmstack/utmstack/backend/modules/opensearch/domain"
)

type policyUsecase struct {
	ismRepo  connectors.ISMRepository
	registry *domain.IndexPatternRegistry
}

func NewPolicyUsecase(ismRepo connectors.ISMRepository, registry *domain.IndexPatternRegistry) connectors.PolicyUsecase {
	return &policyUsecase{ismRepo: ismRepo, registry: registry}
}

func (u *policyUsecase) GetPolicy(ctx context.Context) (*domain.PolicySettings, error) {
	ip, err := u.ismRepo.GetPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("policyUsecase.GetPolicy: %w", err)
	}
	if ip == nil {
		return nil, nil
	}

	openState, err := findState(ip, domain.StateOpen)
	if err != nil {
		return nil, fmt.Errorf("policyUsecase.GetPolicy: %w", err)
	}
	if len(openState.Transitions) == 0 {
		return nil, fmt.Errorf("policyUsecase.GetPolicy: open state has no transitions")
	}

	t0 := openState.Transitions[0]

	var deleteAfter string
	if t0.Conditions != nil {
		deleteAfter = t0.Conditions.MinIndexAge
	}

	settings := &domain.PolicySettings{
		SnapshotActive: t0.StateName == domain.StateSafeDelete,
		DeleteAfter:    deleteAfter,
	}
	return settings, nil
}

func (u *policyUsecase) UpdatePolicy(ctx context.Context, settings domain.PolicySettings) (*domain.UpdateManagedIndexPolicyResponse, error) {
	ip, err := u.ismRepo.GetPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("policyUsecase.UpdatePolicy: get current: %w", err)
	}
	if ip == nil {
		return nil, fmt.Errorf("policyUsecase.UpdatePolicy: policy %s not found", domain.PolicyID)
	}

	changed, err := anyChange(ip, settings)
	if err != nil {
		return nil, fmt.Errorf("policyUsecase.UpdatePolicy: %w", err)
	}

	if changed {
		if err := mutateStates(ip, settings); err != nil {
			return nil, fmt.Errorf("policyUsecase.UpdatePolicy: mutate: %w", err)
		}

		if err := u.ismRepo.UpdatePolicy(ctx, *ip, ip.SeqNo, ip.PrimaryTerm); err != nil {
			return nil, fmt.Errorf("policyUsecase.UpdatePolicy: persist: %w", err)
		}
	}

	result, err := u.updateManagedIndexPolicy(ctx, settings.SnapshotActive)
	if err != nil {
		return nil, fmt.Errorf("policyUsecase.UpdatePolicy: managed index update: %w", err)
	}
	return result, nil
}

func (u *policyUsecase) IsIndexRemovable(ctx context.Context, indexName string) bool {
	removable, err := u.ismRepo.IsIndexRemovable(ctx, indexName)
	if err != nil {
		_ = catcher.Error("opensearch: policyUsecase.IsIndexRemovable", err, nil)
		return false
	}
	return removable
}

func findState(ip *domain.IndexPolicy, name string) (*domain.State, error) {
	if ip.Policy == nil {
		return nil, fmt.Errorf("policy body is nil")
	}
	for i := range ip.Policy.States {
		if ip.Policy.States[i].Name == name {
			return &ip.Policy.States[i], nil
		}
	}
	return nil, fmt.Errorf("state %q not found in policy", name)
}

func anyChange(ip *domain.IndexPolicy, settings domain.PolicySettings) (bool, error) {
	ingest, err := findState(ip, domain.StateIngest)
	if err != nil {
		return false, fmt.Errorf("anyChange: %w", err)
	}
	backupActive := false
	for _, t := range ingest.Transitions {
		if t.StateName == domain.StateBackup {
			backupActive = true
			break
		}
	}
	if backupActive != settings.SnapshotActive {
		return true, nil
	}

	open, err := findState(ip, domain.StateOpen)
	if err != nil {
		return false, fmt.Errorf("anyChange: %w", err)
	}
	if len(open.Transitions) == 0 {
		return true, nil // unexpected, treat as changed
	}
	t0 := open.Transitions[0]
	if t0.Conditions == nil {
		return settings.DeleteAfter != "", nil
	}
	return t0.Conditions.MinIndexAge != settings.DeleteAfter, nil
}

func mutateStates(ip *domain.IndexPolicy, settings domain.PolicySettings) error {
	ingest, err := findState(ip, domain.StateIngest)
	if err != nil {
		return err
	}
	if len(ingest.Transitions) == 0 {
		ingest.Transitions = make([]domain.Transition, 1)
	}

	open, err := findState(ip, domain.StateOpen)
	if err != nil {
		return err
	}
	if len(open.Transitions) == 0 {
		open.Transitions = make([]domain.Transition, 1)
	}

	if settings.SnapshotActive {
		ingest.Transitions[0].StateName = domain.StateBackup
		ingest.Transitions[0].Conditions = &domain.TransitionCondition{MinIndexAge: "24h"}
		open.Transitions[0].StateName = domain.StateSafeDelete
		open.Transitions[0].Conditions = &domain.TransitionCondition{MinIndexAge: settings.DeleteAfter}
	} else {
		ingest.Transitions[0].StateName = domain.StateOpen
		ingest.Transitions[0].Conditions = &domain.TransitionCondition{MinIndexAge: "24h"}
		open.Transitions[0].StateName = domain.StateDelete
		open.Transitions[0].Conditions = &domain.TransitionCondition{MinIndexAge: settings.DeleteAfter}
	}
	return nil
}

func (u *policyUsecase) updateManagedIndexPolicy(ctx context.Context, snapshotActive bool) (*domain.UpdateManagedIndexPolicyResponse, error) {
	patterns := []string{
		u.registry.Get(domain.SysPatternLogs),
		u.registry.Get(domain.SysPatternAlerts),
	}

	result := &domain.UpdateManagedIndexPolicyResponse{}

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		var transitions [][2]string // [sourceState, destState]
		if snapshotActive {
			transitions = [][2]string{
				{domain.StateIngest, domain.StateIngest},
				{domain.StateOpen, domain.StateBackup},
			}
		} else {
			transitions = [][2]string{
				{domain.StateIngest, domain.StateIngest},
				{domain.StateBackup, domain.StateOpen},
				{domain.StateOpen, domain.StateOpen},
			}
		}

		for _, tr := range transitions {
			req := domain.UpdateManagedIndexPolicyConfiguration{
				PolicyID: domain.PolicyID,
				State:    tr[1],                                 // destination
				Include:  []domain.IncludeState{{State: tr[0]}}, // source
			}
			resp, err := u.ismRepo.ChangePolicyForIndex(ctx, pattern, req)
			if err != nil {
				// Log but don't abort — mirror Java's aggregation behaviour.
				_ = catcher.Error(fmt.Sprintf("opensearch: ChangePolicyForIndex(%s %s→%s)",
					pattern, tr[0], tr[1]), err, nil)
				continue
			}
			if resp != nil {
				result.UpdatedIndices += resp.UpdatedIndices
				result.FailedIndices = append(result.FailedIndices, resp.FailedIndices...)
			}
		}
	}

	result.Failures = len(result.FailedIndices) > 0
	return result, nil
}
