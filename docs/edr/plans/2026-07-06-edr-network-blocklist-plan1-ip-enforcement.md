# EDR Network Blocklist — Plan 1: Feed + IP Enforcement (core)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a working, enforcing IP/CIDR network blocklist for the UTMStack EDR: download ThreatWinds indicators from the UTMStack mirror, block outbound+inbound connections to/from blacklisted IPs via user-mode WFP, and emit branded `network_watcher` events for every block — with a self-lockout allowlist and fail-open-on-crash safety.

**Architecture:** New self-contained package `agent/edr/netblock`, orchestrated by `service.startPipeline` and emitting through the existing spool. Portable core (feed downloader, parser, in-memory store, matcher, allowlist, enforcer diff engine) is TDD'd on macOS; the WFP filter engine and the WFP net-event audit reader are Windows-native behind `//go:build windows` with `!windows` no-op stubs, verified on the test VM. Enforcement is on by default; the DNS/domain reactive path is deferred to Plan 2.

**Tech Stack:** Go 1.25.5, `golang.org/x/sys/windows` (already an indirect dep) for WFP via `fwpuclnt.dll`, GORM + `glebarez/sqlite` (existing `cache`), `shared/http` downloader, `kardianos/service` (existing `service`). No new third-party Go dependency.

## Global Constraints

- **Branding (critical):** every emitted event and every surfaced log/status string says **"UTMStack EDR"**. The strings `clam`/`clamd` **and** `threatwinds` must **never** appear in an event field or a surfaced log line. Internal Go identifiers, config keys, and mirror URLs may reference `threatwinds`. The intel source is surfaced as **"UTMStack Threat Intelligence"**. Labeling audit must stay green (extended below).
- **Go-first.** All code is Go. No new language, no third-party Go dependency beyond `golang.org/x/sys/windows`.
- **No kernel driver, ever.** Enforcement is user-mode WFP only.
- **Own `edr.db`.** New tables go in the EDR's own cache, never the agent's `logs.db`.
- **SYSTEM service.** WFP `FwpmEngineOpen0` and filter add require the SYSTEM privileges the EDR service already runs with.
- **Fail-open.** A crashed/stopped EDR must never leave a machine blackholed — WFP filters live in a dynamic session that auto-releases on handle close.
- **Supported arches:** windows/amd64 **and** windows/arm64. Every task's code must cross-build and `go vet` on both. All portable code also builds/tests on darwin (host).
- **Config key:** the feature's config block is `blocklist` in `edr.json`. Package is `agent/edr/netblock`. Event source is `network_watcher`.
- All Go commands run from `utmstack-v12/agent/`.

**Spec:** `docs/superpowers/specs/2026-07-06-edr-network-blocklist-design.md`

---

## Mirror contract (external dependency — assumed by this plan)

The UTMStack platform mirror serves, at base URL `<mirror>/feeds/v1/`, per `(level, name)` where `level ∈ {level1,level2,level3}` and `name ∈ {ip}` (Plan 1 scope):

- **Accumulative:** `GET download/list/{level}/accumulative/{name}` → `application/gzip`, a gzip'd UTF-8 body of **one indicator value per line** (an IP `1.2.3.4` / `2001:db8::1` or a CIDR `10.0.0.0/8`), blank lines and `#`-comments ignored.
- **Daily:** `GET download/list/{level}/daily/{name}` → `application/x-ndjson`, one JSON object per line: `{"value":"1.2.3.4","type":"ip","op":"add"}` where `op ∈ {add,del}` and `type ∈ {ip,cidr}`.
- **Checksum:** `GET download/checksum?level={level}&type={type}&name={name}` → text body `sha256hex` for the corresponding artifact.

This normalized shape is the contract the parser (Task 7) targets; confirm/hand it to the platform team (spec §10.1).

---

## File structure (Plan 1)

| File | Responsibility |
|---|---|
| `edr/config/config.go` (modify) | Add `BlocklistConfig`, default, load-overlay, `BlocklistDir` path const. |
| `edr/event/event.go` (modify) | Add `SourceNetwork`, network fields, `NewNetworkEvent`. |
| `edr/cache/netblock.go` (create) | `IndicatorRecord`, `NetBlockRecord`, `FeedState` + accessors. |
| `edr/cache/cache.go` (modify) | Add the three records to `AutoMigrate`. |
| `edr/netblock/indicator.go` (create) | `Indicator` value type + normalization. |
| `edr/netblock/matcher.go` (create) | Pure IP/CIDR membership. |
| `edr/netblock/allowlist.go` (create) | Never-block set (built-ins + private + configured). |
| `edr/netblock/store.go` (create) | In-memory hot-swappable indicator set. |
| `edr/netblock/parse.go` (create) | Accumulative gzip + daily ndjson → indicators. |
| `edr/netblock/feed.go` (create) | Mirror downloader + scheduler + checksum. |
| `edr/netblock/enforcer.go` (create) | `Blocker` interface + diff engine. |
| `edr/netblock/wfp_windows.go` / `wfp_other.go` (create) | WFP filter engine / no-op stub. |
| `edr/netblock/connaudit_windows.go` / `connaudit_other.go` (create) | WFP net-event reader / no-op stub. |
| `edr/netblock/manager.go` (create) | Orchestrator + status health. |
| `edr/service/service.go` (modify) | Wire `netblock` into `startPipeline`; status fields; `program` handle. |
| `edr/manage.go` (modify) | `config set blocklist.*`, `allow network`, `block` verbs. |
| `*_test.go` | Sibling tests for every portable file. |

---

## Task 1: Config — `BlocklistConfig`

**Files:**
- Modify: `edr/config/config.go`
- Test: `edr/config/config_blocklist_test.go` (create)

**Interfaces:**
- Produces: `config.BlocklistConfig` struct; `EDRConfig.Blocklist BlocklistConfig`; `config.BlocklistDir string`.

- [ ] **Step 1: Write the failing test**

```go
// edr/config/config_blocklist_test.go
package config

import "testing"

func TestDefaultBlocklist(t *testing.T) {
	d := Default()
	b := d.Blocklist
	if !b.Enabled || !b.Enforce {
		t.Fatalf("blocklist should default enabled+enforcing, got %+v", b)
	}
	if !b.AllowPrivateRanges {
		t.Fatal("private ranges must be allowlisted by default")
	}
	if b.Direction != "both" {
		t.Fatalf("direction default = both, got %q", b.Direction)
	}
	if len(b.Levels) != 1 || b.Levels[0] != 1 {
		t.Fatalf("default levels = [1], got %v", b.Levels)
	}
	if b.RefreshHours != 6 {
		t.Fatalf("refresh default 6, got %d", b.RefreshHours)
	}
	if len(b.IndicatorTypes) == 0 || b.IndicatorTypes[0] != "ip" {
		t.Fatalf("indicator types must include ip, got %v", b.IndicatorTypes)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/config/ -run TestDefaultBlocklist -v`
Expected: FAIL (compile error — `Blocklist` undefined).

- [ ] **Step 3: Add the struct, path const, default, and load-overlay**

In `edr/config/config.go`, add near the other path constants (with `BlocklistDir` derived like the others):

```go
// BlocklistDir stages downloaded ThreatWinds feed artifacts.
var BlocklistDir = filepath.Join(InstallDir, "blocklist")
```

Add the struct (after `RansomwareConfig`):

```go
// BlocklistConfig configures the network blocklist (ThreatWinds via the UTMStack
// mirror). Enabled+enforcing by default; safety rails (allowlist, private-range
// exemption, fail-open) are mandatory because it can drop live traffic.
type BlocklistConfig struct {
	Enabled            bool     `json:"enabled"`              // default true
	Enforce            bool     `json:"enforce"`              // default true (false = detect-only)
	MirrorBaseURL      string   `json:"blocklist_mirror"`     // empty = derive from Server
	Levels             []int    `json:"levels"`               // default [1]
	IndicatorTypes     []string `json:"indicator_types"`      // default ["ip"] in Plan 1
	Direction          string   `json:"direction"`            // "both"|"out"|"in"
	RefreshHours       int      `json:"refresh_hours"`        // default 6
	AllowPrivateRanges bool     `json:"allow_private_ranges"` // default true
	DomainBlockTTLMin  int      `json:"domain_block_ttl_min"` // Plan 2; default 30
}
```

Add the field to `EDRConfig` (after `Ransomware`):

```go
	// Blocklist is the network blocklist block (nested; copied wholesale by Load
	// when present on disk).
	Blocklist BlocklistConfig `json:"blocklist"`
```

Seed it in `Default()` (inside the returned struct literal):

```go
		Blocklist: BlocklistConfig{
			Enabled:            true,
			Enforce:            true,
			Levels:             []int{1},
			IndicatorTypes:     []string{"ip"},
			Direction:          "both",
			RefreshHours:       6,
			AllowPrivateRanges: true,
			DomainBlockTTLMin:  30,
		},
```

Overlay in `Load()` (mirror the Ransomware presence-detect + wholesale copy; add after the Ransomware overlay block):

```go
	if onDisk.Blocklist.Enabled || onDisk.Blocklist.RefreshHours != 0 ||
		len(onDisk.Blocklist.Levels) != 0 || onDisk.Blocklist.Direction != "" {
		c.Blocklist = onDisk.Blocklist
		if len(c.Blocklist.Levels) == 0 {
			c.Blocklist.Levels = []int{1}
		}
		if len(c.Blocklist.IndicatorTypes) == 0 {
			c.Blocklist.IndicatorTypes = []string{"ip"}
		}
		if c.Blocklist.Direction == "" {
			c.Blocklist.Direction = "both"
		}
		if c.Blocklist.RefreshHours == 0 {
			c.Blocklist.RefreshHours = 6
		}
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/config/ -run TestDefaultBlocklist -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/config/config.go edr/config/config_blocklist_test.go
git commit -m "feat(edr): add blocklist config block (enabled+enforcing default)"
```

---

## Task 2: Event model — `network_watcher`

**Files:**
- Modify: `edr/event/event.go`
- Test: `edr/event/event_network_test.go` (create)

**Interfaces:**
- Produces: `event.SourceNetwork`; `Event.{RemoteIP,RemotePort,Domain,Direction,Indicator}`; `event.NewNetworkEvent(action, direction, remoteIP string, port int, domain, indicator string, p *ProcInfo) Event`.

- [ ] **Step 1: Write the failing test**

