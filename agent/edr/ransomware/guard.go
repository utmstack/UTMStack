package ransomware

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

// feedRetryBackoff is the delay before restarting a file feed that exited
// unexpectedly. A package var so tests can shorten it.
var feedRetryBackoff = 5 * time.Second

type treeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
}
type appender interface{ Append(string) error }
type quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
}
type incidentSink interface{ StoreIncident(cache.RansomwareIncident) error }
type canarySet interface{ Contains(path string) bool }

// GuardDeps bundles the guard's collaborators (interfaces where a fake is
// useful, concrete where trivial) so the escalation logic is unit-testable.
type GuardDeps struct {
	Cfg       config.EDRConfig
	Table     *proctable.Table
	Resp      treeResponder
	Spool     appender
	Quar      quarantiner
	Incidents incidentSink
	Canaries  canarySet
	Hash      func(string) (string, error)
	Now       func() time.Time
	NewID     func() string
	// Trusted reports whether a process image is allowlisted; when it returns
	// true the guard emits no evidence for that process (canary touches and
	// T1490 commands are ignored). nil = nothing trusted.
	Trusted func(image string) bool
}

// isTrusted resolves a PID's image via the process table and reports whether it
// is allowlisted. A trusted process is never scored.
func (g *Guard) isTrusted(pid int) bool {
	if g.deps.Trusted == nil {
		return false
	}
	if p, ok := g.deps.Table.Get(pid); ok {
		return g.deps.Trusted(p.Image)
	}
	return false
}

// Guard fuses ransomware sensors into per-process risk and runs the escalation
// ladder against the existing response arm.
type Guard struct {
	cfg         config.EDRConfig
	deps        GuardDeps
	scorer      *Scorer
	feedHealthy atomic.Bool

	// Hot-reloadable policy: the response mode and command allowlist can change
	// at runtime (CLI edit → service reload) without restarting the guard. The
	// scorer thresholds/decay are fixed at construction (structural).
	pmu          sync.RWMutex
	responseMode string
	commands     []string
}

// SetPolicy hot-applies a changed response mode and command allowlist (called by
// the service on a config reload). Thresholds/decay are structural and unaffected.
func (g *Guard) SetPolicy(responseMode string, commands []string) {
	g.pmu.Lock()
	defer g.pmu.Unlock()
	g.responseMode = responseMode
	g.commands = commands
}

func (g *Guard) commandAllowlist() []string {
	g.pmu.RLock()
	defer g.pmu.RUnlock()
	return g.commands
}

func (g *Guard) responseModeLive() string {
	g.pmu.RLock()
	defer g.pmu.RUnlock()
	return g.responseMode
}

func NewGuard(deps GuardDeps) *Guard {
	rc := deps.Cfg.Ransomware
	if deps.Now == nil {
		deps.Now = time.Now
	}
	half := float64(rc.DecayHalfLifeMs) / 1000.0
	sc := NewScorer(float64(rc.SuspendThreshold), float64(rc.KillThreshold), half, deps.Now)
	return &Guard{
		cfg: deps.Cfg, deps: deps, scorer: sc,
		responseMode: rc.ResponseMode,
		commands:     deps.Cfg.Allowlist.Commands,
	}
}

// OnProcStart runs the T1490 command-rule sensor for a new process. Attribution
// is to the culprit (the parent that spawned the recovery command, skipping
// shell hosts), because vssadmin/wbadmin/etc. are the encryptor's children.
func (g *Guard) OnProcStart(pid, ppid int, image, cmdline string, gen int64) {
	if !g.cfg.Ransomware.Enabled {
		return
	}
	rule, ok := MatchT1490(image, cmdline, g.commandAllowlist())
	if !ok {
		return
	}
	culprit := g.culpritPID(ppid)
	// A trusted process (e.g. sanctioned backup/DR tooling) that legitimately
	// runs recovery commands must not be scored.
	if g.isTrusted(culprit) {
		return
	}
	ev := Evidence{PID: culprit, Gen: g.genOf(culprit, 0), Kind: KindT1490,
		Weight: DefaultWeights[KindT1490], Detail: rule, TS: g.deps.Now()}
	g.handle(g.scorer.Add(ev), culprit)
}

// OnFileEvent runs the canary sensor for a per-process file op.
func (g *Guard) OnFileEvent(fe FileEvent) {
	if !g.cfg.Ransomware.Enabled {
		return
	}
	if !g.deps.Canaries.Contains(fe.Path) {
		return // v1: only canary touches produce evidence from the file stream
	}
	// A trusted process (backup/sync/indexer) touching a canary is not scored.
	if g.isTrusted(fe.PID) {
		return
	}
	base := fe.Path
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	ev := Evidence{PID: fe.PID, Gen: g.genOf(fe.PID, 0), Kind: KindCanary,
		Weight: DefaultWeights[KindCanary], Detail: base, TS: g.deps.Now()}
	g.handle(g.scorer.Add(ev), fe.PID)
}

func (g *Guard) genOf(pid int, gen int64) int64 {
	if gen != 0 {
		return gen
	}
	if p, ok := g.deps.Table.Get(pid); ok {
		return p.StartTS
	}
	return 0
}

