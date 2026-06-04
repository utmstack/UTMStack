// Package as_400 implements the AS_400 module kind (IBM AS/400 / IBM i).
// The legacy Java impl was ModuleIbmAs400 — same enum value, different file
// name in the legacy backend, but the same three config rows.
package as_400

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "AS_400"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return []domain.ModuleConfigurationKey{
		{
			GroupID: groupID, ConfKey: "collector.as400.user",
			ConfName: "UserName", ConfDescription: "The AS400 user's name.",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "collector.as400.password",
			ConfName: "Password", ConfDescription: "The AS400 user's password.",
			ConfDataType: domain.ConfTypePassword, ConfRequired: true,
		},
		{
			GroupID: groupID, ConfKey: "collector.as400.hostname",
			ConfName: "Hostname", ConfDescription: "The AS400's hostname or IP address.",
			ConfDataType: domain.ConfTypeText, ConfRequired: true,
		},
	}
}