```go
// edr/event/event_network_test.go
package event

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewNetworkEventBranding(t *testing.T) {
	e := NewNetworkEvent(ActionBlocked, "outbound", "1.2.3.4", 443, "", "1.2.3.4", nil)
	js, err := e.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatal(err)
	}
	if m["source"] != "network_watcher" {
		t.Fatalf("source = %v", m["source"])
	}
	if m["engine"] != "UTMStack EDR" || m["product"] != "UTMStack EDR" {
		t.Fatalf("branding wrong: %v", m)
	}
	if m["remote_ip"] != "1.2.3.4" || m["direction"] != "outbound" {
		t.Fatalf("fields wrong: %v", m)
	}
	if strings.Contains(strings.ToLower(js), "threatwinds") ||
		strings.Contains(strings.ToLower(js), "clam") {
		t.Fatalf("event leaked a vendor string: %s", js)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/event/ -run TestNewNetworkEventBranding -v`
Expected: FAIL (compile — `SourceNetwork`/`NewNetworkEvent` undefined).

- [ ] **Step 3: Add the source const, fields, and constructor**

Add to the source const block: `SourceNetwork = "network_watcher"`.

Add to the `Event` struct (after `Signature`):

```go
	RemoteIP   string    `json:"remote_ip,omitempty"`
	RemotePort int       `json:"remote_port,omitempty"`
	Domain     string    `json:"domain,omitempty"`
	Direction  string    `json:"direction,omitempty"` // "outbound" | "inbound"
	Indicator  string    `json:"indicator,omitempty"` // matched blocklist value
```

Add the constructor:

```go
// NewNetworkEvent builds a branded network_watcher event for a connection to/from
// a blacklisted indicator. action is blocked|detected; direction is outbound|inbound.
// domain is set only when the match came via DNS (Plan 2). p may be nil when the
// responsible process is unknown.
func NewNetworkEvent(action, direction, remoteIP string, port int, domain, indicator string, p *ProcInfo) Event {
	var pc *ProcInfo
	if p != nil {
		c := *p
		pc = &c
	}
	return Event{
		Source:     SourceNetwork,
		Action:     action,
		Verdict:    "malicious",
		Direction:  direction,
		RemoteIP:   remoteIP,
		RemotePort: port,
		Domain:     domain,
		Indicator:  indicator,
		Process:    pc,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/event/ -run TestNewNetworkEventBranding -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/event/event.go edr/event/event_network_test.go
git commit -m "feat(edr): add network_watcher event type"
```

---

## Task 3: Cache tables

**Files:**
- Create: `edr/cache/netblock.go`
- Modify: `edr/cache/cache.go` (AutoMigrate list)
- Test: `edr/cache/netblock_test.go` (create)

**Interfaces:**
- Produces: `cache.IndicatorRecord`, `cache.NetBlockRecord`, `cache.FeedState`; methods `(*Cache) PutIndicators([]IndicatorRecord) error`, `ListIndicators() ([]IndicatorRecord, error)`, `RecordNetBlock(NetBlockRecord) error`, `SaveFeedState(FeedState) error`, `GetFeedState(feed string) (FeedState, error)`.

- [ ] **Step 1: Write the failing test**

```go
// edr/cache/netblock_test.go
package cache

import (
	"path/filepath"
	"testing"
	"time"
)

func openTmp(t *testing.T) *Cache {
	t.Helper()
	c, err := Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestIndicatorRoundTrip(t *testing.T) {
	c := openTmp(t)
	now := time.Now().UTC()
	in := []IndicatorRecord{
		{Value: "1.2.3.4", Type: "ip", Level: 1, ListName: "level1/ip", FirstSeen: now, LastSeen: now},
		{Value: "10.0.0.0/8", Type: "cidr", Level: 1, ListName: "level1/ip", FirstSeen: now, LastSeen: now},
	}
	if err := c.PutIndicators(in); err != nil {
		t.Fatal(err)
	}
	got, err := c.ListIndicators()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
}

func TestFeedState(t *testing.T) {
	c := openTmp(t)
	if err := c.SaveFeedState(FeedState{Feed: "level1/ip", IndicatorNum: 5, Checksum: "abc"}); err != nil {
		t.Fatal(err)
	}
	fs, err := c.GetFeedState("level1/ip")
	if err != nil {
		t.Fatal(err)
	}
	if fs.IndicatorNum != 5 || fs.Checksum != "abc" {
		t.Fatalf("bad feedstate: %+v", fs)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/cache/ -run 'TestIndicatorRoundTrip|TestFeedState' -v`
Expected: FAIL (compile — types/methods undefined).

- [ ] **Step 3: Create `netblock.go` and register the tables**

```go
// edr/cache/netblock.go
package cache

import (
	"time"

	"gorm.io/gorm/clause"
)

type IndicatorRecord struct {
	Value     string `gorm:"primaryKey"`
	Type      string `gorm:"primaryKey"` // ip | cidr | domain | hostname
	Level     int
	ListName  string
	FirstSeen time.Time
	LastSeen  time.Time
	Removed   bool
}

type NetBlockRecord struct {
	ID          uint `gorm:"primaryKey"`
	Time        time.Time
	Direction   string
	RemoteIP    string
	RemotePort  int
	PID         int
	ProcessPath string
	Matched     string
	MatchedType string
	Level       int
	Action      string
	Domain      string
}

type FeedState struct {
	Feed         string `gorm:"primaryKey"`
	LastFullSync time.Time
	LastDelta    time.Time
	Checksum     string
	IndicatorNum int
}

func (c *Cache) PutIndicators(recs []IndicatorRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(recs) == 0 {
		return nil
	}
	return c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&recs).Error
}

func (c *Cache) DeleteIndicator(value, typ string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Delete(&IndicatorRecord{}, "value = ? AND type = ?", value, typ).Error
}

func (c *Cache) ListIndicators() ([]IndicatorRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []IndicatorRecord
	err := c.db.Where("removed = ?", false).Find(&out).Error
	return out, err
}

func (c *Cache) RecordNetBlock(r NetBlockRecord) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if r.Time.IsZero() {
		r.Time = time.Now().UTC()
	}
	return c.db.Create(&r).Error
}

func (c *Cache) SaveFeedState(f FeedState) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&f).Error
}

func (c *Cache) GetFeedState(feed string) (FeedState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var f FeedState
	err := c.db.First(&f, "feed = ?", feed).Error
	return f, err
}
```

