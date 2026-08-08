package connectors

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type FrameworkUsecase interface {
	ListControls(ctx context.Context) []dto.ControlResponse
	GetControl(ctx context.Context, id string) (*dto.ControlResponse, error)

	ListFrameworks(ctx context.Context) []dto.FrameworkResponse
	GetFramework(ctx context.Context, key string) (*dto.FrameworkResponse, error)

	CreateControl(ctx context.Context, c domain.Control) (*domain.Control, error)
	UpdateControl(ctx context.Context, c domain.Control) (*domain.Control, error)
	DeleteControl(ctx context.Context, id string) error

	CreateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error)
	UpdateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error)
	DeleteFramework(ctx context.Context, key string) error
}

type ReportBrand struct {
	Name     string // product/company name (defaults to "UTMStack")
	LogoPath string // absolute path to a PNG/JPG report logo on disk ("" → none)
	// CoverPath is the full-bleed image behind the cover page, configured with
	// the rest of the branding. Empty falls back to a plain typographic cover.
	CoverPath string
	AccentHex string // accent color, CSS hex (e.g. "#6366f1"); "" → default
	// PreparedBy names whoever asked for the document. A report says who
	// produced it; that is half of what makes it evidence.
	PreparedBy string
}

type BrandingProvider interface {
	ReportBrand(ctx context.Context) ReportBrand
}

type EvaluatorUsecase interface {
	Evaluate(ctx context.Context, frameworkKey string, windowDays int) (*dto.ReportResponse, error)
	Get(ctx context.Context, frameworkKey string) (*dto.ReportResponse, error)
	List(ctx context.Context) ([]dto.ReportMeta, error)
	Delete(ctx context.Context, frameworkKey string) error
	EditControl(ctx context.Context, editedBy, frameworkKey, controlID string, req dto.EditControlRequest) (*dto.ReportResponse, error)
	// preparedBy names whoever asked for the document — a report says who
	// produced it, and that is half of what makes it evidence. A scheduled run
	// has no requester and passes "".
	PDF(ctx context.Context, frameworkKey, preparedBy string) ([]byte, string, error)
	History(ctx context.Context, frameworkKey string, from, to time.Time) ([]dto.ScorePoint, error)
	HistoryPDF(ctx context.Context, frameworkKey, preparedBy string, day time.Time) ([]byte, string, error)
}

type ScheduleUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.CreateScheduleRequest) (*dto.ScheduleResponse, error)
	Update(ctx context.Context, userID uuid.UUID, req dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*dto.ScheduleResponse, error)
	ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]dto.ScheduleResponse, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
