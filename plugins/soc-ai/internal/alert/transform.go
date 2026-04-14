package alert

import (
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"

	structpb "google.golang.org/protobuf/types/known/structpb"
)

// Clean removes fields that are not needed for LLM analysis or contain sensitive data.
// It tracks which fields were anonymized so the LLM knows not to make assumptions about placeholder values.
func Clean(alert schema.AlertFields) schema.AlertFields {
	alert.Events = nil
	alert.TagRulesApplied = nil
	alert.DeduplicatedBy = nil

	var anonymized []string

	if alert.Target != nil {
		if alert.Target.User != "" {
			alert.Target.User = config.FakeUserName
			anonymized = append(anonymized, "target.user")
		}
		if alert.Target.Email != "" {
			alert.Target.Email = config.FakeEmail
			anonymized = append(anonymized, "target.email")
		}
	}

	if alert.Adversary != nil {
		if alert.Adversary.User != "" {
			alert.Adversary.User = config.FakeUserName
			anonymized = append(anonymized, "adversary.user")
		}
		if alert.Adversary.Email != "" {
			alert.Adversary.Email = config.FakeEmail
			anonymized = append(anonymized, "adversary.email")
		}
	}

	if alert.LastEvent != nil {
		if alert.LastEvent.Target != nil && alert.LastEvent.Target.User != "" {
			alert.LastEvent.Target.User = config.FakeUserName
			anonymized = append(anonymized, "lastEvent.target.user")
		}
		if alert.LastEvent.Target != nil && alert.LastEvent.Target.Email != "" {
			alert.LastEvent.Target.Email = config.FakeEmail
			anonymized = append(anonymized, "lastEvent.target.email")
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
						anonymized = append(anonymized, "lastEvent.log."+key)
					}
				default:
					continue
				}
			}
		}
	}

	if len(anonymized) > 0 {
		alert.AnonymizedFields = anonymized
	}

	return alert
}

// ToAlertFields converts a plugins.Alert to schema.AlertFields
func ToAlertFields(alert *plugins.Alert) schema.AlertFields {
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
		Timestamp:      alert.Timestamp,
		Status:         1,
		StatusLabel:    "Automatic review",
		Severity:       severityN,
		SeverityLabel:  severityLabel,
		Reference:      alert.References,
		DeduplicatedBy: alert.DeduplicateBy,
		GroupedBy:      alert.GroupBy,
		LastEvent: func() *plugins.Event {
			l := len(alert.Events)
			if l == 0 {
				return nil
			}
			return alert.Events[l-1]
		}(),
	}

	// Set fields from embedded plugins.Alert
	a.Id = alert.Id
	a.Name = alert.Name
	a.Category = alert.Category
	a.Description = alert.Description
	a.Technique = alert.Technique
	a.DataType = alert.DataType
	a.DataSource = alert.DataSource
	a.Adversary = alert.Adversary
	a.Target = alert.Target
	a.Events = alert.Events
	a.Impact = alert.Impact
	a.ImpactScore = alert.ImpactScore

	return a
}
