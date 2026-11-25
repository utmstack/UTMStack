package service

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/association"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/extractor"
	"github.com/utmstack/utmstack/backend/modules/threatintel/contribution/models"
)

type AlertProcessor struct {
	alerts         AlertSource
	fieldExtractor *extractor.FieldExtractor
	entityBuilder  *EntityBuilder
}

func NewAlertProcessor(
	alerts AlertSource,
	fieldExtractor *extractor.FieldExtractor,
	entityBuilder *EntityBuilder,
) *AlertProcessor {
	return &AlertProcessor{
		alerts:         alerts,
		fieldExtractor: fieldExtractor,
		entityBuilder:  entityBuilder,
	}
}

func (p *AlertProcessor) ProcessAlertWithAssociations(
	ctx context.Context,
	incidentAlert *models.IncidentAlert,
	incident *models.Incident,
	associationBuilder *association.AssociationBuilder,
) error {
	alert, err := p.alerts.AlertByID(ctx, incidentAlert.AlertID)
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
