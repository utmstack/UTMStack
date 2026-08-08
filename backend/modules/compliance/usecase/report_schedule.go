package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type scheduleUsecase struct {
	repo       connectors.ScheduleRepository
	frameworks *FrameworkStore
	ent        *Entitlement
}

func NewScheduleUsecase(repo connectors.ScheduleRepository, frameworks *FrameworkStore, ent *Entitlement) connectors.ScheduleUsecase {
	return &scheduleUsecase{repo: repo, frameworks: frameworks, ent: ent}
}

func (u *scheduleUsecase) frameworkLocked(ctx context.Context, key string) bool {
	fw, ok := u.frameworks.Get(ctx, key)
	return ok && u.ent.FrameworkLocked(fw, u.frameworks.IsSystem(key))
}

func (u *scheduleUsecase) Create(ctx context.Context, userID uuid.UUID, req dto.CreateScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.ScheduleString); err != nil {
		return nil, domain.ErrInvalidCron
	}
	if u.frameworkLocked(ctx, req.FrameworkKey) {
		return nil, domain.ErrFrameworkLocked
	}
	now := time.Now().UTC()
	s := &domain.ReportSchedule{
		UserID:            userID,
		FrameworkKey:      req.FrameworkKey,
		ScheduleString:    req.ScheduleString,
		WindowDays:        windowDaysOrDefault(req.WindowDays),
		To:                req.To,
		Cc:                req.Cc,
		LastExecutionDate: now,
	}

	next, err := nextCronTime(s.ScheduleString, now)
	if err != nil {
		return nil, domain.ErrInvalidCron
	}
	s.NextExecutionDate = next
	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return toScheduleResp(s), nil
}

func (u *scheduleUsecase) Update(ctx context.Context, userID uuid.UUID, req dto.UpdateScheduleRequest) (*dto.ScheduleResponse, error) {
	if err := validateCron(req.ScheduleString); err != nil {
		return nil, domain.ErrInvalidCron
	}
	if u.frameworkLocked(ctx, req.FrameworkKey) {
		return nil, domain.ErrFrameworkLocked
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
	existing.WindowDays = windowDaysOrDefault(req.WindowDays)
	existing.To = req.To
	existing.Cc = req.Cc
	next, err := nextCronTime(existing.ScheduleString, time.Now().UTC())
	if err != nil {
		return nil, domain.ErrInvalidCron
	}
	existing.NextExecutionDate = next
	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return toScheduleResp(existing), nil
}

func (u *scheduleUsecase) GetByID(ctx context.Context, id uuid.UUID) (*dto.ScheduleResponse, error) {
	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, domain.ErrScheduleNotFound
	}
	return toScheduleResp(s), nil
}

func (u *scheduleUsecase) ListByUser(ctx context.Context, userID uuid.UUID, f dto.ScheduleFilters) ([]dto.ScheduleResponse, int64, error) {
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

func (u *scheduleUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if s == nil {
		return domain.ErrScheduleNotFound
	}
	return u.repo.Delete(ctx, id)
}

func toScheduleResp(s *domain.ReportSchedule) *dto.ScheduleResponse {
	return &dto.ScheduleResponse{
		ID:                s.ID,
		UserID:            s.UserID,
		FrameworkKey:      s.FrameworkKey,
		ScheduleString:    s.ScheduleString,
		WindowDays:        s.WindowDays,
		To:                s.To,
		Cc:                s.Cc,
		LastExecutionDate: s.LastExecutionDate,
	}
}

func windowDaysOrDefault(d int) int {
	if d <= 0 {
		return domain.DefaultWindowDays
	}
	return d
}

func splitEmails(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

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
	ticker := time.NewTicker(time.Minute)
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

	now := time.Now().UTC()
	schedules, err := s.scheduleRepo.ListDue(tenancy.WithAllTenants(ctx), now)
	if err != nil {
		_ = catcher.Error("compliance: listing due schedules failed", err, nil)
		return
	}
	for _, sched := range schedules {
		next, err := nextCronTime(sched.ScheduleString, sched.NextExecutionDate)
		if err != nil {
			_ = catcher.Error("compliance: unusable cron on schedule", err, map[string]any{"scheduleId": sched.ID})
			continue
		}

		tenantCtx := authz.WithTenantID(ctx, sched.TenantID.String())
		claimed, err := s.scheduleRepo.ClaimDue(tenantCtx, sched.ID, sched.NextExecutionDate, sched.NextExecutionDate, next)
		if err != nil {
			_ = catcher.Error("compliance: claiming schedule failed", err, map[string]any{"scheduleId": sched.ID})
			continue
		}
		if !claimed {
			continue
		}
		s.deliver(tenantCtx, &sched, next)
	}
}

func (s *ReportScheduler) deliver(ctx context.Context, sched *domain.ReportSchedule, next time.Time) {
	if _, err := s.evaluator.Evaluate(ctx, sched.FrameworkKey, sched.WindowDays); err != nil {
		_ = catcher.Error("compliance: scheduled evaluation failed", err, map[string]any{"scheduleId": sched.ID, "framework": sched.FrameworkKey})
		return
	}
	pdf, _, err := s.evaluator.PDF(ctx, sched.FrameworkKey, "")
	if err != nil {
		_ = catcher.Error("compliance: PDF generation failed", err, map[string]any{"scheduleId": sched.ID, "framework": sched.FrameworkKey})
		return
	}
	to, cc := splitEmails(sched.To), splitEmails(sched.Cc)
	if len(to) == 0 {
		return
	}
	subject := fmt.Sprintf("UTMStack Compliance Report - %s - %s", sched.FrameworkKey, time.Now().UTC().Format("2006-01-02"))
	if err := s.mail.SendComplianceReport(ctx, to, cc, subject, pdf); err != nil {
		_ = catcher.Error("compliance: email delivery failed", err, map[string]any{"scheduleId": sched.ID})
	}
}

func validateCron(expr string) error {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return fmt.Errorf("cron must have 5 fields, got %d", len(fields))
	}
	return nil
}

func nextCronTime(expr string, last time.Time) (time.Time, error) {
	if err := validateCron(expr); err != nil {
		return time.Time{}, err
	}
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
	if strings.HasPrefix(field, "*/") {
		var step int
		if _, err := fmt.Sscanf(field[2:], "%d", &step); err == nil && step > 0 {
			return value%step == 0
		}
	}
	for _, part := range strings.Split(field, ",") {
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
