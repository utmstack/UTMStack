package service

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/association"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/extractor"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

type AlertProcessor struct {
	opensearchClient *client.OpenSearchClient
	fieldExtractor   *extractor.FieldExtractor
	entityBuilder    *EntityBuilder
}

func NewAlertProcessor(
	opensearchClient *client.OpenSearchClient,
	fieldExtractor *extractor.FieldExtractor,
	entityBuilder *EntityBuilder,
) *AlertProcessor {
	return &AlertProcessor{
		opensearchClient: opensearchClient,
		fieldExtractor:   fieldExtractor,
		entityBuilder:    entityBuilder,
	}
}

func (p *AlertProcessor) ProcessAlertWithAssociations(
	ctx context.Context,
	incidentAlert *models.IncidentAlert,
	incident *models.Incident,
	associationBuilder *association.AssociationBuilder,
) error {
	alert, err := p.opensearchClient.GetAlertByID(ctx, incidentAlert.AlertID)
	if err != nil {
		return catcher.Error("failed to get alert", err, map[string]any{
			"alert_id":    incidentAlert.AlertID,
			"incident_id": incident.ID,
		})
	}

	for _, event := range alert.Events {
		eventFields := p.fieldExtractor.ExtractFromEvent(event)

		p.entityBuilder.MapAndRegisterFieldsToEntities(
			eventFields,
			incident,
			alert,
			event,
			associationBuilder,
		)
	}

	return nil
}
