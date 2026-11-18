package ransomware

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/shared/logger"
)

// feedRetryBackoff is the delay before restarting a file feed that exited
// unexpectedly. A package var so tests can shorten it.
var feedRetryBackoff = 5 * time.Second

// containedTTL bounds trailing-event deduplication after a confirmed culprit
// kill, or after a newer native birth proves the original culprit gone. Failed
// responses to a possibly live culprit release their reservation for fresh
// evidence. In-flight response reservations do not expire under this TTL.
// A newly observed process identity is eligible immediately after PID reuse.
const containedTTL = 30 * time.Second

type treeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
}
type appender interface{ Append(string) error }
type quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
}
type incidentSink interface {
	StoreIncident(cache.RansomwareIncident) error
}
type canarySet interface{ Contains(path string) bool }
type canaryIdentitySet interface {
	ContainsFileIdentity(device, inode uint64) bool
}

func (g *Guard) isCanaryEvent(fe FileEvent) bool {
	if fe.FileIdentityRequired {
		identified, ok := g.deps.Canaries.(canaryIdentitySet)
		return ok && identified.ContainsFileIdentity(fe.Device, fe.Inode)
	}
	return g.deps.Canaries.Contains(fe.Path)
}

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
	// LiveProcess reads a running process's image path and parent from the
	// operating system. The WMI-fed table polls once per second, so a process
	// that tampers right after it starts is not there yet. nil = table only.
	LiveProcess func(pid int) (image string, ppid int, err error)
}

// procOf returns the table entry for pid, filling a missing image and parent
// from the live process when the table has not recorded it yet.
func (g *Guard) procOf(pid int) proctable.Proc {
	p, ok := g.deps.Table.Get(pid)
	if (ok && p.Image != "") || g.deps.LiveProcess == nil {
		return p
	}
	if image, ppid, err := g.deps.LiveProcess(pid); err == nil && image != "" {
		p.PID, p.Image = pid, image
		if p.PPID == 0 {
			p.PPID = ppid
		}
	}
	return p
}

// isTrusted resolves a PID's image (table first, then the live process) and
// reports whether it is allowlisted. A trusted process is never scored.
func (g *Guard) isTrusted(pid int) bool {
	if g.deps.Trusted == nil {
		return false
	}
	if p := g.procOf(pid); p.Image != "" {
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
	fmu         sync.RWMutex
	liveFeed    FileActivityFeed
	smu         sync.Mutex
	suspended   map[int]int64

	// Hot-reloadable policy: the response mode and command allowlist can change
	// at runtime (CLI edit → service reload) without restarting the guard. The
	// scorer thresholds/decay are fixed at construction (structural).
	pmu          sync.RWMutex
	responseMode string
	commands     []string

	// contained reserves in-flight responses and deduplicates confirmed kills.
	// A failed attempt releases its matching reservation for fresh evidence.
	cmu         sync.Mutex
	contained   map[int]containedMark
	claimSerial uint64
}

type containedMark struct {
	pending    bool
	token      uint64 // distinguishes overlapping attempts even for a reused PID
	osIdentity int64
	gen        int64 // process generation (proctable StartTS) at containment; 0 = unknown
	at         time.Time
}

// A blocked response retains its reservation; expiry cannot authorize a second
// concurrent response. The response owner remains responsible for returning.
func (m containedMark) expired(now time.Time) bool {
	return !m.pending && now.Sub(m.at) > containedTTL
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
		contained:    map[int]containedMark{},
		suspended:    map[int]int64{},
	}
}

// isContained reports whether pid has an in-flight response or a retained
// completed-instance deduplication mark.
func (g *Guard) isContained(pid int) bool {
	g.cmu.Lock()
	defer g.cmu.Unlock()
	return g.containedLocked(pid)
}

func (g *Guard) containedLocked(pid int) bool {
	m, ok := g.contained[pid]
	if !ok {
		return false
	}
	if m.expired(g.deps.Now()) {
		delete(g.contained, pid)
		return false
	}
	// The responder drops a killed PID from the table, so a table entry with a
	// different generation is a new process that reused the PID.
	if m.osIdentity != 0 {
		if p, ok := g.deps.Table.Get(pid); ok && p.OSIdentity != 0 && p.OSIdentity != m.osIdentity {
			delete(g.contained, pid)
			return false
		}
	} else if m.gen != 0 {
		if p, ok := g.deps.Table.Get(pid); ok && p.StartTS != 0 && p.StartTS != m.gen {
			delete(g.contained, pid)
			return false
		}
	}
	return true
}

