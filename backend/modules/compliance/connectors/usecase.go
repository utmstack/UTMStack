package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type FrameworkUsecase interface {
	ListControls(ctx context.Context) []domain.Control
	GetControl(ctx context.Context, id string) (*domain.Control, error)

	ListFrameworks(ctx context.Context) []domain.Framework
	GetFramework(ctx context.Context, key string) (*domain.Framework, error)

	CreateControl(ctx context.Context, c domain.Control) (*domain.Control, error)
	UpdateControl(ctx context.Context, c domain.Control) (*domain.Control, error)
	DeleteControl(ctx context.Context, id string) error
	SetControlEnabled(ctx context.Context, id string, enabled bool) error

	CreateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error)
	UpdateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error)
	DeleteFramework(ctx context.Context, key string) error
	SetFrameworkEnabled(ctx context.Context, key string, enabled bool) error
}

type EvaluatorUsecase interface {
	EvaluateFramework(ctx context.Context, frameworkKey string) (*domain.Report, error)
	GenerateReport(ctx context.Context, frameworkKey string) (*domain.Report, error) // evaluate + store snapshot
	ListReports(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error)
	GetReport(ctx context.Context, id string) (*domain.ReportSnapshot, error)
}

type ScheduleUsecase interface {
	Create(ctx context.Context, userID int64, req dto.CreateScheduleRequest) (*dto.ScheduleResponse, error)
	Update(ctx context.Context, userID int64, req dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.ScheduleResponse, error)
	ListByUser(ctx context.Context, userID int64, f dto.ScheduleFilters) ([]dto.ScheduleResponse, int64, error)
	Delete(ctx context.Context, id int64) error
}
