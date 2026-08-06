package compliance

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	mail_domain "github.com/utmstack/utmstack/backend/internal/mail/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/handler"
	"github.com/utmstack/utmstack/backend/modules/compliance/repository"
	"github.com/utmstack/utmstack/backend/modules/compliance/usecase"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
	"gorm.io/gorm"
)

// Leases is the slice of pkg/joblease used to keep peer replicas from
// re-evaluating the same (tenant, framework) — each pair is claimed by one
// replica per interval. Kept as a local interface so tests can stub it.
type Leases interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

// TenantLister returns the active tenants the background eval loop should
// iterate over. It is passed as a func so modules.go can supply a closure that
// resolves the tenant module late — compliance is constructed before tenant.
// nil is legal — on-prem / no tenancy plane → the loop falls back to the
// legacy behaviour (one pass under the module's own ctx).
type TenantLister func(ctx context.Context) ([]string, error)

type Module struct {
	frameworkH *handler.FrameworkHandler
	reportH    *handler.ReportHandler
	scheduleH  *handler.ScheduleHandler
	scheduler  *usecase.ReportScheduler
	coverage   *usecase.CoverageIndex

	frameworkUC       connectors.FrameworkUsecase
	evaluatorUC       connectors.EvaluatorUsecase
	scheduleUC        connectors.ScheduleUsecase
	tenantFrameworks  connectors.TenantFrameworkRepository
	frameworkStore    *usecase.FrameworkStore
	entitlement       *usecase.Entitlement
	tenantLister      TenantLister // nil → skip the tenant iteration (on-prem)
	leases            Leases       // nil → no replica coordination (single-instance)
	evalInterval      time.Duration
}

func (m *Module) GetFrameworkUsecase() connectors.FrameworkUsecase { return m.frameworkUC }
func (m *Module) GetEvaluatorUsecase() connectors.EvaluatorUsecase { return m.evaluatorUC }
func (m *Module) GetScheduleUsecase() connectors.ScheduleUsecase   { return m.scheduleUC }

func NewModule(db *gorm.DB, events repository.Reader, mailSvc mail_connectors.MailService, brand connectors.BrandingProvider, tenants TenantLister, leases Leases, isEnterprise func() bool) *Module {
	scheduleRepo := repository.NewScheduleRepository(db)
	overrideRepo := repository.NewControlStatusOverrideRepository(db)
	noteRepo := repository.NewControlNoteRepository(db)
	tenantFrameworkRepo := repository.NewTenantFrameworkRepository(db)

	root := env.String("COMPLIANCE_DIR", "/workdir/compliance", false)
	src := env.String("COMPLIANCE_SRC_DIR", "/utmstack/compliance", false)
	controlsRoot := filepath.Join(root, "controls")
	frameworksRoot := filepath.Join(root, "frameworks")
	_ = usecase.SeedSystemOverlay(filepath.Join(src, "controls"), filepath.Join(controlsRoot, usecase.SystemSubdir))
	_ = usecase.SeedSystemOverlay(filepath.Join(src, "frameworks"), filepath.Join(frameworksRoot, usecase.SystemSubdir))

	controlStore := usecase.NewControlStore(filepath.Join(controlsRoot, usecase.SystemSubdir), filepath.Join(controlsRoot, usecase.UserSubdir))
	frameworkStore := usecase.NewFrameworkStore(filepath.Join(frameworksRoot, usecase.SystemSubdir), filepath.Join(frameworksRoot, usecase.UserSubdir))
	if err := controlStore.Load(); err != nil {
		_ = catcher.Error("compliance: loading control library failed", err, map[string]any{"dir": controlsRoot})
	}
	if err := frameworkStore.Load(); err != nil {
		_ = catcher.Error("compliance: loading frameworks failed", err, map[string]any{"dir": frameworksRoot})
	}

	// Coverage index: control → enabled correlation rules that tag it (from the
	// eventprocessing rules overlay), powering the coverage/activity dimensions.
	coverageIdx := usecase.NewCoverageIndex(env.String("EVENT_PROCESSING_RULES_DIR", "/workdir/rules", false))
	if err := coverageIdx.Load(); err != nil {
		_ = catcher.Error("compliance: loading rule coverage index failed", err, nil)
	}

	entitlement := usecase.NewEntitlement(isEnterprise)
	frameworkUC := usecase.NewFrameworkUsecase(controlStore, frameworkStore, tenantFrameworkRepo, entitlement)
	evaluatorUC := usecase.NewEvaluator(controlStore, frameworkStore, repository.NewOpenSearchSQL(), coverageIdx, repository.NewCHAlerts(events), repository.NewReportStore(db), overrideRepo, noteRepo, tenantFrameworkRepo, brand, entitlement)
	scheduleUC := usecase.NewScheduleUsecase(scheduleRepo, frameworkStore, entitlement)

	mailSender := &mailSender{svc: mailSvc}
	scheduler := usecase.NewReportScheduler(scheduleRepo, evaluatorUC, mailSender)

	// How often to re-evaluate every framework's compliance posture and store a
	// fresh snapshot, so the UI always shows the latest state. Default 24h.
	evalHours := env.Int("COMPLIANCE_EVAL_INTERVAL_HOURS", 24, false)
	if evalHours < 1 {
		evalHours = 24
	}

	return &Module{
		frameworkH:       handler.NewFrameworkHandler(frameworkUC),
		reportH:          handler.NewReportHandler(evaluatorUC),
		scheduleH:        handler.NewScheduleHandler(scheduleUC),
		scheduler:        scheduler,
		coverage:         coverageIdx,
		frameworkUC:      frameworkUC,
		evaluatorUC:      evaluatorUC,
		scheduleUC:       scheduleUC,
		tenantFrameworks: tenantFrameworkRepo,
		frameworkStore:   frameworkStore,
		entitlement:      entitlement,
		tenantLister:     tenants,
		leases:           leases,
		evalInterval:     time.Duration(evalHours) * time.Hour,
	}
}

