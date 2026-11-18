//go:build linux || windows

package service

import (
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/responder"
)

func cleanupResponse() { _ = responder.CloseSuspensions() }
func killObserved(resp *responder.Responder, p proctable.Proc, sig string) {
	if p.OSIdentity > 0 {
		_, _ = resp.KillTreeIdentity(p.PID, p.OSIdentity, sig)
	}
}
