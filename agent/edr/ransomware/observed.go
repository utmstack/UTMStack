package ransomware

import (
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"strings"
)

type identityResponder interface {
	SuspendIdentity(int, int64) error
	ResumeIdentity(int, int64) error
	KillTreeIdentity(int, int64, string) ([]int, error)
}

func (g *Guard) suspendObservation(p proctable.Proc) error {
	if !p.IdentityRequired && p.OSIdentity == 0 {
		return g.deps.Resp.Suspend(p.PID)
	}
	r, ok := g.deps.Resp.(identityResponder)
	if !ok || p.OSIdentity <= 0 {
		return errors.New("ransomware original process identity unavailable")
	}
	g.smu.Lock()
	defer g.smu.Unlock()
	if id, ok := g.suspended[p.PID]; ok && id == p.OSIdentity {
		return nil
	}
	if len(g.suspended) >= 1024 {
		return errors.New("ransomware suspension ownership limit")
	}
	if err := r.SuspendIdentity(p.PID, p.OSIdentity); err != nil {
		return err
	}
	g.suspended[p.PID] = p.OSIdentity
	return nil
}
func (g *Guard) resumeObservation(p proctable.Proc) error {
	if !p.IdentityRequired && p.OSIdentity == 0 {
		return g.deps.Resp.Resume(p.PID)
	}
	r, ok := g.deps.Resp.(identityResponder)
	if !ok {
		return errors.New("ransomware identity-bound resume unavailable")
	}
	g.smu.Lock()
	defer g.smu.Unlock()
	if id, exists := g.suspended[p.PID]; !exists || p.OSIdentity <= 0 || id != p.OSIdentity {
		return nil
	}
	delete(g.suspended, p.PID)
	return r.ResumeIdentity(p.PID, p.OSIdentity)
}
func (g *Guard) killObservation(p proctable.Proc, sig string) ([]int, error) {
	if !p.IdentityRequired && p.OSIdentity == 0 {
		return g.deps.Resp.KillTree(p.PID, sig)
	}
	r, ok := g.deps.Resp.(identityResponder)
	if !ok || p.OSIdentity <= 0 {
		return nil, errors.New("ransomware original process identity unavailable")
	}
	return r.KillTreeIdentity(p.PID, p.OSIdentity, sig)
}
func (g *Guard) isContainedObservation(p proctable.Proc) bool {
	g.cmu.Lock()
	defer g.cmu.Unlock()
	if m, ok := g.contained[p.PID]; ok && m.osIdentity != 0 {
		if m.expired(g.deps.Now()) || m.osIdentity != p.OSIdentity {
			delete(g.contained, p.PID)
			return false
		}
		return true
	}
	return g.containedLocked(p.PID)
}

// OnObservedProcess preserves native creation identity for command evidence.
// A parent PID is only followed while its captured birth predates the child.
func (g *Guard) OnObservedProcess(p proctable.Proc) {
	if !p.IdentityRequired && p.OSIdentity == 0 {
		g.OnProcStart(p.PID, p.PPID, p.Image, p.Cmdline, p.StartTS)
		return
	}
	g.observeCommand(p, false)
}

// OnObservedSnapshot evaluates freshly captured current metadata without
// asserting that it belonged to an earlier queued process-start/exec event.
// Callers must supply their current pidfd/native-handle-validated observation.
func (g *Guard) OnObservedSnapshot(p proctable.Proc) {
	if p.OSIdentity <= 0 {
		return
	}
	p.IdentityRequired = true
	g.observeCommand(p, true)
}

func (g *Guard) observeCommand(p proctable.Proc, current bool) {
	if !g.cfg.Ransomware.Enabled || p.OSIdentity <= 0 {
		return
	}
	rule, ok := MatchT1490(p.Image, p.Cmdline, g.commandAllowlist())
	if !ok {
		return
	}
	if current {
		rule = "current_command_observation:" + rule
	}
	parent := p.PPID
	childIdentity := p.OSIdentity
	var culprit proctable.Proc
	for hops := 0; hops < 4; hops++ {
		candidate, ok := g.deps.Table.Get(parent)
		if !ok || candidate.OSIdentity <= 0 || candidate.OSIdentity > childIdentity {
			return
		}
		culprit = candidate
		culprit.IdentityRequired = true
		base := strings.ToLower(candidate.Image)
		if i := strings.LastIndexAny(base, `/\`); i >= 0 {
			base = base[i+1:]
		}
		if base != "cmd.exe" && base != "powershell.exe" && base != "pwsh.exe" && base != "conhost.exe" && base != "wscript.exe" && base != "cscript.exe" {
			break
		}
		parent = candidate.PPID
		childIdentity = candidate.OSIdentity
	}
	if culprit.PID <= maxKernelPID || g.isContainedObservation(culprit) || (g.deps.Trusted != nil && culprit.Image != "" && g.deps.Trusted(culprit.Image)) {
		return
	}
	ev := Evidence{PID: culprit.PID, Gen: culprit.OSIdentity, Kind: KindT1490, Weight: DefaultWeights[KindT1490], Detail: rule, TS: g.deps.Now()}
	g.handleObserved(g.scorer.Add(ev), culprit)
}
