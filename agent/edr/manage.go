package main

import (
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

// usageText is the exact help body; printUsage and the --json help verb both
// use it so the two never drift.
func usageText() string {
	return `UTMStack EDR — management CLI

Lifecycle (run via the agent, needs server creds):
  utmstack_agent enable-edr | disable-edr | edr-status

Module commands (utmstack_edr ...):
  status                         Show the running service status
  scan <path>                    On-demand scan a file
  config show                    Print the effective configuration
  config get <key>               Get one value (dotted, e.g. ransomware.response_mode)
  config set <key> <value>       Set a scalar knob (validated)

  allow path    add|remove|list [value]   File paths the watcher skips
  allow process add|remove|list [value]   Processes/services the EDR won't act on
  allow command add|remove|list [value]   T1490 command substrings exempted
  allow network add|remove|list [value]   IPs/CIDRs the network blocklist never blocks

Network blocklist knobs (config set <key> <value>):
  blocklist.enabled true|false            Enable the network blocklist
  blocklist.enforce true|false            Enforce (drop) vs detect-only
  blocklist.allow_private_ranges true|false  Never block RFC1918/private IPs
  blocklist.refresh_hours <n>             Feed refresh interval in hours (>= 0)
  blocklist.direction both|out|in         Which traffic direction to block

  quarantine list                List quarantined items (id | date | detection | path | state)
  quarantine restore <id>        Release an item back to its original path
  quarantine purge   <id>        Permanently delete one quarantined item (logged)
  quarantine purge --expired     Purge items older than quarantine_retention_days now

Built-in allowlists (backup/imaging/DR/sync + Windows VSS/Search, and the EDR's
own working tree) always apply on top of your config and cannot be removed.
`
}

func printUsage() {
	fmt.Print(usageText())
}

// ---- config ----

func runConfig(args []string) {
	if len(args) == 0 {
		failUsage("usage: utmstack_edr config show|get <key>|set <key> <value>")
	}
	switch args[0] {
	case "show":
		if jsonMode {
			cfg, err := config.Load()
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", cfg)
			return
		}
		cfg, err := config.Load()
		exitOn(err)
		b, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(b))
	case "get":
		if len(args) < 2 {
			failUsage("usage: utmstack_edr config get <key>")
		}
		if jsonMode {
			data, err := configGetData(args[1])
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
			return
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
			failUsage("usage: utmstack_edr config set <key> <value>")
		}
		val := strings.Join(args[2:], " ")
		if jsonMode {
			data, err := configSetData(args[1], val)
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
			return
		}
		cfg, err := config.Load()
		exitOn(err)
		if err := applySet(&cfg, args[1], val); err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		exitOn(config.Save(cfg))
		fmt.Printf("set %s = %s (%s)\n", args[1], val, reloadNote)
	default:
		failUsage("usage: utmstack_edr config show|get <key>|set <key> <value>")
	}
}

// configGetData is the --json core of `config get <key>`: it returns the
// dotted-path value without printing.
func configGetData(key string) (interface{}, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	v, ok := getByPath(cfg, key)
	if !ok {
		return nil, fmt.Errorf("unknown key: %s", key)
	}
	return map[string]interface{}{"key": key, "value": v}, nil
}

