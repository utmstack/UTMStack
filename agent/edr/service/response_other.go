//go:build !linux && !windows

package service

import (
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/responder"
)

func cleanupResponse() {}
func killObserved(resp *responder.Responder, p proctable.Proc, sig string) {
	_, _ = resp.KillTree(p.PID, sig)
}
