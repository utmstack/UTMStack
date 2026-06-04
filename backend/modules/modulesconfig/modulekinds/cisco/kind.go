// Package cisco implements the CISCO module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package cisco

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "CISCO"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