// configSetData is the --json core of `config set <key> <value>`: it validates,
// persists, and returns the applied value without printing.
func configSetData(key, val string) (interface{}, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if err := applySet(&cfg, key, val); err != nil {
		return nil, err
	}
	if err := config.Save(cfg); err != nil {
		return nil, err
	}
	return map[string]interface{}{"key": key, "value": val, "applied": true}, nil
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
			return "", false
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
		return setPositiveInt(&cfg.ScanConcurrency, val)
	case "sig_update_hours":
		return setPositiveInt(&cfg.SigUpdateHours, val)
	case "engine_tier_override":
		if val != "" && val != "constrained" && val != "standard" && val != "server" {
			return fmt.Errorf("engine_tier_override must be constrained|standard|server (or empty to auto-derive)")
		}
		cfg.TierOverride = val
	case "signature_mirror":
		cfg.SigMirror = val
	case "signature_fallback":
		if val != "" && val != "cdn" && val != "none" {
			return fmt.Errorf("signature_fallback must be empty, \"cdn\", or \"none\"")
		}
		cfg.SignatureFallback = val
	case "sensors.file_watcher":
		return setSensor(&cfg.Sensors.FileWatcher, val)
	case "sensors.process_guard":
		return setSensor(&cfg.Sensors.ProcessGuard, val)
	case "sensors.amsi":
		return setSensor(&cfg.Sensors.AMSI, val)
	case "sensors.behavioral":
		return setSensor(&cfg.Sensors.Behavioral, val)
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

// ---- allow (unified whitelists) ----

func runAllow(args []string) {
	if len(args) < 2 {
		failUsage("usage: utmstack_edr allow path|process|command|network add|remove|list [value]")
	}
	category, action := args[0], args[1]
	if jsonMode {
		switch action {
		case "add":
			if len(args) < 3 {
				failUsage(fmt.Sprintf("usage: utmstack_edr allow %s add <value>", category))
			}
			data, err := allowAddData(category, strings.Join(args[2:], " "))
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
		case "remove":
			if len(args) < 3 {
				failUsage(fmt.Sprintf("usage: utmstack_edr allow %s remove <value>", category))
			}
			data, err := allowRemoveData(category, strings.Join(args[2:], " "))
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
		case "list":
			data, err := allowListData(category)
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
		default:
			failJSON("action must be add|remove|list")
		}
		return
	}
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

// allowListData is the --json core of `allow <cat> list`: it returns the
// entries without printing.
func allowListData(category string) (interface{}, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	list := allowListPtr(&cfg, category)
	if list == nil {
		return nil, fmt.Errorf("category must be path|process|command|network")
	}
	entries := make([]string, len(*list))
	copy(entries, *list)
	return map[string]interface{}{
		"category": category,
		"entries":  entries,
		"count":    len(entries),
	}, nil
}

// allowAddData is the --json core of `allow <cat> add <val>`: it validates,
// persists, and reports the new count (plus already_present on a duplicate)
// without printing.
func allowAddData(category, val string) (interface{}, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	list := allowListPtr(&cfg, category)
	if list == nil {
		return nil, fmt.Errorf("category must be path|process|command|network")
	}
	if category == "network" {
		if _, errAddr := netip.ParseAddr(val); errAddr != nil {
			if _, errPfx := netip.ParsePrefix(val); errPfx != nil {
				return nil, fmt.Errorf("network entry must be an IP or CIDR")
			}
		}
	}
	present := containsFold(*list, val)
	if !present {
		*list = append(*list, val)
		if err := config.Save(cfg); err != nil {
			return nil, err
		}
	}
	data := map[string]interface{}{
		"category": category,
		"added":    val,
		"count":    len(*list),
	}
	if present {
		data["already_present"] = true
	}
	return data, nil
}

// allowRemoveData is the --json core of `allow <cat> remove <val>`: it
// persists and reports the new count, or an error when the entry is missing,
// without printing.
func allowRemoveData(category, val string) (interface{}, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	list := allowListPtr(&cfg, category)
	if list == nil {
		return nil, fmt.Errorf("category must be path|process|command|network")
	}
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
		return nil, fmt.Errorf("not found in %s allowlist: %s", category, val)
	}
	*list = out
	if err := config.Save(cfg); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"category": category,
		"removed":  val,
		"count":    len(*list),
	}, nil
}

// allowListPtr returns the *[]string backing the given allowlist category
// (nil for an unknown category).
func allowListPtr(cfg *config.EDRConfig, category string) *[]string {
	switch category {
	case "path":
		return &cfg.Allowlist.Paths
	case "process":
		return &cfg.Allowlist.Processes
	case "command":
		return &cfg.Allowlist.Commands
	case "network":
		return &cfg.Allowlist.Networks
	default:
		return nil
	}
}

