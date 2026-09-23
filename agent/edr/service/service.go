package service

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"path/filepath"

	"github.com/google/uuid"
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/edr/amsi"
	"github.com/utmstack/UTMStack/agent/edr/behavioral"
	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/feed"
	"github.com/utmstack/UTMStack/agent/edr/guard"
	"github.com/utmstack/UTMStack/agent/edr/netblock"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/agent/edr/procwatch"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/ransomware"
	"github.com/utmstack/UTMStack/agent/edr/responder"
	"github.com/utmstack/UTMStack/agent/edr/scanner"
	"github.com/utmstack/UTMStack/agent/edr/watcher"
	"github.com/utmstack/UTMStack/shared/logger"
)

type program struct {
	cancel   context.CancelFunc
	eng      *engine.Engine
	canaries *ransomware.Manager // set by startPipeline; read by writeStatus
	rwGuard  *ransomware.Guard   // set by startPipeline; read by writeStatus
	netblock *netblock.Manager   // set by startPipeline; read by writeStatus
	sigFeed  *feed.Feed          // set by startPipeline; read by writeStatus

	// Periodic full-disk scan (set by startPipeline). The runner reuses the
	// shared worker pool + exclusions; it never owns a second scan path.
	sched         *orchestrator.ScheduledScanner
	schedSchedule string // raw schedule string handed to each Trigger

	// Live-reload state (set by startPipeline). The service polls edr.json's
	// mtime and hot-applies allowlist + response-mode changes without a restart,
	// by re-Set-ing these swappable matchers and calling rwGuard.SetPolicy.
	pathEx   *orchestrator.Excluder // file-path exclusion matcher (watcher/orchestrator)
	procEx   *orchestrator.Excluder // trusted-process matcher (guard + ransomware)
	pathBase []string               // built-in path exclusions (structural; user paths appended on reload)
	procBase []string               // built-in trusted processes (structural; user processes appended on reload)
	cfgMTime int64                  // last-seen edr.json modtime (unix nanos)
}

func (p *program) Start(s service.Service) error { go p.run(); return nil }

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	// Unregister the AMSI provider so a stopped EDR doesn't leave script hosts
	// pointing at an absent provider.
	_ = amsi.Unregister()
	if p.eng != nil {
		_ = p.eng.Stop()
	}
	// Stop any in-flight scheduled scan walk (it keeps the shared pool alive for
	// the real-time watcher; this only ends the walk).
	if p.sched != nil {
		p.sched.Cancel()
	}
	return nil
}

type statusDoc struct {
	Running       bool   `json:"running"`
	EngineHealthy bool   `json:"engine_healthy"`
	SigDBVersion  string `json:"sigdb_version"`
	UpdatedAt     string `json:"updated_at"`
	Product       string `json:"product"`

	// Signature-update freshness (Task 16 — auditable update posture).
	SignatureSource     string `json:"signature_source"`
	SignatureLastUpdate string `json:"signature_last_update"`

	RansomwareEnabled     bool   `json:"ransomware_enabled"`
	RansomwareMode        string `json:"ransomware_mode"`
	CanaryCount           int    `json:"ransomware_canary_count"`
	RansomwareFeedHealthy bool   `json:"ransomware_feed_healthy"`

	// Network blocklist health (Task 13 — auditable enforcement posture).
	BlocklistEnabled             bool  `json:"blocklist_enabled"`
	BlocklistEnforce             bool  `json:"blocklist_enforce"`
	BlocklistIndicators          int   `json:"blocklist_indicators"`
	BlocklistActiveFilters       int   `json:"blocklist_active_filters"`
	BlocklistEnforcementDegraded bool  `json:"blocklist_enforcement_degraded"`
	BlocklistFeedStale           bool  `json:"blocklist_feed_stale"`
	BlocklistFeedAgeSec          int64 `json:"blocklist_feed_age_sec"`
	BlocklistRecentBlocks        int   `json:"blocklist_recent_blocks"`

	// Effective sensor toggles + allowlist sizes (auditable management posture).
	SensorFileWatcher  bool `json:"sensor_file_watcher"`
	SensorProcessGuard bool `json:"sensor_process_guard"`
	SensorAMSI         bool `json:"sensor_amsi"`
	SensorBehavioral   bool `json:"sensor_behavioral"`
	AllowPaths         int  `json:"allowlist_paths"`
	AllowProcesses     int  `json:"allowlist_processes"`
	AllowCommands      int  `json:"allowlist_commands"`

	// Effective engine tuning (ClamAV Configuration Guide §8.5 — auditable posture).
	EngineTier       string `json:"engine_tier"`
	EngineResident   bool   `json:"engine_resident"`
	EngineThreads    int    `json:"engine_threads"`
	EngineMaxFileMB  int    `json:"engine_max_file_mb"`
	EngineMaxScanMB  int    `json:"engine_max_scan_mb"`
	EngineReload     bool   `json:"engine_concurrent_reload"`
	EngineTuningNote string `json:"engine_tuning_note"`
}

