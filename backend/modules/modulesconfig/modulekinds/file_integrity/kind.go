// Package file_integrity implements the FILE_INTEGRITY module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package file_integrity

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "FILE_INTEGRITY"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
