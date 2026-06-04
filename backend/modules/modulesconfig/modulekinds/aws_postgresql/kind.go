// Package aws_postgresql implements the AWS_POSTGRESQL module kind. It has no default
// configuration keys — the integration is a pure log-source / agent module,
// configured through the collector itself rather than the panel.
package aws_postgresql

import (
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const Name = "AWS_POSTGRESQL"

func New() connectors.ModuleKind {
	return &baseline.Defaults{KindName: Name}
}