In `edr/cache/cache.go`, extend the `AutoMigrate(...)` call to include `&IndicatorRecord{}, &NetBlockRecord{}, &FeedState{}`. (Verify `c.db` and `c.mu` are the field names used elsewhere in `cache.go`; match them.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/cache/ -run 'TestIndicatorRoundTrip|TestFeedState' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/cache/netblock.go edr/cache/cache.go edr/cache/netblock_test.go
git commit -m "feat(edr): add netblock cache tables (indicators, blocks, feed state)"
```

---

## Task 4: Indicator type + matcher

**Files:**
- Create: `edr/netblock/indicator.go`, `edr/netblock/matcher.go`
- Test: `edr/netblock/matcher_test.go`

**Interfaces:**
- Produces: `netblock.Indicator{Value,Type string,Level int}`; `netblock.ParseIndicator(value, typ string, level int) (Indicator, error)`; `netblock.Matcher` with `Add(Indicator)`, `MatchIP(netip.Addr) (Indicator, bool)`, `Len() int`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/matcher_test.go
package netblock

import (
	"net/netip"
	"testing"
)

func TestMatcherIPAndCIDR(t *testing.T) {
	m := NewMatcher()
	for _, s := range []struct {
		v, ty string
	}{{"1.2.3.4", "ip"}, {"10.0.0.0/8", "cidr"}, {"2001:db8::/32", "cidr"}} {
		ind, err := ParseIndicator(s.v, s.ty, 1)
		if err != nil {
			t.Fatalf("parse %s: %v", s.v, err)
		}
		m.Add(ind)
	}
	cases := []struct {
		ip   string
		want bool
	}{
		{"1.2.3.4", true},
		{"1.2.3.5", false},
		{"10.9.9.9", true},   // inside 10/8
		{"11.0.0.1", false},  // outside
		{"2001:db8::dead", true},
		{"2001:dc8::1", false},
	}
	for _, c := range cases {
		addr := netip.MustParseAddr(c.ip)
		_, ok := m.MatchIP(addr)
		if ok != c.want {
			t.Errorf("MatchIP(%s) = %v, want %v", c.ip, ok, c.want)
		}
	}
}

func TestParseIndicatorRejectsGarbage(t *testing.T) {
	if _, err := ParseIndicator("not-an-ip", "ip", 1); err == nil {
		t.Fatal("expected error for bad ip")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestMatcher -v`
Expected: FAIL (package/types undefined).

- [ ] **Step 3: Implement `indicator.go` and `matcher.go`**

```go
// edr/netblock/indicator.go
package netblock

import (
	"fmt"
	"net/netip"
	"strings"
)

// Indicator is one normalized blocklist entry.
type Indicator struct {
	Value string // canonical string form
	Type  string // ip | cidr | domain | hostname
	Level int
}

// ParseIndicator validates and canonicalizes a raw feed value. Plan 1 handles
// ip and cidr; domain/hostname are accepted (lowercased, trimmed) for Plan 2.
func ParseIndicator(value, typ string, level int) (Indicator, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return Indicator{}, fmt.Errorf("empty indicator")
	}
	switch typ {
	case "ip":
		a, err := netip.ParseAddr(v)
		if err != nil {
			return Indicator{}, fmt.Errorf("bad ip %q: %w", v, err)
		}
		return Indicator{Value: a.String(), Type: "ip", Level: level}, nil
	case "cidr":
		p, err := netip.ParsePrefix(v)
		if err != nil {
			return Indicator{}, fmt.Errorf("bad cidr %q: %w", v, err)
		}
		return Indicator{Value: p.Masked().String(), Type: "cidr", Level: level}, nil
	case "domain", "hostname":
		return Indicator{Value: strings.ToLower(strings.TrimSuffix(v, ".")), Type: typ, Level: level}, nil
	default:
		return Indicator{}, fmt.Errorf("unknown indicator type %q", typ)
	}
}
```

```go
// edr/netblock/matcher.go
package netblock

import "net/netip"

// Matcher answers IP membership against single IPs and CIDR prefixes. Not safe
// for concurrent mutation; build once, then read (the Store swaps whole Matchers).
type Matcher struct {
	exact    map[netip.Addr]Indicator
	prefixes []prefixEntry
}

type prefixEntry struct {
	p   netip.Prefix
	ind Indicator
}

func NewMatcher() *Matcher {
	return &Matcher{exact: map[netip.Addr]Indicator{}}
}

func (m *Matcher) Add(ind Indicator) {
	switch ind.Type {
	case "ip":
		if a, err := netip.ParseAddr(ind.Value); err == nil {
			m.exact[a] = ind
		}
	case "cidr":
		if p, err := netip.ParsePrefix(ind.Value); err == nil {
			m.prefixes = append(m.prefixes, prefixEntry{p: p.Masked(), ind: ind})
		}
	}
}

func (m *Matcher) Len() int { return len(m.exact) + len(m.prefixes) }

// MatchIP returns the matched indicator (exact wins over prefix) and true if the
// address is blacklisted.
func (m *Matcher) MatchIP(addr netip.Addr) (Indicator, bool) {
	if ind, ok := m.exact[addr]; ok {
		return ind, true
	}
	for _, pe := range m.prefixes {
		if pe.p.Contains(addr) {
			return pe.ind, true
		}
	}
	return Indicator{}, false
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run 'TestMatcher|TestParseIndicator' -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/indicator.go edr/netblock/matcher.go edr/netblock/matcher_test.go
git commit -m "feat(edr): netblock indicator type + IP/CIDR matcher"
```

---

## Task 5: Allowlist (self-lockout prevention)

**Files:**
- Create: `edr/netblock/allowlist.go`
- Test: `edr/netblock/allowlist_test.go`

**Interfaces:**
- Consumes: `config.BlocklistConfig`, `config.Allowlist`.
- Produces: `netblock.Allowlist` with `Allowed(netip.Addr) bool`; `netblock.BuildAllowlist(cfg config.EDRConfig, sys SystemNets) *Allowlist`; `netblock.SystemNets{Resolvers []netip.Addr, Gateways []netip.Addr, ServerHosts []netip.Addr}`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/allowlist_test.go
package netblock

import (
	"net/netip"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestAllowlistDefaults(t *testing.T) {
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = true
	al := BuildAllowlist(cfg, SystemNets{
		Resolvers: []netip.Addr{netip.MustParseAddr("8.8.8.8")},
		Gateways:  []netip.Addr{netip.MustParseAddr("192.168.1.1")},
	})
	must := map[string]bool{
		"127.0.0.1":   true, // loopback
		"::1":         true,
		"10.1.2.3":    true, // RFC1918 (private ranges on)
		"192.168.1.1": true, // gateway
		"8.8.8.8":     true, // resolver
		"169.254.1.1": true, // link-local
		"9.9.9.9":     false,
	}
	for ip, want := range must {
		if al.Allowed(netip.MustParseAddr(ip)) != want {
			t.Errorf("Allowed(%s) = %v, want %v", ip, !want, want)
		}
	}
}

func TestAllowlistPrivateOff(t *testing.T) {
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = false
	al := BuildAllowlist(cfg, SystemNets{})
	if al.Allowed(netip.MustParseAddr("10.1.2.3")) {
		t.Fatal("10.1.2.3 must NOT be allowed when private ranges off")
	}
	if !al.Allowed(netip.MustParseAddr("127.0.0.1")) {
		t.Fatal("loopback always allowed")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestAllowlist -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `allowlist.go`**

```go
// edr/netblock/allowlist.go
package netblock

import (
	"net/netip"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// SystemNets are host-derived addresses that must never be blocked (resolved by
// the caller at startup so this stays pure/testable).
type SystemNets struct {
	Resolvers   []netip.Addr
	Gateways    []netip.Addr
	ServerHosts []netip.Addr // UTMStack platform / mirror IPs
}

// Allowlist is the never-block set: built-in critical ranges, optional private
// ranges, host system nets, and admin-configured IP/CIDR entries.
type Allowlist struct {
	exact    map[netip.Addr]struct{}
	prefixes []netip.Prefix
}

var builtinAlways = []string{
	"127.0.0.0/8", "::1/128", // loopback
	"169.254.0.0/16", "fe80::/10", // link-local
	"224.0.0.0/4", "ff00::/8", // multicast
	"255.255.255.255/32",
}

var privateRanges = []string{
	"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "fc00::/7",
}

func BuildAllowlist(cfg config.EDRConfig, sys SystemNets) *Allowlist {
	al := &Allowlist{exact: map[netip.Addr]struct{}{}}
	addPfx := func(list []string) {
		for _, s := range list {
			if p, err := netip.ParsePrefix(s); err == nil {
				al.prefixes = append(al.prefixes, p.Masked())
			}
		}
	}
	addPfx(builtinAlways)
	if cfg.Blocklist.AllowPrivateRanges {
		addPfx(privateRanges)
	}
	for _, a := range sys.Resolvers {
		al.exact[a] = struct{}{}
	}
	for _, a := range sys.Gateways {
		al.exact[a] = struct{}{}
	}
	for _, a := range sys.ServerHosts {
		al.exact[a] = struct{}{}
	}
	// Admin-configured network exemptions live in the unified allowlist (Networks).
	for _, s := range cfg.Allowlist.Networks {
		s = strings.TrimSpace(s)
		if a, err := netip.ParseAddr(s); err == nil {
			al.exact[a] = struct{}{}
			continue
		}
		if p, err := netip.ParsePrefix(s); err == nil {
			al.prefixes = append(al.prefixes, p.Masked())
		}
	}
	return al
}

func (a *Allowlist) Allowed(addr netip.Addr) bool {
	if _, ok := a.exact[addr]; ok {
		return true
	}
	for _, p := range a.prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
```

In `edr/config/config.go`, add a `Networks []string json:"networks,omitempty"` field to the `Allowlist` struct (network IP/CIDR exemptions), alongside the existing `Paths`/`Processes`/`Commands`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestAllowlist -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/allowlist.go edr/netblock/allowlist_test.go edr/config/config.go
git commit -m "feat(edr): netblock self-lockout allowlist"
```

---

## Task 6: Store (hot-swappable indicator set)

**Files:**
- Create: `edr/netblock/store.go`
- Test: `edr/netblock/store_test.go`

**Interfaces:**
- Produces: `netblock.Store` with `Swap([]Indicator)`, `Snapshot() *Matcher`, `Match(netip.Addr) (Indicator, bool)`, `Count() int`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/store_test.go
package netblock

import (
	"net/netip"
	"sync"
	"testing"
)

func TestStoreSwapAtomic(t *testing.T) {
	s := NewStore()
	i1, _ := ParseIndicator("1.2.3.4", "ip", 1)
	s.Swap([]Indicator{i1})
	if _, ok := s.Match(netip.MustParseAddr("1.2.3.4")); !ok {
		t.Fatal("expected match after first swap")
	}
	// Concurrent readers during a swap must never see a torn state.
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					s.Match(netip.MustParseAddr("5.6.7.8"))
				}
			}
		}()
	}
	i2, _ := ParseIndicator("5.6.7.8", "ip", 1)
	for i := 0; i < 100; i++ {
		s.Swap([]Indicator{i1, i2})
		s.Swap([]Indicator{i1})
	}
	close(stop)
	wg.Wait()
	if s.Count() != 1 {
		t.Fatalf("final count = %d, want 1", s.Count())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestStoreSwap -race -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `store.go`**

```go
// edr/netblock/store.go
package netblock

import (
	"net/netip"
	"sync/atomic"
)

// Store holds the active blocklist as an immutable *Matcher swapped atomically on
// each feed refresh, so readers never lock and never see a torn set.
type Store struct {
	cur atomic.Pointer[Matcher]
}

func NewStore() *Store {
	s := &Store{}
	s.cur.Store(NewMatcher())
	return s
}

func (s *Store) Swap(inds []Indicator) {
	m := NewMatcher()
	for _, ind := range inds {
		m.Add(ind)
	}
	s.cur.Store(m)
}

func (s *Store) Snapshot() *Matcher { return s.cur.Load() }

func (s *Store) Match(addr netip.Addr) (Indicator, bool) {
	return s.cur.Load().MatchIP(addr)
}

func (s *Store) Count() int { return s.cur.Load().Len() }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestStoreSwap -race -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/store.go edr/netblock/store_test.go
git commit -m "feat(edr): netblock atomic hot-swap store"
```

---

## Task 7: Parser (accumulative gzip + daily ndjson)

**Files:**
- Create: `edr/netblock/parse.go`
- Test: `edr/netblock/parse_test.go`

**Interfaces:**
- Produces: `netblock.ParseAccumulative(r io.Reader, level int) ([]Indicator, error)` (r = gzip stream); `netblock.DailyOp{Value,Type,Op string}`; `netblock.ParseDaily(r io.Reader) ([]DailyOp, error)`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/parse_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func gz(s string) *bytes.Reader {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(s))
	w.Close()
	return bytes.NewReader(b.Bytes())
}

func TestParseAccumulative(t *testing.T) {
	body := "1.2.3.4\n# comment\n\n10.0.0.0/8\n2001:db8::1\n"
	inds, err := ParseAccumulative(gz(body), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(inds) != 3 {
		t.Fatalf("want 3, got %d: %+v", len(inds), inds)
	}
	// CIDR auto-detected from the slash.
	var sawCIDR bool
	for _, i := range inds {
		if i.Type == "cidr" && i.Value == "10.0.0.0/8" {
			sawCIDR = true
		}
	}
	if !sawCIDR {
		t.Fatal("10.0.0.0/8 not parsed as cidr")
	}
}

func TestParseDaily(t *testing.T) {
	body := `{"value":"1.2.3.4","type":"ip","op":"add"}
{"value":"9.9.9.9","type":"ip","op":"del"}
{"garbage"
{"value":"10.0.0.0/8","type":"cidr","op":"add"}`
	ops, err := ParseDaily(strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 3 { // malformed line skipped, not fatal
		t.Fatalf("want 3 ops, got %d", len(ops))
	}
	if ops[1].Op != "del" {
		t.Fatalf("op[1] = %+v", ops[1])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestParse -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `parse.go`**

```go
// edr/netblock/parse.go
package netblock

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"io"
	"strings"
)

// ParseAccumulative reads a gzip'd newline-delimited value list. A value with a
// "/" is a CIDR, otherwise an IP. Blank lines and #-comments are ignored.
// Unparseable lines are skipped (a poisoned line must not drop the whole feed).
func ParseAccumulative(r io.Reader, level int) ([]Indicator, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	var out []Indicator
	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		typ := "ip"
		if strings.Contains(line, "/") {
			typ = "cidr"
		}
		if ind, err := ParseIndicator(line, typ, level); err == nil {
			out = append(out, ind)
		}
	}
	return out, sc.Err()
}

// DailyOp is one incremental change from the daily ndjson delta.
type DailyOp struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Op    string `json:"op"` // add | del
}

// ParseDaily reads ndjson delta lines. Malformed lines are skipped.
func ParseDaily(r io.Reader) ([]DailyOp, error) {
	var out []DailyOp
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var op DailyOp
		if err := json.Unmarshal([]byte(line), &op); err != nil {
			continue
		}
		if op.Value == "" || (op.Op != "add" && op.Op != "del") {
			continue
		}
		if op.Type == "" {
			op.Type = "ip"
		}
		out = append(out, op)
	}
	return out, sc.Err()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestParse -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/parse.go edr/netblock/parse_test.go
git commit -m "feat(edr): netblock feed parsers (accumulative + daily)"
```

---

## Task 8: Feed downloader + scheduler

**Files:**
- Create: `edr/netblock/feed.go`
- Test: `edr/netblock/feed_test.go`

**Interfaces:**
- Consumes: `config.EDRConfig`, `cache.*`, `ParseAccumulative`, `ParseDaily`, `Store`.
- Produces: `netblock.Feed` with `NewFeed(cfg, cache, store, applyFn) *Feed`; `(*Feed).Run(ctx)`; overridable `Fetch func(url string) (io.ReadCloser, error)` and `FetchChecksum func(url string) (string, error)` fields for tests; `(*Feed).UpdateOnce(ctx) error`. `applyFn func([]Indicator)` is called after each successful refresh (lets the manager drive the enforcer).

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/feed_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func sha(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

func TestFeedBootstrapAndChecksum(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Blocklist.Levels = []int{1}
	store := NewStore()
	var applied []Indicator
	f := NewFeed(cfg, c, store, func(inds []Indicator) { applied = inds })

	var gzbuf bytes.Buffer
	w := gzip.NewWriter(&gzbuf)
	w.Write([]byte("1.2.3.4\n5.6.7.8\n"))
	w.Close()
	body := gzbuf.Bytes()

	f.Fetch = func(url string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	f.FetchChecksum = func(url string) (string, error) { return sha(body), nil }

	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.Count() != 2 {
		t.Fatalf("store count = %d, want 2", store.Count())
	}
	if len(applied) != 2 {
		t.Fatalf("apply called with %d indicators", len(applied))
	}
}

func TestFeedRejectsBadChecksum(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	store := NewStore()
	f := NewFeed(config.Default(), c, store, func([]Indicator) {})
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	gw.Write([]byte("1.2.3.4\n"))
	gw.Close()
	f.Fetch = func(string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(b.Bytes())), nil }
	f.FetchChecksum = func(string) (string, error) { return "deadbeef", nil } // wrong
	if err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("expected checksum-mismatch error")
	}
	if store.Count() != 0 {
		t.Fatal("store must stay empty on checksum failure")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestFeed -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `feed.go`**

```go
// edr/netblock/feed.go
package netblock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

// Feed pulls ThreatWinds indicators from the UTMStack mirror on a cadence and
// hot-swaps the store. applyFn runs after each successful refresh.
type Feed struct {
	cfg     config.EDRConfig
	cache   *cache.Cache
	store   *Store
	applyFn func([]Indicator)

	// overridable for tests
	Fetch         func(url string) (io.ReadCloser, error)
	FetchChecksum func(url string) (string, error)
}

func NewFeed(cfg config.EDRConfig, c *cache.Cache, store *Store, applyFn func([]Indicator)) *Feed {
	f := &Feed{cfg: cfg, cache: c, store: store, applyFn: applyFn}
	f.Fetch = f.httpFetch
	f.FetchChecksum = f.httpChecksum
	return f
}

func (f *Feed) baseURL() string {
	if f.cfg.Blocklist.MirrorBaseURL != "" {
		return strings.TrimRight(f.cfg.Blocklist.MirrorBaseURL, "/")
	}
	return strings.TrimRight(f.cfg.Server, "/") + "/feeds/v1"
}

func (f *Feed) skipTLS() bool { return f.cfg.SkipCertValidate }

func (f *Feed) client() *http.Client {
	c := &http.Client{Timeout: 5 * time.Minute}
	if f.skipTLS() {
		c.Transport = &http.Transport{TLSClientConfig: insecureTLS()}
	}
	return c
}

func (f *Feed) httpFetch(url string) (io.ReadCloser, error) {
	resp, err := f.client().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}

func (f *Feed) httpChecksum(url string) (string, error) {
	resp, err := f.client().Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum %s: status %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return strings.TrimSpace(string(b)), err
}

// Run drives the immediate-then-ticker refresh loop (freshclam pattern).
func (f *Feed) Run(ctx context.Context) {
	if err := f.UpdateOnce(ctx); err != nil {
		logWarn("blocklist initial feed update failed: %v", err)
	}
	hrs := f.cfg.Blocklist.RefreshHours
	if hrs <= 0 {
		hrs = 6
	}
	t := time.NewTicker(time.Duration(hrs) * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := f.UpdateOnce(ctx); err != nil {
				logWarn("blocklist feed update failed: %v", err)
			}
		}
	}
}

// UpdateOnce pulls the accumulative snapshot for every configured (level, ip)
// feed, verifies checksum, rebuilds the store, and persists indicators. (Daily
// deltas layer on in Task N via UpdateOnce reuse; Plan 1 bootstraps full each cycle
// for the modest IP list — deltas are a later optimization.)
func (f *Feed) UpdateOnce(ctx context.Context) error {
	var all []Indicator
	seen := map[string]struct{}{}
	for _, lvl := range f.cfg.Blocklist.Levels {
		level := fmt.Sprintf("level%d", lvl)
		url := fmt.Sprintf("%s/download/list/%s/accumulative/ip", f.baseURL(), level)
		ckURL := fmt.Sprintf("%s/download/checksum?level=%s&type=accumulative&name=ip", f.baseURL(), level)

		body, err := f.Fetch(url)
		if err != nil {
			return err
		}
		raw, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			return err
		}
		want, err := f.FetchChecksum(ckURL)
		if err != nil {
			return err
		}
		if want != "" {
			got := sha256.Sum256(raw)
			if hex.EncodeToString(got[:]) != want {
				return fmt.Errorf("checksum mismatch for %s", url)
			}
		}
		inds, err := ParseAccumulative(bytesReader(raw), lvl)
		if err != nil {
			return err
		}
		for _, ind := range inds {
			key := ind.Type + "|" + ind.Value
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, ind)
		}
		_ = f.cache.SaveFeedState(cache.FeedState{
			Feed: level + "/ip", LastFullSync: time.Now().UTC(),
			Checksum: want, IndicatorNum: len(inds),
		})
	}
	f.store.Swap(all)
	f.persist(all)
	if f.applyFn != nil {
		f.applyFn(all)
	}
	return nil
}