// claimContainment reserves a token for one response attempt. Evidence arriving
// concurrently from file and process feeds cannot start a duplicate response.
func (g *Guard) claimContainment(pid int, gen int64, identity int64) (uint64, bool) {
	g.cmu.Lock()
	defer g.cmu.Unlock()
	if identity > 0 {
		if m, ok := g.contained[pid]; ok && m.osIdentity == identity && !m.expired(g.deps.Now()) {
			return 0, false
		}
		delete(g.contained, pid)
	} else if g.containedLocked(pid) {
		return 0, false
	}
	g.markLocked(pid, gen)
	m := g.contained[pid]
	m.osIdentity = identity
	m.pending = true
	g.contained[pid] = m
	return m.token, true
}

func (g *Guard) markContained(pid int, gen int64) {
	g.cmu.Lock()
	defer g.cmu.Unlock()
	g.markLocked(pid, gen)
}

func (g *Guard) markLocked(pid int, gen int64) {
	now := g.deps.Now()
	for p, m := range g.contained {
		if m.expired(now) {
			delete(g.contained, p)
		}
	}
	g.claimSerial++
	g.contained[pid] = containedMark{gen: gen, at: now, token: g.claimSerial}
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
	if g.isTrusted(culprit) || g.isContained(culprit) {
		return
	}
	ev := Evidence{PID: culprit, Gen: g.genOf(culprit, 0), Kind: KindT1490,
		Weight: DefaultWeights[KindT1490], Detail: rule, TS: g.deps.Now()}
	g.handle(g.scorer.Add(ev), culprit)
}

