// Package palo_alto implements the PALO_ALTO module kind (Palo Alto firewall
// logs piped through an Azure Event Hub).
package palo_alto

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "PALO_ALTO"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return baseline.EventHubKeys(groupID, "Event Hub Connection", "Consumer Group", "Storage Container", "Storage Connection")
}