func (f *Feed) persist(inds []Indicator) {
	now := time.Now().UTC()
	recs := make([]cache.IndicatorRecord, 0, len(inds))
	for _, ind := range inds {
		recs = append(recs, cache.IndicatorRecord{
			Value: ind.Value, Type: ind.Type, Level: ind.Level,
			ListName: fmt.Sprintf("level%d/ip", ind.Level), FirstSeen: now, LastSeen: now,
		})
	}
	if err := f.cache.PutIndicators(recs); err != nil {
		logWarn("blocklist persist indicators: %v", err)
	}
}
```

Add small helpers in a new `edr/netblock/util.go`:

```go
// edr/netblock/util.go
package netblock

import (
	"bytes"
	"crypto/tls"
	"io"

	"github.com/utmstack/UTMStack/agent/shared/logger"
)

func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func insecureTLS() *tls.Config { return &tls.Config{InsecureSkipVerify: true} }

// logWarn/logInfo funnel through the shared logger; surfaced strings never name
// the intel vendor (branding rule) — they say "blocklist".
func logWarn(format string, a ...any) { logger.Error("[UTMStack EDR] "+format, a...) }
func logInfo(format string, a ...any) { logger.Info("[UTMStack EDR] "+format, a...) }
```

(Confirm `shared/logger` exposes `Info`/`Error` with a `(format string, args ...any)` signature — match the existing call sites in `edr/`. Adjust if the project's logger differs.)

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestFeed -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/feed.go edr/netblock/util.go edr/netblock/feed_test.go
git commit -m "feat(edr): netblock mirror feed downloader + scheduler"
```

---

## Task 9: Enforcer interface + diff engine

**Files:**
- Create: `edr/netblock/enforcer.go`
- Test: `edr/netblock/enforcer_test.go`

**Interfaces:**
- Produces: `netblock.Blocker` interface `{ AddIP(netip.Addr, dir string) error; RemoveIP(netip.Addr) error; Reset() error; Count() int }`; `netblock.Enforcer` with `NewEnforcer(b Blocker, dir string) *Enforcer`, `(*Enforcer).Apply(desired []netip.Addr) error` (diffs against last-applied), `(*Enforcer).Clear() error`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/enforcer_test.go
package netblock

import (
	"net/netip"
	"sort"
	"testing"
)

type fakeBlocker struct {
	set map[netip.Addr]string
}

func newFake() *fakeBlocker { return &fakeBlocker{set: map[netip.Addr]string{}} }
func (f *fakeBlocker) AddIP(a netip.Addr, dir string) error { f.set[a] = dir; return nil }
func (f *fakeBlocker) RemoveIP(a netip.Addr) error          { delete(f.set, a); return nil }
func (f *fakeBlocker) Reset() error                         { f.set = map[netip.Addr]string{}; return nil }
func (f *fakeBlocker) Count() int                           { return len(f.set) }
func (f *fakeBlocker) keys() []string {
	var k []string
	for a := range f.set {
		k = append(k, a.String())
	}
	sort.Strings(k)
	return k
}

