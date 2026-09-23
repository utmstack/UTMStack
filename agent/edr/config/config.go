package config

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/shared/fs"
)

const ServiceName = "UTMStackEDR"

var (
	InstallDir    = fs.GetExecutablePath()
	ConfigFile    = filepath.Join(InstallDir, "edr.json")
	StatusFile    = filepath.Join(InstallDir, "status.json")
	DBFile        = filepath.Join(InstallDir, "edr.db")
	SpoolDir      = filepath.Join(InstallDir, "edr-spool")
	SpoolFile     = filepath.Join(SpoolDir, "events.ndjson")
	QuarantineDir = filepath.Join(InstallDir, "quarantine")
	EngineDir     = filepath.Join(InstallDir, "engine")
	LogFile       = filepath.Join(InstallDir, "logs", "utmstack_edr.log")
	// BlocklistDir stages downloaded ThreatWinds feed artifacts.
	BlocklistDir = filepath.Join(InstallDir, "blocklist")
)

const (
	// Agent-facing mirror plane on the UTMStack server: the platform serves
	// the EDR mirror tree (signature databases + threat-intel feeds) as static
	// files on the dependencies port. Path segments deliberately avoid
	// engine/intel vendor names — these URLs surface in warning logs.
	MirrorPort     = "9001"
	MirrorBasePath = "/private/edr"
	// MirrorHTTPPort: the EDR server container serves the signature mirror over
	// plain HTTP on this port for self-signed (skip_cert_validate) deployments,
	// where freshclam cannot validate TLS. CVD signature integrity holds
	// regardless of transport, so the plain-HTTP channel is safe here.
	MirrorHTTPPort = "9002"
)

// Allowlist unifies the three "what does the EDR ignore" lists into one block —
// previously scattered as top-level `exclusions` / `trusted_processes` and a
// nested `ransomware.command_allowlist`. This is the single home for false-
// positive tuning; the CLI's `allow path|process|command` verbs edit it.
//
// IMPORTANT: these user entries are ADDITIVE to the EDR's built-in, ships-with-
// the-system lists (the self-exclusions and the backup/imaging/DR/sync trusted-
// process defaults). User config never replaces the built-ins.
type Allowlist struct {
	// Paths: file paths the real-time watcher skips (never scanned/quarantined).
	// Directory (excludes that subtree on a path boundary) or glob; case- and
	// separator-insensitive.
	Paths []string `json:"paths"`
	// Processes: process/service images the EDR won't act on — the process guard
	// won't scan/suspend/kill them, and the ransomware guard won't score their
	// behavioral signals. Full path / directory subtree / base name / glob.
	Processes []string `json:"processes"`
	// Commands: substrings that exempt a command line from the T1490 recovery-
	// tampering rules (e.g. a backup product's distinctive switch or name).
	Commands []string `json:"commands"`
	// Networks: IP/CIDR exemptions the network blocklist never blocks (self-
	// lockout prevention). Feeds netblock.BuildAllowlist alongside the built-in
	// critical/private ranges and host system nets.
	Networks []string `json:"networks,omitempty"`
}

// Sensors toggles each detector independently so an admin can disable one noisy
// or problematic sensor (e.g. the chatty behavioral telemetry) without disabling
// the whole module. Pointer fields: nil/absent = ON (see the *On() accessors),
// so an omitted block leaves everything enabled. (The ransomware guard has its
// own `ransomware.enabled`.) These are structural — a change takes effect on the
// next service (re)start.
type Sensors struct {
	FileWatcher  *bool `json:"file_watcher,omitempty"`
	ProcessGuard *bool `json:"process_guard,omitempty"`
	AMSI         *bool `json:"amsi,omitempty"`
	Behavioral   *bool `json:"behavioral,omitempty"`
}

func sensorOn(p *bool) bool { return p == nil || *p }

func (s Sensors) FileWatcherOn() bool  { return sensorOn(s.FileWatcher) }
func (s Sensors) ProcessGuardOn() bool { return sensorOn(s.ProcessGuard) }
func (s Sensors) AMSIOn() bool         { return sensorOn(s.AMSI) }
func (s Sensors) BehavioralOn() bool   { return sensorOn(s.Behavioral) }

