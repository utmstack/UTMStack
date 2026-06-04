package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

const (
	schedulerInterval          = 30 * time.Second
	schedulerInitialDelay      = 10 * time.Second
	schedulerMaxAlertsPerBatch = 1000
)

type Scheduler struct {
	alertRepo connectors.AlertRepository
	ruleRepo  connectors.AlertTagRuleRepository
	tagRepo   connectors.AlertTagRepository
	histRepo  connectors.HistoryRecorder
}

func NewScheduler(
	alertRepo connectors.AlertRepository,
	ruleRepo connectors.AlertTagRuleRepository,
	tagRepo connectors.AlertTagRepository,
	histRepo connectors.HistoryRecorder,
) *Scheduler {
	return &Scheduler{
		alertRepo: alertRepo,
		ruleRepo:  ruleRepo,
		tagRepo:   tagRepo,
		histRepo:  histRepo,
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	logger.Info("alerts scheduler: starting (initial delay 10s)")

	select {
	case <-time.After(schedulerInitialDelay):
	case <-ctx.Done():
		logger.Info("alerts scheduler: cancelled during initial delay — stopped")
		return
	}

	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	logger.Info("alerts scheduler: running")

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			logger.Info("alerts scheduler: context cancelled — stopped")
			return
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	count, err := s.alertRepo.CountByStatus(ctx, int(domain.AlertStatusAutomaticReview))
	if err != nil {
		logger.Error("alerts scheduler: CountByStatus error: " + err.Error())
		return
	}
	if count == 0 {
		return
	}

	// TODO(module-33): wire real mapping when network_scan is ported.
	if err := s.alertRepo.AssignAssetGroups(ctx, nil); err != nil {
		logger.Error("alerts scheduler: AssignAssetGroups error: " + err.Error())
	}

	rules, err := s.ruleRepo.FindAllActive(ctx)
	if err != nil {
		logger.Error("alerts scheduler: FindAllActive error: " + err.Error())
		return
	}

	evaluationStart := time.Now().UTC().Truncate(time.Second)

	for _, rule := range rules {
		tagNames, resolveErr := s.resolveTagNames(ctx, rule.RuleAppliedTags)
		if resolveErr != nil {
			logger.Error("alerts scheduler: resolveTagNames error for rule " + rule.RuleName + ": " + resolveErr.Error())
			continue
		}

		oldAlerts, searchErr := s.alertRepo.SearchByRuleConditionsAndTimestamp(
			ctx, rule, evaluationStart, schedulerMaxAlertsPerBatch)
		if searchErr != nil {
			logger.Error("alerts scheduler: SearchByRuleConditionsAndTimestamp error: " + searchErr.Error())
		}

		if applyErr := s.alertRepo.ApplyTagRule(ctx, rule, tagNames, evaluationStart); applyErr != nil {
			logger.Error("alerts scheduler: ApplyTagRule error for rule " + rule.RuleName + ": " + applyErr.Error())
			continue
		}

		if len(oldAlerts) > 0 && s.histRepo != nil {
			s.logAutomaticTagChange(ctx, oldAlerts, tagNames, rule.RuleName)
		}
	}

	oldBeforeRelease, searchErr := s.alertRepo.SearchByStatusAndTimestamp(
		ctx, int(domain.AlertStatusAutomaticReview), evaluationStart, schedulerMaxAlertsPerBatch)
	if searchErr != nil {
		logger.Error("alerts scheduler: SearchByStatusAndTimestamp (pre-release) error: " + searchErr.Error())
	}

	if err := s.alertRepo.ReleaseToOpen(ctx, evaluationStart); err != nil {
		logger.Error("alerts scheduler: ReleaseToOpen error: " + err.Error())
	}

	if len(oldBeforeRelease) > 0 && s.histRepo != nil {
		s.logAutomaticStatusChange(ctx, oldBeforeRelease, evaluationStart)
	}
}

func (s *Scheduler) logAutomaticTagChange(ctx context.Context, alerts []domain.UtmAlert, tags []string, ruleName string) {
	tagsStr := strings.Join(tags, ",") // Java uses "," separator — MINOR-2 fix
	msg := fmt.Sprintf(domain.MsgTagsAutomatic, tagsStr, ruleName)

	entries := make([]connectors.HistoryEntry, 0, len(alerts))
	for _, a := range alerts {
		entries = append(entries, connectors.HistoryEntry{
			AlertID:  a.ID,
			User:     "system",
			Action:   domain.ActionUpdateTags,
			Message:  msg,
			NewValue: fmt.Sprintf(`{"tags":%q}`, tagsStr),
			At:       time.Now().UTC(),
		})
	}
	if err := s.histRepo.Record(ctx, entries); err != nil {
		logger.Error("alerts scheduler: logAutomaticTagChange error: " + err.Error())
	}
}

func (s *Scheduler) logAutomaticStatusChange(ctx context.Context, alerts []domain.UtmAlert, evaluationStart time.Time) {
	oldLabel := domain.StatusName(domain.AlertStatusAutomaticReview)
	newLabel := domain.StatusName(domain.AlertStatusOpen)
	obs := "This alert has been evaluated by the tag rules engine"
	msg := fmt.Sprintf(domain.MsgStatusSystem, oldLabel, newLabel, obs)

	now := time.Now().UTC()
	entries := make([]connectors.HistoryEntry, 0, len(alerts))
	for _, a := range alerts {
		newVal, _ := json.Marshal(map[string]any{
			"status":            int(domain.AlertStatusOpen),
			"statusLabel":       newLabel,
			"statusObservation": obs,
		})
		entries = append(entries, connectors.HistoryEntry{
			AlertID:  a.ID,
			User:     "system",
			Action:   domain.ActionUpdateStatus,
			Message:  msg,
			NewValue: string(newVal),
			At:       now,
		})
	}
	if err := s.histRepo.Record(ctx, entries); err != nil {
		logger.Error("alerts scheduler: logAutomaticStatusChange error: " + err.Error())
	}
}

func (s *Scheduler) resolveTagNames(ctx context.Context, csv string) ([]string, error) {
	ids := csvToInt64Slice(csv)
	if len(ids) == 0 {
		return nil, nil
	}

	tags, err := s.tagRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(tags))
	for _, t := range tags {
		if strings.TrimSpace(t.TagName) != "" {
			names = append(names, t.TagName)
		}
	}
	return names, nil
}