func TestEnforcerDiff(t *testing.T) {
	fb := newFake()
	e := NewEnforcer(fb, "both")
	a := netip.MustParseAddr("1.1.1.1")
	b := netip.MustParseAddr("2.2.2.2")
	cc := netip.MustParseAddr("3.3.3.3")

	if err := e.Apply([]netip.Addr{a, b}); err != nil {
		t.Fatal(err)
	}
	if fb.Count() != 2 {
		t.Fatalf("after first apply want 2, got %d", fb.Count())
	}
	// Drop b, add c: exactly one remove + one add, a untouched.
	if err := e.Apply([]netip.Addr{a, cc}); err != nil {
		t.Fatal(err)
	}
	got := fb.keys()
	if len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "3.3.3.3" {
		t.Fatalf("after diff apply got %v", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestEnforcerDiff -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `enforcer.go`**

```go
// edr/netblock/enforcer.go
package netblock

import "net/netip"

// Blocker is the platform enforcement surface (WFP on Windows, no-op elsewhere).
type Blocker interface {
	AddIP(addr netip.Addr, dir string) error
	RemoveIP(addr netip.Addr) error
	Reset() error
	Count() int
}

// Enforcer reconciles a desired IP set against what is currently applied, issuing
// the minimal add/remove calls to the Blocker.
type Enforcer struct {
	b       Blocker
	dir     string
	applied map[netip.Addr]struct{}
}

func NewEnforcer(b Blocker, dir string) *Enforcer {
	return &Enforcer{b: b, dir: dir, applied: map[netip.Addr]struct{}{}}
}

func (e *Enforcer) Apply(desired []netip.Addr) error {
	want := make(map[netip.Addr]struct{}, len(desired))
	for _, a := range desired {
		want[a] = struct{}{}
		if _, ok := e.applied[a]; !ok {
			if err := e.b.AddIP(a, e.dir); err != nil {
				return err
			}
			e.applied[a] = struct{}{}
		}
	}
	for a := range e.applied {
		if _, ok := want[a]; !ok {
			if err := e.b.RemoveIP(a); err != nil {
				return err
			}
			delete(e.applied, a)
		}
	}
	return nil
}

func (e *Enforcer) Clear() error {
	if err := e.b.Reset(); err != nil {
		return err
	}
	e.applied = map[netip.Addr]struct{}{}
	return nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestEnforcerDiff -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/enforcer.go edr/netblock/enforcer_test.go
git commit -m "feat(edr): netblock enforcer diff engine + Blocker interface"
```

---

## Task 10: WFP Blocker (Windows) + no-op stub

**Files:**
- Create: `edr/netblock/wfp_windows.go`, `edr/netblock/wfp_other.go`
- Test: `edr/netblock/wfp_other_test.go` (host build only)

**Interfaces:**
- Produces: `netblock.NewOSBlocker(sublayerName string) (Blocker, error)` — real WFP on windows, no-op elsewhere.

**Notes:** This is Windows-native and cannot run on macOS. The stub keeps the package building/testing on the host; the real code is compile-checked via `GOOS=windows go build` and validated on the VM (Task 15). WFP calls use `golang.org/x/sys/windows` `NewLazySystemDLL("fwpuclnt.dll")`. Filters live in a **dynamic session** (auto-released on process exit → fail-open) and a dedicated sublayer.

- [ ] **Step 1: Write the stub test (host)**

```go
// edr/netblock/wfp_other_test.go
//go:build !windows

package netblock

import (
	"net/netip"
	"testing"
)

func TestNoopBlocker(t *testing.T) {
	b, err := NewOSBlocker("UTMStackEDR-Blocklist")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddIP(netip.MustParseAddr("1.2.3.4"), "both"); err != nil {
		t.Fatal(err)
	}
	if b.Count() != 0 { // no-op counts nothing
		t.Fatalf("noop count = %d", b.Count())
	}
	if err := b.Reset(); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestNoopBlocker -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement the no-op stub**

```go
// edr/netblock/wfp_other.go
//go:build !windows

package netblock

import "net/netip"

type noopBlocker struct{}

func NewOSBlocker(sublayerName string) (Blocker, error) { return &noopBlocker{}, nil }

func (noopBlocker) AddIP(netip.Addr, string) error { return nil }
func (noopBlocker) RemoveIP(netip.Addr) error      { return nil }
func (noopBlocker) Reset() error                   { return nil }
func (noopBlocker) Count() int                     { return 0 }
```

- [ ] **Step 4: Run the stub test**

Run: `go test ./edr/netblock/ -run TestNoopBlocker -v`
Expected: PASS.

- [ ] **Step 5: Implement the Windows WFP blocker**

```go
// edr/netblock/wfp_windows.go
//go:build windows

package netblock

import (
	"fmt"
	"net/netip"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// --- fwpuclnt.dll bindings -------------------------------------------------

var (
	modFwpuclnt = windows.NewLazySystemDLL("fwpuclnt.dll")

	procEngineOpen        = modFwpuclnt.NewProc("FwpmEngineOpen0")
	procEngineClose       = modFwpuclnt.NewProc("FwpmEngineClose0")
	procSubLayerAdd       = modFwpuclnt.NewProc("FwpmSubLayerAdd0")
	procFilterAdd         = modFwpuclnt.NewProc("FwpmFilterAdd0")
	procFilterDeleteById  = modFwpuclnt.NewProc("FwpmFilterDeleteById0")
	procTransactionBegin  = modFwpuclnt.NewProc("FwpmTransactionBegin0")
	procTransactionCommit = modFwpuclnt.NewProc("FwpmTransactionCommit0")
	procTransactionAbort  = modFwpuclnt.NewProc("FwpmTransactionAbort0")
)

const (
	rpcCAuthnWinNT       = 10
	fwpmSessionFlagDyn   = 0x00000001 // FWPM_SESSION_FLAG_DYNAMIC
	fwpActionBlock       = 0x00000001 | 0x00001000 // FWP_ACTION_BLOCK
	fwpMatchEqual        = 0
	fwpUint32            = 8  // FWP_UINT32
	fwpV6AddrMask        = 20 // FWP_V6_ADDR_MASK
	fwpByteArray16Type   = 11 // FWP_BYTE_ARRAY16_TYPE
	fwpV4AddrMask        = 19 // FWP_V4_ADDR_MASK
)

// Well-known layer + condition GUIDs (windows headers: fwpmu.h / fwpmtypes.h).
var (
	layerConnectV4    = mustGUID("c38d57d1-05a7-4c33-904f-7fbceee60e82")
	layerConnectV6    = mustGUID("4a72393b-319f-44bc-84c3-ba54dcb3b6b4")
	layerRecvAcceptV4 = mustGUID("e1cd9fe7-f4b5-4273-96c0-592e487b8650")
	layerRecvAcceptV6 = mustGUID("a3b42c97-9f04-4672-b87b-cee8514723b0")
	condRemoteAddress = mustGUID("b235ae9a-1d64-49b8-a44c-5ff3d9095045")
	edrSublayerKey    = mustGUID("6f3b8e21-1c4a-4d77-9a2e-utm00blockli")
)

func mustGUID(s string) windows.GUID {
	g, err := windows.GUIDFromString("{" + s + "}")
	if err != nil {
		// edrSublayerKey uses a private, fixed GUID; generate a valid one at build
		// time and paste it here (see Task 10 note). Panic on a malformed literal.
		panic("netblock: bad GUID literal " + s + ": " + err.Error())
	}
	return g
}

// FWPM_DISPLAY_DATA0
type displayData struct {
	name        *uint16
	description *uint16
}

// FWPM_SUBLAYER0 (trimmed to fields we set)
type sublayer0 struct {
	subLayerKey  windows.GUID
	displayData  displayData
	flags        uint32
	providerKey  uintptr
	providerData windows.GUID // reused as blob placeholder; zeroed
	weight       uint16
	_            [6]byte
}

// FWP_VALUE0 / FWP_CONDITION_VALUE0
type fwpValue struct {
	typ  uint32
	_    uint32
	data uintptr // pointer or inline uint32
}

// FWPM_FILTER_CONDITION0
type filterCondition struct {
	fieldKey       windows.GUID
	matchType      uint32
	conditionValue fwpValue
}

// FWP_V4_ADDR_AND_MASK
type v4AddrAndMask struct {
	addr uint32
	mask uint32
}

// FWP_V6_ADDR_AND_MASK
type v6AddrAndMask struct {
	addr         [16]byte
	prefixLength byte
	_            [3]byte
}

// FWPM_FILTER0 (trimmed) — action + single remote-address condition.
type filter0 struct {
	filterKey           windows.GUID
	displayData         displayData
	flags               uint32
	providerKey         uintptr
	providerData        [16]byte
	layerKey            windows.GUID
	subLayerKey         windows.GUID
	weight              fwpValue
	numFilterConditions uint32
	filterConditions    *filterCondition
	action              filterAction
	context             uint64
	reserved            uintptr
	filterID            uint64
	effectiveWeight     fwpValue
}

type filterAction struct {
	actionType uint32
	filterType windows.GUID // union: calloutKey/filterType — zeroed for BLOCK
}

// --- Blocker impl ----------------------------------------------------------

type wfpBlocker struct {
	mu      sync.Mutex
	engine  windows.Handle
	filters map[netip.Addr][]uint64 // addr -> filter IDs (v4/v6 × layers)
}

func NewOSBlocker(sublayerName string) (Blocker, error) {
	b := &wfpBlocker{filters: map[netip.Addr][]uint64{}}
	if err := b.open(sublayerName); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *wfpBlocker) open(name string) error {
	// Dynamic session → all filters auto-removed when this handle closes (fail-open).
	type session0 struct {
		sessionKey           windows.GUID
		displayData          displayData
		flags                uint32
		txnWaitTimeoutMs     uint32
		processID            uint32
		sid                  uintptr
		username             *uint16
		kernelMode           int32
	}
	sess := session0{flags: fwpmSessionFlagDyn}
	r, _, _ := procEngineOpen.Call(0, rpcCAuthnWinNT, 0,
		uintptr(unsafe.Pointer(&sess)), uintptr(unsafe.Pointer(&b.engine)))
	if r != 0 {
		return fmt.Errorf("FwpmEngineOpen0: 0x%x", r)
	}
	np, _ := windows.UTF16PtrFromString(name)
	sl := sublayer0{subLayerKey: edrSublayerKey, displayData: displayData{name: np}, weight: 0x8000}
	r, _, _ = procSubLayerAdd.Call(uintptr(b.engine), uintptr(unsafe.Pointer(&sl)), 0)
	if r != 0 && r != 0x80320009 /* FWP_E_ALREADY_EXISTS */ {
		return fmt.Errorf("FwpmSubLayerAdd0: 0x%x", r)
	}
	return nil
}

func layersFor(dir string, v6 bool) []windows.GUID {
	var out []windows.GUID
	if dir == "both" || dir == "out" {
		if v6 {
			out = append(out, layerConnectV6)
		} else {
			out = append(out, layerConnectV4)
		}
	}
	if dir == "both" || dir == "in" {
		if v6 {
			out = append(out, layerRecvAcceptV6)
		} else {
			out = append(out, layerRecvAcceptV4)
		}
	}
	return out
}

func (b *wfpBlocker) AddIP(addr netip.Addr, dir string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.filters[addr]; ok {
		return nil
	}
	var ids []uint64
	for _, layer := range layersFor(dir, addr.Is6()) {
		id, err := b.addFilter(addr, layer)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	b.filters[addr] = ids
	return nil
}

func (b *wfpBlocker) addFilter(addr netip.Addr, layer windows.GUID) (uint64, error) {
	cond := filterCondition{fieldKey: condRemoteAddress, matchType: fwpMatchEqual}
	if addr.Is4() {
		v4 := v4AddrAndMask{addr: be32(addr.As4()), mask: 0xffffffff}
		cond.conditionValue = fwpValue{typ: fwpV4AddrMask, data: uintptr(unsafe.Pointer(&v4))}
		return b.commitFilter(layer, &cond, unsafe.Pointer(&v4))
	}
	a16 := addr.As16()
	v6 := v6AddrAndMask{addr: a16, prefixLength: 128}
	cond.conditionValue = fwpValue{typ: fwpV6AddrMask, data: uintptr(unsafe.Pointer(&v6))}
	return b.commitFilter(layer, &cond, unsafe.Pointer(&v6))
}

func (b *wfpBlocker) commitFilter(layer windows.GUID, cond *filterCondition, keep unsafe.Pointer) (uint64, error) {
	np, _ := windows.UTF16PtrFromString("UTMStack EDR blocklist")
	f := filter0{
		displayData:         displayData{name: np},
		layerKey:            layer,
		subLayerKey:         edrSublayerKey,
		weight:              fwpValue{typ: fwpUint32, data: 0x8000},
		numFilterConditions: 1,
		filterConditions:    cond,
		action:              filterAction{actionType: fwpActionBlock},
	}
	var id uint64
	r, _, _ := procFilterAdd.Call(uintptr(b.engine), uintptr(unsafe.Pointer(&f)), 0,
		uintptr(unsafe.Pointer(&id)))
	_ = keep // ensure the condition buffer outlives the call
	if r != 0 {
		return 0, fmt.Errorf("FwpmFilterAdd0: 0x%x", r)
	}
	return id, nil
}

func (b *wfpBlocker) RemoveIP(addr netip.Addr) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, id := range b.filters[addr] {
		procFilterDeleteById.Call(uintptr(b.engine), uintptr(id))
	}
	delete(b.filters, addr)
	return nil
}

func (b *wfpBlocker) Reset() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for addr, ids := range b.filters {
		for _, id := range ids {
			procFilterDeleteById.Call(uintptr(b.engine), uintptr(id))
		}
		delete(b.filters, addr)
	}
	return nil
}

func (b *wfpBlocker) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.filters)
}

func be32(b [4]byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}
```

> **Implementer note (Task 10):** WFP struct layouts must match the native headers exactly. The structs above mirror `fwpmtypes.h`; validate field sizes with a Windows build and adjust any padding/`_` fields until `FwpmFilterAdd0` returns `0`. Replace `edrSublayerKey`'s placeholder literal with a real generated GUID (e.g. `powershell New-Guid`) before first VM run — the placeholder `utm00blockli` segment is intentionally invalid so the panic fires if it's forgotten. This is expected VM-iteration territory (matches the project's Windows-native model); the diff engine + fail-open contract are what the plan pins down, not byte-perfect struct offsets on first paste.

- [ ] **Step 6: Cross-build both arches + run host tests**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./edr/netblock/
GOOS=windows GOARCH=arm64 go build ./edr/netblock/
go test ./edr/netblock/ -v
```
Expected: windows builds succeed; host tests PASS (stub path).

- [ ] **Step 7: Commit**

```bash
git add edr/netblock/wfp_windows.go edr/netblock/wfp_other.go edr/netblock/wfp_other_test.go
git commit -m "feat(edr): WFP dynamic-filter blocker (windows) + noop stub"
```

---

## Task 11: WFP net-event audit reader (Windows) + stub

**Files:**
- Create: `edr/netblock/connaudit_windows.go`, `edr/netblock/connaudit_other.go`
- Test: `edr/netblock/connaudit_other_test.go`

**Interfaces:**
- Produces: `netblock.BlockEvent{RemoteIP netip.Addr, RemotePort int, PID int, ProcessPath string, Direction string}`; `netblock.NewConnAudit() ConnAudit`; `ConnAudit` interface `{ Run(ctx context.Context, out chan<- BlockEvent) }`.

**Purpose:** turn actual WFP drops into detection events with process attribution. On Windows this subscribes to WFP net events (`FwpmNetEventSubscribe0`, filtered to `FWPM_NET_EVENT_TYPE_CLASSIFY_DROP`), reads the blocked 5-tuple + PID + app-id, and emits `BlockEvent`s. The `!windows` stub blocks until ctx cancel.

- [ ] **Step 1: Write the stub test (host)**

```go
// edr/netblock/connaudit_other_test.go
//go:build !windows

package netblock

import (
	"context"
	"testing"
	"time"
)

func TestNoopConnAuditStops(t *testing.T) {
	ca := NewConnAudit()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { ca.Run(ctx, make(chan BlockEvent)); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("conn audit did not stop on ctx cancel")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestNoopConnAudit -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement the shared type + stub**

```go
// edr/netblock/connaudit.go
package netblock

import (
	"context"
	"net/netip"
)

// BlockEvent describes one enforced WFP drop, for detection eventing.
type BlockEvent struct {
	RemoteIP    netip.Addr
	RemotePort  int
	PID         int
	ProcessPath string
	Direction   string
}

// ConnAudit surfaces WFP drop events. Real on windows; no-op elsewhere.
type ConnAudit interface {
	Run(ctx context.Context, out chan<- BlockEvent)
}
```

```go
// edr/netblock/connaudit_other.go
//go:build !windows

package netblock

import "context"

type noopConnAudit struct{}

func NewConnAudit() ConnAudit { return noopConnAudit{} }

func (noopConnAudit) Run(ctx context.Context, out chan<- BlockEvent) { <-ctx.Done() }
```

- [ ] **Step 4: Run the stub test**

Run: `go test ./edr/netblock/ -run TestNoopConnAudit -v`
Expected: PASS.

- [ ] **Step 5: Implement the Windows net-event reader**

```go
// edr/netblock/connaudit_windows.go
//go:build windows

package netblock

import (
	"context"
	"net/netip"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procNetEventSub2   = modFwpuclnt.NewProc("FwpmNetEventSubscribe2")
	procNetEventUnsub  = modFwpuclnt.NewProc("FwpmNetEventUnsubscribe0")
)

// FWPM_NET_EVENT_SUBSCRIPTION0 (trimmed) — no template = all events; we filter in
// the callback for classify-drop within our sublayer.
type netEventSubscription struct {
	enumTemplate uintptr
	flags        uint32
	sessionKey   windows.GUID
}

type winConnAudit struct {
	mu     sync.Mutex
	engine windows.Handle
	handle windows.Handle
}

func NewConnAudit() ConnAudit { return &winConnAudit{} }

func (w *winConnAudit) Run(ctx context.Context, out chan<- BlockEvent) {
	// Reuse a dedicated dynamic engine handle for the subscription lifetime.
	var eng windows.Handle
	type session0 struct {
		sessionKey       windows.GUID
		displayData      displayData
		flags            uint32
		txnWaitTimeoutMs uint32
		processID        uint32
		sid              uintptr
		username         *uint16
		kernelMode       int32
	}
	sess := session0{flags: fwpmSessionFlagDyn}
	r, _, _ := procEngineOpen.Call(0, rpcCAuthnWinNT, 0,
		uintptr(unsafe.Pointer(&sess)), uintptr(unsafe.Pointer(&eng)))
	if r != 0 {
		logWarn("conn audit engine open: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.engine = eng
	defer procEngineClose.Call(uintptr(eng))

	cb := windows.NewCallback(func(ctxPtr uintptr, event *netEvent1) uintptr {
		be, ok := decodeNetEvent(event)
		if ok {
			select {
			case out <- be:
			default: // never block the ETW/WFP callback
			}
		}
		return 0
	})

	sub := netEventSubscription{}
	var subHandle windows.Handle
	r, _, _ = procNetEventSub2.Call(uintptr(eng), uintptr(unsafe.Pointer(&sub)),
		cb, 0, uintptr(unsafe.Pointer(&subHandle)))
	if r != 0 {
		logWarn("net event subscribe: 0x%x", r)
		<-ctx.Done()
		return
	}
	w.handle = subHandle
	<-ctx.Done()
	procNetEventUnsub.Call(uintptr(eng), uintptr(subHandle))
}

// netEvent1 mirrors FWPM_NET_EVENT1 header fields we read (timestamp/flags/ip
// version/protocol/local+remote addr+port) plus the classify-drop union pointer.
type netEvent1 struct {
	header  netEventHeader
	typ     uint32
	dropPtr uintptr // *FWPM_NET_EVENT_CLASSIFY_DROP1 when typ == classify-drop
}

type netEventHeader struct {
	timestamp     windows.Filetime
	flags         uint32
	ipVersion     uint32
	ipProtocol    uint8
	_             [3]byte
	localAddrV4   uint32
	remoteAddrV4  uint32
	localAddrV6   [16]byte
	remoteAddrV6  [16]byte
	localPort     uint16
	remotePort    uint16
	scopeID       uint32
	appID         windows.FwpByteBlob
	userID        uintptr
}

type classifyDrop1 struct {
	filterID     uint64
	layerID      uint16
	_            [6]byte
	reauthReason uint32
	origDir      uint32
	msFwpDir     uint32
	isLoopback   int32
}

const netEventTypeClassifyDrop = 3 // FWPM_NET_EVENT_TYPE_CLASSIFY_DROP

func decodeNetEvent(e *netEvent1) (BlockEvent, bool) {
	if e == nil || e.typ != netEventTypeClassifyDrop {
		return BlockEvent{}, false
	}
	h := e.header
	var ip netip.Addr
	if h.ipVersion == 0 { // FWP_IP_VERSION_V4
		ip = netip.AddrFrom4([4]byte{
			byte(h.remoteAddrV4 >> 24), byte(h.remoteAddrV4 >> 16),
			byte(h.remoteAddrV4 >> 8), byte(h.remoteAddrV4)})
	} else {
		ip = netip.AddrFrom16(h.remoteAddrV6)
	}
	dir := "outbound"
	if e.dropPtr != 0 {
		d := (*classifyDrop1)(unsafe.Pointer(e.dropPtr))
		if d.msFwpDir == 1 { // FWP_DIRECTION_INBOUND
			dir = "inbound"
		}
	}
	be := BlockEvent{
		RemoteIP:    ip,
		RemotePort:  int(h.remotePort),
		Direction:   dir,
		ProcessPath: appIDPath(h.appID),
	}
	return be, true
}

// appIDPath converts the FWP_BYTE_BLOB app id (device-path UTF16) to a string.
func appIDPath(blob windows.FwpByteBlob) string {
	if blob.Size == 0 || blob.Data == nil {
		return ""
	}
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(blob.Data)), blob.Size/2)
	return windows.UTF16ToString(u16)
}
```

> **Implementer note (Task 11):** as with Task 10, `FWPM_NET_EVENT*` layouts are header-sensitive; validate on the VM and adjust padding. PID: `FWPM_NET_EVENT1`/`2` carry the process id in later revisions — if the subscribed struct version doesn't surface a PID, resolve it from the app-id path or leave `PID=0` and rely on `ProcessPath` (the ransomware guard's device-path→drive-letter helper can normalize the `\device\harddiskvolumeN\...` form). Emit events with whatever attribution WFP provides; never block the callback.

- [ ] **Step 6: Cross-build + host tests**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./edr/netblock/
GOOS=windows GOARCH=arm64 go build ./edr/netblock/
go test ./edr/netblock/ -run TestNoopConnAudit -v
```
Expected: builds succeed; test PASS.

- [ ] **Step 7: Commit**

```bash
git add edr/netblock/connaudit.go edr/netblock/connaudit_windows.go edr/netblock/connaudit_other.go edr/netblock/connaudit_other_test.go
git commit -m "feat(edr): WFP net-event drop audit reader (windows) + noop stub"
```

---

## Task 12: Manager (orchestrator + status health)

**Files:**
- Create: `edr/netblock/manager.go`
- Test: `edr/netblock/manager_test.go`

**Interfaces:**
- Consumes: everything above; `event.Spool` (append), `cache.*`.
- Produces: `netblock.Deps{Cfg config.EDRConfig, Cache *cache.Cache, Spool spooler, Sys SystemNets}`; `spooler` interface `{ Append(string) error }`; `netblock.NewManager(Deps) *Manager`; `(*Manager).Run(ctx)`; `(*Manager).Health() Health`; `netblock.Health{Enabled,Enforce bool, Indicators, ActiveFilters, RecentBlocks, AllowlistedDropped int, FeedAgeSec int64, FeedStale bool}`.

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/manager_test.go
package netblock

import (
	"context"
	"net/netip"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type memSpool struct {
	mu    sync.Mutex
	lines []string
}

func (m *memSpool) Append(s string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lines = append(m.lines, s)
	return nil
}

func TestManagerBlocksAndEmits(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = false
	sp := &memSpool{}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp})

	// Directly drive the desired-set application (no live feed in the test).
	i, _ := ParseIndicator("9.9.9.9", "ip", 1)
	m.applyIndicators([]Indicator{i})
	if m.Health().ActiveFilters != 1 {
		t.Fatalf("active filters = %d, want 1 (noop blocker counts via enforcer)", m.Health().ActiveFilters)
	}

	// A drop audit event → a branded network_watcher event + a cache record.
	m.onBlockEvent(BlockEvent{RemoteIP: netip.MustParseAddr("9.9.9.9"), RemotePort: 443, Direction: "outbound"})
	if len(sp.lines) != 1 {
		t.Fatalf("want 1 spooled event, got %d", len(sp.lines))
	}

	// Allowlisted target must be dropped from the desired set.
	cfg2 := config.Default()
	cfg2.Blocklist.AllowPrivateRanges = true
	m2 := NewManager(Deps{Cfg: cfg2, Cache: c, Spool: &memSpool{}})
	priv, _ := ParseIndicator("10.1.2.3", "ip", 1)
	m2.applyIndicators([]Indicator{priv})
	if m2.Health().ActiveFilters != 0 {
		t.Fatal("allowlisted indicator must not become a filter")
	}
	if m2.Health().AllowlistedDropped != 1 {
		t.Fatalf("allowlisted-dropped = %d, want 1", m2.Health().AllowlistedDropped)
	}
	_ = time.Second
}
```

> Note: with the host no-op blocker `Count()` returns 0, so `applyIndicators` tracks its own desired-count for `ActiveFilters` health (via the enforcer's applied set). The test asserts against the enforcer-tracked count, which is real on all platforms.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestManagerBlocks -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `manager.go`**

```go
// edr/netblock/manager.go
package netblock

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

