package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type evaluator struct {
	controls   *ControlStore
	frameworks *FrameworkStore
	coverage   *CoverageIndex
	events     connectors.EventCounter
	store      connectors.ReportStore
	scores     connectors.ReportScoreStore
	schedules  connectors.ScheduleRepository
	brand      connectors.BrandingProvider // optional; nil → default UTMStack branding
	ent        *Entitlement
}

func NewEvaluator(
	controls *ControlStore,
	frameworks *FrameworkStore,
	coverage *CoverageIndex,
	events connectors.EventCounter,
	store connectors.ReportStore,
	scores connectors.ReportScoreStore,
	schedules connectors.ScheduleRepository,
	brand connectors.BrandingProvider,
	ent *Entitlement,
) connectors.EvaluatorUsecase {
	return &evaluator{
		controls: controls, frameworks: frameworks, coverage: coverage,
		events: events, store: store, scores: scores, schedules: schedules,
		brand: brand, ent: ent,
	}
}

func (e *evaluator) brandFor(ctx context.Context, preparedBy string) connectors.ReportBrand {
	var b connectors.ReportBrand
	if e.brand != nil {
		b = e.brand.ReportBrand(ctx)
	}
	b.PreparedBy = preparedBy
	return b
}

// Evaluate runs the framework and replaces the standing report.
func (e *evaluator) Evaluate(ctx context.Context, frameworkKey string, windowDays int) (*dto.ReportResponse, error) {
	fw, ok := e.frameworks.Get(ctx, frameworkKey)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	if e.ent.FrameworkLocked(fw, e.frameworks.IsSystem(fw.Key)) {
		return nil, domain.ErrFrameworkLocked
	}

	now := time.Now().UTC()
	from, to := windowOf(now, e.resolveWindow(ctx, frameworkKey, windowDays))

	// The previous report is read before anything is computed: its human edits
	// have to survive this run, and a nightly schedule would otherwise erase
	// every justification written the day before.
	prev, prevBody := e.previous(ctx, frameworkKey)

	ids := controlIDsOf(fw)
	activity, err := e.events.CountByRuleNames(ctx, e.ruleNames(ids), from, to)
	if err != nil {
		_ = catcher.Error("compliance: activity query failed", err, map[string]any{"framework": frameworkKey})
		activity = map[string]int64{}
	}
	results := e.runChecks(ctx, e.collectChecks(ctx, ids), from, to)

	body := dto.ReportBody{Controls: make([]dto.ControlRow, 0, len(ids))}
	for _, id := range ids {
		row := e.controlRow(ctx, id, results, activity)
		inheritEdit(&row, prevBody.Controls)
		body.Controls = append(body.Controls, row)
	}
	body.Sections = sectionsOf(fw, body.Controls)
	recompute(&body)

	rep := &domain.Report{
		FrameworkKey:    fw.Key,
		FrameworkName:   fw.Name,
		FrameworkSource: fw.Source,
		GeneratedAt:     now,
		WindowFrom:      from,
		WindowTo:        to,
	}
	if prev != nil {
		rep.ID, rep.Version = prev.ID, prev.Version
	}
	if err := e.save(ctx, rep, &body); err != nil {
		return nil, err
	}
	e.recordPoint(ctx, rep, &body)

	return response(rep, &body), nil
}

// resolveWindow: what the caller asked, else the framework's schedule, else the
// default. Borrowing the schedule's window is what keeps a manual run and the
// mailed report describing the same period.
func (e *evaluator) resolveWindow(ctx context.Context, frameworkKey string, windowDays int) int {
	if windowDays > 0 {
		return windowDays
	}
	if e.schedules != nil {
		if s, err := e.schedules.ForFramework(ctx, frameworkKey); err == nil && s != nil && s.WindowDays > 0 {
			return s.WindowDays
		}
	}
	return domain.DefaultWindowDays
}

func (e *evaluator) previous(ctx context.Context, frameworkKey string) (*domain.Report, dto.ReportBody) {
	var body dto.ReportBody
	rep, err := e.store.Get(ctx, frameworkKey)
	if err != nil {
		if !errors.Is(err, domain.ErrReportNotFound) {
			_ = catcher.Error("compliance: reading the previous report failed", err, map[string]any{"framework": frameworkKey})
		}
		return nil, body
	}
	if len(rep.Body) > 0 {
		if err := json.Unmarshal(rep.Body, &body); err != nil {
			_ = catcher.Error("compliance: previous report body is unreadable", err, map[string]any{"framework": frameworkKey})
		}
	}
	return rep, body
}

func (e *evaluator) save(ctx context.Context, rep *domain.Report, body *dto.ReportBody) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	rep.Body = raw
	rep.Score = body.Summary.CompliantPct
	return e.store.Save(ctx, rep)
}

// recordPoint writes the day's point on the chart. A failure here is logged
// rather than returned: the report is the deliverable, the chart is a
// by-product, and losing one point must not lose the run.
func (e *evaluator) recordPoint(ctx context.Context, rep *domain.Report, body *dto.ReportBody) {
	if e.scores == nil {
		return
	}
	err := e.scores.Upsert(ctx, &domain.ReportScore{
		FrameworkKey: rep.FrameworkKey,
		Day:          rep.GeneratedAt,
		GeneratedAt:  rep.GeneratedAt,
		Score:        body.Summary.CompliantPct,
		Total:        body.Summary.Total,
		Evaluated:    body.Summary.Evaluated,
		Compliant:    body.Summary.Compliant,
		Body:         rep.Body,
	})
	if err != nil {
		_ = catcher.Error("compliance: recording the score point failed", err, map[string]any{"framework": rep.FrameworkKey})
	}
}