// ---- quarantine ----

func runQuarantine(args []string) {
	if len(args) == 0 {
		failUsage("usage: utmstack_edr quarantine list|restore <id>|purge <id>|purge --expired")
	}
	if jsonMode {
		switch args[0] {
		case "list":
			data, err := quarListData()
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
			return
		case "restore":
			if len(args) < 2 {
				failUsage("usage: utmstack_edr quarantine restore <id>")
			}
			if err := restoreData(args[1]); err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", map[string]interface{}{"restored": args[1]})
			return
		case "purge":
			data, err := quarPurgeData(args[1:])
			if err != nil {
				failJSON(err.Error())
			}
			emitJSON(true, "", data)
			return
		default:
			failJSON("unknown quarantine action: " + args[0])
		}
	}
	switch args[0] {
	case "list":
		c, err := cache.Open(config.DBFile)
		exitOn(err)
		defer c.Close()
		recs, err := c.ListQuarantine()
		exitOn(err)
		if len(recs) == 0 {
			fmt.Println("(quarantine is empty)")
			return
		}
		sort.Slice(recs, func(i, j int) bool { return recs[i].QuarantinedAt.After(recs[j].QuarantinedAt) })
		fmt.Printf("%-38s  %-20s  %-26s  %-8s  %s\n", "ID", "DATE (UTC)", "DETECTION", "STATE", "ORIGINAL PATH")
		for _, r := range recs {
			fmt.Printf("%-38s  %-20s  %-26s  %-8s  %s\n",
				r.QuarantineID, r.QuarantinedAt.Format("2006-01-02 15:04:05"),
				truncate(r.Detection, 26), quarState(r), r.OriginalPath)
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
		fmt.Println("usage: utmstack_edr quarantine list|restore <id>|purge <id>|purge --expired")
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

// quarItemData maps one quarantine record to its --json list item.
func quarItemData(r cache.QuarantineRecord) map[string]interface{} {
	return map[string]interface{}{
		"id":        r.QuarantineID,
		"date":      r.QuarantinedAt.Format("2006-01-02 15:04:05"),
		"detection": r.Detection,
		"path":      r.OriginalPath,
		"state":     quarState(r),
	}
}

// quarListData is the --json core of `quarantine list`: the full record set
// (newest first, like the human table) mapped to list items, without printing.
func quarListData() (interface{}, error) {
	c, err := cache.Open(config.DBFile)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	recs, err := c.ListQuarantine()
	if err != nil {
		return nil, err
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].QuarantinedAt.After(recs[j].QuarantinedAt) })
	items := make([]map[string]interface{}, len(recs))
	for i, r := range recs {
		items[i] = quarItemData(r)
	}
	return map[string]interface{}{"items": items, "count": len(items)}, nil
}

// quarPurgeData is the --json core of `quarantine purge <id>` /
// `quarantine purge --expired`, without printing. rest is the args after
// "purge".
func quarPurgeData(rest []string) (interface{}, error) {
	cfg, _ := config.Load()
	c, err := cache.Open(config.DBFile)
	if err != nil {
		return nil, err
	}
	defer c.Close()
	store, err := quarantine.New(cfg.QuarantineDir, c)
	if err != nil {
		return nil, err
	}
	if len(rest) >= 1 && rest[0] == "--expired" {
		n, err := store.PurgeExpired(cfg.QuarantineDays, time.Now())
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"purged_count": n}, nil
	}
	if len(rest) < 1 {
		return nil, fmt.Errorf("usage: utmstack_edr quarantine purge <id> | purge --expired")
	}
	if err := store.Purge(rest[0]); err != nil {
		return nil, fmt.Errorf("purge error: %w", err)
	}
	return map[string]interface{}{"purged": rest[0]}, nil
}
