// Package soc_ai implements the SOC_AI module kind (UTMStack's GPT-based
// alert analyst). The model select carries the legacy verbatim options JSON
// so the panel keeps rendering the same dropdown of GPT model names.
package soc_ai

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "SOC_AI"

const (
	confDataTypeBool   = "bool"
	confDataTypeSelect = "select"
)

const modelOptions = `[` +
	`{"value": "gpt-4", "label": "GPT-4"},` +
	`{"value": "gpt-4-0613", "label": "GPT-4 (0613)"},` +
	`{"value": "gpt-4-32k", "label": "GPT-4 32K"},` +
	`{"value": "gpt-4-32k-0613", "label": "GPT-4 32K (0613)"},` +
	`{"value": "gpt-4-turbo", "label": "GPT-4 Turbo"},` +
	`{"value": "gpt-4o", "label": "GPT-4 Omni"},` +
	`{"value": "gpt-4o-mini", "label": "GPT-4 Omni Mini"},` +
	`{"value": "gpt-4.1", "label": "GPT-4.1"},` +
	`{"value": "gpt-4.1-mini", "label": "GPT-4.1 Mini"},` +
	`{"value": "gpt-4.1-nano", "label": "GPT-4.1 Nano"},` +
	`{"value": "gpt-3.5-turbo", "label": "GPT-3.5 Turbo"},` +
	`{"value": "gpt-3.5-turbo-0613", "label": "GPT-3.5 Turbo (0613)"},` +
	`{"value": "gpt-3.5-turbo-16k", "label": "GPT-3.5 Turbo 16K"},` +
	`{"value": "gpt-3.5-turbo-16k-0613", "label": "GPT-3.5 Turbo 16K (0613)"}` +
	`]`

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "utmstack.socai.incidentCreation",
			ConfName:        "Automatic Incident creation",
			ConfDescription: `If set to "true", the system will create incidents based on analysis of alerts.`,
			ConfDataType:    confDataTypeBool, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.changeAlertStatus",
			ConfName:        "Change Alert Status",
			ConfDescription: `If set to "true", SOC Ai will automatically change the status of alerts. Analysts should investigate those with the status "In Review".`,
			ConfDataType:    confDataTypeBool, ConfRequired: false,
		},
		{
			GroupID: groupID, ConfKey: "utmstack.socai.model",
			ConfName:        "Select AI Model",
			ConfDescription: "Choose the AI model that SOC AI will use to analyze alerts.",
			ConfDataType:    confDataTypeSelect, ConfRequired: true,
			ConfOptions: modelOptions,
		},
	}
}
