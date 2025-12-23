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
		return nil, "", fmt.Errorf("validation failed for type %s: %w", entityType, err)
	}

	attrs := entities.Attributes{}
	if !attrs.SetAttribute(entityType, validatedValue) {
		return nil, "", fmt.Errorf("failed to set attribute for type %s", entityType)
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

	tags := []string{
		"utmstack",
		"incident-" + context.IncidentID,
	}
	if context.DataType != "" {
		tags = append(tags, "datasource-"+context.DataType)
	}

	entity := &entities.Entity{
		Type:         entityType,
		Attributes:   attrs,
		Reputation:   reputation,
		Tags:         tags,
		VisibleBy:    []string{"utmstack"},
		Associations: nil,
	}

	entityID := fmt.Sprintf("%s-%s", entityType, hash)

	catcher.Info("entity built successfully", map[string]any{
		"entity_type": entityType,
		"entity_id":   entityID,
		"reputation":  reputation,
	})

	return entity, entityID, nil
}

type EntityEnrichmentContext struct {
	IncidentID     string
	Severity       int
	DataType       string
	Country        string
	City           string
	Latitude       *float64
	Longitude      *float64
	ASO            string
	AccuracyRadius *float64
}

func calculateReputation(severity int) int {
	if severity >= 7 {
		return -3
	} else if severity >= 4 {
		return -1
	}
	return 0
}
