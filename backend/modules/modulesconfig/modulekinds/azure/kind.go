// Package azure implements the AZURE module kind. Azure logs are streamed
// through an Event Hub; the panel collects the Event Hub connection string,
// consumer group, and a storage account for checkpointing.
package azure

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "AZURE"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return baseline.EventHubKeys(
		groupID,
		"Event Hub Shared access policies - Connection string",
		"Consumer Group Name",
		"Storage Container Name",
		"Storage account connection string with key",
	)
}
