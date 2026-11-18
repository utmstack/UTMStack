package procwatch

import "github.com/utmstack/UTMStack/agent/edr/linuxproc"

type observedProcess struct {
	identity int64
	image    string
}

// New images executed by an existing PID are observed too. A polling monitor
// cannot see processes that start and exit entirely between two snapshots.
func processChanges(previous map[int]observedProcess, current []linuxproc.Process, baseline bool) (map[int]observedProcess, []ProcStart) {
	next := make(map[int]observedProcess, len(current))
	var changes []ProcStart
	for _, p := range current {
		key := observedProcess{p.StartTicks, p.Image}
		next[p.PID] = key
		if old, ok := previous[p.PID]; !baseline && (!ok || old != key) {
			changes = append(changes, ProcStart{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline, OSIdentity: p.StartTicks, IdentityRequired: true})
		}
	}
	return next, changes
}
