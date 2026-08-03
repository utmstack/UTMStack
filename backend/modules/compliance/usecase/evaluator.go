package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type evaluator struct {
	controls   *ControlStore
	frameworks *FrameworkStore
	sql        connectors.OpenSearchSQL
	coverage   *CoverageIndex
	alerts     connectors.OpenSearchAlerts
	store      connectors.ReportStore
	overrides  connectors.ControlStatusOverrideRepository // optional; nil → no manual overrides
	notes      connectors.ControlNoteRepository           // optional; nil → no user notes on rows
	brand      connectors.BrandingProvider                // optional; nil → default UTMStack branding
	ent        *Entitlement
}

func NewEvaluator(controls *ControlStore, frameworks *FrameworkStore, sql connectors.OpenSearchSQL, coverage *CoverageIndex, alerts connectors.OpenSearchAlerts, store connectors.ReportStore, overrides connectors.ControlStatusOverrideRepository, notes connectors.ControlNoteRepository, brand connectors.BrandingProvider, ent *Entitlement) connectors.EvaluatorUsecase {
	return &evaluator{controls: controls, frameworks: frameworks, sql: sql, coverage: coverage, alerts: alerts, store: store, overrides: overrides, notes: notes, brand: brand, ent: ent}
}

func (e *evaluator) reportBrand(ctx context.Context) connectors.ReportBrand {
	if e.brand == nil {
		return connectors.ReportBrand{}
	}
	return e.brand.ReportBrand(ctx)
}

// GenerateReport evaluates the framework and persists a snapshot (history). The
// snapshot is the report DATA — the frontend renders it and the PDF path renders
// it; no binary PDF is stored.
func (e *evaluator) GenerateReport(ctx context.Context, key string) (*domain.Report, error) {
	rep, err := e.EvaluateFramework(ctx, key)
	if err != nil {
		return nil, err
	}
	snap := &domain.ReportSnapshot{
		ID:            uuid.NewString(),
		FrameworkKey:  rep.FrameworkKey,
		FrameworkName: rep.FrameworkName,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Score:         rep.Summary.CompliantPct,
		Report:        *rep,
	}
	if err := e.store.Save(ctx, snap); err != nil {
		// Don't fail the report if persistence fails — the live report is still useful.
		_ = catcher.Error("compliance: saving report snapshot failed", err, map[string]any{"framework": key})
	}
	return rep, nil
}

func (e *evaluator) ListReports(ctx context.Context, frameworkKey string, limit int) ([]domain.ReportSnapshotMeta, error) {
	metas, err := e.store.List(ctx, frameworkKey, limit)
	if err != nil {
		return nil, err
	}
	out := metas[:0]
	for _, m := range metas {
		if !e.frameworkKeyLocked(m.FrameworkKey) {
			out = append(out, m)
		}
	}
	return out, nil
}

func (e *evaluator) GetReport(ctx context.Context, id string) (*domain.ReportSnapshot, error) {
	snap, err := e.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if snap != nil && e.frameworkKeyLocked(snap.FrameworkKey) {
		return nil, domain.ErrFrameworkLocked
	}
	return snap, nil
}

func (e *evaluator) DeleteReport(ctx context.Context, id string) error {
	return e.store.Delete(ctx, id)
}

func (e *evaluator) frameworkKeyLocked(key string) bool {
	fw, ok := e.frameworks.Get(key)
	return ok && e.ent.FrameworkLocked(fw)
}

func (e *evaluator) FrameworkReportPDF(ctx context.Context, key string) ([]byte, string, error) {
	rep, err := e.EvaluateFramework(ctx, key)
	if err != nil {
		return nil, "", err
	}
	pdf, err := renderReportPDF(*rep, e.reportBrand(ctx))
	return pdf, rep.FrameworkName, err
}

