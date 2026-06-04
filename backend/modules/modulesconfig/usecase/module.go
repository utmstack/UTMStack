package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/dto"
)

var needsRestartModules = map[string]struct{}{
	"AWS_IAM_USER": {},
	"AZURE":        {},
	"GCP":          {},
	"SOPHOS":       {},
}

type moduleUsecase struct {
	repo            connectors.ModuleRepository
	factory         connectors.ModuleFactory
	indexPatternTog connectors.IndexPatternToggler
	logstashTog     connectors.LogstashFilterToggler
	menuTog         connectors.MenuToggler
}

func NewModuleUsecase(
	repo connectors.ModuleRepository,
	factory connectors.ModuleFactory,
	ip connectors.IndexPatternToggler,
	ls connectors.LogstashFilterToggler,
	menu connectors.MenuToggler,
) connectors.ModuleUsecase {
	return &moduleUsecase{
		repo:            repo,
		factory:         factory,
		indexPatternTog: ip,
		logstashTog:     ls,
		menuTog:         menu,
	}
}

func (u *moduleUsecase) ActivateDeactivate(ctx context.Context, req dto.ModuleActivationRequest) (*dto.ModuleResponse, error) {
	if req.ActivationStatus == nil {
		return nil, fmt.Errorf("%w: activationStatus is required", domain.ErrInvalidActivation)
	}
	active := *req.ActivationStatus

	module, err := u.repo.GetByServerAndName(ctx, req.ServerID, req.ModuleName)
	if err != nil {
		return nil, err
	}

	module.ModuleActive = active
	if err := u.repo.Save(ctx, module); err != nil {
		return nil, err
	}

	if err := u.applyCascade(ctx, req.ModuleName, active); err != nil {
		return nil, err
	}

	fresh, err := u.repo.GetByID(ctx, module.ID)
	if err != nil {
		return nil, err
	}

	if u.eventProc != nil {
		if perr := u.eventProc.UpdateModule(ctx, fresh.ModuleName, dto.ToEventProcessorPayload(*fresh)); perr != nil {
			return nil, fmt.Errorf("event-processor publish failed: %w", perr)
		}
	}

	resp := dto.FromModule(*fresh, false)
	return &resp, nil
}

func (u *moduleUsecase) applyCascade(ctx context.Context, moduleName string, active bool) error {
	activeCount, err := u.repo.CountActiveByName(ctx, moduleName)
	if err != nil {
		return err
	}
	if (!active && activeCount > 0) || (active && activeCount > 1) {
		return nil
	}

	if err := u.indexPatternTog.SetActiveByModule(ctx, moduleName, active); err != nil {
		return fmt.Errorf("index pattern cascade: %w", err)
	}
	if err := u.logstashTog.SetActiveByModule(ctx, moduleName, active); err != nil {
		return fmt.Errorf("logstash filter cascade: %w", err)
	}
	if err := u.menuTog.SetActiveByModule(ctx, moduleName, active); err != nil {
		return fmt.Errorf("menu cascade: %w", err)
	}
	return nil
}

func (u *moduleUsecase) List(ctx context.Context, filter connectors.ModuleListFilter) ([]dto.ModuleResponse, int64, error) {
	rows, total, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ModuleResponse, 0, len(rows))
	for _, m := range rows {
		out = append(out, dto.FromModule(m, false))
	}
	return out, total, nil
}

func (u *moduleUsecase) GetByID(ctx context.Context, id int64) (*dto.ModuleResponse, error) {
	m, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.FromModule(*m, false)
	return &resp, nil
}

func (u *moduleUsecase) Details(ctx context.Context, serverID int64, moduleName string, reveal bool) (*dto.ModuleResponse, error) {
	m, err := u.repo.GetByServerAndName(ctx, serverID, moduleName)
	if err != nil {
		return nil, err
	}
	resp := dto.FromModule(*m, reveal)
	return &resp, nil
}

func (u *moduleUsecase) CheckRequirements(ctx context.Context, serverID int64, moduleName string) (*dto.CheckRequirementsResponse, error) {
	kind, ok := u.factory.Get(moduleName)
	if !ok {
		return nil, fmt.Errorf("%w: %s", domain.ErrModuleKindNotPorted, moduleName)
	}
	checks, err := kind.CheckRequirements(ctx, serverID)
	if err != nil {
		return nil, err
	}
	status := domain.RequirementOK
	for _, c := range checks {
		if c.CheckStatus == domain.RequirementFail {
			status = domain.RequirementFail
			break
		}
	}
	if checks == nil {
		checks = []domain.ModuleRequirement{}
	}
	return &dto.CheckRequirementsResponse{Status: status, Checks: checks}, nil
}

func (u *moduleUsecase) Categories(ctx context.Context, serverID *int64) ([]string, error) {
	return u.repo.Categories(ctx, serverID)
}

func (u *moduleUsecase) IsActive(ctx context.Context, moduleName string) (bool, error) {
	count, err := u.repo.CountActiveByName(ctx, moduleName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func markNeedsRestart(module *domain.UtmModule) {
	_, restart := needsRestartModules[module.ModuleName]
	module.NeedsRestart = restart
}