// OnFileEvent runs the canary sensor for a per-process file op. Only an op that
// modifies the canary is evidence; opening or reading one is not.
func (g *Guard) OnFileEvent(fe FileEvent) {
	if !g.cfg.Ransomware.Enabled || !fe.Op.Modifying() {
		return
	}
	if !g.isCanaryEvent(fe) {
		return // v1: only canary tampering produces evidence from the file stream
	}
	// A trusted process (backup/sync/indexer) touching a canary is not scored,
	// and an already-contained process's trailing events are not re-scored.
	if fe.IdentityRequired || fe.OSIdentity > 0 {
		if fe.OSIdentity <= 0 {
			// Retain the canary alert when an origin exited before kernel
			// pidfd delivery. Numeric PID reuse never authorizes a response.
			p := proctable.Proc{PID: fe.PID, IdentityRequired: true}
			d := Decision{Escalation: EscKill, Score: DefaultWeights[KindCanary], TopSignal: "canary_identity_unavailable", Kinds: []SignalKind{KindCanary}}
			g.emit(event.ActionRansomwareSuspected, event.ProcInfo{PID: fe.PID}, d, "critical")
			g.record("identity_unavailable", fe.PID, p, d, "")
			return
		}
		p := proctable.Proc{PID: fe.PID, PPID: fe.PPID, Image: fe.Image, Cmdline: fe.Cmdline, OSIdentity: fe.OSIdentity, IdentityRequired: true}
		if prior, ok := g.deps.Table.Get(fe.PID); ok && prior.OSIdentity == fe.OSIdentity {
			p.StartTS = prior.StartTS
			if p.Image == "" {
				p.Image = prior.Image
			}
			if p.Cmdline == "" {
				p.Cmdline = prior.Cmdline
			}
		}
		if p.Image != "" && g.deps.Trusted != nil && g.deps.Trusted(p.Image) {
			return
		}
		if g.isContainedObservation(p) {
			return
		}
		base := fe.Path
		if i := strings.LastIndexAny(base, `/\`); i >= 0 {
			base = base[i+1:]
		}
		ev := Evidence{PID: p.PID, Gen: p.OSIdentity, Kind: KindCanary, Weight: DefaultWeights[KindCanary], Detail: base, TS: g.deps.Now()}
		g.handleObserved(g.scorer.Add(ev), p)
		return
	}
	if g.isTrusted(fe.PID) || g.isContained(fe.PID) {
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
func (g *Guard) handle(d Decision, pid int) { p := g.procOf(pid); p.PID = pid; g.handleObserved(d, p) }

func (g *Guard) handleObserved(d Decision, p proctable.Proc) {
	pid := p.PID
	if d.Escalation == EscNone {
		return
	}
	// Resolve the culprit before any kill: the live lookup needs the process.
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
	// kill mode takes no containment action until the kill threshold. The event
	// only claims suspicion; the incident records whether the suspend worked.
	if d.Escalation == EscSuspend {
		if g.isContained(pid) {
			return
		}
		outcome := "suspected"
		if mode == "suspend" {
			outcome = "suspended"
			if err := g.suspendObservation(p); err != nil {
				outcome = "suspend_failed"
				logger.Error("UTMStack EDR: ransomware suspend %d: %v", pid, err)
			}
		}
		g.emit(event.ActionRansomwareSuspected, pinfo, d, "high")
		g.record(outcome, pid, p, d, "")
		return
	}

	// Reserve the response while it is in flight. Only a confirmed culprit kill
	// retains the completed mark; a later fresh event may retry a failed kill.
	token, claimed := g.claimContainment(pid, p.StartTS, p.OSIdentity)
	if !claimed {
		return
	}
	// Capture observed child identities before the response. The responder
	// removes entries for processes it kills; later PID lookups can be replacements.
	children := map[int]proctable.Proc{}
	for _, c := range g.deps.Table.Descendants(pid) {
		if cp, ok := g.deps.Table.Get(c); ok {
			children[c] = cp
		}
	}
	if mode == "suspend" {
		_ = g.suspendObservation(p) // freeze the suspect to reduce further encryption during the kill
	}
	killed, killErr := g.killObservation(p, "ransomware:"+d.TopSignal)
	if p.IdentityRequired {
		if killErr != nil || !slices.Contains(killed, pid) {
			_ = g.resumeObservation(p)
		} else {
			g.smu.Lock()
			delete(g.suspended, pid)
			g.smu.Unlock()
		}
	}
	if killErr != nil {
		logger.Error("UTMStack EDR: ransomware kill tree %d: %v", pid, killErr)
	}
	for _, k := range killed {
		if k != pid {
			if child, captured := children[k]; captured {
				g.completeDescendant(child)
			}
		}
	}
	qid := ""
	quarFailed := false
	switch {
	case p.IdentityRequired:
		// A behavior verdict is tied to a process instance, not an immutable
		// executable snapshot. Do not move a potentially replaced pathname.
		logger.Info("UTMStack EDR: ransomware process %d response recorded; executable path is not identity-pinned for quarantine", pid)
	case p.Image == "":
		logger.Info("UTMStack EDR: ransomware pid %d: image unknown, nothing to quarantine", pid)
	case isWindowsDirImage(p.Image):
		logger.Info("UTMStack EDR: ransomware pid %d: %s is an operating-system image under the Windows directory; image not quarantined", pid, p.Image)
	default:
		sha := ""
		if g.deps.Hash != nil {
			sha, _ = g.deps.Hash(p.Image)
		}
		if id, err := g.deps.Quar.Quarantine(p.Image, sha, "ransomware:"+d.TopSignal); err == nil {
			qid = id
		} else {
			quarFailed = true
			var contained *quarantine.ContainmentError
			if errors.As(err, &contained) {
				qid = contained.QuarantineID
			}
			logger.Error("UTMStack EDR: ransomware quarantine %s: %v", p.Image, err)
		}
	}
	// Honest reporting (project rule): the event and the incident only claim
	// "contained" if the kill actually terminated the culprit itself and
	// quarantine did not fail. responder.KillTree logs+swallows per-PID kill
	// errors and returns nil, so the honest signal is whether `pid` is in the
	// `killed` slice — not killErr. Killing only a child (e.g. its conhost.exe)
	// leaves the encryptor running.
	action, evAction := "contained", event.ActionRansomwareContained
	if killErr != nil || !slices.Contains(killed, pid) || quarFailed {
		action, evAction = "contain_failed", event.ActionRansomwareContainFailed
	}
	g.emit(evAction, pinfo, d, "critical")
	g.record(action, pid, p, d, qid)
	g.finishContainment(p, token, slices.Contains(killed, pid))
}

// completeDescendant uses only the pre-response observation. Numeric PID-only
// response results cannot authorize erasing a replacement's claim or evidence.
// The response API does not return child native handles/identities, so an
// uncaptured or unidentified child deliberately acquires no deduplication mark.
func (g *Guard) completeDescendant(p proctable.Proc) {
	gen := p.StartTS
	if p.OSIdentity > 0 {
		gen = p.OSIdentity
	}
	if gen <= 0 {
		return
	}
	g.cmu.Lock()
	defer g.cmu.Unlock()
	if current, exists := g.deps.Table.Get(p.PID); exists {
		if p.OSIdentity > 0 && current.OSIdentity != p.OSIdentity {
			return
		}
		if p.OSIdentity == 0 && (current.OSIdentity != 0 || current.StartTS != p.StartTS) {
			return
		}
		if p.StartTS > 0 && current.StartTS != p.StartTS {
			return
		}
	}
	if mark, exists := g.contained[p.PID]; exists {
		if !mark.expired(g.deps.Now()) || mark.osIdentity > p.OSIdentity || mark.gen > p.StartTS {
			return
		}
	}
	g.markLocked(p.PID, p.StartTS)
	mark := g.contained[p.PID]
	mark.osIdentity = p.OSIdentity
	g.contained[p.PID] = mark
	g.scorer.ForgetGeneration(p.PID, gen)
}

// finishContainment cannot erase a later claim or a newer process's evidence.
// A table entry born after this native observation proves its original process
// is gone; retain its trailing-event dedup instead of retrying an obsolete PID.
func (g *Guard) finishContainment(p proctable.Proc, token uint64, culpritKilled bool) {
	g.cmu.Lock()
	defer g.cmu.Unlock()
	mark, ok := g.contained[p.PID]
	if !ok || mark.token != token {
		return
	}
	mark.pending = false
	mark.at = g.deps.Now()
	g.contained[p.PID] = mark
	gen := p.StartTS
	if p.IdentityRequired || p.OSIdentity > 0 {
		gen = p.OSIdentity
	}
	g.scorer.ForgetGeneration(p.PID, gen)
	if culpritKilled {
		return
	}
	if p.OSIdentity > 0 {
		if current, ok := g.deps.Table.Get(p.PID); ok && current.OSIdentity > p.OSIdentity {
			return
		}
	}
	delete(g.contained, p.PID)
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
func (g *Guard) FeedHealthy() bool {
	if !g.feedHealthy.Load() {
		return false
	}
	g.fmu.RLock()
	f := g.liveFeed
	g.fmu.RUnlock()
	if h, ok := f.(interface{ Health() error }); ok {
		return h.Health() == nil
	}
	return true
}

// Health combines live feed readiness with sticky evidence-admission failures.
// FeedHealthy alone intentionally describes only the operating-system stream.
func (g *Guard) Health() error {
	var feedErr error
	if !g.feedHealthy.Load() {
		feedErr = errors.New("ransomware file feed is not running")
	} else {
		g.fmu.RLock()
		f := g.liveFeed
		g.fmu.RUnlock()
		if h, ok := f.(interface{ Health() error }); ok {
			feedErr = h.Health()
		}
	}
	return errors.Join(feedErr, g.scorer.Health())
}

// Run is the long-lived driver: subscribe to the file-activity feed and route
// each considered event to OnFileEvent. Process-start evidence arrives via
// OnProcStart (called from the procwatch dispatch hook). selfPID excludes the
// EDR's own file ops; excluded filters self/high-churn paths.
func (g *Guard) Run(ctx context.Context, feed FileActivityFeed, selfPID int, excluded func(string) bool) error {
	g.fmu.Lock()
	g.liveFeed = feed
	g.fmu.Unlock()
	if f, ok := feed.(decoyAware); ok {
		f.setInterest(g.deps.Canaries.Contains)
	}
	if f, ok := feed.(interface {
		setIdentityInterest(func(uint64, uint64) bool)
	}); ok {
		identified, ok := g.deps.Canaries.(canaryIdentitySet)
		if !ok {
			return errors.New("fanotify canary feed requires pinned canary file identities")
		}
		f.setIdentityInterest(identified.ContainsFileIdentity)
	}
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// A feed may retain a callback after Run returns. Give each attempt
		// its own admission lifetime so retries cannot reactivate old sinks.
		attemptCtx, cancelAttempt := context.WithCancel(ctx)
		sink := func(fe FileEvent) {
			if attemptCtx.Err() != nil {
				return
			}
			// A verified Unix file identity remains a canary through a rename or
			// hard link, including in an otherwise excluded directory.
			admitted := fe.FileIdentityRequired && g.isCanaryEvent(fe) &&
				fe.PID != selfPID && fe.PID > maxKernelPID && fe.Op.Modifying()
			if (admitted || Consider(fe, g.deps.Canaries.Contains, excluded, selfPID)) && attemptCtx.Err() == nil {
				g.OnFileEvent(fe)
			}
		}
		g.feedHealthy.Store(true)
		err := feed.Run(attemptCtx, sink)
		cancelAttempt()
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