// RansomwareConfig configures the behavioral ransomware guard (v1 core).
type RansomwareConfig struct {
	Enabled          bool     `json:"enabled"`
	ResponseMode     string   `json:"response_mode"`      // "alert" | "suspend" | "kill"
	CanaryDirs       []string `json:"canary_dirs"`        // extra subtrees to seed (beyond volume roots + profile)
	CanaryPerDir     int      `json:"canary_per_dir"`     // decoys planted per directory
	SuspendThreshold int      `json:"suspend_threshold"`  // score to suspend a suspect
	KillThreshold    int      `json:"kill_threshold"`     // score to kill+quarantine
	DecayHalfLifeMs  int      `json:"decay_half_life_ms"` // per-PID score half-life
	// CommandAllowlist is DEPRECATED — migrated into Allowlist.Commands by Load()
	// and no longer written by Save(). Kept only so existing edr.json files that
	// still carry it keep working.
	CommandAllowlist []string `json:"command_allowlist,omitempty"`
	UseETW           bool     `json:"use_etw"` // true = ETW file attribution (v1 default)
}

// BlocklistConfig configures the network blocklist (ThreatWinds via the UTMStack
// mirror). Enabled+enforcing by default; safety rails (allowlist, private-range
// exemption, fail-open) are mandatory because it can drop live traffic.
type BlocklistConfig struct {
	Enabled            bool     `json:"enabled"`              // default true
	Enforce            bool     `json:"enforce"`              // default true (false = inert in Plan 1 (loads intel, no WFP filters, no telemetry); observation lands in Plan 2)
	MirrorBaseURL      string   `json:"blocklist_mirror"`     // empty = derive from Server
	Levels             []int    `json:"levels"`               // default [1]
	IndicatorTypes     []string `json:"indicator_types"`      // default ["ip"] in Plan 1
	Direction          string   `json:"direction"`            // "both"|"out"|"in"
	RefreshHours       int      `json:"refresh_hours"`        // default 6
	FullResyncHours    int      `json:"full_resync_hours"`    // default 24; how often a cycle does a full accumulative re-sync instead of a daily delta
	AllowPrivateRanges bool     `json:"allow_private_ranges"` // default true
	DomainBlockTTLMin  int      `json:"domain_block_ttl_min"` // Plan 2; default 30
}

// ScheduledScanConfig configures the periodic full-disk scan. The walk
// reuses the real-time pipeline: enqueued files are scanned by the same worker
// pool, cache, engine, and exclusions. These are structural — a change takes
// effect on the next service (re)start.
type ScheduledScanConfig struct {
	// Enabled turns the scheduled scan on. Default false.
	Enabled bool `json:"enabled"`
	// Schedule is the period between runs. Only the form "every:<duration>" is
	// supported (e.g. "every:24h"); any other form is rejected at trigger time
	// with a clear error. Default "every:24h".
	Schedule string `json:"schedule"`
	// Paths are the roots to walk each run. Empty = all fixed volumes (same
	// rule as WatchVolumes).
	Paths []string `json:"paths,omitempty"`
	// ThrottleMs pauses the walk that long after every BatchSize enqueues so a
	// full-disk sweep cannot starve the real-time watcher queue. 0 = no throttle.
	ThrottleMs int `json:"throttle_ms"`
	// BatchSize is the number of files enqueued before a ThrottleMs pause.
	// Default 256.
	BatchSize int `json:"batch_size"`
}

// EDRConfig is the operational configuration read by the EDR service and
// writable by the agent / local CLI. It lives entirely in <install>/edr.json —
// a single source of truth, hand-editable and CLI-editable.
type EDRConfig struct {
	Enabled          bool                `json:"enabled"`
	Server           string              `json:"server"` // written by the agent on enable
	SkipCertValidate bool                `json:"skip_cert_validate"`
	WatchVolumes     []string            `json:"watch_volumes"` // Plan 2; empty = all fixed NTFS
	ScheduledScan    ScheduledScanConfig `json:"scheduled_scan"`
	Allowlist        Allowlist           `json:"allowlist"` // unified FP-tuning lists (see Allowlist)
	Sensors          Sensors             `json:"sensors"`   // per-detector on/off toggles

	// Exclusions / TrustedProcesses are DEPRECATED top-level lists, superseded by
	// Allowlist.{Paths,Processes}. Load() migrates them into Allowlist and Save()
	// stops writing them (omitempty). Retained so pre-existing edr.json keeps
	// working.
	Exclusions       []string `json:"exclusions,omitempty"`
	TrustedProcesses []string `json:"trusted_processes,omitempty"`

	FailMode         string `json:"fail_mode"`          // open | closed
	SuspendOnLaunch  bool   `json:"suspend_on_launch"`  // Plan 3
	SuspendTimeoutMs int    `json:"suspend_timeout_ms"` // Plan 3
	QuarantineDir    string `json:"quarantine_dir"`
	QuarantineDays   int    `json:"quarantine_retention_days"` // 0 = keep forever
	ClamdAddr        string `json:"clamd_addr"`
	ClamdPath        string `json:"clamd_path"` // optional explicit path for testing
	ScanConcurrency  int    `json:"scan_concurrency"`
	SigUpdateHours   int    `json:"sig_update_hours"`

	// Engine tuning (ClamAV Configuration Guide §8). TierOverride forces a tier
	// ("constrained"|"standard"|"server"); empty = auto-derive from the host.
	TierOverride string `json:"engine_tier_override"`
	// SigMirror selects the signature-update source: "" (default) auto-derives
	// the UTMStack server mirror; "cdn" forces the official public database; any
	// other value is an explicit mirror URL.
	SigMirror string `json:"signature_mirror"`
	// SignatureFallback controls behavior when the private signature mirror is
	// unreachable: "cdn" (default) falls back to the official public database
	// for one cycle; "none" stays mirror-only (air-gapped posture).
	SignatureFallback string `json:"signature_fallback"`
	// PolicyVersion is the version of the last centrally-assigned policy
	// document (policy_set). The status reporter ships it upstream so the
	// server can detect drift between assigned and applied policy.
	PolicyVersion string `json:"policy_version"`

	// Ransomware is the behavioral ransomware guard block (nested; copied wholesale
	// by Load when present on disk).
	Ransomware RansomwareConfig `json:"ransomware"`

	// Blocklist is the network blocklist block (nested; copied wholesale by Load
	// when present on disk).
	Blocklist BlocklistConfig `json:"blocklist"`
}

