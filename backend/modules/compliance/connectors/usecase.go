package connectors

import (
	"context"

	"github.com/google/uuid"

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

type ReportBrand struct {
	Name      string // product/company name (defaults to "UTMStack")
	LogoPath  string // absolute path to a PNG/JPG report logo on disk ("" → none)
	AccentHex string // accent color, CSS hex (e.g. "#6366f1"); "" → default
}

type BrandingProvider interface {
	ReportBrand(ctx context.Context) ReportBrand
}

type EvaluatorUsecase interface {
	EvaluateFramework(ctx context.Context, frameworkKey string) (*domain.Report, error)
	GenerateReport(ctx context.Context, frameworkKey string) (*domain.Report, error) // evaluate + store snapshot
	ListReports(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error)
	GetReport(ctx context.Context, id string) (*domain.ReportSnapshot, error)
	DeleteReport(ctx context.Context, id string) error
	FrameworkReportPDF(ctx context.Context, frameworkKey string) ([]byte, string, error) // live eval → PDF + framework name
	SnapshotPDF(ctx context.Context, id string) ([]byte, string, error)                  // stored snapshot → PDF + framework name
	SetStatusOverride(ctx context.Context, frameworkKey, controlID, status, reason string) error
	ClearStatusOverride(ctx context.Context, frameworkKey, controlID string) error
	SetControlNote(ctx context.Context, frameworkKey, controlID, note string) error
	ClearControlNote(ctx context.Context, frameworkKey, controlID string) error
}

type ScheduleUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateScheduleRequest) (*dto.ScheduleResponse, error)
	Update(ctx context.Context, userID uuid.UUID, req dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error)
	GetByID(ctx context.Context, id int64) (*dto.ScheduleResponse, error)
	ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]dto.ScheduleResponse, int64, error)
	Delete(ctx context.Context, id int64) error
}
