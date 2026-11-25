package main

import (
	"context"

	alerts_connectors "github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	alerts_usecase "github.com/utmstack/utmstack/backend/modules/alerts/usecase"
	ep_connectors "github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	ep_dto "github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	notif_connectors "github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	notif_domain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"

	"github.com/threatwinds/go-sdk/catcher"
)

// floodRulesPageSize is how many rules a name search returns. The name matches by part, and the guard
// keeps only the exact ones, so this has to be well above any realistic number of look-alikes.
const floodRulesPageSize = 1000

// floodRuleSwitch lets the flood guard find and disable a tenant's correlation rules through the
// event processing module, scoped to that tenant the way a request from it would be.
type floodRuleSwitch struct {
	rules ep_connectors.CorrelationRuleUsecase
}

var _ alerts_connectors.FloodRuleSwitch = floodRuleSwitch{}

func (s floodRuleSwitch) FindByName(ctx context.Context, tenantID, name string) ([]alerts_connectors.FloodRule, error) {
	result, err := s.rules.List(authz.WithTenantID(ctx, tenantID), ep_dto.CorrelationRuleFilters{RuleName: name, Size: floodRulesPageSize})
	if err != nil {
		return nil, err
	}

	out := make([]alerts_connectors.FloodRule, 0, len(result.Items))
	for _, r := range result.Items {
		out = append(out, alerts_connectors.FloodRule{RelPath: r.RelPath, Name: r.RuleName, Active: r.RuleActive})
	}

	return out, nil
}

func (s floodRuleSwitch) Deactivate(ctx context.Context, tenantID, relPath string) (bool, error) {
	return s.rules.SetActive(authz.WithTenantID(ctx, tenantID), relPath, false)
}

// floodNotifier puts the warning in the tenant's notification bell.
type floodNotifier struct {
	producer notif_connectors.Producer
}

var _ alerts_connectors.FloodNotifier = floodNotifier{}

func (n floodNotifier) Warn(ctx context.Context, tenantID, message string) error {
	return n.producer.Notify(authz.WithTenantID(ctx, tenantID), notif_domain.SourceSystem, notif_domain.TypeWarning, message)
}

// floodConfig reads the guard's settings from the pipeline folder on every pass. A file that cannot be
// read leaves the defaults in force and says why.
func floodConfig(dir string) func() alerts_usecase.FloodConfig {
	return func() alerts_usecase.FloodConfig {
		cfg, err := alerts_usecase.LoadFloodConfig(dir)
		if err != nil {
			_ = catcher.Error("flood guard: cannot read its configuration file, using the defaults", err, map[string]any{"dir": dir})
		}

		return cfg
	}
}
