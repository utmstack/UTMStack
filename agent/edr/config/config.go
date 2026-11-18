package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
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
	// Paths: trusted locations the real-time sensors skip. Files there are never
	// scanned or quarantined by the real-time watcher, and programs whose image
	// is there are skipped by the launch guard (no scan/suspend/kill) and by
	// process telemetry. This is the documented false-positive control for a
	// trusted tool tree. The ransomware guard ignores file changes there except
	// canary tampering, but still scores such programs and can stop them and
	// quarantine their image (only Processes exempts them). An on-demand scan
	// ignores this list.
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

// ScanHeuristics controls heuristic alerts independently of signature matches.
// An omitted switch keeps the current default-on behavior. These settings
// take effect when the service constructs a new engine after restart.
type ScanHeuristics struct {
	PDFObfuscatedName     *bool `json:"pdf_obfuscated_name,omitempty"`
	PartitionIntersection *bool `json:"partition_intersection,omitempty"`
	BrokenExecutable      *bool `json:"broken_executable,omitempty"`
}

func (h ScanHeuristics) PDFObfuscatedNameOn() bool { return sensorOn(h.PDFObfuscatedName) }

// The incumbent endpoint ran ClamAV with --alert-broken=yes. Turning this
// off also relaxes ClamAV's PE header checks before built-in unpacking.
func (h ScanHeuristics) BrokenExecutableOn() bool { return sensorOn(h.BrokenExecutable) }

// ClamAV's partition intersection alert is opt-in, unlike its PDF heuristic.
func (h ScanHeuristics) PartitionIntersectionOn() bool {
	return h.PartitionIntersection != nil && *h.PartitionIntersection
}

// A misspelled heuristic switch must not silently leave a different response
// policy in effect.
func (h *ScanHeuristics) UnmarshalJSON(data []byte) error {
	type heuristics ScanHeuristics
	var decoded heuristics
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&decoded); err != nil {
		return fmt.Errorf("scan_heuristics: %w", err)
	}
	*h = ScanHeuristics(decoded)
	return nil
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

// ScanLimits overrides the host-derived analysis budgets. Zero selects the
// derived default; it never disables a limit. Limits apply to each input and
// its recursively decoded children, and an exceeded budget is incomplete
// analysis rather than a clean verdict, except MaxPartitions, which follows
// ClamAV's stop-after-N behavior. These budgets do not cap process RSS.
type ScanLimits struct {
	MaxFileSizeMB int `json:"max_file_size_mb"`
	MaxScanSizeMB int `json:"max_scan_size_mb"`
	MaxRecursion  int `json:"max_recursion"`
	MaxFiles      int `json:"max_files"`
	MaxPartitions int `json:"max_partitions"`
	MaxScanTimeMS int `json:"max_scan_time_ms"`
}

// UnmarshalJSON is strict within the new scan_limits block so a misspelled
// budget cannot silently leave a larger host-derived default in effect.
// Unknown legacy top-level fields remain accepted for configuration migration.
func (s *ScanLimits) UnmarshalJSON(data []byte) error {
	type limits ScanLimits
	var decoded limits
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err := d.Decode(&decoded); err != nil {
		return fmt.Errorf("scan_limits: %w", err)
	}
	*s = ScanLimits(decoded)
	return nil
}

// Hard resource ceilings are shared by validation and native-engine tuning.
// Explicit overrides can change a tier's defaults but cannot remove these
// bounds or increase them on larger hosts.
const (
	MaxScanConcurrency = 16
	MaxFileSizeMB      = 200
	MaxScanSizeMB      = 400
	MaxRecursion       = 16
	MaxFiles           = 10000
	MaxPartitions      = 1000
	MaxScanTimeMS      = 120000
)

