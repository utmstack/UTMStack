package service

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/association"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/mapper"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/models"
)

type EntityBuilder struct {
	entityMapper *mapper.EntityMapper
}

func NewEntityBuilder(entityMapper *mapper.EntityMapper) *EntityBuilder {
	return &EntityBuilder{
		entityMapper: entityMapper,
	}
}

func (b *EntityBuilder) MapAndRegisterFieldsToEntities(
	fields []*models.FlattenedField,
	incident *models.Incident,
	alert *models.Alert,
	event *models.Event,
	associationBuilder *association.AssociationBuilder,
) {
	for _, field := range fields {
		entityType, matched := b.entityMapper.MapFieldToEntityType(field)
		if !matched {
			continue
		}

		sourceField := ""
		var sideContext *models.Side
		var sourceType string

		if strings.Contains(field.Path, ".origin") {
			sourceField = "origin"
			sourceType = "event.origin"
			sideContext = event.Origin
		} else if strings.Contains(field.Path, ".target") {
			sourceField = "target"
			sourceType = "event.target"
			sideContext = event.Target
		}

		enrichmentCtx := b.buildEnrichmentContext(incident, alert, sideContext, event.ID, sourceType)
		entity, entityID, err := b.entityMapper.BuildEntity(entityType, field.Value, enrichmentCtx)
		if err != nil {
			catcher.Error("failed to build entity", err, map[string]any{
				"entity_type": entityType,
				"field_path":  field.Path,
				"event_id":    event.ID,
			})
			continue
		}

		assocContext := association.AssociationContext{
			AlertID:     alert.ID,
			EventID:     event.ID,
			IncidentID:  fmt.Sprintf("%d", incident.ID),
			SourceField: sourceField,
		}
		associationBuilder.RegisterEntity(entity, entityID, field.Path, assocContext)
	}
}

func (b *EntityBuilder) buildEnrichmentContext(
	incident *models.Incident,
	alert *models.Alert,
	side *models.Side,
	eventID string,
	sourceType string,
) mapper.EntityEnrichmentContext {
	ctx := mapper.EntityEnrichmentContext{
		IncidentID: fmt.Sprintf("%d", incident.ID),
		AlertID:    alert.ID,
		EventID:    eventID,
		Severity:   incident.Severity,
		DataType:   alert.DataType,
		SourceType: sourceType,
	}

	if side != nil && side.Geolocation != nil {
		geo := side.Geolocation
		ctx.Country = geo.Country
		ctx.City = geo.City
		ctx.ASO = geo.ASO

		if geo.Latitude != 0.0 || geo.Longitude != 0.0 {
			ctx.Latitude = &geo.Latitude
			ctx.Longitude = &geo.Longitude
		}

		if geo.Accuracy > 0 {
			accuracy := float64(geo.Accuracy)
			ctx.AccuracyRadius = &accuracy
		}
	}

	return ctx
}
