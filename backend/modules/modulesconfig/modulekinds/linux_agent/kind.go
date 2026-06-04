// Package linux_agent implements the LINUX_AGENT module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package linux_agent

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "LINUX_AGENT"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
