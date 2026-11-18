package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
)

// This file is the EDR module's local management CLI — the single endpoint for
// inspecting and changing the module's configuration and quarantine store. It
// only ever edits <install>/edr.json (still fully hand-editable) and reads
// edr.db; the running service polls edr.json and hot-applies reloadable settings
// (allowlists — incl. the network never-block list — ransomware response mode, and
// blocklist.enforce) within a few seconds. Structural settings (enabled,
// blocklist.enabled, sensors, watched volumes, engine) take effect on the next
// disable-edr/enable-edr.

const reloadNote = "saved. The running service applies reloadable settings (allowlists, response mode, blocklist.enforce) within a few seconds; structural changes (enabled, blocklist.enabled, sensors, engine) need disable-edr/enable-edr."

func printUsage() {
	fmt.Print(`UTMStack EDR — management CLI

Lifecycle (run via the agent, needs server creds):
  utmstack_agent enable-edr | disable-edr | edr-status

Module commands (utmstack_edr ...):
  version [--json]                Show package and native scanner versions
  engine-check                     Validate local native signature generation (JSON)
  status                         Show the running service status
  scan <path>                    On-demand scan a file
  scan --dry-run --json [--db <dir>] [--workers N] <path>...
                                 Analyze files and folders and print one JSON
                                 line per file; nothing is quarantined or recorded
  config show                    Print the effective configuration
  config get <key>               Get one value (dotted, e.g. ransomware.response_mode)
  config set <key> <value>       Set a scalar knob (validated)

  allow path    add|remove|list [value]   Paths the watcher, launch guard and process
                                          telemetry skip (see the note below)
  allow process add|remove|list [value]   Processes/services the EDR won't act on
  allow command add|remove|list [value]   T1490 command substrings exempted
  allow network add|remove|list [value]   IPs/CIDRs the network blocklist never blocks

Network blocklist knobs (config set <key> <value>):
  blocklist.enabled true|false            Enable the network blocklist
  blocklist.enforce true|false            Enforce (drop) vs detect-only
  blocklist.allow_private_ranges true|false  Never block RFC1918/private IPs
  blocklist.refresh_hours <n>             Feed refresh interval in hours (>= 0)
  blocklist.direction both|out|in         Which traffic direction to block

Linux native protection (restart the service after changing these):
  linux_protection.process_events auto|connector|poll
  linux_protection.network_enforcement true|false       Opt in to leased nftables
  linux_protection.fanotify.enabled true|false          Opt in to scoped file access
  linux_protection.fanotify.mounts '["/data"]'           Explicit actual mount points
  linux_protection.fanotify.roots '["/data/protected"]'  Protected directories
  linux_protection.fanotify.mode notify|permission_exec|permission_open
  linux_protection.fanotify.fail_closed true|false       Decision on incomplete scans
  linux_protection.fanotify.deadline_ms <100..5000>
  linux_protection.fanotify.workers <1..8>
  linux_protection.fanotify.max_pending <workers..64>
  File access monitoring requires file_watcher enabled. Linux ransomware uses
  verified canaries within these explicit roots and declared mount identities.

Linux finite host isolation (requires root and nftables):
  network isolate --management <IP,IP> [--resolver <IP,IP>]
                  [--allow <CIDR,CIDR>] [--duration 5m]
  network status                         Verify the actual kernel isolation rules
  network release                        Remove only this module's isolation table
  Management addresses must resolve positively; duration must be <=1h. Rules
  allow exact management IPs and DNS port 53, and expire if the process exits.

Scan budgets (config set <key> <value>; restart the service to apply):
  scan_concurrency <1..16>                Worker ceiling shared by engine callers
  scan_limits.max_file_size_mb <0..200>   Per-object MiB (including decoded files)
  scan_limits.max_scan_size_mb <0..400>   Total MiB analyzed per input
  scan_limits.max_recursion <0..16>       Nested-container recursion budget
  scan_limits.max_files <0..10000>        Object budget per input
  scan_limits.max_scan_time_ms <0..120000> Scan and worker-wait time budget
  Zero selects a host-derived default; it never disables a limit. Exhausted
  budgets produce incomplete analysis. These budgets do not cap process memory.

Heuristic alerts (restart the service to apply):
  scan_heuristics.pdf_obfuscated_name true|false
      Alert on hex-escaped common PDF dictionary names. Enabled by default;
      this is a heuristic, including for otherwise benign PDFs.

  config set scheduled_scan.paths '["/absolute/path"]'   Set scheduled roots (JSON array)
  config set scheduled_scan.enabled true                Enable bounded periodic sweeps
  config set scheduled_scan.run_on_start true           Sweep at service startup
  config set scheduled_scan.interval_minutes 240        Set interval after each sweep
  quarantine list                List quarantined items (id | date | detection | path | state)
  quarantine inspect <id>        Capture recovery journals as JSON without replay or DB access
  quarantine hold <id> --inspection <saved-json> --reason <text>
                           Retain a reviewed restore case without replay or DB access
  quarantine release-hold <id> --hold <decision-uuid> --inspection <saved-json> --reason <text>
                           Retain hold evidence and authorize normal recovery of the original operation
  quarantine cancel-restore <id> --hold <decision-uuid> --inspection <saved-json> --reason <text>
                           Archive the original operation while retaining the active hold
  quarantine inspect-fresh <id> --hold <decision-uuid> --destination <absolute-local-path>
                           Inspect the retained snapshot and a new destination without authorizing publication
  quarantine approve-fresh <id> --operation <new-uuid> --inspection <saved-fresh-json> --reason <text>
                           Record independent approval while keeping the original retention hold
  quarantine execute-fresh <id> --approval <saved-approval-json> --reason <text>
  quarantine inspect-fresh-cancel <id> --operation <new-operation-uuid>
  quarantine inspect-retirement <id> --hold <original-hold-uuid> --stage <retained-original-stage>
  quarantine retire-cancelled <id> --inspection <saved-retirement-json> --reason <text> [--stage <same-renamed-original-stage>]
      Inspect the exact incomplete fresh operation without opening its database
  quarantine cancel-fresh <id> --operation <new-operation-uuid> --inspection <saved-cancellation-json> --reason <text>
      Archive the same reviewed fresh operation and preserve its retention fence
                           Execute or retry that exact new operation; retain original evidence and hold
  quarantine recover             Replay interrupted admissions; report retained conflicts
  quarantine restore <id>        Release an item back to its original path
  quarantine purge   <id>        Permanently delete one quarantined item (logged)
  quarantine purge --expired     Purge items older than quarantine_retention_days now

An allowlisted path covers files and programs: the real-time watcher never scans
or quarantines files there, and programs started from there are not scanned,
suspended or stopped at launch and are left out of process telemetry. The
ransomware guard ignores file changes there except canary tampering, but still
scores those programs and can stop them and quarantine their image (exempt one
with 'allow process'). An on-demand 'scan <path>' ignores the allowlist.

Built-in allowlists (backup/imaging/DR/sync + Windows VSS/Search, and the EDR's
own working tree) always apply on top of your config and cannot be removed.
`)
}

