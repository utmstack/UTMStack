package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type scheduleUsecase struct{ repo connectors.ScheduleRepository }

func NewScheduleUsecase(repo connectors.ScheduleRepository) connectors.ScheduleUsecase {
	return &scheduleUsecase{repo: repo}
}

func (u *scheduleUsecase) Create(ctx context.Context, userID int64, req dto.CreateScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.ScheduleString); err != nil {
		return nil, domain.ErrInvalidCron
	}
	s := &domain.UtmComplianceReportSchedule{
		UserID:            userID,
		FrameworkKey:      req.FrameworkKey,
		ScheduleString:    req.ScheduleString,
		Recipients:        req.Recipients,
		LastExecutionDate: time.Now().UTC(),
	}
	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return toScheduleResp(s), nil
}

func (u *scheduleUsecase) Update(ctx context.Context, userID int64, req dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.ScheduleString); err != nil {
		return nil, domain.ErrInvalidCron
	}
	existing, err := u.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrScheduleNotFound
	}
	existing.FrameworkKey = req.FrameworkKey
	existing.ScheduleString = req.ScheduleString
	existing.Recipients = req.Recipients
	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return toScheduleResp(existing), nil
}

func (u *scheduleUsecase) GetByID(ctx context.Context, id int64) (*dto.ScheduleResponse, error) {
	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.ErrScheduleNotFound
	}
	return toScheduleResp(s), nil
}

func (u *scheduleUsecase) ListByUser(ctx context.Context, userID int64, f dto.ScheduleFilters) ([]dto.ScheduleResponse, int64, error) {
	items, total, err := u.repo.ListByUser(ctx, userID, f)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.ScheduleResponse, len(items))
	for i, s := range items {
		out[i] = *toScheduleResp(&s)
	}
	return out, total, nil
}

func (u *scheduleUsecase) Delete(ctx context.Context, id int64) error {
	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return domain.ErrScheduleNotFound
	}
	return u.repo.Delete(ctx, id)
}

func toScheduleResp(s *domain.UtmComplianceReportSchedule) *dto.ScheduleResponse {
	return &dto.ScheduleResponse{
		ID:                s.ID,
		UserID:            s.UserID,
		FrameworkKey:      s.FrameworkKey,
		ScheduleString:    s.ScheduleString,
		Recipients:        s.Recipients,
		LastExecutionDate: s.LastExecutionDate,
	}
}

// ── Report scheduler background job ──────────────────────────────────────────

// ReportScheduler polls DB every 5s and fires PDF+email delivery for due schedules.
type ReportScheduler struct {
	scheduleRepo connectors.ScheduleRepository
	evaluator    connectors.EvaluatorUsecase
	mail         connectors.MailSender
}

func NewReportScheduler(
	scheduleRepo connectors.ScheduleRepository,
	evaluator connectors.EvaluatorUsecase,
	mail connectors.MailSender,
) *ReportScheduler {
	return &ReportScheduler{
		scheduleRepo: scheduleRepo,
		evaluator:    evaluator,
		mail:         mail,
	}
}

func (s *ReportScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.run(ctx)
		}
	}
}

func (s *ReportScheduler) run(ctx context.Context) {
	schedules, err := s.scheduleRepo.ListAll(ctx)
	if err != nil {
		_ = catcher.Error("compliance: listing schedules failed", err, nil)
		return
	}
	for _, sched := range schedules {
		next, err := nextCronTime(sched.ScheduleString, sched.LastExecutionDate)
		if err != nil {
			continue
		}
		if time.Now().UTC().Before(next) {
			continue
		}
		s.deliver(ctx, &sched, next)
	}
}

func (s *ReportScheduler) deliver(ctx context.Context, sched *domain.UtmComplianceReportSchedule, next time.Time) {
	report, err := s.evaluator.EvaluateFramework(ctx, sched.FrameworkKey)
	if err != nil {
		_ = catcher.Error("compliance: framework evaluation failed", err, map[string]any{"scheduleId": sched.ID, "framework": sched.FrameworkKey})
		return
	}
	pdf, err := renderReportPDF(*report)
	if err != nil {
		_ = catcher.Error("compliance: PDF generation failed", err, map[string]any{"scheduleId": sched.ID})
		return
	}
	subject := fmt.Sprintf("UTMStack Compliance Report - %s - %s", sched.FrameworkKey, time.Now().UTC().Format("2006-01-02"))
	// Email each recipient (comma-separated). The mail sender is a no-op until the
	// platform mail gateway is wired.
	for _, to := range strings.Split(sched.Recipients, ",") {
		to = strings.TrimSpace(to)
		if to == "" {
			continue
		}
		if err := s.mail.SendComplianceReport(ctx, to, subject, pdf); err != nil {
			_ = catcher.Error("compliance: email delivery failed", err, map[string]any{"scheduleId": sched.ID, "to": to})
		}
	}
	sched.LastExecutionDate = next
	if err := s.scheduleRepo.Update(ctx, sched); err != nil {
		_ = catcher.Error("compliance: updating last execution date failed", err, map[string]any{"scheduleId": sched.ID})
	}
}

// validateCron does a basic 5-field UNIX cron validation.
func validateCron(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return fmt.Errorf("cron must have 5 fields, got %d", len(fields))
	}
	return nil
}

// nextCronTime computes the next execution time after `last` for a 5-field cron.
// Implementation uses a simple minute-step search (mirrors Java CronExpression semantics).
func nextCronTime(expr string, last time.Time) (time.Time, error) {
	if err := validateCron(expr); err != nil {
		return time.Time{}, err
	}
	// Advance by 1 minute steps up to 1 year.
	t := last.Add(time.Minute).Truncate(time.Minute)
	for i := 0; i < 525960; i++ {
		if cronMatches(expr, t) {
			return t, nil
		}
		t = t.Add(time.Minute)
	}
	return time.Time{}, fmt.Errorf("no next time found for cron %q", expr)
}

func cronMatches(expr string, t time.Time) bool {
	fields := strings.Fields(expr)
	return matchField(fields[0], t.Minute()) &&
		matchField(fields[1], t.Hour()) &&
		matchField(fields[2], t.Day()) &&
		matchField(fields[3], int(t.Month())) &&
		matchField(fields[4], int(t.Weekday()))
}

func matchField(field string, value int) bool {
	if field == "*" {
		return true
	}
	// handle */n step
	if strings.HasPrefix(field, "*/") {
		var step int
		if _, err := fmt.Sscanf(field[2:], "%d", &step); err == nil && step > 0 {
			return value%step == 0
		}
	}
	// handle list a,b,c
	for _, part := range strings.Split(field, ",") {
		// handle range a-b
		if strings.Contains(part, "-") {
			var lo, hi int
			if _, err := fmt.Sscanf(part, "%d-%d", &lo, &hi); err == nil {
				if value >= lo && value <= hi {
					return true
				}
			}
		} else {
			var v int
			if _, err := fmt.Sscanf(part, "%d", &v); err == nil && v == value {
				return true
			}
		}
	}
	return false
}