func (e *evaluator) SnapshotPDF(ctx context.Context, id string) ([]byte, string, error) {
	snap, err := e.GetReport(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if snap == nil {
		return nil, "", domain.ErrReportNotFound
	}
	pdf, err := renderReportPDF(snap.Report, e.reportBrand(ctx))
	return pdf, snap.FrameworkName, err
}

func (e *evaluator) SetStatusOverride(ctx context.Context, frameworkKey, controlID, status, reason string) error {
	if e.overrides == nil {
		return fmt.Errorf("status overrides not configured")
	}
	if frameworkKey == "" || controlID == "" {
		return domain.ErrInvalidID
	}
	if _, ok := e.frameworks.Get(frameworkKey); !ok {
		return domain.ErrFrameworkNotFound
	}
	if !domain.ValidStatus(status) {
		return domain.ErrInvalidStatus
	}

	if err := e.notes.Upsert(ctx, &domain.UtmComplianceControlNote{
		FrameworkKey: frameworkKey,
		ControlID:    controlID,
		Note:         reason,
	}); err != nil {
		return err
	}

	return e.overrides.Upsert(ctx, &domain.UtmComplianceControlStatusOverride{
		FrameworkKey: frameworkKey,
		ControlID:    controlID,
		Status:       status,
		Reason:       reason,
	})
}

func (e *evaluator) ClearStatusOverride(ctx context.Context, frameworkKey, controlID string) error {
	if e.overrides == nil {
		return nil
	}
	if frameworkKey == "" || controlID == "" {
		return domain.ErrInvalidID
	}
	return e.overrides.Delete(ctx, frameworkKey, controlID)
}

func (e *evaluator) SetControlNote(ctx context.Context, frameworkKey, controlID, note string) error {
	if e.notes == nil {
		return fmt.Errorf("control notes not configured")
	}
	if frameworkKey == "" || controlID == "" {
		return domain.ErrInvalidID
	}
	if _, ok := e.frameworks.Get(frameworkKey); !ok {
		return domain.ErrFrameworkNotFound
	}
	// Empty / whitespace-only note is a delete — keep the table tidy.
	if strings.TrimSpace(note) == "" {
		return e.notes.Delete(ctx, frameworkKey, controlID)
	}
	return e.notes.Upsert(ctx, &domain.UtmComplianceControlNote{
		FrameworkKey: frameworkKey,
		ControlID:    controlID,
		Note:         note,
	})
}

func (e *evaluator) ClearControlNote(ctx context.Context, frameworkKey, controlID string) error {
	if e.notes == nil {
		return nil
	}
	if frameworkKey == "" || controlID == "" {
		return domain.ErrInvalidID
	}
	return e.notes.Delete(ctx, frameworkKey, controlID)
}

// activityWindow is how far back the activity dimension counts alerts.
const activityWindow = -30 * 24 * time.Hour

// EvaluateFramework resolves the framework's requirements to controls, evaluates
// each control once (cached), and assembles the report with a rollup.
func (e *evaluator) EvaluateFramework(ctx context.Context, key string) (*domain.Report, error) {
	fw, ok := e.frameworks.Get(key)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	if e.ent.FrameworkLocked(fw) {
		return nil, domain.ErrFrameworkLocked
	}

	now := time.Now().UTC()
	report := domain.Report{
		FrameworkKey:  fw.Key,
		FrameworkName: fw.Name,
		GeneratedAt:   now.Format("2006-01-02 15:04 UTC"),
	}
	since := now.Add(activityWindow).Format(time.RFC3339)
	cache := map[string]domain.ReportControlRow{}

	// Load manual overrides once per evaluation; missing / errored → treat as no overrides.
	var overrides map[string]string
	if e.overrides != nil {
		if m, err := e.overrides.ListByFramework(ctx, fw.Key); err == nil {
			overrides = m
		} else {
			_ = catcher.Error("compliance: loading status overrides failed", err, map[string]any{"framework": fw.Key})
		}
	}
	// Notes are surfaced on report rows; errored → treat as no notes.
	var notes map[string]string
	if e.notes != nil {
		if m, err := e.notes.ListByFramework(ctx, fw.Key); err == nil {
			notes = m
		} else {
			_ = catcher.Error("compliance: loading control notes failed", err, map[string]any{"framework": fw.Key})
		}
	}

	for _, sec := range fw.Sections {
		rs := domain.ReportSection{Name: sec.Name}
		seen := map[string]bool{}
		for _, req := range sec.Requirements {
			for _, cid := range req.SatisfiedBy {
				if seen[cid] {
					continue
				}
				seen[cid] = true
				row, cached := cache[cid]
				if !cached {
					row = e.evalControl(ctx, cid, since)
					if s, ok := overrides[cid]; ok && domain.ValidStatus(s) && s != row.Status {
						row.Evidence = fmt.Sprintf("manual override (was %s)", row.Status)
						row.Status = s
						row.Overridden = true
					}
					if n, ok := notes[cid]; ok {
						row.Note = n
					}
					cache[cid] = row
				}
				rs.Controls = append(rs.Controls, row)
				tally(&report.Summary, row.Status)
			}
		}
		report.Sections = append(report.Sections, rs)
	}

	finalizeSummary(&report.Summary)
	return &report, nil
}

// evalControl evaluates one control across the three dimensions: ① coverage (how
// many enabled rules tag it), ② activity (alerts from those rules in the window),
// ③ analysis (its checks). Governance → out of scope; analysis is authoritative
// when present; otherwise coverage decides; otherwise pending.
func (e *evaluator) evalControl(ctx context.Context, id, since string) domain.ReportControlRow {
	row := domain.ReportControlRow{ControlID: id, Name: id}
	c, ok := e.controls.Get(id)
	if !ok {
		row.Status = domain.StatusNotCovered
		row.Evidence = "control not found in library"
		return row
	}
	row.Name = c.Name

	if c.EffectiveScope() == domain.ScopeGovernance {
		row.Status = domain.StatusOutOfScope
		row.Evidence = "governance control — not evaluated from logs"
		return row
	}

	// ① coverage + ② activity.
	rules := e.coverage.Rules(id)
	row.Coverage = len(rules)
	if len(rules) > 0 {
		if n, err := e.alerts.CountByRuleNames(ctx, rules, since); err == nil {
			row.Activity = int(n)
		}
	}

	// ③ analysis — authoritative when the control has runnable checks.
	checks := c.RunnableChecks()
	if len(checks) > 0 {
		passed, failed := 0, 0
		var firstFail string
		for _, ch := range checks {
			hits, err := e.sql.RunCheck(ctx, ch.SQL)
			pass := false
			detail := ""
			if err != nil {
				detail = "error: " + err.Error()
			} else {
				pass = passesRule(ch, hits)
				detail = fmt.Sprintf("hits=%d %s", hits, ch.Rule)
			}
			if pass {
				passed++
			} else {
				failed++
				if firstFail == "" {
					firstFail = ch.Name + " (" + detail + ")"
				}
			}
		}
		compliant := failed == 0
		if c.EffectiveStrategy() == domain.StrategyAny {
			compliant = passed > 0
		}
		if compliant {
			row.Status = domain.StatusCompliant
			row.Evidence = fmt.Sprintf("%d/%d checks passed", passed, len(checks))
		} else {
			row.Status = domain.StatusNonCompliant
			if firstFail != "" {
				row.Evidence = firstFail
			} else {
				row.Evidence = fmt.Sprintf("%d/%d checks failed", failed, len(checks))
			}
		}
		return row
	}

	// No analysis → fall back to coverage.
	if len(rules) > 0 {
		row.Status = domain.StatusCompliant
		row.Evidence = fmt.Sprintf("covered by %d rule(s), %d alerts in window", len(rules), row.Activity)
		return row
	}

	row.Status = domain.StatusPending
	row.Evidence = "no check or rule coverage"
	return row
}

// passesRule applies a check's rule to the hit count.
func passesRule(ch domain.Check, hits int64) bool {
	rv := int64(1)
	if ch.RuleValue != nil {
		rv = int64(*ch.RuleValue)
	}
	switch ch.Rule {
	case domain.RuleNoHitsAllowed:
		return hits == 0
	case domain.RuleMinHitsRequired, domain.RuleMatchFieldValue:
		return hits >= rv
	case domain.RuleThresholdMax:
		return hits <= rv
	default:
		return false
	}
}

func tally(s *domain.ReportSummary, status string) {
	s.Total++
	switch status {
	case domain.StatusCompliant:
		s.Compliant++
	case domain.StatusNonCompliant:
		s.NonCompliant++
	case domain.StatusAtRisk:
		s.AtRisk++
	case domain.StatusNotCovered:
		s.NotCovered++
	case domain.StatusOutOfScope:
		s.OutOfScope++
	case domain.StatusPending:
		s.Pending++
	}
}

func finalizeSummary(s *domain.ReportSummary) {
	evaluated := s.Compliant + s.NonCompliant + s.AtRisk + s.Pending
	if evaluated > 0 {
		s.CompliantPct = (s.Compliant  / evaluated) * 100
	}
}