func Default() EDRConfig {
	conc := runtime.NumCPU()
	if conc > 4 {
		conc = 4
	}
	if conc < 1 {
		conc = 1
	}
	return EDRConfig{
		Enabled:  false,
		FailMode: "open",
		// Default OFF: suspend-on-launch narrows the §6.2 race but freezes every
		// launched process while it is scanned; without a kernel driver that is
		// risky on a busy host. The safe default is detect-and-kill. When enabled,
		// the guard's rails apply (never suspend OS processes; bound by
		// SuspendTimeoutMs).
		SuspendOnLaunch:   false,
		SuspendTimeoutMs:  5000,
		QuarantineDir:     QuarantineDir,
		QuarantineDays:    30,
		ClamdAddr:         "127.0.0.1:3310",
		ScanConcurrency:   conc,
		SigUpdateHours:    4,
		SignatureFallback: "cdn",
		ScheduledScan: ScheduledScanConfig{
			// Disabled by default; enabled with a sane 24h period.
			Schedule:  "every:24h",
			BatchSize: 256,
		},
		// Sensors left zero: all *bool nil ⇒ every sensor ON by default.
		Ransomware: RansomwareConfig{
			Enabled:          true, // on by default; response_mode "suspend" is reversible
			ResponseMode:     "suspend",
			CanaryPerDir:     1,
			SuspendThreshold: 50,
			KillThreshold:    100,
			DecayHalfLifeMs:  10000,
			UseETW:           true,
		},
		Blocklist: BlocklistConfig{
			Enabled:            true,
			Enforce:            true,
			Levels:             []int{1},
			IndicatorTypes:     []string{"ip", "domain", "hostname"},
			Direction:          "both",
			RefreshHours:       6,
			FullResyncHours:    24,
			AllowPrivateRanges: true,
			DomainBlockTTLMin:  30,
		},
	}
}

