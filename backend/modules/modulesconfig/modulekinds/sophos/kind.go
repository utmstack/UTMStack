// Package sophos implements the SOPHOS module kind. Sophos Central is consumed
// over its REST API; the panel collects a client id + secret pair and the
// collector polls Sophos for events.
package sophos

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "SOPHOS"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "sophos_client_id",
			ConfName: "Client Id", ConfDescription: "Configure Sophos Central Client Id",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "sophos_x_api_key",
			ConfName: "Client Secret", ConfDescription: "Configure Sophos Central Client Secret",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
	}
}
