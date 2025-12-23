package extractor

import (
	"fmt"
	"strings"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
)

type FieldExtractor struct{}

func NewFieldExtractor() *FieldExtractor {
	return &FieldExtractor{}
}

func (e *FieldExtractor) ExtractFromAlert(alert *models.Alert) []*models.FlattenedField {
	fields := make([]*models.FlattenedField, 0, 4)

	if alert.Source != nil {
		fields = append(fields, e.extractFromHost(alert.Source, "alert.source")...)
	}

	if alert.Destination != nil {
		fields = append(fields, e.extractFromHost(alert.Destination, "alert.destination")...)
	}

	catcher.Info("extracted fields from alert", map[string]any{
		"alert_id":    alert.ID,
		"field_count": len(fields),
	})

	return fields
}

func (e *FieldExtractor) extractFromHost(host *models.Host, prefix string) []*models.FlattenedField {
	fields := make([]*models.FlattenedField, 0, 5)

	if host.IP != "" && !isInvalidValue(host.IP) {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".ip",
			Key:   "ip",
			Value: host.IP,
		})
	}

	if host.Host != "" && !isInvalidValue(host.Host) {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".host",
			Key:   "hostname",
			Value: host.Host,
		})
	}

	if host.User != "" && !isInvalidValue(host.User) {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".user",
			Key:   "username",
			Value: host.User,
		})
	}

	if host.Port != 0 {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".port",
			Key:   "port",
			Value: host.Port,
		})
	}

	if host.ASN != 0 && !isInvalidValue(host.ASN) {
		fields = append(fields, &models.FlattenedField{
			Path:  prefix + ".asn",
			Key:   "asn",
			Value: host.ASN,
		})
	}

	return fields
}

func isInvalidValue(value any) bool {
	strValue := fmt.Sprintf("%v", value)
	if strings.TrimSpace(strValue) == "" {
		return true
	}

	invalidValues := []string{
		"-",
		"N/A",
		"n/a",
		"unknown",
		"null",
		"(null)",
		"none",
		"0",
		"-1",
		"0.0.0.0",
		"255.255.255.255",
		"127.0.0.1",
		"localhost",
	}

	for _, invalid := range invalidValues {
		if strings.EqualFold(strValue, invalid) {
			return true
		}
	}

	return false
}
