package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"sync"
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
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
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
	"github.com/utmstack/UTMStack/agent/edr/scheduled"
	"github.com/utmstack/UTMStack/agent/edr/watcher"
	"github.com/utmstack/UTMStack/shared/logger"
)

type program struct {
	cancel         context.CancelFunc
	done           chan struct{}
	instanceID     string
	startedAt      time.Time
	serviceSHA256  string
	eng            *engine.Engine
	amsiPipe       *amsi.PipeServer
	amsiRegistered bool
	canaries       *ransomware.Manager // set by startPipeline; read by writeStatus
	rwGuard        *ransomware.Guard   // set by startPipeline; read by writeStatus
	netblock       *netblock.Manager   // set by startPipeline; read by writeStatus
	sigFeed        *feed.Feed          // set by startPipeline; read by writeStatus
	schedule       *scheduled.Scheduler
	scanQueue      *orchestrator.Orchestrator
	eventSpool     *event.Spool
	fileAccess     fanotify.Runner
	backgroundCap  int
	nativeWG       sync.WaitGroup // permission owners must finish before engine/cache shutdown
	sensorHealth   map[string]interface{ Health() error }

	// Live-reload state (set by startPipeline). The service polls edr.json's
	// mtime and hot-applies allowlist + response-mode changes without a restart,
	// by re-Set-ing these swappable matchers and calling rwGuard.SetPolicy.
	pathEx   *orchestrator.Excluder // file-path exclusion matcher (watcher/orchestrator)
	procEx   *orchestrator.Excluder // trusted-process matcher (guard + ransomware)
	pathBase []string               // built-in path exclusions (structural; user paths appended on reload)
	procBase []string               // built-in trusted processes (structural; user processes appended on reload)
	cfgMTime int64                  // last-seen edr.json modtime (unix nanos)
}

func (p *program) Start(s service.Service) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan struct{})
	go func() { defer close(p.done); p.run(ctx) }()
	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	// Unregister the AMSI provider so a stopped EDR doesn't leave script hosts
	// pointing at an absent provider.
	_ = amsi.Unregister()
	if p.done != nil {
		select {
		case <-p.done:
		case <-time.After(5 * time.Second):
			return errors.New("EDR shutdown still waiting for protection workers")
		}
	}
	return nil
}

type statusDoc struct {
	Platform             string                 `json:"platform"`
	InstanceID           string                 `json:"instance_id"`
	ProcessID            int                    `json:"process_id"`
	StartedAt            string                 `json:"started_at"`
	ServiceSHA256        string                 `json:"service_executable_sha256"`
	AMSILocalReady       bool                   `json:"amsi_local_ready"`
	SensorStates         map[string]sensorState `json:"sensor_states"`
	ScheduledScan        *scheduled.Status      `json:"scheduled_scan,omitempty"`
	ScanQueue            *orchestrator.Stats    `json:"scan_queue,omitempty"`
	BackgroundSlots      int                    `json:"background_scan_concurrency,omitempty"`
	EventSpool           *event.SpoolStats      `json:"event_spool,omitempty"`
	FileAccess           *fanotify.Stats        `json:"file_access,omitempty"`
	Running              bool                   `json:"running"`
	EngineHealthy        bool                   `json:"engine_healthy"`
	SigDBVersion         string                 `json:"sigdb_version"`
	UpdatedAt            string                 `json:"updated_at"`
	Product              string                 `json:"product"`
	EngineImplementation string                 `json:"engine_implementation"`
	EngineCompatibility  string                 `json:"engine_compatibility"`
	SignatureCount       int                    `json:"signature_count"`
	// Loaded rules that need analysis this engine does not perform, by
	// reason. Each makes an object incomplete only where it could apply.
	SignatureUnknownRules    int            `json:"signature_unknown_rules"`
	SignatureUnknownByReason map[string]int `json:"signature_unknown_by_reason,omitempty"`

	// Signature-update freshness (Task 16 — auditable update posture).
	// SignatureLastUpdate is the last successful refresh, falling back to the
	// serving generation's activation time after a restart.
	SignatureSource              string `json:"signature_source"`
	SignatureLastUpdate          string `json:"signature_last_update"`
	SignatureGenerationActivated string `json:"signature_generation_activated,omitempty"`
	SignatureUpdateState         string `json:"signature_update_state,omitempty"` // feed.Update*
	SignatureUpdateError         string `json:"signature_update_error,omitempty"`

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

	// Effective native scan budgets. Legacy resident/reload status fields are
	// observed engine-loaded state and atomic-reload support, not RAM guarantees.
	EngineTier          string `json:"engine_tier"`
	EngineResident      bool   `json:"engine_resident"`
	EngineThreads       int    `json:"engine_threads"`
	EngineMaxFileMB     int    `json:"engine_max_file_mb"`
	EngineMaxScanMB     int    `json:"engine_max_scan_mb"`
	EngineMaxRecursion  int    `json:"engine_max_recursion"`
	EngineMaxFiles      int    `json:"engine_max_files"`
	EngineMaxScanTimeMS int    `json:"engine_max_scan_time_ms"`
	EngineReload        bool   `json:"engine_concurrent_reload"`
	EngineTuningNote    string `json:"engine_tuning_note"`
}