// ---- config ----

func runConfig(args []string) {
	if len(args) == 0 {
		fmt.Println("usage: utmstack_edr config show|get <key>|set <key> <value>")
		os.Exit(1)
	}
	switch args[0] {
	case "show":
		cfg, err := config.Load()
		exitOn(err)
		b, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(b))
	case "get":
		if len(args) < 2 {
			fmt.Println("usage: utmstack_edr config get <key>")
			os.Exit(1)
		}
		cfg, err := config.Load()
		exitOn(err)
		v, ok := getByPath(cfg, args[1])
		if !ok {
			fmt.Printf("unknown key: %s\n", args[1])
			os.Exit(1)
		}
		fmt.Println(v)
	case "set":
		if len(args) < 3 {
			fmt.Println("usage: utmstack_edr config set <key> <value>")
			os.Exit(1)
		}
		cfg, err := config.Load()
		exitOn(err)
		if err := applySet(&cfg, args[1], strings.Join(args[2:], " ")); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		exitOn(config.Save(cfg))
		fmt.Printf("set %s = %s (%s)\n", args[1], strings.Join(args[2:], " "), reloadNote)
	default:
		fmt.Println("usage: utmstack_edr config show|get <key>|set <key> <value>")
		os.Exit(1)
	}
}

// omittedDefaults holds the value of keys whose JSON field is left out while it
// holds its default (omitempty), so `config get` reports that default instead
// of calling the key unknown. An unset sensor is on.
var omittedDefaults = map[string]string{
	"signature_db_dir":                    "",
	"signature_public_key":                "",
	"allowlist.networks":                  "",
	"sensors.file_watcher":                "true",
	"sensors.process_guard":               "true",
	"sensors.amsi":                        "true",
	"sensors.behavioral":                  "true",
	"scan_heuristics.pdf_obfuscated_name": "true",
}

