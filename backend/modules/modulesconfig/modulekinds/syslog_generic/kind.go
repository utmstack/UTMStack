// Package syslog_generic implements the SYSLOG_GENERIC module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package syslog_generic

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "SYSLOG_GENERIC"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
