package usecase

import (
	"context"
	"fmt"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"

	os_connectors "github.com/utmstack/utmstack/backend/modules/opensearch/connectors"
	os_dto "github.com/utmstack/utmstack/backend/modules/opensearch/dto"
)

type moduleUsecase struct {
	repo          connectors.ModuleRepository
	tenantFileTog connectors.TenantFileToggler
	opensearch    os_connectors.IndexPatternUsecase
}

func NewModuleUsecase(
	repo connectors.ModuleRepository,
	tenantFile connectors.TenantFileToggler,
	opensearch os_connectors.IndexPatternUsecase,
) connectors.ModuleUsecase {
	return &moduleUsecase{
		repo:          repo,
		tenantFileTog: tenantFile,
		opensearch: opensearch,
	}
}

var _ connectors.ModuleUsecase = (*moduleUsecase)(nil)

func (u *moduleUsecase) ActivateDeactivate(ctx context.Context, req dto.ModuleActivationRequest) (*dto.ModuleResponse, error) {
	if req.ActivationStatus == nil {
		return nil, fmt.Errorf("%w: activationStatus is required", domain.ErrInvalidActivation)
	}
	active := *req.ActivationStatus

	module, err := u.repo.GetByName(ctx, req.ModuleName)
	if err != nil {
		return nil, err
	}

	module.ModuleActive = active
	if err := u.repo.Save(ctx, module); err != nil {
		return nil, err
	}

	if err := u.tenantFileTog.SetActiveByModule(ctx, req.ModuleName, active); err != nil {
		return nil, fmt.Errorf("tenant file toggle: %w", err)
	}

	fresh, err := u.repo.GetByID(ctx, module.ID)
	if err != nil {
		return nil, err
	}

	results,count,err :=u.opensearch.List(ctx,os_dto.IndexPatternFilters{
		ModuleEquals: &module.ModuleName,
	})
	if err != nil {
		return nil,err
	}
	if count > 0 && len(results)>0 {
		//if module has index patterns
		pattern:=results[0]

		if _,err := u.opensearch.Update(ctx,os_dto.UpdateIndexPatternRequest{
			ID: pattern.ID,
			Pattern: pattern.Pattern,
			PatternModule: pattern.PatternModule,
			IsActive: &fresh.ModuleActive,
			PatternSystem:pattern.PatternSystem,
		});err!=nil{
			catcher.Error("failed to toggle is active for module's index pattern, rolling back", err, map[string]any{"module":module.ModuleName,"pattern":pattern} )
			module.ModuleActive = !active
			if err_ := u.repo.Save(ctx, module); err_ != nil {
				catcher.Error("failed to change module status after index pattern activation toggle", err, map[string]any{"module":module.ModuleName,"pattern":pattern} )
				return nil, err
			}
			return nil,err
		}
	}
	resp := dto.FromModule(*fresh)
	return &resp, nil
}

func (u *moduleUsecase) List(ctx context.Context, filter connectors.ModuleListFilter) ([]dto.ModuleResponse, int64, error) {
	rows, total, err := u.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ModuleResponse, 0, len(rows))
	for _, m := range rows {
		out = append(out, dto.FromModule(m))
	}
	return out, total, nil
}

func (u *moduleUsecase) GetByID(ctx context.Context, id int64) (*dto.ModuleResponse, error) {
	m, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.FromModule(*m)
	return &resp, nil
}

func (u *moduleUsecase) GetByName(ctx context.Context, moduleName string) (*dto.ModuleResponse, error) {
	m, err := u.repo.GetByName(ctx, moduleName)
	if err != nil {
		return nil, err
	}
	resp := dto.FromModule(*m)
	return &resp, nil
}

func (u *moduleUsecase) Categories(ctx context.Context) ([]string, error) {
	return u.repo.Categories(ctx)
}

func (u *moduleUsecase) DataTypes(ctx context.Context) ([]dto.DataTypeOption, error) {
	modules, err := u.repo.DataTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DataTypeOption, 0, len(modules))
	for i := range modules {
		m := &modules[i]
		out = append(out, dto.DataTypeOption{
			DataType:   m.DataType,
			Name:       m.PrettyName,
			ModuleName: m.ModuleName,
			Active:     m.ModuleActive,
			IsSystem:   m.IsSystem,
		})
	}
	return out, nil
}

func (u *moduleUsecase) IsActive(ctx context.Context, moduleName string) (bool, error) {
	count, err := u.repo.CountActiveByName(ctx, moduleName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (u *moduleUsecase) Create(ctx context.Context, req dto.CreateModuleRequest) (*dto.ModuleResponse, error) {
	m := &domain.UtmModule{
		ModuleName:        req.ModuleName,
		DataType:          req.DataType,
		PrettyName:        req.PrettyName,
		ModuleDescription: req.ModuleDescription,
		ModuleIcon:        req.ModuleIcon,
		ModuleCategory:    req.ModuleCategory, // technology — user's choice
		IngestType:        "forwarder",        // custom integrations always ingest via the forwarder
		ModuleActive:      false,
		IsSystem:          false,
	}


	if err := u.repo.Save(ctx, m); err != nil {
		return nil, err
	}

	if _,err:= u.opensearch.Create(ctx,os_dto.CreateIndexPatternRequest{
		PatternModule: &req.DataType,
	});err!=nil{
		//opensearch insertion fails, rolling back database
		if err:=u.repo.Delete(ctx,m.ID);err!=nil{
		  catcher.Error("failed to remove module after index pattern creation fail", err, map[string]any{"module":m.ModuleName} )
		}
		return nil, err
	}

	resp := dto.FromModule(*m)
	return &resp, nil
}

func (u *moduleUsecase) Update(ctx context.Context, id int64, req dto.UpdateModuleRequest) (*dto.ModuleResponse, error) {
	m, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.IsSystem {
		return nil, domain.ErrSystemModule
	}
	m.PrettyName = req.PrettyName
	m.ModuleDescription = req.ModuleDescription
	m.ModuleIcon = req.ModuleIcon
	m.ModuleCategory = req.ModuleCategory // technology — user's choice
	m.IngestType = "forwarder"            // custom integrations always ingest via the forwarder
	if err := u.repo.Save(ctx, m); err != nil {
		return nil, err
	}
	resp := dto.FromModule(*m)
	return &resp, nil
}

func (u *moduleUsecase) Delete(ctx context.Context, id int64) error {
	m, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if m.IsSystem {
		return domain.ErrSystemModule
	}

	results,count,err :=u.opensearch.List(ctx,os_dto.IndexPatternFilters{
		ModuleEquals: &m.ModuleName,
	})
	if err != nil {
		return err
	}
	if count == 0 || len(results)==0{
		return fmt.Errorf("module %s index pattern not found.",m.ModuleName)
	}

	err= u.repo.Delete(ctx, id)
	if err!=nil{
		return err
	}

	pattern:=results[0]
	if err:=u.opensearch.Delete(ctx,pattern.ID);err!=nil{
		//if index patter deletion fails integration must be restored
		if err:=u.repo.Save(ctx,m);err!=nil{
			catcher.Error("failed to restore module after index pattern remove fail", err, map[string]any{"module":m.ModuleName} )
		}
		return err
	}

	return nil
}