// getByPath returns the dotted-path value from the config (via its JSON form) as
// a printable string.
func getByPath(cfg config.EDRConfig, key string) (string, bool) {
	b, _ := json.Marshal(cfg)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	var cur any = m
	for _, part := range strings.Split(key, ".") {
		mm, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		cur, ok = mm[part]
		if !ok {
			v, known := omittedDefaults[key]
			return v, known
		}
	}
	switch v := cur.(type) {
	case nil:
		return "", true
	case []any:
		parts := make([]string, len(v))
		for i, e := range v {
			parts[i] = fmt.Sprint(e)
		}
		return "[" + strings.Join(parts, ", ") + "]", true
	default:
		return fmt.Sprint(v), true
	}
}

// applySet validates and applies a scalar config change. Only safe, documented
// keys are settable; unknown or malformed values are rejected (so a typo can't
// silently corrupt the config).
func applySet(cfg *config.EDRConfig, key, val string) error {
	if strings.HasPrefix(key, "linux_protection.") {
		return setLinuxProtection(cfg, key, val)
	}
	if strings.HasPrefix(key, "scheduled_scan.") {
		return setScheduledScan(cfg, key, val)
	}
	switch key {
	case "enabled":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Enabled = b
	case "fail_mode":
		if val != "open" && val != "closed" {
			return fmt.Errorf("fail_mode must be open|closed")
		}
		cfg.FailMode = val
	case "suspend_on_launch":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.SuspendOnLaunch = b
	case "suspend_timeout_ms":
		return setPositiveInt(&cfg.SuspendTimeoutMs, val)
	case "quarantine_retention_days":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			return fmt.Errorf("quarantine_retention_days must be an integer >= 0 (0 = keep forever)")
		}
		cfg.QuarantineDays = n
	case "scan_concurrency":
		return setScanResource(cfg, key, val)
	case "scan_limits.max_file_size_mb", "scan_limits.max_scan_size_mb", "scan_limits.max_recursion", "scan_limits.max_files", "scan_limits.max_scan_time_ms":
		return setScanResource(cfg, key, val)
	case "sig_update_hours":
		return setPositiveInt(&cfg.SigUpdateHours, val)
	case "engine_tier_override":
		if val != "" && val != "constrained" && val != "standard" && val != "server" {
			return fmt.Errorf("engine_tier_override must be constrained|standard|server (or empty to auto-derive)")
		}
		cfg.TierOverride = val
	case "signature_mirror":
		cfg.SigMirror = val
	case "signature_db_dir":
		dir, err := config.NormalizeSignatureDBDir(val)
		if err != nil {
			return err
		}
		cfg.SignatureDBDir = dir
	case "signature_public_key":
		key, err := base64.StdEncoding.DecodeString(val)
		if err != nil || len(key) != ed25519.PublicKeySize {
			return fmt.Errorf("signature_public_key must be a base64 Ed25519 public key")
		}
		cfg.SignaturePublicKey = val
	case "sensors.file_watcher":
		return setSensor(&cfg.Sensors.FileWatcher, val)
	case "sensors.process_guard":
		return setSensor(&cfg.Sensors.ProcessGuard, val)
	case "sensors.amsi":
		return setSensor(&cfg.Sensors.AMSI, val)
	case "sensors.behavioral":
		return setSensor(&cfg.Sensors.Behavioral, val)
	case "scan_heuristics.pdf_obfuscated_name":
		return setSensor(&cfg.ScanHeuristics.PDFObfuscatedName, val)
	case "ransomware.enabled":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Ransomware.Enabled = b
	case "ransomware.response_mode":
		if val != "alert" && val != "suspend" && val != "kill" {
			return fmt.Errorf("ransomware.response_mode must be alert|suspend|kill")
		}
		cfg.Ransomware.ResponseMode = val
	case "ransomware.suspend_threshold":
		return setPositiveInt(&cfg.Ransomware.SuspendThreshold, val)
	case "ransomware.kill_threshold":
		return setPositiveInt(&cfg.Ransomware.KillThreshold, val)
	case "ransomware.decay_half_life_ms":
		return setPositiveInt(&cfg.Ransomware.DecayHalfLifeMs, val)
	case "ransomware.canary_per_dir":
		return setPositiveInt(&cfg.Ransomware.CanaryPerDir, val)
	case "blocklist.enabled":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Blocklist.Enabled = b
	case "blocklist.enforce":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Blocklist.Enforce = b
	case "blocklist.allow_private_ranges":
		b, err := parseBool(val)
		if err != nil {
			return err
		}
		cfg.Blocklist.AllowPrivateRanges = b
	case "blocklist.refresh_hours":
		n, err := strconv.Atoi(val)
		if err != nil || n < 0 {
			return fmt.Errorf("refresh_hours must be a non-negative integer")
		}
		cfg.Blocklist.RefreshHours = n
	case "blocklist.direction":
		if val != "both" && val != "out" && val != "in" {
			return fmt.Errorf("direction must be both|out|in")
		}
		cfg.Blocklist.Direction = val
	default:
		return fmt.Errorf("unknown or read-only key %q (use `allow ...` for whitelists; `config show` to list values)", key)
	}
	return nil
}