// ValidateScanResources rejects invalid hand-edited and CLI configurations
// alike. Validation occurs before defaults are overlaid so negative values
// cannot silently disappear. The engine also clamps direct callers defensively.
func (c EDRConfig) ValidateScanResources() error {
	if err := c.ScheduledScan.Validate(); err != nil {
		return err
	}
	for _, v := range []struct {
		key    string
		n, max int
	}{
		{"scan_concurrency", c.ScanConcurrency, MaxScanConcurrency},
		{"scan_limits.max_file_size_mb", c.ScanLimits.MaxFileSizeMB, MaxFileSizeMB},
		{"scan_limits.max_scan_size_mb", c.ScanLimits.MaxScanSizeMB, MaxScanSizeMB},
		{"scan_limits.max_recursion", c.ScanLimits.MaxRecursion, MaxRecursion},
		{"scan_limits.max_files", c.ScanLimits.MaxFiles, MaxFiles},
		{"scan_limits.max_partitions", c.ScanLimits.MaxPartitions, MaxPartitions},
		{"scan_limits.max_scan_time_ms", c.ScanLimits.MaxScanTimeMS, MaxScanTimeMS},
	} {
		if v.n < 0 || v.n > v.max {
			return fmt.Errorf("%s must be an integer from 0 to %d (0 = host default)", v.key, v.max)
		}
	}
	if c.ScanLimits.MaxFileSizeMB > 0 && c.ScanLimits.MaxScanSizeMB > 0 && c.ScanLimits.MaxFileSizeMB > c.ScanLimits.MaxScanSizeMB {
		return errors.New("scan_limits.max_file_size_mb must not exceed max_scan_size_mb")
	}
	switch c.TierOverride {
	case "", "constrained", "standard", "server":
	default:
		return errors.New("engine_tier_override must be constrained|standard|server (or empty to auto-derive)")
	}
	return nil
}