// Load reads edr.json, filling missing/zero fields from Default() and migrating
// the deprecated scattered whitelists into the unified Allowlist.
func Load() (EDRConfig, error) {
	c := Default()
	if !fs.Exists(ConfigFile) {
		return c, nil
	}
	var onDisk EDRConfig
	if err := fs.ReadJSON(ConfigFile, &onDisk); err != nil {
		return c, err
	}
	// Overlay non-zero on-disk values onto defaults.
	if onDisk.FailMode != "" {
		c.FailMode = onDisk.FailMode
	}
	if onDisk.ClamdAddr != "" {
		c.ClamdAddr = onDisk.ClamdAddr
	}
	if onDisk.ScanConcurrency > 0 {
		c.ScanConcurrency = onDisk.ScanConcurrency
	}
	if onDisk.QuarantineDir != "" {
		c.QuarantineDir = onDisk.QuarantineDir
	}
	if onDisk.QuarantineDays != 0 {
		c.QuarantineDays = onDisk.QuarantineDays
	}
	c.Enabled = onDisk.Enabled
	c.Server = onDisk.Server
	c.SkipCertValidate = onDisk.SkipCertValidate
	c.SuspendOnLaunch = onDisk.SuspendOnLaunch
	if onDisk.SuspendTimeoutMs > 0 {
		c.SuspendTimeoutMs = onDisk.SuspendTimeoutMs
	}
	c.ClamdPath = onDisk.ClamdPath
	c.WatchVolumes = onDisk.WatchVolumes
	c.Sensors = onDisk.Sensors
	c.TierOverride = onDisk.TierOverride
	c.SigMirror = onDisk.SigMirror
	if onDisk.SignatureFallback != "" {
		c.SignatureFallback = onDisk.SignatureFallback
	}
	c.PolicyVersion = onDisk.PolicyVersion
	// RansomwareConfig has slice fields, so it is not comparable with != against a
	// zero literal. Detect an on-disk block via a reliable non-zero scalar and copy
	// the whole block; an absent block leaves the Default() nested values intact.
	if onDisk.Ransomware.ResponseMode != "" || onDisk.Ransomware.Enabled || onDisk.Ransomware.KillThreshold != 0 {
		c.Ransomware = onDisk.Ransomware
	}
	// BlocklistConfig also has slice fields, so detect an on-disk block via a
	// reliable non-zero signal and copy it wholesale, then backfill any missing
	// required defaults. An absent block leaves the Default() nested values intact.
	if onDisk.Blocklist.Enabled || onDisk.Blocklist.RefreshHours != 0 ||
		len(onDisk.Blocklist.Levels) != 0 || onDisk.Blocklist.Direction != "" {
		c.Blocklist = onDisk.Blocklist
		if len(c.Blocklist.Levels) == 0 {
			c.Blocklist.Levels = []int{1}
		}
		if len(c.Blocklist.IndicatorTypes) == 0 {
			c.Blocklist.IndicatorTypes = []string{"ip", "domain", "hostname"}
		}
		if c.Blocklist.Direction == "" {
			c.Blocklist.Direction = "both"
		}
		if c.Blocklist.RefreshHours == 0 {
			c.Blocklist.RefreshHours = 6
		}
		if c.Blocklist.FullResyncHours == 0 {
			c.Blocklist.FullResyncHours = 24
		}
	}
	// ScheduledScanConfig has a slice field (Paths), so detect an on-disk block
	// via a reliable non-zero signal and copy it wholesale, then backfill the
	// batch-size default. An absent block leaves the Default() nested values
	// intact.
	if onDisk.ScheduledScan.Enabled || onDisk.ScheduledScan.Schedule != "" ||
		onDisk.ScheduledScan.ThrottleMs != 0 || onDisk.ScheduledScan.BatchSize != 0 ||
		len(onDisk.ScheduledScan.Paths) != 0 {
		c.ScheduledScan = onDisk.ScheduledScan
		if c.ScheduledScan.Schedule == "" {
			c.ScheduledScan.Schedule = "every:24h"
		}
		if c.ScheduledScan.BatchSize == 0 {
			c.ScheduledScan.BatchSize = 256
		}
	}

	// Unify whitelists: start from the on-disk allowlist block, then fold in any
	// legacy scattered lists so old files keep working. The deprecated fields are
	// then cleared so a subsequent Save() migrates the file to the unified form.
	c.Allowlist = onDisk.Allowlist
	c.Allowlist.Paths = mergeUnique(c.Allowlist.Paths, onDisk.Exclusions)
	c.Allowlist.Processes = mergeUnique(c.Allowlist.Processes, onDisk.TrustedProcesses)
	c.Allowlist.Commands = mergeUnique(c.Allowlist.Commands, onDisk.Ransomware.CommandAllowlist)
	c.Exclusions = nil
	c.TrustedProcesses = nil
	c.Ransomware.CommandAllowlist = nil

	return c, nil
}

func Save(c EDRConfig) error {
	if err := fs.CreateDirIfNotExist(InstallDir); err != nil {
		return err
	}
	// Never write the deprecated whitelist fields — the unified Allowlist is
	// canonical. (Load already nils them, but be defensive against a caller that
	// set them directly.)
	c.Exclusions = nil
	c.TrustedProcesses = nil
	c.Ransomware.CommandAllowlist = nil
	return fs.WriteJSON(ConfigFile, c)
}

// mergeUnique appends any entries of extra not already present in base
// (case-insensitive), preserving order. Used to fold legacy lists into the
// unified Allowlist without introducing duplicates.
func mergeUnique(base, extra []string) []string {
	seen := make(map[string]bool, len(base))
	for _, b := range base {
		seen[strings.ToLower(strings.TrimSpace(b))] = true
	}
	for _, e := range extra {
		k := strings.ToLower(strings.TrimSpace(e))
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		base = append(base, e)
	}
	return base
}
