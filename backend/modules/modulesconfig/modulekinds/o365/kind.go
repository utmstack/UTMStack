// Package o365 implements the O365 module kind. The cloud-environment select
// field is optional; its options JSON is the legacy verbatim string so the
// panel renders the same dropdown labels.
package o365

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "O365"

// confDataTypeSelect is the legacy "select" data type — a dropdown the panel
// renders from confOptions (JSON). It is not a sensitive type.
const confDataTypeSelect = "select"

const cloudEnvironmentOptions = `[` +
	`{ "value": "Commercial", "label": "Commercial - Azure commercial global (Default)" },` +
	`{ "value": "GCC", "label": "GCC - US Government Community Cloud" },` +
	`{ "value": "GCCHigh", "label": "GCC High - US Government Community Cloud High (DoD IL4)" },` +
	`{ "value": "DoD", "label": "DoD - US Department of Defense (DoD IL5)" }` +
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
			GroupID: groupID, ConfKey: "office365_client_id",
			ConfName: "Client ID", ConfDescription: "Configure Office365 client ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "office365_client_secret",
			ConfName: "Client Secret", ConfDescription: "Configure Office365 client secret",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "office365_tenant_id",
			ConfName: "Tenant ID", ConfDescription: "Configure Office365 tenant ID",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "office365_cloud_environment",
			ConfName: "Cloud Environment", ConfDescription: "Select the Microsoft cloud environment for Office 365 integration.",
			ConfDataType: confDataTypeSelect, ConfRequired: false,
			ConfValue:   "Commercial",
			ConfOptions: cloudEnvironmentOptions,
		},
	}
}
