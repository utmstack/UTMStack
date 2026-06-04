// Package deceptive_bytes implements the DECEPTIVE_BYTES module kind.
package deceptive_bytes

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "DECEPTIVE_BYTES"

type kind struct {
	baseline.Defaults
}

func New() connectors.ModuleKind {
	return &kind{Defaults: baseline.Defaults{KindName: Name}}
}

func (k *kind) ConfigurationKeys(groupID int64) []domain.ModuleConfigurationKey {
	return baseline.EventHubKeys(groupID, "Event Hub Connection", "Consumer Group", "Storage Container", "Storage Connection")
}