func setScheduledScan(cfg *config.EDRConfig, key, val string) error {
	next := cfg.ScheduledScan
	var err error
	switch key {
	case "scheduled_scan.enabled":
		next.Enabled, err = parseBool(val)
	case "scheduled_scan.run_on_start":
		next.RunOnStart, err = parseBool(val)
	case "scheduled_scan.paths":
		err = json.Unmarshal([]byte(val), &next.Paths)
	case "scheduled_scan.interval_minutes":
		next.IntervalMinutes, err = strconv.Atoi(val)
	case "scheduled_scan.max_files":
		next.MaxFiles, err = strconv.Atoi(val)
	case "scheduled_scan.max_entries":
		next.MaxEntries, err = strconv.Atoi(val)
	case "scheduled_scan.max_depth":
		next.MaxDepth, err = strconv.Atoi(val)
	case "scheduled_scan.max_duration_minutes":
		next.MaxDurationMinutes, err = strconv.Atoi(val)
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", key, err)
	}
	if err = next.Validate(); err != nil {
		return err
	}
	cfg.ScheduledScan = next
	return nil
}

func setScanResource(cfg *config.EDRConfig, key, val string) error {
	n, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("%s must be an integer", key)
	}
	next := *cfg
	switch key {
	case "scan_concurrency":
		if n < 1 {
			return fmt.Errorf("scan_concurrency must be an integer from 1 to %d", config.MaxScanConcurrency)
		}
		next.ScanConcurrency = n
	case "scan_limits.max_file_size_mb":
		next.ScanLimits.MaxFileSizeMB = n
	case "scan_limits.max_scan_size_mb":
		next.ScanLimits.MaxScanSizeMB = n
	case "scan_limits.max_recursion":
		next.ScanLimits.MaxRecursion = n
	case "scan_limits.max_files":
		next.ScanLimits.MaxFiles = n
	case "scan_limits.max_scan_time_ms":
		next.ScanLimits.MaxScanTimeMS = n
	}
	if err := next.ValidateScanResources(); err != nil {
		return err
	}
	*cfg = next
	return nil
}