func (p *program) run(ctx context.Context) {
	defer func() { writeStoppedStatus(p.scanQueue, p.eventSpool) }()
	p.instanceID = uuid.NewString()
	p.startedAt = time.Now().UTC()
	if hash, err := runningExecutableSHA256(); err != nil {
		logger.Error("UTMStack EDR: hashing running executable for release readiness: %v", err)
	} else {
		p.serviceSHA256 = hash
	}
	cfg, err := config.Load()
	if err != nil {
		logger.Error("UTMStack EDR: config load: %v", err)
		return
	}
	if ctx.Err() != nil {
		return
	}

	c, err := cache.Open(config.DBFile)
	if err != nil {
		logger.Error("UTMStack EDR: cache open: %v", err)
		return
	}
	defer c.Close()

	p.eng = engine.New(cfg)
	defer p.eng.Stop()
	defer func() {
		if p.eventSpool != nil {
			_ = p.eventSpool.Close()
		}
	}()
	defer cleanupResponse()
	defer func() {
		if p.canaries != nil {
			if err := p.canaries.Close(); err != nil {
				logger.Error("UTMStack EDR: closing canary identity pins: %v", err)
			}
		}
	}()
	defer p.nativeWG.Wait()
	// Upgraded endpoints still hold the retired updater's downloads; move them
	// aside once so every reload does not try to compile them.
	if n, err := engine.RetireLegacyUpdaterFiles(p.eng.DatabaseDir()); err != nil {
		logger.Error("UTMStack EDR: moving retired signature updater files: %v", err)
	} else if n > 0 {
		logger.Info("UTMStack EDR: moved %d retired signature updater files out of the signature directory", n)
	}
	if err := p.eng.EnsureRunning(); err != nil {
		logger.Error("UTMStack EDR: %v", err)
	}
	// Retry loading a database if initial provisioning has not completed.
	p.goPipeline("engine-supervisor", func() { p.eng.Supervise(ctx) })

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
	writeStatus(cfg, p.eng, p.canaries, p.rwGuard, p.netblock, p.sigFeed, p)
	for {
		select {
		case <-ctx.Done():
			logger.Info("UTMStack EDR: stopping")
			return
		case <-ticker.C:
			writeStatus(cfg, p.eng, p.canaries, p.rwGuard, p.netblock, p.sigFeed, p)
		case <-reloadTicker.C:
			p.maybeReload(&cfg)
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
func (p *program) maybeReload(active *config.EDRConfig) {
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
	// Status must reflect only the policy actually applied live. Keep structural
	// values from startup until a restart activates them.
	if active != nil {
		active.Allowlist = cfg.Allowlist
		active.Ransomware.ResponseMode = cfg.Ransomware.ResponseMode
		active.Allowlist.Commands = cfg.Allowlist.Commands
		if p.netblock != nil {
			active.Blocklist = cfg.Blocklist
		}
	}
	logger.Info("UTMStack EDR: config reloaded — allowlists (%d paths, %d processes, %d commands) and response mode=%q applied live; structural changes (enable/sensors/volumes/engine) require a restart",
		len(cfg.Allowlist.Paths), len(cfg.Allowlist.Processes), len(cfg.Allowlist.Commands), cfg.Ransomware.ResponseMode)
}

func unknownRuleStatus(doc *statusDoc, byReason map[string]int) {
	doc.SignatureUnknownRules = 0
	for _, n := range byReason {
		doc.SignatureUnknownRules += n
	}
	doc.SignatureUnknownByReason = byReason
}

func writeStatus(cfg config.EDRConfig, eng *engine.Engine, canaries *ransomware.Manager, rwGuard *ransomware.Guard, nb *netblock.Manager, sf *feed.Feed, p *program) {
	ver := eng.Version()
	healthy := ver != ""
	tn := eng.Tuning()
	cc := 0
	if canaries != nil {
		cc = canaries.Count()
	}
	doc := statusDoc{
		Platform:   runtime.GOOS,
		InstanceID: p.instanceID, ProcessID: os.Getpid(), StartedAt: p.startedAt.Format(time.RFC3339Nano),
		ServiceSHA256: p.serviceSHA256,
		Running:       true, EngineHealthy: healthy, SigDBVersion: ver,
		EngineImplementation: engine.ImplementationVersion,
		EngineCompatibility:  "partial-signature-and-format-support",
		SignatureCount:       eng.SignatureCount(),
		UpdatedAt:            time.Now().UTC().Format(time.RFC3339), Product: "UTMStack EDR",
		EngineTier: string(tn.Tier), EngineResident: healthy,
		EngineThreads: tn.MaxThreads, EngineMaxFileMB: tn.MaxFileSizeMB, EngineMaxScanMB: tn.MaxScanSizeMB,
		EngineMaxRecursion: tn.MaxRecursion, EngineMaxFiles: tn.MaxFiles, EngineMaxScanTimeMS: tn.MaxScanTimeMS,
		EngineReload: true, EngineTuningNote: tn.Reason,
		BackgroundSlots:   p.backgroundCap,
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
	if p.amsiRegistered && p.amsiPipe != nil && p.amsiPipe.Health() == nil {
		doc.AMSILocalReady = true
	}
	doc.SensorStates = platformSensors(cfg, runtime.GOOS, p.sensorHealth)
	effective := func(name string) bool {
		st := doc.SensorStates[name]
		return cfg.Enabled && st.Requested && st.Supported
	}
	doc.SensorFileWatcher = effective("file_watcher")
	doc.SensorProcessGuard = effective("process_guard")
	doc.SensorAMSI = effective("amsi")
	doc.SensorBehavioral = effective("behavioral")
	doc.RansomwareEnabled = effective("ransomware")
	if p.schedule != nil {
		st := p.schedule.Status()
		doc.ScheduledScan = &st
	}
	if p.fileAccess != nil {
		st := p.fileAccess.Stats()
		doc.FileAccess = &st
	}
	if p.scanQueue != nil {
		st := p.scanQueue.Snapshot()
		doc.ScanQueue = &st
	}
	if p.eventSpool != nil {
		st := p.eventSpool.Snapshot()
		doc.EventSpool = &st
	}
	unknownRuleStatus(&doc, eng.UnknownRules())
	doc.SignatureSource = feed.SignatureSource(cfg)
	sequence, hash := eng.FeedState()
	signatureStatus(&doc, sf, feed.GenerationActivatedAt(eng.DatabaseDir(), sequence, hash))
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
	if err := writeStatusFile(b); err != nil {
		logger.Error("UTMStack EDR: writing runtime status: %v", err)
	}
}

// signatureStatus fills the signature-update fields of doc. activated is when
// the serving signed generation was activated (zero when none); sf is nil
// while the module is disabled.
func signatureStatus(doc *statusDoc, sf *feed.Feed, activated time.Time) {
	last := activated
	if !activated.IsZero() {
		doc.SignatureGenerationActivated = activated.UTC().Format(time.RFC3339)
	}
	if sf != nil {
		doc.SignatureUpdateError = sf.LastError()
		doc.SignatureUpdateState = sf.UpdateState()
		if t := sf.LastSuccess(); t.After(last) {
			last = t
		}
	}
	if !last.IsZero() {
		doc.SignatureLastUpdate = last.UTC().Format(time.RFC3339)
	}
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
		_ = sp.Close()
		return
	}
	p.eventSpool = sp
	scanCfg, backgroundLimit := backgroundScanConfig(cfg, p.eng.Tuning().MaxThreads)
	p.backgroundCap = backgroundLimit
	sc := scanner.New(scanCfg, c, sp, store, p.eng)
	// Exclude our install, signatures, spool and quarantine activity from
	// monitoring. Container decoding stays in memory; it creates no temp files.
	// User allowlists are layered on these structural exclusions.
	pathBase := []string{config.InstallDir, config.EngineDir, config.SpoolDir, cfg.QuarantineDir}
	if cfg.SignatureDBDir != "" {
		pathBase = append(pathBase, cfg.SignatureDBDir)
	}
	ex := orchestrator.NewExcluder(append(append([]string{}, pathBase...), cfg.Allowlist.Paths...))
	orch := orchestrator.New(sc, ex, min(cfg.ScanConcurrency, scanCfg.ScanConcurrency), 4096)

	// Trusted-process allowlist: a conservative BUILT-IN set of FP-prone workloads
	// (backup/imaging/DR/sync + Windows VSS/Search) — the ships-with-the-system
	// list — plus any user `allowlist.processes`. User entries are ADDITIVE; the
	// built-ins are never dropped. Matched by image path/subtree/basename. Used to
	// skip the process-kill guard and to suppress ransomware scoring for these
	// whole processes — what lets a user quiet a false positive without disabling
	// the EDR.
	procBase := defaultTrustedProcesses()
	trusted := orchestrator.NewExcluder(append(append([]string{}, procBase...), cfg.Allowlist.Processes...))
	// The process-launch guard and process telemetry skip images in the EDR's own
	// tree, under allowlist.paths, AND trusted processes. A path allowlist entry
	// therefore exempts the programs started from it as well as its files: that
	// is the intended false-positive control for a trusted tool tree, not a gap.
	// T1490 command matching runs before this filter, and ransomware scoring only
	// honors the trusted-process list.
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
		for _, observed := range tab.FindProcessesByImage(path) {
			killObserved(resp, observed, "file_watcher:"+sig)
		}
	})
	orch.Start(ctx)
	p.scanQueue = orch
	p.goNative("scan-workers", func() { <-ctx.Done(); orch.Stop() })
	p.sensorHealth = make(map[string]interface{ Health() error })
	if cfg.ScheduledScan.Enabled {
		sched, err := scheduled.New(cfg.ScheduledScan, func(ctx context.Context, path string) error {
			for {
				if err := ctx.Err(); err != nil {
					return err
				}
				if orch.Enqueue(orchestrator.FileEvent{Path: path, Op: "scheduled", Source: event.SourceScheduledScan}) {
					return nil
				}
				t := time.NewTimer(25 * time.Millisecond)
				select {
				case <-ctx.Done():
					t.Stop()
					return ctx.Err()
				case <-t.C:
				}
			}
		}, ex.Excluded)
		if err != nil {
			logger.Error("UTMStack EDR: scheduled scan: %v", err)
		} else {
			p.schedule = sched
			p.goPipeline("scheduled-scan", func() { sched.Run(ctx) })
		}
	}

	// Sensor toggles (config.Sensors): each detector can be turned off
	// independently so one noisy/problematic sensor can be disabled without
	// disabling the whole module. These are structural — a change applies on
	// restart.
	vols := cfg.WatchVolumes
	if len(vols) == 0 {
		vols = defaultFixedVolumes()
	}
	if cfg.Sensors.FileWatcherOn() && (runtime.GOOS == "windows" || runtime.GOOS == "linux") {
		w := watcher.New(vols, sinkAdapter{orch}, c, ex.Excluded)
		if h, ok := any(w).(interface{ Health() error }); ok {
			p.sensorHealth["file_watcher"] = h
		}
		p.goPipeline("watcher", func() { w.Run(ctx) })
	} else {
		logger.Info("UTMStack EDR: file_watcher sensor disabled by config")
	}
	p.startFilePrevention(ctx, cfg, p.eng, sp, orch, ex.Excluded)

	// Process-kill (§6.2 race): one process-creation watcher feeds the table + a
	// guard that suspends an unknown launch, expedites its scan, and kills the
	// whole tree on a malicious verdict. The scanner (sc) is shared with the file
	// path so a verdict is cached once and reused. (tab/resp created above.)
	// When the process_guard sensor is off, the guard is not wired (g=nil) — but
	// the process watcher may still run to feed ransomware/behavioral.
	var g *guard.Guard
	if cfg.Sensors.ProcessGuardOn() && (runtime.GOOS == "windows" || runtime.GOOS == "linux") {
		g = guard.New(cfg, sc, resp)
	} else {
		logger.Info("UTMStack EDR: process_guard sensor disabled by config")
	}

	// Ransomware guard: canary tripwires + T1490 command rules → per-PID scorer →
	// suspend/kill/quarantine. Attribution for file signals comes from the ETW
	// file-activity feed; T1490 attribution comes from the procwatch dispatch hook.
	var rwGuard *ransomware.Guard
	var onProc func(procwatch.ProcStart)
	var onSnapshot func(procwatch.ProcStart)
	if cfg.Ransomware.Enabled && (runtime.GOOS == "windows" || runtime.GOOS == "linux" && cfg.LinuxProtection.Fanotify.Enabled) {
		mgr := ransomware.NewManager(c)
		canaryDirs := append(defaultCanaryDirs(), cfg.Ransomware.CanaryDirs...)
		var canaryErr error
		if runtime.GOOS == "linux" {
			canaryDirs = linuxCanaryScope(cfg)
			canaryErr = mgr.SetMountScope(cfg.LinuxProtection.Fanotify.Mounts)
		}
		if canaryErr == nil {
			canaryErr = mgr.LoadCovered(canaryDirs)
		}
		if canaryErr == nil {
			n, err := mgr.Plant(canaryDirs, cfg.Ransomware.CanaryPerDir)
			canaryErr = err
			logger.Info("UTMStack EDR: planted %d canaries", n)
		}
		if mgr.Count() == 0 {
			canaryErr = errors.Join(canaryErr, errors.New("no verified canaries in monitored scope"))
		}
		if canaryErr != nil {
			logger.Error("UTMStack EDR: canary coverage: %v", canaryErr)
		}
		p.canaries = mgr
		rwGuard = ransomware.NewGuard(ransomware.GuardDeps{
			Cfg: cfg, Table: tab, Resp: resp, Spool: sp, Quar: store, Incidents: c,
			Canaries: mgr, Hash: scanner.SHA256File, Now: time.Now, NewID: uuid.NewString,
			Trusted: trusted.Excluded, LiveProcess: responder.LiveProcess,
		})
		p.rwGuard = rwGuard
		onProc = func(ps procwatch.ProcStart) {
			if observed, ok := tab.Get(ps.PID); ok && observed.OSIdentity == ps.OSIdentity {
				rwGuard.OnObservedProcess(observed)
			}
		}
		onSnapshot = func(ps procwatch.ProcStart) {
			if observed, ok := tab.Get(ps.PID); ok && observed.OSIdentity == ps.OSIdentity {
				rwGuard.OnObservedSnapshot(observed)
			}
		}
		selfPID := os.Getpid()
		var fileFeed ransomware.FileActivityFeed = ransomware.NewFeed()
		if runtime.GOOS == "linux" {
			fc := nativeFanotifyConfig(cfg, p.eng.Tuning().MaxFileSizeMB)
			fc.Mode = "notify"
			ff, err := ransomware.NewFanotifyFeed(fc)
			if err != nil {
				p.sensorHealth["ransomware"] = fixedHealth{errors.Join(err, canaryErr)}
				fileFeed = nil
			} else {
				fileFeed = ff
				p.sensorHealth["ransomware"] = combinedHealth{ff, fixedHealth{canaryErr}}
			}
		}
		if fileFeed != nil {
			checks := combinedHealth{fixedHealth{canaryErr}, rwGuard}
			if h, ok := fileFeed.(interface{ Health() error }); ok {
				checks = append(checks, h)
			} else {
				checks = append(checks, unverifiedFeed{})
			}
			// Windows-only: the Kernel-Registry ETW feed delivers per-process
			// registry mutations for the recovery-tamper sensor. Its health is
			// folded into the same ransomware sensor-health aggregate as the file
			// feed; on non-Windows the feed is a healthy no-op.
			regFeed := ransomware.NewRegistryFeed()
			if h, ok := regFeed.(interface{ Health() error }); ok {
				checks = append(checks, h)
			} else {
				checks = append(checks, unverifiedFeed{})
			}
			p.sensorHealth["ransomware"] = checks
			p.goNative("ransomware-registry", func() { _ = regFeed.Run(ctx, func(re ransomware.RegistryEvent) { rwGuard.OnRegistryEvent(re) }) })
			p.goNative("ransomware", func() { _ = rwGuard.Run(ctx, fileFeed, selfPID, ex.Excluded) })
		}
	}

	// Network blocklist: ThreatWinds-derived IP/CIDR indicators → WFP enforcement +
	// connection audit. Feed refreshes from the server mirror; enforcement and
	// detect-only mode are governed by the blocklist config block. The never-block
	// allowlist is seeded with host-derived system nets (server IPs, and on Windows
	// resolvers/gateways) so we can never cut the host off from the platform.
	if cfg.Blocklist.Enabled && (runtime.GOOS == "windows" || runtime.GOOS == "linux" && cfg.LinuxProtection.NetworkEnforcement) {
		nb := netblock.NewManager(netblock.Deps{
			Cfg:   cfg,
			Cache: c,
			Spool: sp,
			Sys:   resolveSystemNets(cfg),
		})
		p.netblock = nb
		p.sensorHealth["network_enforcement"] = networkHealth{nb}
		p.goNative("netblock", func() { nb.Run(ctx) })
		p.startNetworkSafetyRefresh(ctx, nb)
	}
	// The process-creation watcher runs if any consumer needs it: the process
	// guard, the ransomware T1490 hook, or behavioral process telemetry. The
	// dispatch gets a nil guard / nil telemetry spool for disabled sensors, so
	// each consumer is independently gated while the single WMI feed is shared.
	behavioralOn := cfg.Sensors.BehavioralOn() && (runtime.GOOS == "windows" || runtime.GOOS == "linux")
	var telemetrySpool *event.Spool
	if behavioralOn {
		telemetrySpool = sp // process-creation telemetry (§4.9) only when behavioral is on
	}
	if g != nil || rwGuard != nil || behavioralOn {
		dispatch := procwatch.NewDispatch(tab, g, telemetrySpool, procSkip, onProc)
		dispatch.SetSnapshotHook(onSnapshot)
		pw := newProcessWatcher(cfg, dispatch)
		checks := combinedHealth{dispatch}
		if h, ok := any(pw).(interface{ Health() error }); ok {
			checks = append(checks, h)
		} else {
			checks = append(checks, unverifiedFeed{})
		}
		p.sensorHealth["process_guard"] = checks
		p.sensorHealth["behavioral"] = checks
		p.goPipeline("procwatch", func() { pw.Run(ctx) })
	}

	p.sigFeed = feed.New(cfg, c, p.eng)
	p.sigFeed.SetSpool(sp)
	// Recover the refresh baseline, so a first refresh that fails after a
	// restart is not reported as unavailable updates while a verified
	// generation serves.
	p.sigFeed.RestoreHealth(p.eng.DatabaseDir())
	p.goPipeline("feed", func() { p.sigFeed.Run(ctx) })

	// Scripts/fileless (AMSI): named-pipe scan server + provider registration.
	if cfg.Sensors.AMSIOn() && runtime.GOOS == "windows" {
		amsiSc := amsi.NewScanner(cfg, sp, p.eng)
		p.amsiPipe = amsi.NewPipeServer(amsiSc)
		goSafe("amsi-pipe", func() { p.amsiPipe.Run(ctx) })
		dll := filepath.Join(config.InstallDir, "utmstack_amsi.dll")
		if err := amsi.Register(dll); err != nil {
			p.sensorHealth["amsi"] = fixedHealth{err}
			logger.Error("UTMStack EDR: AMSI provider registration failed: %v", err)
		} else {
			p.amsiRegistered = true
			p.sensorHealth["amsi"] = combinedHealth{p.amsiPipe, unverifiedFeed{}}
		}
	} else {
		logger.Info("UTMStack EDR: amsi sensor disabled by config")
	}

	// Behavioral telemetry: PowerShell script-block (4104) logs (+ process
	// telemetry above). Chatty — this is the sensor most likely to be turned off.
	if behavioralOn && runtime.GOOS == "windows" {
		goSafe("behavioral", func() { behavioral.NewPSLogReader(sp).Run(ctx) })
	} else {
		logger.Info("UTMStack EDR: behavioral sensor disabled by config")
	}

	// Quarantine retention: periodically purge quarantined items older than the
	// configured retention (0 = keep forever). Makes quarantine_retention_days real.
	p.goPipeline("quarantine-retention", func() { runRetention(ctx, store, cfg.QuarantineDays) })

	logger.Info("UTMStack EDR: real-time protection started (%d volumes, %d workers)", len(vols), cfg.ScanConcurrency)
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

func (p *program) goNative(name string, fn func()) {
	p.nativeWG.Add(1)
	goSafe(name, func() { defer p.nativeWG.Done(); fn() })
}

// Linux pipeline users of shared event/cache state finish before it closes.
// Other platforms retain their existing lifecycle pending native qualification.
func (p *program) goPipeline(name string, fn func()) {
	if runtime.GOOS == "linux" {
		p.goNative(name, fn)
	} else {
		goSafe(name, fn)
	}
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
