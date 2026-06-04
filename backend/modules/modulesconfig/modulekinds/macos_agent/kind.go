// Package macos_agent implements the MACOS_AGENT module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package macos_agent

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "MACOS_AGENT"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
