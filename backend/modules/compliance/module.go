package compliance

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"gorm.io/gorm"

	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	mail_domain "github.com/utmstack/utmstack/backend/internal/mail/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/handler"
	"github.com/utmstack/utmstack/backend/modules/compliance/repository"
	"github.com/utmstack/utmstack/backend/modules/compliance/usecase"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type Module struct {
	scores        connectors.ReportScoreStore
	bodyRetention time.Duration

	frameworkH *handler.FrameworkHandler
	reportH    *handler.ReportHandler
	scheduleH  *handler.ScheduleHandler
	scheduler  *usecase.ReportScheduler
	coverage   *usecase.CoverageIndex

	frameworkUC connectors.FrameworkUsecase
	evaluatorUC connectors.EvaluatorUsecase
	scheduleUC  connectors.ScheduleUsecase
}

func (m *Module) GetFrameworkUsecase() connectors.FrameworkUsecase { return m.frameworkUC }
func (m *Module) GetEvaluatorUsecase() connectors.EvaluatorUsecase { return m.evaluatorUC }
func (m *Module) GetScheduleUsecase() connectors.ScheduleUsecase   { return m.scheduleUC }

func NewModule(db *gorm.DB, events repository.Reader, mailSvc mail_connectors.MailService, brand connectors.BrandingProvider, isEnterprise func() bool) *Module {
	src := env.String("COMPLIANCE_SRC_DIR", "/utmstack/compliance", false)
	root := env.String("COMPLIANCE_DIR", "/workdir/compliance", false)

	controlStore := usecase.NewControlStore(
		filepath.Join(src, "controls"),
		filepath.Join(root, "controls", usecase.UserSubdir),
	)
	frameworkStore := usecase.NewFrameworkStore(
		filepath.Join(src, "frameworks"),
		filepath.Join(root, "frameworks", usecase.UserSubdir),
	)
	if err := controlStore.Load(); err != nil {
		_ = catcher.Error("compliance: loading control library failed", err, map[string]any{"src": src})
	}
	if err := frameworkStore.Load(); err != nil {
		_ = catcher.Error("compliance: loading frameworks failed", err, map[string]any{"src": src})
	}

	coverageIdx := usecase.NewCoverageIndex(env.String("EVENT_PROCESSING_RULES_DIR", "/workdir/rules", false))
	if err := coverageIdx.Load(); err != nil {
		_ = catcher.Error("compliance: loading rule coverage index failed", err, nil)
	}

	scheduleRepo := repository.NewScheduleRepository(db)
	scoreStore := repository.NewReportScoreStore(db)
	entitlement := usecase.NewEntitlement(isEnterprise)
	frameworkUC := usecase.NewFrameworkUsecase(controlStore, frameworkStore, entitlement)
	evaluatorUC := usecase.NewEvaluator(
		controlStore, frameworkStore, coverageIdx,
		repository.NewEventCounter(events),
		repository.NewReportStore(db),
		scoreStore,
		scheduleRepo, brand, entitlement,
	)
	scheduleUC := usecase.NewScheduleUsecase(scheduleRepo, frameworkStore, entitlement)
	scheduler := usecase.NewReportScheduler(scheduleRepo, evaluatorUC, &mailSender{svc: mailSvc})

	retentionDays := env.Int("COMPLIANCE_REPORT_BODY_RETENTION_DAYS", 730, false)
	if retentionDays < 1 {
		retentionDays = 730
	}

	return &Module{
		scores:        scoreStore,
		bodyRetention: time.Duration(retentionDays) * 24 * time.Hour,

		frameworkH:  handler.NewFrameworkHandler(frameworkUC),
		reportH:     handler.NewReportHandler(evaluatorUC),
		scheduleH:   handler.NewScheduleHandler(scheduleUC),
		scheduler:   scheduler,
		coverage:    coverageIdx,
		frameworkUC: frameworkUC,
		evaluatorUC: evaluatorUC,
		scheduleUC:  scheduleUC,
	}
}

func (m *Module) Start(ctx context.Context) {
	go m.scheduler.Start(ctx)
	go m.reloadCoverage(ctx)
	go m.pruneReportBodies(ctx)
}

func (m *Module) pruneReportBodies(ctx context.Context) {
	if m.scores == nil {
		return
	}
	prune := func() {
		n, err := m.scores.PruneBodies(tenancy.WithAllTenants(ctx), time.Now().UTC().Add(-m.bodyRetention))
		if err != nil {
			_ = catcher.Error("compliance: pruning stored report bodies failed", err, nil)
			return
		}
		if n > 0 {
			catcher.Info("compliance: released stored report bodies past retention", map[string]any{"points": n})
		}
	}
	prune()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			prune()
		}
	}
}

func (m *Module) reloadCoverage(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.coverage.Load(); err != nil {
				_ = catcher.Error("compliance: reloading rule coverage index failed", err, nil)
			}
		}
	}
}

// ── Mail sender wrapper ───────────────────────────────────────────────────────

type mailSender struct{ svc mail_connectors.MailService }

func (m *mailSender) SendComplianceReport(ctx context.Context, to, cc []string, subject string, pdfData []byte) error {
	if m.svc == nil {
		_ = catcher.Error("compliance: mail service not configured", nil, nil)
		return nil
	}
	attachment := mail_domain.Attatchment{
		Filename:    fmt.Sprintf("Compliance_Report_%s.pdf", time.Now().UTC().Format("20060102_150405")),
		ContentType: "application/pdf",
		Bytes:       pdfData,
	}
	body := fmt.Sprintf("<html><body><p>%s</p></body></html>", subject)
	return m.svc.SendMail(ctx, to, cc, subject, body, []mail_domain.Attatchment{attachment})
}

var _ connectors.MailSender = (*mailSender)(nil)