type spooler interface{ Append(string) error }

type Deps struct {
	Cfg   config.EDRConfig
	Cache *cache.Cache
	Spool spooler
	Sys   SystemNets
}

type Health struct {
	Enabled, Enforce   bool
	Indicators         int
	ActiveFilters      int
	RecentBlocks       int
	AllowlistedDropped int
	FeedAgeSec         int64
	FeedStale          bool
}

type Manager struct {
	cfg      config.EDRConfig
	cache    *cache.Cache
	spool    spooler
	store    *Store
	allow    *Allowlist
	enf      *Enforcer
	feed     *Feed
	audit    ConnAudit

	lastSync     atomic.Int64 // unix sec
	recentBlocks atomic.Int64
	dropped      atomic.Int64
	desired      atomic.Int64
	mu           sync.Mutex
}

func NewManager(d Deps) *Manager {
	blk, err := NewOSBlocker(config.ServiceName + "-Blocklist")
	if err != nil {
		logWarn("blocklist WFP init failed, enforcement disabled: %v", err)
		blk, _ = newDisabledBlocker()
	}
	m := &Manager{
		cfg:   d.Cfg,
		cache: d.Cache,
		spool: d.Spool,
		store: NewStore(),
		allow: BuildAllowlist(d.Cfg, d.Sys),
		enf:   NewEnforcer(blk, d.Cfg.Blocklist.Direction),
		audit: NewConnAudit(),
	}
	m.feed = NewFeed(d.Cfg, d.Cache, m.store, m.applyIndicators)
	return m
}

