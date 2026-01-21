package main

import (
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

func cleanAlerts(alert schema.AlertFields) schema.AlertFields {
	alert.ParentID = nil
	alert.Events = nil
	alert.TagRulesApplied = nil
	alert.DeduplicatedBy = nil

	if alert.Target != nil {
		if alert.Target.User != "" {
			alert.Target.User = config.FakeUserName
		}
		if alert.Target.Email != "" {
			alert.Target.Email = config.FakeEmail
		}
	}

	if alert.LastEvent != nil {
		if alert.LastEvent.Target != nil && alert.LastEvent.Target.User != "" {
			alert.LastEvent.Target.User = config.FakeUserName
		}
		if alert.LastEvent.Target != nil && alert.LastEvent.Target.Email != "" {
			alert.LastEvent.Target.Email = config.FakeEmail
		}

		if alert.LastEvent.Log != nil {
			for key, val := range alert.LastEvent.Log {
				switch v := val.Kind.(type) {
				case *structpb.Value_StringValue:
					original := v.StringValue
					cleaned := original
					for _, pattern := range config.SensitivePatterns {
						re := pattern.GetRegexp()
						cleaned = re.ReplaceAllString(cleaned, pattern.FakeValue)
					}
					if cleaned != original {
						alert.LastEvent.Log[key] = structpb.NewStringValue(cleaned)
					}
				default:
					continue
				}
			}
		}
	}
	return alert
}

func alertToAlertFields(alert *plugins.Alert) schema.AlertFields {
	var severityN int
	var severityLabel string
	switch alert.Severity {
	case "low":
		severityN = 1
		severityLabel = "Low"
	case "medium":
		severityN = 2
		severityLabel = "Medium"
	case "high":
		severityN = 3
		severityLabel = "High"
	default:
		severityN = 1
		severityLabel = "Low"
	}

	a := schema.AlertFields{
		Timestamp:     alert.Timestamp,
		ID:            alert.Id,
		Status:        1,
		StatusLabel:   "Automatic review",
		Name:          alert.Name,
		Category:      alert.Category,
		Severity:      severityN,
		SeverityLabel: severityLabel,
		Description:   alert.Description,
		Technique:     alert.Technique,
		Reference:     alert.References,
		DataType:      alert.DataType,
		DataSource:    alert.DataSource,
		Adversary:     alert.Adversary,
		Target:        alert.Target,
		LastEvent: func() *plugins.Event {
			l := len(alert.Events)
			if l == 0 {
				return nil
			}
			return alert.Events[l-1]
		}(),
		Events:         alert.Events,
		Impact:         alert.Impact,
		ImpactScore:    alert.ImpactScore,
		DeduplicatedBy: alert.DeduplicateBy,
		GroupedBy:      alert.GroupBy,
	}

	return a
}