// ---- allow (unified whitelists) ----

func runAllow(args []string) {
	if len(args) < 2 {
		fmt.Println("usage: utmstack_edr allow path|process|command|network add|remove|list [value]")
		os.Exit(1)
	}
	category, action := args[0], args[1]
	cfg, err := config.Load()
	exitOn(err)

	var list *[]string
	switch category {
	case "path":
		list = &cfg.Allowlist.Paths
	case "process":
		list = &cfg.Allowlist.Processes
	case "command":
		list = &cfg.Allowlist.Commands
	case "network":
		list = &cfg.Allowlist.Networks
	default:
		fmt.Println("category must be path|process|command|network")
		os.Exit(1)
	}

	switch action {
	case "list":
		if len(*list) == 0 {
			fmt.Printf("(no user %s entries; built-in defaults still apply)\n", category)
			return
		}
		for _, e := range *list {
			fmt.Println(e)
		}
	case "add":
		if len(args) < 3 {
			fmt.Printf("usage: utmstack_edr allow %s add <value>\n", category)
			os.Exit(1)
		}
		val := strings.Join(args[2:], " ")
		if category == "network" {
			if _, errAddr := netip.ParseAddr(val); errAddr != nil {
				if _, errPfx := netip.ParsePrefix(val); errPfx != nil {
					fmt.Println("network entry must be an IP or CIDR")
					os.Exit(1)
				}
			}
		}
		if containsFold(*list, val) {
			fmt.Printf("%s already allowlisted: %s\n", category, val)
			return
		}
		*list = append(*list, val)
		exitOn(config.Save(cfg))
		fmt.Printf("added %s allowlist: %s (%s)\n", category, val, reloadNote)
	case "remove":
		if len(args) < 3 {
			fmt.Printf("usage: utmstack_edr allow %s remove <value>\n", category)
			os.Exit(1)
		}
		val := strings.Join(args[2:], " ")
		out := (*list)[:0]
		removed := false
		for _, e := range *list {
			if strings.EqualFold(strings.TrimSpace(e), strings.TrimSpace(val)) {
				removed = true
				continue
			}
			out = append(out, e)
		}
		if !removed {
			fmt.Printf("not found in %s allowlist: %s\n", category, val)
			os.Exit(1)
		}
		*list = out
		exitOn(config.Save(cfg))
		fmt.Printf("removed %s allowlist: %s (%s)\n", category, val, reloadNote)
	default:
		fmt.Println("action must be add|remove|list")
		os.Exit(1)
	}
}

// ---- quarantine ----

