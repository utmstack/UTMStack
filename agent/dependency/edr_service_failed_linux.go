//go:build linux

package dependency

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/kardianos/service"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
)

// Kardianos reports a failed systemd unit as an unknown-state error. Recovery
// still owns that installed, stopped unit and must be able to restore its
// previous signed executable. Only this positively observed state overrides
// the original error; unavailable, masked and malformed units keep it.
func edrFailedServiceInstalled() bool {
	if service.Platform() != "linux-systemd" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", "show", edrconfig.ServiceName+".service", "--no-pager", "--property=LoadState", "--property=ActiveState").Output()
	return err == nil && edrLoadedFailedSystemdUnit(string(out))
}

func edrLoadedFailedSystemdUnit(output string) bool {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		return false
	}
	return lines[0] == "LoadState=loaded" && lines[1] == "ActiveState=failed" ||
		lines[1] == "LoadState=loaded" && lines[0] == "ActiveState=failed"
}
