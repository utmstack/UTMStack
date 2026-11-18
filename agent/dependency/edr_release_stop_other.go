//go:build !windows

package dependency

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/svc"
)

func stopEDRReleaseService(string) error { return svc.Stop(config.ServiceName) }