func runQuarantine(args []string) {
	if len(args) == 0 {
		fmt.Println("usage: utmstack_edr quarantine list|recover|restore <id>|purge <id>|purge --expired")
		os.Exit(1)
	}
	switch args[0] {
	case "inspect":
		os.Exit(runQuarantineInspection(args[1:], os.Stdout, os.Stderr))
	case "list":
		c, err := cache.Open(config.DBFile)
		exitOn(err)
		defer c.Close()
		cfg, err := config.Load()
		exitOn(err)
		store, err := quarantine.New(cfg.QuarantineDir, c)
		exitOn(err)
		for _, r := range store.RecoveryStatus() {
			if r.Status == "needs_attention" {
				fmt.Printf("Recovery needed for %s (%s): %s\n", r.QuarantineID, r.OriginalPath, r.Detail)
			}
		}
		recs, err := c.ListQuarantine()
		exitOn(err)
		if len(recs) == 0 {
			fmt.Println("(quarantine is empty)")
			return
		}
		sort.Slice(recs, func(i, j int) bool { return recs[i].QuarantinedAt.After(recs[j].QuarantinedAt) })
		fmt.Printf("%-38s  %-20s  %-26s  %-8s  %-12s  %s\n", "ID", "DATE (UTC)", "DETECTION", "STATE", "RAW RETAINED", "ORIGINAL PATH")
		for _, r := range recs {
			rawRetained := "no"
			if r.SourceCaptureDir != "" && !r.Purged {
				rawRetained = "yes"
			}
			fmt.Printf("%-38s  %-20s  %-26s  %-8s  %-12s  %s\n",
				r.QuarantineID, r.QuarantinedAt.Format("2006-01-02 15:04:05"),
				truncate(r.Detection, 26), quarState(r), rawRetained, r.OriginalPath)
		}
	case "recover":
		cfg, err := config.Load()
		exitOn(err)
		c, err := cache.Open(config.DBFile)
		exitOn(err)
		defer c.Close()
		store, err := quarantine.New(cfg.QuarantineDir, c)
		exitOn(err)
		results := store.RecoveryStatus()
		if len(results) == 0 {
			fmt.Println("No pending quarantine operations.")
			return
		}
		pending := false
		for _, r := range results {
			operation := r.Operation
			if operation == "" {
				operation = "admission"
			}
			fmt.Printf("%s  %s  %s  %s\n", r.QuarantineID, operation, r.Status, r.Detail)
			pending = pending || r.Status == "needs_attention"
		}
		if pending {
			os.Exit(1)
		}
	case "restore":
		if len(args) < 2 {
			fmt.Println("usage: utmstack_edr quarantine restore <id>")
			os.Exit(1)
		}
		runRestore(args[1])
	case "purge":
		cfg, _ := config.Load()
		c, err := cache.Open(config.DBFile)
		exitOn(err)
		defer c.Close()
		store, err := quarantine.New(cfg.QuarantineDir, c)
		exitOn(err)
		if len(args) >= 2 && args[1] == "--expired" {
			n, err := store.PurgeExpired(cfg.QuarantineDays, time.Now())
			exitOn(err)
			fmt.Printf("UTMStack EDR: purged %d expired quarantined item(s)\n", n)
			return
		}
		if len(args) < 2 {
			fmt.Println("usage: utmstack_edr quarantine purge <id> | purge --expired")
			os.Exit(1)
		}
		if err := store.Purge(args[1]); err != nil {
			fmt.Println("purge error:", err)
			os.Exit(1)
		}
		fmt.Println("UTMStack EDR: purged", args[1])
	default:
		fmt.Println("usage: utmstack_edr quarantine list|recover|restore <id>|purge <id>|purge --expired")
		os.Exit(1)
	}
}

// ---- helpers ----

func exitOn(err error) {
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "on", "yes", "1", "enabled":
		return true, nil
	case "false", "off", "no", "0", "disabled":
		return false, nil
	}
	return false, fmt.Errorf("expected a boolean (true/false/on/off), got %q", s)
}

func setPositiveInt(dst *int, val string) error {
	n, err := strconv.Atoi(val)
	if err != nil || n <= 0 {
		return fmt.Errorf("expected a positive integer, got %q", val)
	}
	*dst = n
	return nil
}

func setSensor(dst **bool, val string) error {
	b, err := parseBool(val)
	if err != nil {
		return err
	}
	*dst = &b
	return nil
}

func containsFold(list []string, v string) bool {
	for _, e := range list {
		if strings.EqualFold(strings.TrimSpace(e), strings.TrimSpace(v)) {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

func quarState(r cache.QuarantineRecord) string {
	switch {
	case r.Purged:
		return "purged"
	case r.Restored:
		return "restored"
	default:
		return "held"
	}
}