func (e *evaluator) Get(ctx context.Context, frameworkKey string) (*dto.ReportResponse, error) {
	rep, err := e.store.Get(ctx, frameworkKey)
	if err != nil {
		return nil, err
	}
	var body dto.ReportBody
	if len(rep.Body) > 0 {
		if err := json.Unmarshal(rep.Body, &body); err != nil {
			return nil, err
		}
	}
	return response(rep, &body), nil
}

func (e *evaluator) List(ctx context.Context) ([]dto.ReportMeta, error) {
	rows, err := e.store.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ReportMeta, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.ReportMeta{
			ID:            r.ID,
			FrameworkKey:  r.FrameworkKey,
			FrameworkName: r.FrameworkName,
			GeneratedAt:   r.GeneratedAt,
			Score:         r.Score,
		})
	}
	return out, nil
}

func (e *evaluator) Delete(ctx context.Context, frameworkKey string) error {
	return e.store.Delete(ctx, frameworkKey)
}

func (e *evaluator) EditControl(ctx context.Context, editedBy, frameworkKey, controlID string, req dto.EditControlRequest) (*dto.ReportResponse, error) {
	// An empty status means the caller is annotating, not overriding.
	if req.Status != "" && !validStatus(req.Status) {
		return nil, domain.ErrInvalidStatus
	}
	rep, err := e.store.Get(ctx, frameworkKey)
	if err != nil {
		return nil, err
	}
	var body dto.ReportBody
	if len(rep.Body) > 0 {
		if err := json.Unmarshal(rep.Body, &body); err != nil {
			return nil, err
		}
	}

	idx := -1
	for i := range body.Controls {
		if body.Controls[i].ControlID == controlID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, domain.ErrControlNotFound
	}

	now := time.Now().UTC()
	row := &body.Controls[idx]
	row.Note = req.Note
	row.EditedBy = editedBy
	row.EditedAt = &now

	if req.Status == "" || req.Status == row.EngineStatus {
		// A note that leaves the verdict alone is an annotation. Recording it
		// as an override would put a verdict on the row nobody gave, and would
		// have it go stale the day the engine changes its mind.
		row.Status = row.EngineStatus
		row.OriginalStatus = ""
	} else {
		// OriginalStatus is what the engine said at the moment of the override,
		// which is what lets a later run notice the ground moved underneath it.
		row.OriginalStatus = row.EngineStatus
		row.Status = req.Status
	}

	recompute(&body)
	if err := e.save(ctx, rep, &body); err != nil {
		return nil, err
	}
	return response(rep, &body), nil
}

func response(rep *domain.Report, body *dto.ReportBody) *dto.ReportResponse {
	return &dto.ReportResponse{
		ID:              rep.ID,
		FrameworkKey:    rep.FrameworkKey,
		FrameworkName:   rep.FrameworkName,
		FrameworkSource: rep.FrameworkSource,
		GeneratedAt:     rep.GeneratedAt,
		WindowFrom:      rep.WindowFrom,
		WindowTo:        rep.WindowTo,
		Summary:         body.Summary,
		Sections:        body.Sections,
		Controls:        body.Controls,
	}
}

func (e *evaluator) History(ctx context.Context, frameworkKey string, from, to time.Time) ([]dto.ScorePoint, error) {
	rows, err := e.scores.History(ctx, frameworkKey, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ScorePoint, 0, len(rows))
	for _, r := range rows {
		out = append(out, dto.ScorePoint{
			Day:         r.Day,
			GeneratedAt: r.GeneratedAt,
			Score:       r.Score,
			Total:       r.Total,
			Evaluated:   r.Evaluated,
			Compliant:   r.Compliant,
			HasDocument: r.HasBody,
		})
	}
	return out, nil
}

// PDF renders the standing report.
func (e *evaluator) PDF(ctx context.Context, frameworkKey, preparedBy string) ([]byte, string, error) {
	rep, err := e.Get(ctx, frameworkKey)
	if err != nil {
		return nil, "", err
	}
	pdf, err := renderReportPDF(*rep, e.brandFor(ctx, preparedBy))
	return pdf, rep.FrameworkName, err
}

func (e *evaluator) HistoryPDF(ctx context.Context, frameworkKey, preparedBy string, day time.Time) ([]byte, string, error) {
	raw, err := e.scores.Body(ctx, frameworkKey, day)
	if err != nil {
		return nil, "", err
	}
	if len(raw) == 0 {
		return nil, "", domain.ErrReportNotFound
	}
	var body dto.ReportBody
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, "", err
	}
	rep := dto.ReportResponse{
		FrameworkKey: frameworkKey,
		GeneratedAt:  day,
		Summary:      body.Summary,
		Sections:     body.Sections,
		Controls:     body.Controls,
	}
	if fw, ok := e.frameworks.Get(ctx, frameworkKey); ok {
		rep.FrameworkName, rep.FrameworkSource = fw.Name, fw.Source
	}
	pdf, err := renderReportPDF(rep, e.brandFor(ctx, preparedBy))
	return pdf, rep.FrameworkName, err
}