// EDRConfig is the operational configuration read by the EDR service and
// writable by the agent / local CLI. It lives entirely in <install>/edr.json —
// a single source of truth, hand-editable and CLI-editable.
type EDRConfig struct {
	Enabled          bool           `json:"enabled"`
	Server           string         `json:"server"` // written by the agent on enable
	SkipCertValidate bool           `json:"skip_cert_validate"`
	WatchVolumes     []string       `json:"watch_volumes"`   // Windows volumes or Linux absolute scoped roots
	Allowlist        Allowlist      `json:"allowlist"`       // unified FP-tuning lists (see Allowlist)
	Sensors          Sensors        `json:"sensors"`         // per-detector on/off toggles
	ScanHeuristics   ScanHeuristics `json:"scan_heuristics"` // per-analysis heuristic alerts
	// CollectMetadata enables root PRECLASS JSON hooks and legacy target-13
	// rules, matching ClamAV's opt-in --gen-json scan mode. It is off by
	// default; ordinary bytecode JSON queries remain inactive.
	CollectMetadata bool          `json:"collect_metadata"`
	ScheduledScan   ScheduledScan `json:"scheduled_scan"`

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
	// The legacy daemon transport keys of older edr.json files are deliberately
	// not fields: Load ignores them and the next Save drops them. The native
	// engine neither launches a helper nor opens a scanning socket.
	// SignatureDBDir must be absolute (see NormalizeSignatureDBDir); empty means
	// <install>/engine/database.
	SignatureDBDir     string     `json:"signature_db_dir,omitempty"`
	SignaturePublicKey string     `json:"signature_public_key,omitempty"` // base64 Ed25519 feed signing key
	ScanConcurrency    int        `json:"scan_concurrency"`
	ScanLimits         ScanLimits `json:"scan_limits"`
	SigUpdateHours     int        `json:"sig_update_hours"`

	// Host-adaptive scanning bounds. TierOverride forces a tier
	// ("constrained"|"standard"|"server"); empty = auto-derive from the host.
	TierOverride string `json:"engine_tier_override"`
	// SigMirror selects the signature-update source: "" (default) auto-derives
	// the UTMStack signed native feed; another value is an explicit HTTPS URL.
	// The legacy "cdn" sentinel is rejected by the native updater.
	SigMirror string `json:"signature_mirror"`
	// SignatureFallback is retained for legacy configuration migration only.
	// Native updates never fall back to an unauthenticated or incompatible feed.
	SignatureFallback string `json:"signature_fallback"`

	// Ransomware is the behavioral ransomware guard block (nested; copied wholesale
	// by Load when present on disk).
	Ransomware RansomwareConfig `json:"ransomware"`

	// Blocklist is the network blocklist block (nested; copied wholesale by Load
	// when present on disk).
	Blocklist       BlocklistConfig `json:"blocklist"`
	LinuxProtection LinuxProtection `json:"linux_protection"`
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
		ScanConcurrency:   conc,
		SigUpdateHours:    4,
		SignatureFallback: "none",
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
	f, err := os.Open(ConfigFile)
	if err != nil {
		return c, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return c, err
	}
	if len(raw) > 1<<20 {
		return c, errors.New("EDR configuration exceeds 1 MiB")
	}
	var onDisk EDRConfig
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		return c, err
	}
	var blocks map[string]json.RawMessage
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return c, err
	}
	if err := onDisk.ValidateScanResources(); err != nil {
		return c, err
	}
	if err := onDisk.LinuxProtection.Validate(); err != nil {
		return c, err
	}
	// Overlay non-zero on-disk values onto defaults.
	if onDisk.FailMode != "" {
		c.FailMode = onDisk.FailMode
	}
	if onDisk.ScanConcurrency > 0 {
		c.ScanConcurrency = onDisk.ScanConcurrency
	}
	c.ScanLimits = onDisk.ScanLimits
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
	if dir, err := NormalizeSignatureDBDir(onDisk.SignatureDBDir); err != nil {
		// A hand-edited relative value would resolve differently for the service
		// and the CLI; fall back to the default directory instead.
		logger.Error("UTMStack EDR: ignoring signature_db_dir %q: %v", onDisk.SignatureDBDir, err)
	} else {
		c.SignatureDBDir = dir
	}
	c.SignaturePublicKey = onDisk.SignaturePublicKey
	if onDisk.SigUpdateHours > 0 {
		c.SigUpdateHours = onDisk.SigUpdateHours
	}
	c.WatchVolumes = onDisk.WatchVolumes
	c.Sensors = onDisk.Sensors
	c.ScanHeuristics = onDisk.ScanHeuristics
	c.CollectMetadata = onDisk.CollectMetadata
	c.ScheduledScan = onDisk.ScheduledScan
	c.LinuxProtection = onDisk.LinuxProtection
	c.TierOverride = onDisk.TierOverride
	c.SigMirror = onDisk.SigMirror
	if onDisk.SignatureFallback != "" {
		c.SignatureFallback = onDisk.SignatureFallback
	}
	// Overlay nested blocks onto defaults, respecting explicit false values even
	// when they are the only field present. Scalar truthiness loses disable edits.
	if block, ok := blocks["ransomware"]; ok && string(block) != "null" {
		if err := json.Unmarshal(block, &c.Ransomware); err != nil {
			return c, err
		}
	}
	// An absent block retains defaults; an explicit block preserves false values.
	if block, ok := blocks["blocklist"]; ok && string(block) != "null" {
		if err := json.Unmarshal(block, &c.Blocklist); err != nil {
			return c, err
		}
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
	if err := c.LinuxProtection.Validate(); err != nil {
		return err
	}
	if err := c.ValidateScanResources(); err != nil {
		return err
	}
	if err := fs.CreateDirIfNotExist(InstallDir); err != nil {
		return err
	}
	// Never write the deprecated whitelist fields — the unified Allowlist is
	// canonical. (Load already nils them, but be defensive against a caller that
	// set them directly.)
	c.Exclusions = nil
	c.TrustedProcesses = nil
	c.Ransomware.CommandAllowlist = nil
	encoded, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if len(encoded) > 1<<20 {
		return errors.New("EDR configuration exceeds 1 MiB")
	}
	return os.WriteFile(ConfigFile, encoded, 0600)
}

// NormalizeSignatureDBDir validates a signature_db_dir value. Empty selects the
// default directory. Anything else must be absolute: the Windows service runs
// from %SystemRoot%\System32 while engine-check and scan run from the admin's
// shell, so a relative path names different directories for each (and a bare
// name in the scan exclusions would match any file with that base name).
func NormalizeSignatureDBDir(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}
	if !filepath.IsAbs(dir) {
		return "", errors.New("signature_db_dir must be an absolute directory (or empty for the default)")
	}
	dir = filepath.Clean(dir)
	// The directory is excluded from file scanning; a volume root would
	// exclude the whole drive.
	if filepath.Dir(dir) == dir {
		return "", errors.New("signature_db_dir must not be a volume root")
	}
	return dir, nil
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