// culpritPID walks up past known shell/host interpreters so we blame the
// encryptor, not the cmd.exe/powershell.exe it used to run vssadmin.
func (g *Guard) culpritPID(pid int) int {
	shells := map[string]bool{"cmd.exe": true, "powershell.exe": true, "pwsh.exe": true, "conhost.exe": true, "wscript.exe": true, "cscript.exe": true}
	for hops := 0; hops < 4; hops++ {
		p, ok := g.deps.Table.Get(pid)
		if !ok {
			return pid
		}
		base := strings.ToLower(p.Image)
		if i := strings.LastIndexAny(base, `/\`); i >= 0 {
			base = base[i+1:]
		}
		if !shells[base] || p.PPID == 0 {
			return pid
		}
		pid = p.PPID
	}
	return pid
}

// handle executes the escalation ladder for a scorer decision on culpritPID.
func (g *Guard) handle(d Decision, pid int) {
	if d.Escalation == EscNone {
		return
	}
	p, _ := g.deps.Table.Get(pid)
	pinfo := event.ProcInfo{PID: pid, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline}
	mode := normalizeMode(g.responseModeLive())

	// alert mode: surface the escalation but never suspend/kill/quarantine.
	if mode == "alert" {
		sev := "high"
		if d.Escalation == EscKill {
			sev = "critical"
		}
		g.emit(event.ActionRansomwareSuspected, pinfo, d, sev)
		g.record("alerted", pid, p, d, "")
		return
	}

	// suspend|kill mode, sub-kill escalation: suspend mode freezes the suspect;
	// kill mode takes no containment action until the kill threshold.
	if d.Escalation == EscSuspend {
		suspended := false
		if mode == "suspend" {
			_ = g.deps.Resp.Suspend(pid)
			suspended = true
		}
		g.emit(event.ActionRansomwareSuspected, pinfo, d, "high")
		if suspended {
			g.record("suspended", pid, p, d, "")
		} else {
			g.record("suspected", pid, p, d, "")
		}
		return
	}

	// EscKill in suspend|kill mode: contain.
	if mode == "suspend" {
		_ = g.deps.Resp.Suspend(pid) // freeze the tree to reduce further encryption during the kill
	}
	killed, killErr := g.deps.Resp.KillTree(pid, "ransomware:"+d.TopSignal)
	if killErr != nil {
		logger.Error("UTMStack EDR: ransomware kill tree %d: %v", pid, killErr)
	}
	qid := ""
	quarFailed := false
	if p.Image != "" {
		sha := ""
		if g.deps.Hash != nil {
			sha, _ = g.deps.Hash(p.Image)
		}
		if id, err := g.deps.Quar.Quarantine(p.Image, sha, "ransomware:"+d.TopSignal); err == nil {
			qid = id
		} else {
			quarFailed = true
			logger.Error("UTMStack EDR: ransomware quarantine %s: %v", p.Image, err)
		}
	}
	// Honest reporting (project rule): only claim "contained" if the kill actually
	// terminated something and quarantine did not fail.
	action := "contained"
	// responder.KillTree logs+swallows per-PID kill errors and returns nil, so the
	// honest signal that we terminated nothing is an empty `killed` slice — not killErr.
	if killErr != nil || len(killed) == 0 || quarFailed {
		action = "contain_failed"
	}
	g.emit(event.ActionRansomwareContained, pinfo, d, "critical")
	g.record(action, pid, p, d, qid)
	g.scorer.Forget(pid)
}

// normalizeMode maps a configured response_mode to one of the three known modes,
// defaulting unknown/empty/mis-cased values to the LEAST destructive ("alert")
// so a typo can never cause auto-kill.
func normalizeMode(m string) string {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "kill":
		return "kill"
	case "suspend":
		return "suspend"
	default:
		return "alert"
	}
}

func (g *Guard) emit(action string, p event.ProcInfo, d Decision, sev string) {
	if js, err := event.NewRansomwareEvent(action, p, d.TopSignal, sev).ToJSON(); err == nil {
		_ = g.deps.Spool.Append(js)
	}
}

func (g *Guard) record(action string, pid int, p proctable.Proc, d Decision, qid string) {
	id := "incident"
	if g.deps.NewID != nil {
		id = g.deps.NewID()
	}
	kinds := make([]string, 0, len(d.Kinds))
	for _, k := range d.Kinds {
		kinds = append(kinds, string(k))
	}
	_ = g.deps.Incidents.StoreIncident(cache.RansomwareIncident{
		ID: id, PID: pid, Image: p.Image, Cmdline: p.Cmdline,
		Signals: strings.Join(kinds, ","), Score: int(d.Score), Action: action, QuarantineID: qid,
	})
}

// FeedHealthy reports whether the file-activity feed is currently running.
func (g *Guard) FeedHealthy() bool { return g.feedHealthy.Load() }

// Run is the long-lived driver: subscribe to the file-activity feed and route
// each considered event to OnFileEvent. Process-start evidence arrives via
// OnProcStart (called from the procwatch dispatch hook). selfPID excludes the
// EDR's own file ops; excluded filters self/high-churn paths.
func (g *Guard) Run(ctx context.Context, feed FileActivityFeed, selfPID int, excluded func(string) bool) error {
	sink := func(fe FileEvent) {
		if Consider(fe, g.deps.Canaries.Contains, excluded, selfPID) {
			g.OnFileEvent(fe)
		}
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		g.feedHealthy.Store(true)
		err := feed.Run(ctx, sink)
		g.feedHealthy.Store(false)
		if ctx.Err() != nil {
			return ctx.Err() // clean shutdown
		}
		// Feed exited unexpectedly (e.g. an ETW session drop). Without this
		// supervision the canary sensor would die silently while status still
		// reported protection. Log and restart after a bounded backoff.
		logger.Error("UTMStack EDR: ransomware file feed exited (%v); restarting", err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(feedRetryBackoff):
		}
	}
}
