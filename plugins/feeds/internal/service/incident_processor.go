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
	opensearchClient   *client.OpenSearchClient
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
		opensearchClient:   deps.OpenSearch,
		threadwindsClient:  deps.ThreadWinds,
		alertProcessor:     alertProcessor,
		associationBuilder: associationBuilder,
	}
}

func (p *IncidentProcessor) ProcessIncident(ctx context.Context, incident *models.Incident) (int, error) {
	p.associationBuilder.ClearRegistry()

	incidentAlerts, err := p.backendClient.GetIncidentAlerts(ctx, incident.ID)
	if err != nil {
		return 0, catcher.Error("failed to get incident alerts", err, nil)
	}

	if len(incidentAlerts) == 0 {
		return 0, nil
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

	if len(allEntities) > 0 {
		if err := p.threadwindsClient.IngestBatch(ctx, allEntities); err != nil {
			return 0, catcher.Error("failed to ingest batch", err, map[string]any{
				"incident_id":  incident.ID,
				"entity_count": len(allEntities),
			})
		}
	}

	return len(allEntities), nil
}