func (m *Module) Start(ctx context.Context) {
	go m.scheduler.Start(ctx)
	go m.reloadCoverage(ctx)
	go m.evaluateLoop(ctx)
}

// evaluateLoop re-evaluates every enabled framework's compliance posture on a
// fixed interval (COMPLIANCE_EVAL_INTERVAL_HOURS, default 24h) and stores a fresh
// snapshot for each, so the UI always shows the latest "are we compliant now"
// state without the user triggering an evaluation. It runs an initial pass at
// startup so data is available promptly.
func (m *Module) evaluateLoop(ctx context.Context) {
	m.evaluateAll(ctx)
	ticker := time.NewTicker(m.evalInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.evaluateAll(ctx)
		}
	}
}

// evaluateAll generates (evaluates + snapshots) a report for every framework a
// tenant possesses, across every active tenant on the instance. A framework's
// file being globally `.disabled` at the platform level, or being enterprise-
// locked without the licence, skips it for every tenant. Per-item failures
// are logged and skipped so one bad tenant / framework doesn't stall the sweep.
//
// On-prem (no tenantLister) falls back to the legacy behaviour: one pass over
// every file-enabled framework under the module's own ctx.
func (m *Module) evaluateAll(ctx context.Context) {
	if m.tenantLister == nil {
		m.evaluateForCurrentContext(ctx)
		return
	}
	tenants, err := m.tenantLister(tenancy.WithAllTenants(ctx))
	if err != nil {
		_ = catcher.Error("compliance: listing tenants for scheduled evaluation failed", err, nil)
		return
	}
	for _, tid := range tenants {
		if ctx.Err() != nil {
			return
		}
		tenantCtx := authz.WithTenantID(ctx, tid)
		m.evaluateForCurrentContext(tenantCtx)
	}
}

// evaluateForCurrentContext runs one tenant's sweep — every framework in its
// possession that's not locked and whose file is still available.
func (m *Module) evaluateForCurrentContext(ctx context.Context) {
	keys, err := m.tenantFrameworks.List(ctx)
	if err != nil {
		_ = catcher.Error("compliance: listing possessed frameworks failed", err, map[string]any{"tenant": authz.TenantIDFromContext(ctx)})
		return
	}
	// Empty tenant on-prem — no possession rows. Fall back to the file-level
	// enabled state, matching the legacy loop.
	if len(keys) == 0 && authz.TenantIDFromContext(ctx) == "" {
		for _, fw := range m.frameworkStore.All() {
			if fw.Enabled && !m.entitlement.FrameworkLocked(&fw) {
				keys = append(keys, fw.Key)
			}
		}
	}
	tid := authz.TenantIDFromContext(ctx)
	for _, key := range keys {
		if ctx.Err() != nil {
			return
		}
		fw, ok := m.frameworkStore.Get(key)
		if !ok || !fw.Enabled || m.entitlement.FrameworkLocked(fw) {
			continue
		}
		if !m.claimEval(ctx, tid, key) {
			continue
		}
		if _, err := m.evaluatorUC.GenerateReport(ctx, key); err != nil {
			_ = catcher.Error("compliance: scheduled evaluation failed", err, map[string]any{"framework": key, "tenant": tid})
		}
	}
}

// claimEval tries to take the per-(tenant, framework) lease. A false return
// means a peer replica is doing this evaluation for this cycle — skip it. Any
// acquire error is treated as "not mine" (log + skip) so a joblease outage
// stops the sweep entirely rather than falling back to N-replica writes.
//
// TTL is deliberately much larger than one eval takes and much smaller than
// the tick interval, so a crash mid-run lets the next tick reclaim.
func (m *Module) claimEval(ctx context.Context, tenantID, frameworkKey string) bool {
	if m.leases == nil {
		return true
	}
	name := evalLeaseName(tenantID, frameworkKey)
	mine, err := m.leases.Acquire(ctx, name, evalLeaseTTL)
	if err != nil {
		_ = catcher.Error("compliance: acquire eval lease failed", err, map[string]any{"lease": name})
		return false
	}
	return mine
}

// evalLeaseName packs (tenant, framework) into a job_leases.name that fits
// the column's 64-char cap. Tenant is a UUID (36 chars with dashes); we drop
// dashes to save 4 chars, giving `c:e:<32>:<key>` = 5 + 32 + up-to-20 = up to
// 57 chars. Empty tenant (on-prem) becomes literal `-` so the prefix stays
// distinguishable from a real UUID.
func evalLeaseName(tenantID, frameworkKey string) string {
	tid := tenantID
	if tid == "" {
		tid = "-"
	}
	// strip dashes without allocating a scanner
	compact := make([]byte, 0, len(tid))
	for i := 0; i < len(tid); i++ {
		if tid[i] != '-' {
			compact = append(compact, tid[i])
		}
	}
	return "c:e:" + string(compact) + ":" + frameworkKey
}

const evalLeaseTTL = 15 * time.Minute

// reloadCoverage periodically rebuilds the rule coverage index so newly tagged
// or toggled correlation rules are reflected without a restart.
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

func (m *mailSender) SendComplianceReport(ctx context.Context, toEmail, subject string, pdfData []byte) error {
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
	return m.svc.SendMail(ctx, []string{toEmail}, nil, subject, body, []mail_domain.Attatchment{attachment})
}

// interface assertions
var _ connectors.MailSender = (*mailSender)(nil)
