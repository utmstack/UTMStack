package mapper

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/entities"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

type EntityMapper struct {
	entityTypes map[string]bool
}

func NewEntityMapper() *EntityMapper {
	mapper := &EntityMapper{
		entityTypes: make(map[string]bool),
	}

	for _, def := range entities.Definitions {
		mapper.entityTypes[def.Type] = true
	}

	catcher.Info("entity mapper initialized", map[string]any{
		"total_entity_types": len(mapper.entityTypes),
	})

	return mapper
}

func (m *EntityMapper) MapFieldToEntityType(field *models.FlattenedField) (string, bool) {
	leafKey := normalizeKey(field.Key)

	if m.entityTypes[leafKey] {
		return leafKey, true
	}

	return "", false
}

func normalizeKey(key string) string {
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "_", "-")
	return key
}

func (m *EntityMapper) BuildEntity(entityType string, value any, context EntityEnrichmentContext) (*entities.Entity, string, error) {
	validatedValue, hash, err := entities.ValidateValue(value, entityType)
	if err != nil {
		return nil, "", catcher.Error("validation failed for entity type", err, map[string]any{
			"entity_type": entityType,
		})
	}

	attrs := entities.Attributes{}
	if !attrs.SetAttribute(entityType, validatedValue) {
		return nil, "", catcher.Error("failed to set attribute for entity type", nil, map[string]any{
			"entity_type": entityType,
		})
	}

	if context.Country != "" {
		attrs.Country = &context.Country
	}
	if context.City != "" {
		attrs.City = &context.City
	}
	if context.ASO != "" {
		attrs.Aso = &context.ASO
	}
	if context.Latitude != nil {
		attrs.Latitude = context.Latitude
	}
	if context.Longitude != nil {
		attrs.Longitude = context.Longitude
	}
	if context.AccuracyRadius != nil {
		attrs.AccuracyRadius = context.AccuracyRadius
	}

	reputation := calculateReputation(context.Severity)

	tags := buildEntityTags(context)

	entity := &entities.Entity{
		Type:         entityType,
		Attributes:   attrs,
		Reputation:   reputation,
		Tags:         tags,
		Associations: nil,
	}

	entityID := fmt.Sprintf("%s-%s", entityType, hash)

	return entity, entityID, nil
}

func buildEntityTags(context EntityEnrichmentContext) []string {
	tags := []string{"utmstack"}

	if context.IncidentID != "" {
		tags = append(tags, "incident-"+context.IncidentID)
	}

	if context.DataType != "" {
		tags = append(tags, "datasource-"+context.DataType)
	}

	return tags
}

type EntityEnrichmentContext struct {
	IncidentID string
	AlertID    string
	EventID    string

	Severity int
	DataType string

	SourceType string

	Country        string
	City           string
	Latitude       *float64
	Longitude      *float64
	ASO            string
	AccuracyRadius *float64
}

func calculateReputation(severity int) int {
	switch {
	case severity >= 7:
		return -3
	case severity >= 4:
		return -1
	default:
		return 0
	}
}