func (m *Manager) Run(ctx context.Context) {
	// Startup convergence: re-apply from cached indicators before the first fetch.
	if recs, err := m.cache.ListIndicators(); err == nil && len(recs) > 0 {
		inds := make([]Indicator, 0, len(recs))
		for _, r := range recs {
			if ind, err := ParseIndicator(r.Value, r.Type, r.Level); err == nil {
				inds = append(inds, ind)
			}
		}
		m.store.Swap(inds)
		m.applyIndicators(inds)
	}

	if m.cfg.Blocklist.Enforce {
		blockCh := make(chan BlockEvent, 256)
		go m.audit.Run(ctx, blockCh)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case be := <-blockCh:
					m.onBlockEvent(be)
				}
			}
		}()
	}

	m.feed.Run(ctx) // blocks until ctx.Done()
	_ = m.enf.Clear()
}

// applyIndicators recomputes the desired IP set (allowlist-filtered) and reconciles
// the enforcer. Domain/hostname indicators are ignored in Plan 1.
func (m *Manager) applyIndicators(inds []Indicator) {
	m.lastSync.Store(time.Now().Unix())
	var desired []netip.Addr
	var dropped int
	for _, ind := range inds {
		switch ind.Type {
		case "ip":
			if a, err := netip.ParseAddr(ind.Value); err == nil {
				if m.allow.Allowed(a) {
					dropped++
					continue
				}
				desired = append(desired, a)
			}
		case "cidr":
			// CIDR filters are applied directly by the blocker in Plan 2's
			// prefix path; Plan 1 enforces exact IPs. Expand is out of scope —
			// CIDRs are matched for *detection* via the store, blocked in Plan 2.
		}
	}
	m.dropped.Store(int64(dropped))
	if !m.cfg.Blocklist.Enforce {
		m.desired.Store(0)
		return
	}
	if err := m.enf.Apply(desired); err != nil {
		logWarn("blocklist enforce apply: %v", err)
	}
	m.desired.Store(int64(len(desired)))
}

// onBlockEvent turns a WFP drop into a branded event + audit record.
func (m *Manager) onBlockEvent(be BlockEvent) {
	m.recentBlocks.Add(1)
	ind, _ := m.store.Match(be.RemoteIP)
	var p *event.ProcInfo
	if be.PID != 0 || be.ProcessPath != "" {
		p = &event.ProcInfo{PID: be.PID, Image: be.ProcessPath}
	}
	ev := event.NewNetworkEvent(event.ActionBlocked, be.Direction, be.RemoteIP.String(),
		be.RemotePort, "", ind.Value, p)
	if js, err := ev.ToJSON(); err == nil {
		_ = m.spool.Append(js)
	}
	_ = m.cache.RecordNetBlock(cache.NetBlockRecord{
		Direction: be.Direction, RemoteIP: be.RemoteIP.String(), RemotePort: be.RemotePort,
		PID: be.PID, ProcessPath: be.ProcessPath, Matched: ind.Value, MatchedType: ind.Type,
		Level: ind.Level, Action: event.ActionBlocked,
	})
}

func (m *Manager) Health() Health {
	age := int64(0)
	stale := false
	if ls := m.lastSync.Load(); ls > 0 {
		age = time.Now().Unix() - ls
		stale = age > int64(24*3600) // >24h with no successful sync
	}
	return Health{
		Enabled:            m.cfg.Blocklist.Enabled,
		Enforce:            m.cfg.Blocklist.Enforce,
		Indicators:         m.store.Count(),
		ActiveFilters:      int(m.desired.Load()),
		RecentBlocks:       int(m.recentBlocks.Swap(0)),
		AllowlistedDropped: int(m.dropped.Load()),
		FeedAgeSec:         age,
		FeedStale:          stale,
	}
}
```

Add a tiny always-available disabled blocker (used when WFP init fails) in `util.go`:

```go
// appended to edr/netblock/util.go
type disabledBlocker struct{}

func newDisabledBlocker() (Blocker, error) { return disabledBlocker{}, nil }
func (disabledBlocker) AddIP(a netipAddr, dir string) error { return nil }
func (disabledBlocker) RemoveIP(a netipAddr) error          { return nil }
func (disabledBlocker) Reset() error                        { return nil }
func (disabledBlocker) Count() int                          { return 0 }
```

> Replace `netipAddr` with `netip.Addr` and add the `net/netip` import to `util.go` (kept as a named alias here only to show the method set; use the real type).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run TestManagerBlocks -v && go test ./edr/netblock/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/netblock/manager.go edr/netblock/util.go edr/netblock/manager_test.go
git commit -m "feat(edr): netblock manager (feed→store→enforcer, drop eventing, health)"
```

---

## Task 13: Service wiring + status.json

**Files:**
- Modify: `edr/service/service.go`
- Test: `edr/service/service_blocklist_test.go` (create — status-doc field test, no service start)

**Interfaces:**
- Consumes: `netblock.NewManager`, `netblock.Health`.
- Produces: blocklist health surfaced in `status.json`; `netblock` wired in `startPipeline`.

- [ ] **Step 1: Write the failing test**

```go
// edr/service/service_blocklist_test.go
package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStatusDocHasBlocklist(t *testing.T) {
	// statusDoc must marshal blocklist fields; build one with values set.
	d := statusDoc{}
	d.BlocklistEnabled = true
	d.BlocklistEnforce = true
	d.BlocklistIndicators = 42
	b, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, k := range []string{"blocklist_enabled", "blocklist_enforce", "blocklist_indicators"} {
		if !strings.Contains(s, k) {
			t.Fatalf("status doc missing %q: %s", k, s)
		}
	}
	if strings.Contains(strings.ToLower(s), "threatwinds") {
		t.Fatal("status doc must not name the intel vendor")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/service/ -run TestStatusDocHasBlocklist -v`
Expected: FAIL (fields undefined).

- [ ] **Step 3: Add status fields, program handle, and pipeline wiring**

In `statusDoc` (service.go), add:

