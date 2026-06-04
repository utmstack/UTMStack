// Package windows_events implements the WINDOWS_EVENTS module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package windows_events

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "WINDOWS_EVENTS"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
