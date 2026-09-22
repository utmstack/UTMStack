package service

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/association"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/client"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/models"
)

type IncidentProcessor struct {
	backendClient      *client.BackendClient
	alertClient        *client.AlertClient
	threadwindsClient  *client.ThreadWindsClient
	alertProcessor     *AlertProcessor
	associationBuilder *association.AssociationBuilder
}

func NewIncidentProcessor(
	deps *client.ClientDependencies,
	alertProcessor *AlertProcessor,
	associationBuilder *association.AssociationBuilder,
) *IncidentProcessor {
	return &IncidentProcessor{
		backendClient:      deps.Backend,
		alertClient:        deps.Alerts,
		threadwindsClient:  deps.ThreadWinds,
		alertProcessor:     alertProcessor,
		associationBuilder: associationBuilder,
	}
}

func (p *IncidentProcessor) ProcessIncident(ctx context.Context, incident *models.Incident) (int, bool, error) {
	p.associationBuilder.ClearRegistry()

	incidentAlerts, err := p.backendClient.GetIncidentAlerts(ctx, incident.ID)
	if err != nil {
		return 0, false, catcher.Error("failed to get incident alerts", err, nil)
	}

	if len(incidentAlerts) == 0 {
		return 0, false, nil
	}

	for _, incidentAlert := range incidentAlerts {
		err := p.alertProcessor.ProcessAlertWithAssociations(ctx, incidentAlert, incident, p.associationBuilder)
		if err != nil {
			catcher.Error("failed to process alert", err, map[string]any{
				"alert_id":    incidentAlert.AlertID,
				"incident_id": incident.ID,
			})
			continue
		}
	}

	allEntities := p.associationBuilder.BuildAssociations()

	var outcome client.IngestOutcome
	if len(allEntities) > 0 {
		outcome = p.threadwindsClient.IngestBatch(ctx, allEntities)
		if len(outcome.Failures) > 0 {
			_ = catcher.Error("incident processed with errors", nil, map[string]any{
				"incident_id":  incident.ID,
				"ok_count":     outcome.OKCount,
				"fail_count":   len(outcome.Failures),
				"entity_count": len(allEntities),
			})
		}
	}

	return len(allEntities), outcome.ProvisioningBlocked, nil
}