```go
	BlocklistEnabled      bool  `json:"blocklist_enabled"`
	BlocklistEnforce      bool  `json:"blocklist_enforce"`
	BlocklistIndicators   int   `json:"blocklist_indicators"`
	BlocklistActiveFilters int  `json:"blocklist_active_filters"`
	BlocklistFeedStale    bool  `json:"blocklist_feed_stale"`
	BlocklistFeedAgeSec   int64 `json:"blocklist_feed_age_sec"`
	BlocklistRecentBlocks int   `json:"blocklist_recent_blocks"`
```

On the `program` struct add a handle: `netblock *netblock.Manager`.

In `startPipeline`, after the ransomware block, add:

```go
	if cfg.Blocklist.Enabled {
		nb := netblock.NewManager(netblock.Deps{
			Cfg:   cfg,
			Cache: c,
			Spool: sp,
			Sys:   resolveSystemNets(cfg), // resolvers/gateways/server IPs
		})
		p.netblock = nb
		goSafe("netblock", func() { nb.Run(ctx) })
	}
```

In `writeStatus`, populate the fields from the handle when present:

```go
	if p.netblock != nil {
		h := p.netblock.Health()
		doc.BlocklistEnabled = h.Enabled
		doc.BlocklistEnforce = h.Enforce
		doc.BlocklistIndicators = h.Indicators
		doc.BlocklistActiveFilters = h.ActiveFilters
		doc.BlocklistFeedStale = h.FeedStale
		doc.BlocklistFeedAgeSec = h.FeedAgeSec
		doc.BlocklistRecentBlocks = h.RecentBlocks
	}
```

Add `resolveSystemNets` (platform-split so DNS/gateway lookup only compiles where available; a portable version can read `/etc/resolv.conf`-style config, but on Windows use the adapter config). Minimal portable version + windows override:

```go
// service_netnets_other.go  //go:build !windows
func resolveSystemNets(cfg config.EDRConfig) netblock.SystemNets {
	return netblock.SystemNets{ServerHosts: resolveHostIPs(cfg.Server)}
}

// service_netnets_windows.go  //go:build windows
func resolveSystemNets(cfg config.EDRConfig) netblock.SystemNets {
	sn := netblock.SystemNets{ServerHosts: resolveHostIPs(cfg.Server)}
	sn.Resolvers = windowsDNSServers() // via GetAdaptersAddresses
	sn.Gateways = windowsGateways()
	return sn
}
```

`resolveHostIPs(server string) []netip.Addr` parses the host out of the server URL and `net.LookupIP`s it (portable). `windowsDNSServers`/`windowsGateways` use `GetAdaptersAddresses` (implement minimally; if unavailable, return nil — the built-in loopback/link-local/private defaults still protect the host).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/service/ -run TestStatusDocHasBlocklist -v`
Expected: PASS.

- [ ] **Step 5: Cross-build the whole module both arches**

Run:
```bash
for arch in amd64 arm64; do GOOS=windows GOARCH=$arch go build ./... || exit 1; done
```
Expected: success.

- [ ] **Step 6: Commit**

```bash
git add edr/service/
git commit -m "feat(edr): wire netblock into service pipeline + status.json"
```

---

## Task 14: CLI verbs

**Files:**
- Modify: `edr/manage.go`
- Test: `edr/manage_blocklist_test.go` (create — if `applySet`/verb parsing is unit-testable; else fold assertions into an existing manage test)

**Interfaces:**
- Produces: `config set blocklist.{enabled,enforce,refresh_hours,allow_private_ranges,direction}`; `allow network <ip|cidr> add|remove|list`; `block <ip|cidr> add|remove|list`.

- [ ] **Step 1: Write the failing test**

```go
// edr/manage_blocklist_test.go
package main

import "testing"

func TestApplySetBlocklist(t *testing.T) {
	c := baseTestConfig() // helper returning config.Default(); add if missing
	if err := applySet(&c, "blocklist.enforce", "false"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.Enforce {
		t.Fatal("enforce should be false")
	}
	if err := applySet(&c, "blocklist.refresh_hours", "12"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.RefreshHours != 12 {
		t.Fatalf("refresh = %d", c.Blocklist.RefreshHours)
	}
	if err := applySet(&c, "blocklist.direction", "out"); err != nil {
		t.Fatal(err)
	}
	if c.Blocklist.Direction != "out" {
		t.Fatalf("direction = %q", c.Blocklist.Direction)
	}
	if err := applySet(&c, "blocklist.direction", "sideways"); err == nil {
		t.Fatal("bad direction must error")
	}
}
```

(If `applySet`'s signature differs — e.g. it takes/returns differently — adapt the test to the real signature seen in `manage.go`.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/ -run TestApplySetBlocklist -v`
Expected: FAIL (cases not handled).

- [ ] **Step 3: Add the CLI cases**

In `applySet`'s switch, add:

```go
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
```

Extend the `allow` verb group to accept a `network` subject that appends to `cfg.Allowlist.Networks` (mirror how `allow path` appends to `cfg.Allowlist.Paths`, with an `netip.ParseAddr || netip.ParsePrefix` validation). Add a new `block` verb group that appends manual IP/CIDR entries to a `cfg.Blocklist` manual list — **or**, to avoid a second store, append them into `Allowlist.Networks`'s sibling `cfg.Blocklist.Manual []string` (add the field to `BlocklistConfig`, default nil, merged into the desired set by the manager). Update the CLI usage text (`manage.go` usage block) to document the new verbs. Reuse existing `parseBool`/`strconv` helpers already imported in `manage.go`.

> If you add `Blocklist.Manual []string`, also: (a) seed nothing in `Default()`, (b) copy it in `Load()`'s blocklist overlay, and (c) in `manager.applyIndicators`, merge parsed `Manual` entries into `inds` before the allowlist filter. Add one unit test in `manager_test.go` asserting a manual IP becomes a filter.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/ -run TestApplySetBlocklist -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add edr/manage.go edr/manage_blocklist_test.go
git commit -m "feat(edr): CLI verbs for blocklist config + manual block/allow"
```

---

## Task 15: Full verification + labeling audit + VM acceptance

**Files:** none new — verification task.

- [ ] **Step 1: Whole-module cross-build + vet, both arches**

Run from `utmstack-v12/agent/`:
```bash
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr || exit 1
  GOOS=windows GOARCH=$arch go build ./... || exit 1
  GOOS=windows GOARCH=$arch go vet ./edr/... || exit 1
done
```
Expected: all succeed.

- [ ] **Step 2: Host unit tests + race**

Run:
```bash
go test ./edr/... ./agent/
go test ./edr/netblock/ -race
```
Expected: all PASS.

- [ ] **Step 3: Labeling audit (extended with `threatwinds`)**

Run:
```bash
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/ ; echo "clam exit: $?"
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*threatwinds' edr/ ; echo "tw exit: $?"
grep -rniE 'Source:\s*"[^"]*(clam|threatwinds)' edr/ ; echo "src exit: $?"
```
Expected: no matches (grep exit `1`). Any surfaced string that names `threatwinds`/`clam` is a failure — reword to "UTMStack EDR" / "UTMStack Threat Intelligence".

- [ ] **Step 4: VM acceptance (Windows 10.211.55.12)** — verify final outputs, not just logs.

Provision: point `blocklist_mirror` at a local test HTTP server that serves `download/list/level1/accumulative/ip` (gzip list containing a **safe, controllable test IP** — e.g. a spare host on the VM's LAN or a throwaway public IP you own) plus its checksum. Enable enforcing (default). Then verify:

1. **IP block:** from the VM, `curl`/`Test-NetConnection` to the seeded IP → **connection refused/timed out**; a `network_watcher` `blocked` event with `remote_ip` reaches the platform; a `NetBlockRecord` row exists in `edr.db`.
2. **Allowlist:** add the seeded IP via `utmstack_edr allow network <ip> add`, reload → connection now **succeeds**; status shows `blocklist_allowlisted_dropped` incremented; no filter for it.
3. **Fail-open:** with the IP blocked, `Stop-Service UTMStackEDR` (or kill the process) → the IP is **reachable again** (WFP dynamic filters released). `Start-Service` → filter re-applied from cache, IP blocked again.
4. **Self-lockout guard:** inject the platform/mirror IP and a DNS-server IP as indicators → confirm they are **never** blocked (status `allowlisted_dropped` reflects the drops; connectivity to platform + DNS intact).
5. **Kill-switch:** `utmstack_edr config set blocklist.enabled false`, restart → all filters gone (`netsh wfp show filters` shows no UTMStack sublayer filters), IP reachable.
6. **Refresh:** update the mirror list to drop the IP → after `UpdateOnce` (restart or wait a cycle) the filter is removed; add a new IP → a filter appears.
7. Confirm status.json `blocklist_*` fields populate and `blocklist_indicators` matches the seeded count.

- [ ] **Step 5: Record results + commit any VM fixes**

Document VM findings (mirroring how Plans 1–4 recorded VM bug-fixes). Commit fixes with `fix(edr):` messages. Update the memory note `edr-network-blocklist` with what was validated and any bugs found.

```bash
git add -A
git commit -m "test(edr): VM-validate network blocklist IP enforcement (Plan 1)"
```

---

## Self-review — spec coverage

- Spec §1.1 decisions → Tasks 8/10 (WFP), 12 (manager), 8 (mirror), 1 (on+enforcing default). ✅
- §1.2 scope/limits (IP+CIDR in, URL/DoH/domain out for Plan 1) → Tasks 4/9/12; domain deferred to Plan 2 as spec §9 says. ✅
- §2 package layout → Tasks 4–12 create the files. ✅ (`dnsfeed_*` intentionally deferred to Plan 2.)
- §3 fail-open + sublayer + kill-switch → Task 10 (dynamic session) + Task 12 (`Clear()` on exit) + Task 15 VM steps 3/5. ✅
- §3.1 self-lockout allowlist → Task 5 + Task 13 `resolveSystemNets` + Task 15 step 4. ✅
- §5 feed ingestion (mirror, checksum, cadence, staleness) → Task 8 + Health.FeedStale (Task 12). ✅
- §6 data/event model → Tasks 2, 3. ✅
- §6.3 branding (`threatwinds` audit) → Task 15 step 3 + global constraints. ✅
- §7 config/CLI/status → Tasks 1, 13, 14. ✅
- §8 testing → per-task host tests + Task 15 VM acceptance. ✅

**Known deferrals (Plan 2, by design):** DNS-Client ETW feed, domain/hostname reactive blocking, CIDR *enforcement* (Plan 1 stores CIDRs for match but enforces exact IPs — the WFP condition already supports masks, so Plan 2 extends `AddIP`→`AddPrefix`), daily-delta application (Plan 1 re-bootstraps full each cycle; delta parser exists and is tested, wired in Plan 2).

**Placeholder scan:** the two WFP/ETW "implementer notes" describe VM-iteration on struct offsets — this is the project's established Windows-native model, not a plan placeholder (real API code is present). No `TODO`/`TBD` in code steps.
