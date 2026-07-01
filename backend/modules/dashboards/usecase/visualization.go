package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	os_usecase "github.com/utmstack/utmstack/backend/modules/opensearch/usecase"
)

type visualizationUsecase struct {
	repo connectors.VisualizationRepository
}

func NewVisualizationUsecase(repo connectors.VisualizationRepository) connectors.VisualizationUsecase {
	return &visualizationUsecase{repo: repo}
}

func (u *visualizationUsecase) Create(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error) {
	if v.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(v.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	if err := sanitizeVisualizationSQL(v); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	v.CreatedDate = now
	v.ModifiedDate = now
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) Update(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error) {
	if v.ID == 0 {
		return nil, domain.ErrIDRequired
	}
	existing, err := u.repo.FindByID(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	if err := sanitizeVisualizationSQL(v); err != nil {
		return nil, err
	}
	v.CreatedDate = existing.CreatedDate
	v.SystemOwner = existing.SystemOwner
	v.ModifiedDate = time.Now().UTC()
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) GetByID(ctx context.Context, id uint64) (*domain.Visualization, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *visualizationUsecase) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *visualizationUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}

// sanitizeVisualizationSQL trims and validates a visualization's SQL query
// through the same guard used at query-execution time (opensearch.ValidateSQL:
// SELECT-only, no comments, no DML/DDL keywords, balanced quotes/parens, only
// whitelisted aggregate functions). Placeholders like {{timeFilter}} and
// {{dashboardFilters}} pass through unchanged. Enforcing here on save prevents
// dangerous queries from ever reaching the database.
func sanitizeVisualizationSQL(v *domain.Visualization) error {
	v.SQLQuery = strings.TrimSpace(v.SQLQuery)
	if v.SQLQuery == "" {
		return domain.ErrSQLQueryRequired
	}
	if err := os_usecase.ValidateSQL(v.SQLQuery); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidSQL, err.Error())
	}
	return nil
}
