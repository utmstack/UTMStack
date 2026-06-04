// Package vmware implements the VMWARE module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package vmware

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "VMWARE"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
