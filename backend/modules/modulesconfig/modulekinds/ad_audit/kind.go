// Package ad_audit implements the AD_AUDIT module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package ad_audit

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "AD_AUDIT"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