func (p *program) run() {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("UTMStack EDR: config load: %v", err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	c, err := cache.Open(config.DBFile)
	if err != nil {
		logger.Error("UTMStack EDR: cache open: %v", err)
		return
	}
	defer c.Close()

	p.eng = engine.New(cfg)
	if err := p.eng.EnsureRunning(); err != nil {
		logger.Error("UTMStack EDR: %v", err)
	}
	// Keep the engine alive: restart it if it dies (it can be OOM-killed under load).
	goSafe("engine-supervisor", func() { p.eng.Supervise(ctx) })

	p.cfgMTime = configMTime()
	if cfg.Enabled {
		p.startPipeline(ctx, cfg, c)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	// Poll edr.json for CLI/hand edits and hot-apply the reloadable settings
	// (allowlists + response mode) without a restart.
	reloadTicker := time.NewTicker(3 * time.Second)
	defer reloadTicker.Stop()
	writeStatus(cfg, p.eng, p.canaries, p.rwGuard, p.netblock, p.sigFeed)
	for {
		select {
		case <-ctx.Done():
			logger.Info("UTMStack EDR: stopping")
			return
		case <-ticker.C:
			writeStatus(cfg, p.eng, p.canaries, p.rwGuard, p.netblock, p.sigFeed)
		case <-reloadTicker.C:
			p.maybeReload()
		}
	}
}

// configMTime returns edr.json's modification time in unix nanoseconds, or 0 if
// the file is absent/unstattable.
func configMTime() int64 {
	fi, err := os.Stat(config.ConfigFile)
	if err != nil {
		return 0
	}
	return fi.ModTime().UnixNano()
}

// maybeReload re-reads edr.json when its mtime changed and hot-applies the
// reloadable policy (allowlists, ransomware response mode). Structural settings
// (enabled, sensors, watched volumes, ransomware.enabled, engine, concurrency)
// still take effect on the next service restart — a change to one is logged.
func (p *program) maybeReload() {
	m := configMTime()
	if m == 0 || m == p.cfgMTime {
		return
	}
	p.cfgMTime = m
	cfg, err := config.Load()
	if err != nil {
		logger.Error("UTMStack EDR: config reload failed: %v", err)
		return
	}
	if p.pathEx != nil {
		p.pathEx.Set(append(append([]string{}, p.pathBase...), cfg.Allowlist.Paths...))
	}
	if p.procEx != nil {
		p.procEx.Set(append(append([]string{}, p.procBase...), cfg.Allowlist.Processes...))
	}
	if p.rwGuard != nil {
		p.rwGuard.SetPolicy(cfg.Ransomware.ResponseMode, cfg.Allowlist.Commands)
	}
	if p.netblock != nil {
		p.netblock.Reload(cfg, resolveSystemNets(cfg))
	}
	logger.Info("UTMStack EDR: config reloaded — allowlists (%d paths, %d processes, %d commands) and response mode=%q applied live; structural changes (enable/sensors/volumes/engine) require a restart",
		len(cfg.Allowlist.Paths), len(cfg.Allowlist.Processes), len(cfg.Allowlist.Commands), cfg.Ransomware.ResponseMode)
}

func writeStatus(cfg config.EDRConfig, eng *engine.Engine, canaries *ransomware.Manager, rwGuard *ransomware.Guard, nb *netblock.Manager, sf *feed.Feed) {
	healthy := engine.Ping(cfg.ClamdAddr) == nil
	ver := ""
	if healthy {
		raw, _ := engine.Version(cfg.ClamdAddr)
		ver = sigDBVersion(raw)
	}
	tn := eng.Tuning()
	cc := 0
	if canaries != nil {
		cc = canaries.Count()
	}
	doc := statusDoc{
		Running: true, EngineHealthy: healthy, SigDBVersion: ver,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339), Product: "UTMStack EDR",
		EngineTier: string(tn.Tier), EngineResident: tn.ResidentViable,
		EngineThreads: tn.MaxThreads, EngineMaxFileMB: tn.MaxFileSizeMB, EngineMaxScanMB: tn.MaxScanSizeMB,
		EngineReload: tn.ConcurrentReload, EngineTuningNote: tn.Reason,
		RansomwareEnabled: cfg.Ransomware.Enabled, RansomwareMode: cfg.Ransomware.ResponseMode,
		CanaryCount:           cc,
		RansomwareFeedHealthy: rwGuard != nil && rwGuard.FeedHealthy(),
		SensorFileWatcher:     cfg.Sensors.FileWatcherOn(),
		SensorProcessGuard:    cfg.Sensors.ProcessGuardOn(),
		SensorAMSI:            cfg.Sensors.AMSIOn(),
		SensorBehavioral:      cfg.Sensors.BehavioralOn(),
		AllowPaths:            len(cfg.Allowlist.Paths),
		AllowProcesses:        len(cfg.Allowlist.Processes),
		AllowCommands:         len(cfg.Allowlist.Commands),
	}
	doc.SignatureSource = feed.SignatureSource(cfg)
	if sf != nil {
		if t := sf.LastSuccess(); !t.IsZero() {
			doc.SignatureLastUpdate = t.UTC().Format(time.RFC3339)
		}
	}
	if nb != nil {
		h := nb.Health()
		doc.BlocklistEnabled = h.Enabled
		doc.BlocklistEnforce = h.Enforce
		doc.BlocklistIndicators = h.Indicators
		doc.BlocklistActiveFilters = h.ActiveFilters
		doc.BlocklistEnforcementDegraded = h.EnforcementDegraded
		doc.BlocklistFeedStale = h.FeedStale
		doc.BlocklistFeedAgeSec = h.FeedAgeSec
		doc.BlocklistRecentBlocks = h.RecentBlocks
	}
	b, _ := json.MarshalIndent(doc, "", "  ")
	_ = os.WriteFile(config.StatusFile, b, 0o644)
}

// sigDBVersion extracts just the signature-database version from the engine's
// raw VERSION banner (format: "<engine-name> <engine-ver>/<sigdb-ver>/<date>"),
// deliberately dropping the engine-name field so it never reaches a surfaced
// status file or edr-status output (branding rule).
func sigDBVersion(raw string) string {
	parts := strings.Split(raw, "/")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// startPipeline builds and starts the real-time protection pipeline:
// USN watcher → orchestrator → scanner (quarantine-wired) → spool, plus the
// signature feed. Runs only when the module is enabled.
func (p *program) startPipeline(ctx context.Context, cfg config.EDRConfig, c *cache.Cache) {
	sp, err := event.OpenSpool(config.SpoolFile, 8<<20)
	if err != nil {
		logger.Error("UTMStack EDR: spool open: %v", err)
		return
	}
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		logger.Error("UTMStack EDR: quarantine init: %v", err)
		return
	}
	// Counting wrapper: the orchestrator scans through a thin wrapper that
	// delegates to the real scanner and increments shared ScanStats, so scheduled
	// runs can report detections. This is the ONLY change to scanner construction.
	scanStats := &orchestrator.ScanStats{}
	sc := scanner.New(cfg, c, sp, store)
	counted := orchestrator.NewCountingScanner(sc, scanStats)
	// Always exclude the EDR's own working tree from the watcher. The engine
	// writes archive-extraction artifacts under EngineDir/tmp and we move
	// detections into QuarantineDir; scanning those back is pure noise (spurious
	// events, wasted re-scans, and — since the engine deletes its temp files —
	// misleading "detected" events for files that are already gone). This is the
	// §13 "exclude high-churn paths" rule applied to our own directories.
	// Built-in (structural) path exclusions: the EDR's own working tree, always
	// applied so we never scan our own churn. User `allowlist.paths` are layered
	// on top and can be hot-reloaded.
	pathBase := []string{config.InstallDir, config.EngineDir, config.SpoolDir, cfg.QuarantineDir}
	if cfg.ClamdPath != "" {
		// Also exclude the scan engine's own directory (clamd/freshclam/clamdscan)
		// from file-watch and process telemetry — self-noise + branding.
		pathBase = append(pathBase, filepath.Dir(cfg.ClamdPath))
	}
	ex := orchestrator.NewExcluder(append(append([]string{}, pathBase...), cfg.Allowlist.Paths...))
	// The orchestrator's scanner is the counting wrapper (delegates to the real
	// scanner + bumps ScanStats). The guard keeps the raw scanner.
	orch := orchestrator.New(counted, ex, cfg.ScanConcurrency, 4096)

	// Trusted-process allowlist: a conservative BUILT-IN set of FP-prone workloads
	// (backup/imaging/DR/sync + Windows VSS/Search) — the ships-with-the-system
	// list — plus any user `allowlist.processes`. User entries are ADDITIVE; the
	// built-ins are never dropped. Matched by image path/subtree/basename. Used to
	// skip the process-kill guard and to suppress ransomware scoring for these
	// whole processes — what lets a user quiet a false positive without disabling
	// the EDR.
	procBase := defaultTrustedProcesses()
	trusted := orchestrator.NewExcluder(append(append([]string{}, procBase...), cfg.Allowlist.Processes...))
	// The process-launch guard skips the EDR's own tree, path-excluded images, AND
	// trusted processes.
	procSkip := func(image string) bool { return ex.Excluded(image) || trusted.Excluded(image) }

	// Stash for live reload: the swappable matchers and their built-in bases.
	p.pathEx, p.procEx = ex, trusted
	p.pathBase, p.procBase = pathBase, procBase

	// Process table + responder, created before the orchestrator starts so the
	// file-watcher path can also terminate a malicious executable that is already
	// running (dropped-and-run malware). Registering the callback before Start
	// avoids a data race with the worker goroutines.
	tab := proctable.New()
	resp := responder.New(tab, sp, responder.OSKiller{}, responder.OSSuspender{})
	orch.SetOnMalicious(func(path, sig string) {
		for _, pid := range tab.FindByImage(path) {
			_, _ = resp.KillTree(pid, "file_watcher:"+sig)
		}
	})
	orch.Start(ctx)

	// Sensor toggles (config.Sensors): each detector can be turned off
	// independently so one noisy/problematic sensor can be disabled without
	// disabling the whole module. These are structural — a change applies on
	// restart.
	vols := cfg.WatchVolumes
	if len(vols) == 0 {
		vols = defaultFixedVolumes()
	}
	if cfg.Sensors.FileWatcherOn() {
		w := watcher.New(vols, sinkAdapter{orch}, c, ex.Excluded)
		goSafe("watcher", func() { w.Run(ctx) })
	} else {
		logger.Info("UTMStack EDR: file_watcher sensor disabled by config")
	}

	// Process-kill (§6.2 race): one process-creation watcher feeds the table + a
	// guard that suspends an unknown launch, expedites its scan, and kills the
	// whole tree on a malicious verdict. The scanner (sc) is shared with the file
	// path so a verdict is cached once and reused. (tab/resp created above.)
	// When the process_guard sensor is off, the guard is not wired (g=nil) — but
	// the process watcher may still run to feed ransomware/behavioral.
	var g *guard.Guard
	if cfg.Sensors.ProcessGuardOn() {
		g = guard.New(cfg, sc, resp)
	} else {
		logger.Info("UTMStack EDR: process_guard sensor disabled by config")
	}

	// Ransomware guard: canary tripwires + T1490 command rules → per-PID scorer →
	// suspend/kill/quarantine. Attribution for file signals comes from the ETW
	// file-activity feed; T1490 attribution comes from the procwatch dispatch hook.
	var rwGuard *ransomware.Guard
	var onProc func(procwatch.ProcStart)
	if cfg.Ransomware.Enabled {
		mgr := ransomware.NewManager(c)
		_ = mgr.Load()
		canaryDirs := append(defaultCanaryDirs(), cfg.Ransomware.CanaryDirs...)
		if n, err := mgr.Plant(canaryDirs, cfg.Ransomware.CanaryPerDir); err != nil {
			logger.Error("UTMStack EDR: canary plant: %v", err)
		} else {
			logger.Info("UTMStack EDR: planted %d canaries", n)
		}
		p.canaries = mgr
		rwGuard = ransomware.NewGuard(ransomware.GuardDeps{
			Cfg: cfg, Table: tab, Resp: resp, Spool: sp, Quar: store, Incidents: c,
			Canaries: mgr, Hash: scanner.SHA256File, Now: time.Now, NewID: uuid.NewString,
			Trusted: trusted.Excluded,
		})
		p.rwGuard = rwGuard
		onProc = func(ps procwatch.ProcStart) {
			rwGuard.OnProcStart(ps.PID, ps.PPID, ps.Image, ps.Cmdline, 0)
		}
		selfPID := os.Getpid()
		fileFeed := ransomware.NewFeed()
		goSafe("ransomware", func() { _ = rwGuard.Run(ctx, fileFeed, selfPID, ex.Excluded) })
	}

	// Network blocklist: ThreatWinds-derived IP/CIDR indicators → WFP enforcement +
	// connection audit. Feed refreshes from the server mirror; enforcement and
	// detect-only mode are governed by the blocklist config block. The never-block
	// allowlist is seeded with host-derived system nets (server IPs, and on Windows
	// resolvers/gateways) so we can never cut the host off from the platform.
	if cfg.Blocklist.Enabled {
		nb := netblock.NewManager(netblock.Deps{
			Cfg:   cfg,
			Cache: c,
			Spool: sp,
			Sys:   resolveSystemNets(cfg),
		})
		p.netblock = nb
		goSafe("netblock", func() { nb.Run(ctx) })
	}
	// The process-creation watcher runs if any consumer needs it: the process
	// guard, the ransomware T1490 hook, or behavioral process telemetry. The
	// dispatch gets a nil guard / nil telemetry spool for disabled sensors, so
	// each consumer is independently gated while the single WMI feed is shared.
	behavioralOn := cfg.Sensors.BehavioralOn()
	var telemetrySpool *event.Spool
	if behavioralOn {
		telemetrySpool = sp // process-creation telemetry (§4.9) only when behavioral is on
	}
	if g != nil || cfg.Ransomware.Enabled || behavioralOn {
		pw := procwatch.New(procwatch.NewDispatch(tab, g, telemetrySpool, procSkip, onProc))
		goSafe("procwatch", func() { pw.Run(ctx) })
	}

	p.sigFeed = feed.New(cfg, c)
	p.sigFeed.SetSpool(sp)
	goSafe("feed", func() { p.sigFeed.Run(ctx) })

	// Scripts/fileless (AMSI): named-pipe scan server + provider registration.
	if cfg.Sensors.AMSIOn() {
		amsiSc := amsi.NewScanner(cfg, sp)
		goSafe("amsi-pipe", func() { amsi.NewPipeServer(amsiSc).Run(ctx) })
		dll := filepath.Join(config.InstallDir, "utmstack_amsi.dll")
		if err := amsi.Register(dll); err != nil {
			logger.Error("UTMStack EDR: AMSI provider registration failed: %v", err)
		}
	} else {
		logger.Info("UTMStack EDR: amsi sensor disabled by config")
	}

	// Behavioral telemetry: PowerShell script-block (4104) logs (+ process
	// telemetry above). Chatty — this is the sensor most likely to be turned off.
	if behavioralOn {
		goSafe("behavioral", func() { behavioral.NewPSLogReader(sp).Run(ctx) })
	} else {
		logger.Info("UTMStack EDR: behavioral sensor disabled by config")
	}

	// Quarantine retention: periodically purge quarantined items older than the
	// configured retention (0 = keep forever). Makes quarantine_retention_days real.
	goSafe("quarantine-retention", func() { runRetention(ctx, store, cfg.QuarantineDays) })

	// Y1.6 periodic full-disk scan. Reuses the shared orchestrator pool +
	// exclusions (no second scan path). Paths empty → all fixed volumes (same
	// rule as the watcher). ScheduledScan is structural: a change takes effect on
	// the next module restart (consistent with the other structural keys) — it is
	// NOT hot-reloaded. Failures are contained by goSafe and never take the
	// service down.
	if cfg.ScheduledScan.Enabled {
		ssPaths := cfg.ScheduledScan.Paths
		if len(ssPaths) == 0 {
			ssPaths = vols
		}
		p.sched = orchestrator.NewScheduledScanner(orchestrator.ScheduledScanParams{
			Service:  ctx,
			Orch:     orch,
			Ex:       ex,
			Paths:    ssPaths,
			Batch:    cfg.ScheduledScan.BatchSize,
			Throttle: time.Duration(cfg.ScheduledScan.ThrottleMs) * time.Millisecond,
			Stats:    scanStats,
			Spool:    sp,
		})
		p.schedSchedule = cfg.ScheduledScan.Schedule
		goSafe("scheduled_scan", func() { p.scheduledScanLoop(ctx) })
	} else {
		logger.Info("UTMStack EDR: scheduled scan disabled by config")
	}

	logger.Info("UTMStack EDR: real-time protection started (%d volumes, %d workers)", len(vols), cfg.ScanConcurrency)
}

// scheduledScanLoop runs the periodic full-disk scan: one run immediately on
// start (when enabled), then every parsed `every:` duration. It honors the
// service context on shutdown. A bad schedule string is logged and the loop
// exits (goSafe contains any panic).
func (p *program) scheduledScanLoop(ctx context.Context) {
	period, err := orchestrator.ParseSchedule(p.schedSchedule)
	if err != nil {
		logger.Error("UTMStack EDR: scheduled scan: %v", err)
		return
	}
	// Run once on start, then on each period tick.
	p.sched.Trigger(p.schedSchedule)
	t := time.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.sched.Trigger(p.schedSchedule)
		}
	}
}

