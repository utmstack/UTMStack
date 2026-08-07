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
	alert.DeduplicateBy = nil

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
func ToAlertFields(alert *plugins.Alert) schema.AlertFields {
	a := schema.AlertFields{
		Timestamp: alert.Timestamp,
		Alert:     *alert,
	}
	a.Alert.Timestamp = ""
	if len(alert.Events) > 0 {
		a.LastEvent = alert.Events[len(alert.Events)-1]
	}
	return a
}