// runRetention purges quarantined items older than retentionDays every 6h (and
// once shortly after start). retentionDays <= 0 disables auto-purge.
func runRetention(ctx context.Context, store *quarantine.Store, retentionDays int) {
	if retentionDays <= 0 {
		return
	}
	sweep := func() {
		n, err := store.PurgeExpired(retentionDays, time.Now())
		if err != nil {
			logger.Error("UTMStack EDR: quarantine retention sweep: %v", err)
		} else if n > 0 {
			logger.Info("UTMStack EDR: quarantine retention purged %d item(s) older than %d days", n, retentionDays)
		}
	}
	t := time.NewTicker(6 * time.Hour)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return
	case <-time.After(1 * time.Minute): // initial sweep after startup settles
		sweep()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			sweep()
		}
	}
}

// goSafe runs fn in a goroutine with panic recovery, so a panic in one pipeline
// component (e.g. the WMI process watcher) is logged and contained instead of
// crashing the whole SYSTEM service. Mirrors the agent's goSafe.
func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("UTMStack EDR: panic in %s: %v", name, r)
			}
		}()
		fn()
	}()
}

type sinkAdapter struct{ o *orchestrator.Orchestrator }

func (s sinkAdapter) Enqueue(path, op string) {
	if !s.o.Enqueue(orchestrator.FileEvent{Path: path, Op: op}) {
		logger.Debug(100, "UTMStack EDR: scan queue full, dropped %s", path)
	}
}

func RunService() {
	prg := new(program)
	s, err := service.New(prg, GetConfigServ())
	if err != nil {
		logger.Fatal("UTMStack EDR: create service: %v", err)
	}
	if err := s.Run(); err != nil {
		logger.Fatal("UTMStack EDR: run service: %v", err)
	}
}
