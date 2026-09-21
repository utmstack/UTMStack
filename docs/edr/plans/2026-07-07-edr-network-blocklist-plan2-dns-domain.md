# EDR Network Blocklist — Plan 2: DNS-ETW Domain Reactive Blocking

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Block blacklisted **domains/hostnames** by watching DNS resolutions (Microsoft-Windows-DNS-Client ETW) and reactively adding short-TTL WFP filters for the resolved IPs, plus **CIDR enforcement** and **daily-delta feed application** — building on the Plan 1 IP-enforcement core.

**Architecture:** A new DNS-Client ETW feed (`dnsfeed_windows.go`, mirroring the ransomware guard's `golang-etw` file feed) surfaces `(queryName, resolvedIPs, pid)` on each DNS answer. A pure name matcher tests the query (and its parent domains) against the blocklist; on a hit, a TTL-bounded reactive blocker adds WFP block filters for the resolved IPs and emits a branded `network_watcher` event naming the domain. CIDR indicators graduate from match-only to enforced (WFP mask filters), and the feed applies daily `add|del` deltas instead of re-bootstrapping every cycle. Portable logic is TDD'd on macOS; ETW/WFP edges are `//go:build windows` with `!windows` stubs, VM-validated.

**Tech Stack:** Go 1.25.7, `github.com/0xrawsec/golang-etw/etw` (already a dep, ransomware guard), `golang.org/x/sys/windows` (WFP), existing `netblock` package. No new third-party dependency.

## Global Constraints

- **Branding (critical):** every emitted event field + surfaced log/status string says **"UTMStack EDR"**; intel source is **"UTMStack Threat Intelligence"**. `clam`/`clamd` and `threatwinds` must NEVER appear in an event field or surfaced log line. Labeling audit must stay green: `grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*(clam|threatwinds)' edr/` returns nothing.
- **Go-first.** No new third-party Go dependency beyond what's already in `go.mod` (`golang-etw`, `golang.org/x/sys`).
- **`golang-etw` is a windows-only import** — NEVER `go mod tidy` on the macOS host (it drops the dep from go.mod). Windows ETW code is behind `//go:build windows`.
- **No kernel driver.** Reactive blocking is user-mode WFP filters (Plan 1's `Blocker`), in the same dynamic session (fail-open).
- **Own `edr.db`.** New persisted state (delta cursor) goes in the EDR's cache, never the agent's `logs.db`.
- **Supported arches:** windows/amd64 **and** windows/arm64; all portable code also builds/tests on darwin. Every task cross-builds + vets both windows arches.
- Event source stays `network_watcher`; the `Domain` event field (added in Plan 1) carries the matched name.

**Spec:** `docs/superpowers/specs/2026-07-06-edr-network-blocklist-design.md` (§4 DNS-Client ETW reactive; §9 phasing).
**Builds on Plan 1:** `docs/superpowers/plans/2026-07-06-edr-network-blocklist-plan1-ip-enforcement.md` (implemented + VM-validated, incl. connaudit detection events).

All Go commands run from `utmstack-v12/agent/`.

---

## Existing interfaces this plan consumes (verified)

- `netblock.Indicator{Value, Type string; Level int}`; `ParseIndicator(value, typ string, level int) (Indicator, error)` — already handles `domain`/`hostname` (lowercases, trims trailing dot).
- `netblock.Matcher`: `NewMatcher()`, `Add(Indicator)`, `MatchIP(netip.Addr) (Indicator, bool)`, `Len() int` (exact IP map + CIDR prefixes; `MatchIP`/`Add` `Unmap()` addresses).
- `netblock.Store`: `Swap([]Indicator)`, `Snapshot() *Matcher`, `Match(netip.Addr) (Indicator, bool)`, `Count() int` (atomic pointer swap).
- `netblock.Blocker` interface: `AddIP(netip.Addr, dir string) error`, `RemoveIP(netip.Addr) error`, `Reset() error`, `Count() int`. WFP impl `wfp_windows.go` (dynamic session, `v4AddrAndMask{addr,mask uint32}` / `v6AddrAndMask{addr [16]byte, prefixLength byte}` — masks already in the struct, currently `/32` and `/128`).
- `netblock.Enforcer`: `NewEnforcer(b Blocker, dir string)`, `Apply([]netip.Addr) error`, `Clear() error`, `BlockerCount() int`.
- `netblock.Feed`: `NewFeed(cfg, cache, store, applyFn)`, `UpdateOnce(ctx) error`, `Run(ctx)`; `ParseAccumulative`, `ParseDaily(r) ([]DailyOp,error)`, `DailyOp{Value,Type,Op string}`.
- `netblock.Manager`: `NewManager(Deps)`, `Run(ctx)` (startup convergence → conn-audit consumer → feed.Run), `applyLocked([]Indicator)` (allowlist→enforcer, m.mu held), `onBlockEvent(BlockEvent)`, `Health()`. `Deps{Cfg, Cache, Spool, Sys}`.
- `event.NewNetworkEvent(action, direction, remoteIP string, port int, domain, indicator string, p *ProcInfo) Event`.
- `config.BlocklistConfig`: `IndicatorTypes []string` (Plan 1 default `["ip"]`), `DomainBlockTTLMin int` (default 30), `Enforce`, `Levels`.
- Ransomware ETW pattern (`edr/ransomware/feed_windows.go`): `etw.NewRealTimeSession(name)`, `etw.Provider{GUID,Name,EnableLevel:0xff}`, `session.EnableProvider`, `etw.NewRealTimeConsumer(ctx).FromSessions(session)`, `c.EventCallback = func(e *etw.Event) error{ e.GetPropertyString("X"); e.System.EventID; e.System.Execution.ProcessID }`, `c.Start()`, `c.Err()`, `c.Stop()`. Device-path normalizer `normalizeKernelPath` + `QueryDosDevice` (in the same file).

---

## File structure (Plan 2)

| File | Responsibility |
|---|---|
| `edr/netblock/namematch.go` (create) | Pure domain/hostname suffix matcher (`NameSet`). |
| `edr/netblock/matcher.go` (modify) | Fold `NameSet` into `Matcher`; add `MatchName`; `Add` routes domain/hostname. |
| `edr/netblock/reactive.go` (create) | TTL-bounded reactive IP blocker (add resolved IPs, expire after TTL, injected clock). |
| `edr/netblock/dnsparse.go` (create) | Parse DNS-Client `QueryResults` string → resolved `netip.Addr`s. |
| `edr/netblock/dnsfeed_windows.go` / `dnsfeed_other.go` (create) | DNS-Client ETW feed → `DNSEvent{Name, IPs, PID}`; `!windows` nop. |
| `edr/netblock/wfp_windows.go` (modify) | `AddPrefix(netip.Prefix, dir)` for CIDR mask filters. |
| `edr/netblock/wfp_other.go` (modify) | nop `AddPrefix`. |
| `edr/netblock/enforcer.go` (modify) | `Blocker.AddPrefix`; enforcer applies CIDRs. |
| `edr/netblock/feed.go` (modify) | Daily-delta application + cursor (via cache). |
| `edr/cache/netblock.go` (modify) | `FeedState.LastDelta` already exists; add a `DeltaCursor` accessor if needed. |
| `edr/netblock/manager.go` (modify) | Consume DNS feed → reactive block + domain event; enforce CIDRs; wire in Run. |
| `edr/netblock/netpath.go` (create) | Shared device-path→drive-letter normalizer (extracted, reused by DNS PID path + connaudit). |
| `edr/config/config.go` (modify) | `IndicatorTypes` default `["ip","domain","hostname"]`. |
| `*_test.go` | Sibling tests for every portable file. |

---

## Task 1: Domain/hostname name matcher

**Files:**
- Create: `edr/netblock/namematch.go`, `edr/netblock/namematch_test.go`

**Interfaces:**
- Produces: `netblock.NameSet` with `NewNameSet()`, `Add(name string)`, `Match(query string) (string, bool)`, `Len() int`. `Match` returns the matched blocklist entry. A query matches if it equals a blocklisted name OR is a subdomain of a blocklisted domain (label-boundary suffix match: `c2.evil.com` matches `evil.com`, but `notevil.com` does NOT match `evil.com`).

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/namematch_test.go
package netblock

import "testing"

func TestNameSetMatch(t *testing.T) {
	ns := NewNameSet()
	ns.Add("evil.com")
	ns.Add("bad.example.org")
	cases := []struct {
		q    string
		want bool
	}{
		{"evil.com", true},
		{"c2.evil.com", true},     // subdomain
		{"a.b.evil.com", true},    // deep subdomain
		{"EVIL.COM", true},        // case-insensitive
		{"evil.com.", true},       // trailing dot
		{"notevil.com", false},    // not a label-boundary match
		{"evil.com.attacker.net", false}, // evil.com is a label, but not a suffix-parent here
		{"example.org", false},    // parent of a listed host is not listed
		{"bad.example.org", true},
		{"x.bad.example.org", true},
		{"good.com", false},
	}
	for _, c := range cases {
		_, ok := ns.Match(c.q)
		if ok != c.want {
			t.Errorf("Match(%q) = %v, want %v", c.q, ok, c.want)
		}
	}
	if ns.Len() != 2 {
		t.Fatalf("Len = %d, want 2", ns.Len())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestNameSetMatch -v`
Expected: FAIL (undefined `NewNameSet`).

- [ ] **Step 3: Implement `namematch.go`**

```go
// edr/netblock/namematch.go
package netblock

import "strings"

// NameSet answers domain/hostname membership with label-boundary suffix matching:
// a query matches a listed name if it equals it or is a subdomain of it. Built
// once, then read (the Store swaps whole matchers).
type NameSet struct {
	names map[string]struct{}
}

func NewNameSet() *NameSet { return &NameSet{names: map[string]struct{}{}} }

func canonName(s string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(s), "."))
}

func (n *NameSet) Add(name string) {
	if c := canonName(name); c != "" {
		n.names[c] = struct{}{}
	}
}

func (n *NameSet) Len() int { return len(n.names) }

// Match returns the listed name that the query equals or is a subdomain of.
// Walks the query's parent domains: for "a.b.evil.com" it tests "a.b.evil.com",
// "b.evil.com", "evil.com", "com" — the first that is listed wins.
func (n *NameSet) Match(query string) (string, bool) {
	q := canonName(query)
	if q == "" {
		return "", false
	}
	for {
		if _, ok := n.names[q]; ok {
			return q, true
		}
		i := strings.IndexByte(q, '.')
		if i < 0 {
			return "", false
		}
		q = q[i+1:]
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/netblock/ -run TestNameSetMatch -v`
Expected: PASS.

- [ ] **Step 5: Commit** — SKIP (project policy: user drives git; leave uncommitted).

---

## Task 2: Fold names into the Matcher/Store

**Files:**
- Modify: `edr/netblock/matcher.go`
- Test: `edr/netblock/matcher_test.go` (add a case)

**Interfaces:**
- Consumes: `NameSet` (Task 1).
- Produces: `Matcher.MatchName(query string) (Indicator, bool)`; `Matcher.Add` now routes `domain`/`hostname` into an internal `NameSet`; `Len()` includes names. `Store.MatchName(query) (Indicator, bool)` delegates to the snapshot.

- [ ] **Step 1: Write the failing test** (append to `matcher_test.go`)

```go
func TestMatcherNames(t *testing.T) {
	m := NewMatcher()
	d, _ := ParseIndicator("evil.com", "domain", 2)
	h, _ := ParseIndicator("c2.bad.net", "hostname", 1)
	m.Add(d)
	m.Add(h)
	if ind, ok := m.MatchName("x.evil.com"); !ok || ind.Value != "evil.com" {
		t.Fatalf("MatchName(x.evil.com) = %+v ok=%v", ind, ok)
	}
	if _, ok := m.MatchName("c2.bad.net"); !ok {
		t.Fatal("exact hostname should match")
	}
	if _, ok := m.MatchName("good.com"); ok {
		t.Fatal("good.com must not match")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestMatcherNames -v`
Expected: FAIL (undefined `MatchName`).

- [ ] **Step 3: Implement**

In `edr/netblock/matcher.go`, add a `names *NameSet` field to `Matcher`, initialize it in `NewMatcher`, route `domain`/`hostname` in `Add`, add `MatchName`, and include names in `Len`:

```go
// in the Matcher struct, add:
//	names    *NameSet
// in NewMatcher():
//	return &Matcher{exact: map[netip.Addr]Indicator{}, names: NewNameSet()}

// extend Add's switch with:
	case "domain", "hostname":
		m.names.Add(ind.Value)
		m.nameInd = append(m.nameInd, ind) // keep the Indicator for Match results

// MatchName resolves a DNS query to its blocklist Indicator, if any.
func (m *Matcher) MatchName(query string) (Indicator, bool) {
	matched, ok := m.names.Match(query)
	if !ok {
		return Indicator{}, false
	}
	for _, ind := range m.nameInd {
		if ind.Value == matched {
			return ind, true
		}
	}
	return Indicator{Value: matched, Type: "domain"}, true
}
```

Add the `nameInd []Indicator` field to `Matcher` and include `len(m.names.names)` in `Len()`:

```go
func (m *Matcher) Len() int { return len(m.exact) + len(m.prefixes) + m.names.Len() }
```

In `edr/netblock/store.go`, add:

```go
func (s *Store) MatchName(query string) (Indicator, bool) { return s.cur.Load().MatchName(query) }
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./edr/netblock/ -run 'TestMatcher|TestNameSet|TestStore' -v`
Expected: PASS (existing matcher/store tests still green).

- [ ] **Step 5: Commit** — SKIP.

---

## Task 3: DNS QueryResults parser

**Files:**
- Create: `edr/netblock/dnsparse.go`, `edr/netblock/dnsparse_test.go`

**Interfaces:**
- Produces: `netblock.ParseDNSResults(queryResults string) []netip.Addr` — extracts A (type 1) and AAAA (type 28) addresses from the Microsoft-Windows-DNS-Client `QueryResults` string; ignores CNAMEs (type 5) and malformed entries.

**Note:** DNS-Client event 3008 `QueryResults` is a `;`-separated list, each entry `type: <N> <data>` (e.g. `"type: 5 e.cdn.net;type: 1 1.2.3.4;type: 28 2606:4700::1111;"`). Exact format is VM-confirmed (Task 8).

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/dnsparse_test.go
package netblock

import (
	"net/netip"
	"testing"
)

func TestParseDNSResults(t *testing.T) {
	in := "type: 5 e.cdn.net;type: 1 1.2.3.4;type: 28 2606:4700:4700::1111;type: 1 5.6.7.8;"
	got := ParseDNSResults(in)
	want := []string{"1.2.3.4", "2606:4700:4700::1111", "5.6.7.8"}
	if len(got) != len(want) {
		t.Fatalf("got %d addrs %v, want %d", len(got), got, len(want))
	}
	for i, w := range want {
		if got[i] != netip.MustParseAddr(w) {
			t.Errorf("addr[%d] = %v, want %v", i, got[i], w)
		}
	}
	if len(ParseDNSResults("type: 5 only.cname.com;garbage;")) != 0 {
		t.Fatal("no A/AAAA → empty")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestParseDNSResults -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `dnsparse.go`**

```go
// edr/netblock/dnsparse.go
package netblock

import (
	"net/netip"
	"strings"
)

// ParseDNSResults extracts A/AAAA addresses from a Microsoft-Windows-DNS-Client
// QueryResults string: ";"-separated "type: <N> <data>" records. Type 1 = A,
// type 28 = AAAA. CNAMEs (type 5) and unparseable records are skipped.
func ParseDNSResults(qr string) []netip.Addr {
	var out []netip.Addr
	for _, rec := range strings.Split(qr, ";") {
		rec = strings.TrimSpace(rec)
		if !strings.HasPrefix(rec, "type:") {
			continue
		}
		fields := strings.Fields(rec) // ["type:", "N", "data"]
		if len(fields) < 3 {
			continue
		}
		switch fields[1] {
		case "1", "28": // A or AAAA
			if a, err := netip.ParseAddr(fields[2]); err == nil {
				out = append(out, a.Unmap())
			}
		}
	}
	return out
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/netblock/ -run TestParseDNSResults -v`
Expected: PASS.

- [ ] **Step 5: Commit** — SKIP.

---

## Task 4: TTL-bounded reactive blocker

**Files:**
- Create: `edr/netblock/reactive.go`, `edr/netblock/reactive_test.go`

**Interfaces:**
- Consumes: `Blocker` (Plan 1).
- Produces: `netblock.Reactive` with `NewReactive(b Blocker, dir string, ttl time.Duration, now func() time.Time) *Reactive`; `Block(addrs []netip.Addr)` (adds via Blocker, (re)sets each addr's expiry to now+ttl); `Sweep()` (removes addrs whose expiry passed); `Active() int`. Injected `now` for testing.

**Note:** Reactive-blocked IPs are ephemeral (resolved-for-a-blacklisted-domain). They must NOT collide with the enforcer's feed-driven exact-IP set — Reactive owns its own addr→expiry map and only Blocks/Removes addrs it added. If an addr is already blocked by the feed enforcer, `AddIP` dedups (Plan 1 blocker returns nil for a known addr), so double-blocking is a no-op; Reactive still tracks expiry but its `RemoveIP` on expiry would remove the feed filter too. To avoid that, Reactive skips addrs already on the blocklist store — see Task 9 wiring (the manager only passes non-store IPs to Reactive).

- [ ] **Step 1: Write the failing test**

```go
// edr/netblock/reactive_test.go
package netblock

import (
	"net/netip"
	"testing"
	"time"
)

func TestReactiveTTL(t *testing.T) {
	fb := newFake() // from enforcer_test.go (same package)
	base := time.Unix(1000, 0)
	clk := base
	r := NewReactive(fb, "both", 30*time.Minute, func() time.Time { return clk })

	a := netip.MustParseAddr("1.2.3.4")
	r.Block([]netip.Addr{a})
	if fb.Count() != 1 || r.Active() != 1 {
		t.Fatalf("after Block: blocker=%d active=%d", fb.Count(), r.Active())
	}
	// Before TTL: sweep keeps it.
	clk = base.Add(20 * time.Minute)
	r.Sweep()
	if fb.Count() != 1 {
		t.Fatal("must survive before TTL")
	}
	// Re-block refreshes expiry.
	r.Block([]netip.Addr{a})
	clk = base.Add(40 * time.Minute) // 20m since refresh
	r.Sweep()
	if fb.Count() != 1 {
		t.Fatal("re-block should have refreshed TTL")
	}
	// Past TTL from last refresh: swept.
	clk = base.Add(75 * time.Minute)
	r.Sweep()
	if fb.Count() != 0 || r.Active() != 0 {
		t.Fatalf("must expire: blocker=%d active=%d", fb.Count(), r.Active())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestReactiveTTL -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement `reactive.go`**

```go
// edr/netblock/reactive.go
package netblock

import (
	"net/netip"
	"sync"
	"time"
)

// Reactive holds short-TTL WFP blocks for IPs resolved from blacklisted domains.
// Each Block (re)sets an addr's expiry to now+ttl; Sweep removes expired addrs.
// It owns only the addrs it added (distinct from the feed enforcer's set).
type Reactive struct {
	mu      sync.Mutex
	b       Blocker
	dir     string
	ttl     time.Duration
	now     func() time.Time
	expiry  map[netip.Addr]time.Time
}

func NewReactive(b Blocker, dir string, ttl time.Duration, now func() time.Time) *Reactive {
	return &Reactive{b: b, dir: dir, ttl: ttl, now: now, expiry: map[netip.Addr]time.Time{}}
}

func (r *Reactive) Block(addrs []netip.Addr) {
	r.mu.Lock()
	defer r.mu.Unlock()
	exp := r.now().Add(r.ttl)
	for _, a := range addrs {
		if _, ok := r.expiry[a]; !ok {
			if err := r.b.AddIP(a, r.dir); err != nil {
				continue
			}
		}
		r.expiry[a] = exp
	}
}

func (r *Reactive) Sweep() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	for a, exp := range r.expiry {
		if now.After(exp) {
			_ = r.b.RemoveIP(a)
			delete(r.expiry, a)
		}
	}
}

func (r *Reactive) Active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.expiry)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/netblock/ -run TestReactiveTTL -v`
Expected: PASS.

- [ ] **Step 5: Commit** — SKIP.

---

## Task 5: CIDR enforcement (WFP mask filters)

**Files:**
- Modify: `edr/netblock/enforcer.go`, `edr/netblock/wfp_windows.go`, `edr/netblock/wfp_other.go`
- Test: `edr/netblock/enforcer_test.go` (extend the fake + add a test)

**Interfaces:**
- Produces: `Blocker.AddPrefix(p netip.Prefix, dir string) error`, `Blocker.RemovePrefix(p netip.Prefix) error`; `Enforcer.ApplyPrefixes(prefixes []netip.Prefix) error` (diffs a desired CIDR set like `Apply` does for IPs).

- [ ] **Step 1: Write the failing test**

```go
// add to edr/netblock/enforcer_test.go
func TestEnforcerPrefixDiff(t *testing.T) {
	fb := newFakeP() // prefix-aware fake below
	e := NewEnforcer(fb, "both")
	a := netip.MustParsePrefix("10.0.0.0/8")
	b := netip.MustParsePrefix("192.168.0.0/16")
	cc := netip.MustParsePrefix("172.16.0.0/12")
	if err := e.ApplyPrefixes([]netip.Prefix{a, b}); err != nil {
		t.Fatal(err)
	}
	if fb.pcount() != 2 {
		t.Fatalf("want 2 prefixes, got %d", fb.pcount())
	}
	if err := e.ApplyPrefixes([]netip.Prefix{a, cc}); err != nil {
		t.Fatal(err)
	}
	if fb.pcount() != 2 || !fb.hasP(a) || !fb.hasP(cc) || fb.hasP(b) {
		t.Fatalf("diff wrong: %v", fb.pkeys())
	}
}
```

Extend `fakeBlocker` in `enforcer_test.go` to implement the new methods (add a `pset map[netip.Prefix]string`, `AddPrefix`, `RemovePrefix`, `pcount`, `hasP`, `pkeys`, and a `newFakeP()` constructor that also initializes `set`). Keep the existing `fakeBlocker` methods.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestEnforcerPrefixDiff -v`
Expected: FAIL (undefined `AddPrefix`/`ApplyPrefixes`).

- [ ] **Step 3: Implement**

Extend the `Blocker` interface in `enforcer.go`:

```go
type Blocker interface {
	AddIP(addr netip.Addr, dir string) error
	RemoveIP(addr netip.Addr) error
	AddPrefix(p netip.Prefix, dir string) error
	RemovePrefix(p netip.Prefix) error
	Reset() error
	Count() int
}
```

Add a prefix reconciler to `Enforcer` (mirror `Apply`):

```go
// add an appliedP map to Enforcer (init in NewEnforcer: appliedP: map[netip.Prefix]struct{}{})
func (e *Enforcer) ApplyPrefixes(desired []netip.Prefix) error {
	want := make(map[netip.Prefix]struct{}, len(desired))
	for _, p := range desired {
		want[p] = struct{}{}
		if _, ok := e.appliedP[p]; !ok {
			if err := e.b.AddPrefix(p, e.dir); err != nil {
				return err
			}
			e.appliedP[p] = struct{}{}
		}
	}
	for p := range e.appliedP {
		if _, ok := want[p]; !ok {
			if err := e.b.RemovePrefix(p); err != nil {
				return err
			}
			delete(e.appliedP, p)
		}
	}
	return nil
}
```

Also update `Enforcer.Clear` to reset `appliedP` (and the WFP `Reset` already removes all filters).

In `wfp_windows.go`, implement `AddPrefix`/`RemovePrefix` reusing `addFilter` but with the real mask. The existing `addFilter(addr, layer)` builds a `/32`/`/128` condition; generalize by passing a prefix. Add:

```go
func (b *wfpBlocker) AddPrefix(p netip.Prefix, dir string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := p.Masked()
	if _, ok := b.prefixFilters[key]; ok {
		return nil
	}
	var ids []uint64
	for _, layer := range layersFor(dir, key.Addr().Is6()) {
		id, err := b.addPrefixFilter(key, layer)
		if err != nil {
			return err
		}
		ids = append(ids, id)
	}
	b.prefixFilters[key] = ids
	return nil
}

func (b *wfpBlocker) RemovePrefix(p netip.Prefix) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := p.Masked()
	for _, id := range b.prefixFilters[key] {
		procFilterDeleteById.Call(uintptr(b.engine), uintptr(id))
	}
	delete(b.prefixFilters, key)
	return nil
}

// addPrefixFilter mirrors addFilter but sets the V4/V6 mask from the prefix bits.
func (b *wfpBlocker) addPrefixFilter(p netip.Prefix, layer windows.GUID) (uint64, error) {
	cond := filterCondition{fieldKey: condRemoteAddress, matchType: fwpMatchEqual}
	if p.Addr().Is4() {
		v4 := v4AddrAndMask{addr: be32(p.Addr().As4()), mask: prefixMask4(p.Bits())}
		cond.conditionValue = fwpValue{typ: fwpV4AddrMask, data: uintptr(unsafe.Pointer(&v4))}
		return b.commitFilter(layer, &cond, unsafe.Pointer(&v4))
	}
	a16 := p.Addr().As16()
	v6 := v6AddrAndMask{addr: a16, prefixLength: byte(p.Bits())}
	cond.conditionValue = fwpValue{typ: fwpV6AddrMask, data: uintptr(unsafe.Pointer(&v6))}
	return b.commitFilter(layer, &cond, unsafe.Pointer(&v6))
}

func prefixMask4(bits int) uint32 {
	if bits <= 0 {
		return 0
	}
	if bits >= 32 {
		return 0xffffffff
	}
	return ^uint32(0) << (32 - bits)
}
```

Add `prefixFilters map[netip.Prefix][]uint64` to `wfpBlocker` (init in `NewOSBlocker`), and delete prefix filters in `Reset` too (iterate `prefixFilters`).

In `wfp_other.go`, add nop `AddPrefix`/`RemovePrefix` to `noopBlocker`, and to `disabledBlocker` in `util.go`.

- [ ] **Step 4: Run tests + windows build**

Run:
```bash
go test ./edr/netblock/ -run 'TestEnforcer|TestNoopBlocker' -v
GOOS=windows GOARCH=amd64 go build ./edr/netblock/ && GOOS=windows GOARCH=arm64 go build ./edr/netblock/
```
Expected: host tests PASS; both windows arches compile.

- [ ] **Step 5: Commit** — SKIP.

---

## Task 6: Daily-delta feed application

**Files:**
- Modify: `edr/netblock/feed.go`
- Test: `edr/netblock/feed_test.go` (add a delta test)

**Interfaces:**
- Consumes: `ParseDaily`, `cache.FeedState` (has `LastDelta time.Time`).
- Produces: `Feed.applyDelta(ops []DailyOp)` — merges `add`/`del` ops into the current in-memory indicator set and hot-swaps; `Feed.UpdateOnce` fetches the daily ndjson after the accumulative bootstrap and applies it. Overridable `FetchDaily func(url string) (io.ReadCloser, error)` (defaults to the shared http path).

- [ ] **Step 1: Write the failing test**

```go
// add to edr/netblock/feed_test.go
func TestFeedAppliesDelta(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	store := NewStore()
	i1, _ := ParseIndicator("1.1.1.1", "ip", 1)
	i2, _ := ParseIndicator("2.2.2.2", "ip", 1)
	store.Swap([]Indicator{i1, i2})
	f := NewFeed(config.Default(), c, store, func([]Indicator) {})
	// del 1.1.1.1, add 3.3.3.3
	ops := []DailyOp{
		{Value: "1.1.1.1", Type: "ip", Op: "del"},
		{Value: "3.3.3.3", Type: "ip", Op: "add"},
	}
	f.applyDelta(ops)
	if _, ok := store.Match(netip.MustParseAddr("1.1.1.1")); ok {
		t.Fatal("1.1.1.1 should be removed")
	}
	if _, ok := store.Match(netip.MustParseAddr("3.3.3.3")); !ok {
		t.Fatal("3.3.3.3 should be added")
	}
	if _, ok := store.Match(netip.MustParseAddr("2.2.2.2")); !ok {
		t.Fatal("2.2.2.2 should remain")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestFeedAppliesDelta -v`
Expected: FAIL (undefined `applyDelta`).

- [ ] **Step 3: Implement `applyDelta`**

Add to `feed.go`. Maintain the current indicator set from the store snapshot (the store doesn't expose its indicators; keep a `cur []Indicator` on the Feed, updated in `UpdateOnce`'s swap and `applyDelta`). Simplest: keep `f.cur []Indicator` alongside the store.

```go
// add field: cur []Indicator  (guarded by f.mu sync.Mutex)
func (f *Feed) applyDelta(ops []DailyOp) {
	f.mu.Lock()
	defer f.mu.Unlock()
	idx := map[string]Indicator{} // key = type|value
	for _, ind := range f.cur {
		idx[ind.Type+"|"+ind.Value] = ind
	}
	for _, op := range ops {
		ind, err := ParseIndicator(op.Value, op.Type, 0)
		if err != nil {
			continue
		}
		key := ind.Type + "|" + ind.Value
		switch op.Op {
		case "add":
			idx[key] = ind
		case "del":
			delete(idx, key)
		}
	}
	next := make([]Indicator, 0, len(idx))
	for _, ind := range idx {
		next = append(next, ind)
	}
	f.cur = next
	f.store.Swap(next)
	if f.applyFn != nil {
		f.applyFn(next)
	}
}
```

In `UpdateOnce`, after the accumulative swap, set `f.cur = all` (under `f.mu`) so deltas apply on top. Then fetch the daily ndjson for each `(level, name)` and call `applyDelta` (best-effort; a daily fetch failure logs and keeps the bootstrapped set). Persist `FeedState.LastDelta = now`.

> Add the `f.mu sync.Mutex` and `f.cur []Indicator` fields to the `Feed` struct. Guard `UpdateOnce`'s `f.cur` assignment with `f.mu`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./edr/netblock/ -run 'TestFeed' -v`
Expected: PASS (existing feed tests still green).

- [ ] **Step 5: Commit** — SKIP.

---

## Task 7: Shared device-path normalizer

**Files:**
- Create: `edr/netblock/netpath.go`, `edr/netblock/netpath_windows.go`, `edr/netblock/netpath_other.go`, `edr/netblock/netpath_test.go`

**Interfaces:**
- Produces: `netblock.NormalizeDevicePath(p string) string` — rewrites `\device\harddiskvolumeN\...` → `C:\...` via `QueryDosDevice` (windows) or returns `p` unchanged (other). Used to make connaudit + DNS process paths readable. Pure fallback logic in `netpath.go`; the device-map lookup is platform-split.

- [ ] **Step 1: Write the failing test** (host — exercises the passthrough + non-device paths)

```go
// edr/netblock/netpath_test.go
package netblock

import "testing"

func TestNormalizeDevicePathPassthrough(t *testing.T) {
	// Non-device paths and already-drive-letter paths are returned unchanged.
	for _, p := range []string{`C:\Windows\x.exe`, ``, `System`, `\\?\C:\y`} {
		if got := NormalizeDevicePath(p); got != stripQ(p) {
			t.Errorf("NormalizeDevicePath(%q) = %q", p, got)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestNormalizeDevicePath -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement**

`netpath.go` (portable core + the `stripQ` helper):

```go
// edr/netblock/netpath.go
package netblock

import "strings"

func stripQ(p string) string { return strings.TrimPrefix(p, `\??\`) }
```

`netpath_other.go` (`//go:build !windows`):

```go
//go:build !windows
package netblock

func NormalizeDevicePath(p string) string { return stripQ(p) }
```

`netpath_windows.go` (`//go:build windows`) — copy the `deviceMap`/`buildDeviceMap`/lookup logic from `edr/ransomware/feed_windows.go` (lines ~121-164), renamed:

```go
//go:build windows
package netblock

import (
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

var (
	devMapOnce sync.Once
	devMap     map[string]string
)

func buildDevMap() {
	devMap = map[string]string{}
	var buf [1024]uint16
	for c := byte('A'); c <= 'Z'; c++ {
		letter := string(rune(c)) + ":"
		lp, err := windows.UTF16PtrFromString(letter)
		if err != nil {
			continue
		}
		n, err := windows.QueryDosDevice(lp, &buf[0], uint32(len(buf)))
		if err != nil || n == 0 {
			continue
		}
		if dev := windows.UTF16ToString(buf[:n]); dev != "" {
			devMap[strings.ToLower(dev)] = letter
		}
	}
}

// NormalizeDevicePath rewrites "\Device\HarddiskVolumeN\..." → "C:\...".
func NormalizeDevicePath(p string) string {
	p = stripQ(p)
	if strings.HasPrefix(strings.ToLower(p), `\device\`) {
		devMapOnce.Do(buildDevMap)
		low := strings.ToLower(p)
		for dev, letter := range devMap {
			if strings.HasPrefix(low, dev+`\`) {
				return letter + p[len(dev):]
			}
		}
	}
	return p
}
```

Then in `connaudit_windows.go` `appIDPath`, wrap the result: `return NormalizeDevicePath(windows.UTF16ToString(u16))` so connaudit events get drive-letter process paths (closes a Plan-1 cosmetic follow-up).

- [ ] **Step 4: Run tests + windows build**

Run:
```bash
go test ./edr/netblock/ -run TestNormalizeDevicePath -v
GOOS=windows GOARCH=amd64 go build ./edr/netblock/ && GOOS=windows GOARCH=arm64 go build ./edr/netblock/
```
Expected: host test PASS; windows compiles.

- [ ] **Step 5: Commit** — SKIP.

---

## Task 8: DNS-Client ETW feed

**Files:**
- Create: `edr/netblock/dnsfeed.go`, `edr/netblock/dnsfeed_windows.go`, `edr/netblock/dnsfeed_other.go`, `edr/netblock/dnsfeed_other_test.go`

**Interfaces:**
- Produces: `netblock.DNSEvent{Name string; IPs []netip.Addr; PID int; Process string}`; `netblock.DNSFeed` interface `{ Run(ctx context.Context, sink func(DNSEvent)) error }`; `netblock.NewDNSFeed() DNSFeed` (real ETW on windows; nop on other).

**Note:** DNS-Client provider GUID `{1C95126E-7EEA-49A9-A3FE-A378B03DDB4D}`, event ID **3008** (query completed) carries `QueryName` + `QueryResults`. Event ID and field names are VM-confirmed (Step 6) exactly as the ransomware plan VM-confirmed Kernel-File IDs. Mirrors `edr/ransomware/feed_windows.go`.

- [ ] **Step 1: Write the stub test (host)**

```go
// edr/netblock/dnsfeed_other_test.go
//go:build !windows
package netblock

import (
	"context"
	"testing"
	"time"
)

func TestNoopDNSFeedStops(t *testing.T) {
	f := NewDNSFeed()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = f.Run(ctx, func(DNSEvent) {}); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("nop dns feed did not stop on ctx cancel")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestNoopDNSFeed -v`
Expected: FAIL (undefined).

- [ ] **Step 3: Implement the shared type + stub**

```go
// edr/netblock/dnsfeed.go
package netblock

import (
	"context"
	"net/netip"
)

// DNSEvent is one completed DNS resolution: the queried name and the A/AAAA
// addresses it resolved to, with the querying process.
type DNSEvent struct {
	Name    string
	IPs     []netip.Addr
	PID     int
	Process string
}

// DNSFeed surfaces DNS resolutions. Real on windows (ETW); nop elsewhere.
type DNSFeed interface {
	Run(ctx context.Context, sink func(DNSEvent)) error
}
```

```go
// edr/netblock/dnsfeed_other.go
//go:build !windows
package netblock

import "context"

type nopDNSFeed struct{}

func NewDNSFeed() DNSFeed { return nopDNSFeed{} }

func (nopDNSFeed) Run(ctx context.Context, sink func(DNSEvent)) error {
	<-ctx.Done()
	return ctx.Err()
}
```

- [ ] **Step 4: Run the stub test**

Run: `go test ./edr/netblock/ -run TestNoopDNSFeed -v`
Expected: PASS.

- [ ] **Step 5: Implement the Windows DNS-Client ETW feed**

```go
// edr/netblock/dnsfeed_windows.go
//go:build windows
package netblock

import (
	"context"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
)

const (
	dnsClientProviderName = "Microsoft-Windows-DNS-Client"
	dnsClientProviderGUID = "{1C95126E-7EEA-49A9-A3FE-A378B03DDB4D}"
	dnsQueryCompletedID   = 3008 // carries QueryName + QueryResults
)

type etwDNSFeed struct{}

func NewDNSFeed() DNSFeed { return &etwDNSFeed{} }

func (f *etwDNSFeed) Run(ctx context.Context, sink func(DNSEvent)) error {
	session := etw.NewRealTimeSession("UTMStackEDR-DNSClient")
	defer func() { _ = session.Stop() }()

	prov := etw.Provider{
		GUID:        dnsClientProviderGUID,
		Name:        dnsClientProviderName,
		EnableLevel: 0xff, // filter by event ID in the callback
	}
	if err := session.EnableProvider(prov); err != nil {
		return err
	}

	c := etw.NewRealTimeConsumer(ctx).FromSessions(session)
	c.EventCallback = func(e *etw.Event) error {
		if e.System.EventID != dnsQueryCompletedID {
			return nil
		}
		name, _ := e.GetPropertyString("QueryName")
		if name == "" {
			return nil
		}
		results, _ := e.GetPropertyString("QueryResults")
		ips := ParseDNSResults(results)
		if len(ips) == 0 {
			return nil
		}
		sink(DNSEvent{
			Name:    name,
			IPs:     ips,
			PID:     int(e.System.Execution.ProcessID),
			Process: "", // DNS-Client doesn't carry the image path; PID is the handle
		})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
	logInfo("blocklist DNS feed (ETW) started")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = c.Stop()
			return ctx.Err()
		case <-ticker.C:
			if err := c.Err(); err != nil {
				_ = c.Stop()
				logWarn("blocklist DNS feed (ETW) stopped: %v", err)
				return err
			}
		}
	}
}
```

- [ ] **Step 6: Cross-build both arches + host tests**

Run:
```bash
GOOS=windows GOARCH=amd64 go build ./edr/netblock/
GOOS=windows GOARCH=arm64 go build ./edr/netblock/
go test ./edr/netblock/ -run TestNoopDNSFeed -v
```
Expected: windows compiles; host stub PASS.

> **Implementer note:** the DNS event ID (3008) and property names (`QueryName`, `QueryResults`) are the documented Microsoft-Windows-DNS-Client schema but MUST be VM-confirmed (Task 10) — if `GetPropertyString("QueryResults")` returns empty, dump `e.System.EventID` + available properties on the VM (like the connaudit byte-dump in Plan 1) and adjust the ID/field names. DoH/DoT bypasses this provider (documented spec limit).

- [ ] **Step 7: Commit** — SKIP.

---

## Task 9: Manager wiring — DNS reactive block + CIDR enforcement

**Files:**
- Modify: `edr/netblock/manager.go`
- Test: `edr/netblock/manager_test.go` (add a DNS-path test)

**Interfaces:**
- Consumes: `DNSFeed`, `Reactive`, `Store.MatchName`, `Enforcer.ApplyPrefixes`, `event.NewNetworkEvent` (with `domain`).
- Produces: `Manager.onDNSEvent(DNSEvent)` (name match → reactive-block resolved non-store IPs + emit a domain `network_watcher` event); `applyLocked` also enforces CIDRs via `ApplyPrefixes`; `Run` starts the DNS feed + a Reactive sweep ticker (both gated on `Enforce`).

- [ ] **Step 1: Write the failing test**

```go
// add to edr/netblock/manager_test.go
func TestManagerDNSReactiveBlock(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	sp := &memSpool{}
	m := NewManager(Deps{Cfg: cfg, Cache: c, Spool: sp})
	// Blocklist holds a domain.
	d, _ := ParseIndicator("evil.com", "domain", 2)
	m.store.Swap([]Indicator{d})

	// A resolution of a subdomain to two IPs → event naming the domain.
	m.onDNSEvent(DNSEvent{
		Name: "c2.evil.com",
		IPs:  []netip.Addr{netip.MustParseAddr("9.9.9.9"), netip.MustParseAddr("9.9.9.10")},
		PID:  1234,
	})
	if len(sp.lines) != 1 {
		t.Fatalf("want 1 domain event, got %d", len(sp.lines))
	}
	if !strings.Contains(sp.lines[0], `"domain":"c2.evil.com"`) ||
		!strings.Contains(sp.lines[0], `"indicator":"evil.com"`) {
		t.Fatalf("event missing domain/indicator: %s", sp.lines[0])
	}
	// A resolution of a NON-blocklisted name → no event, no block.
	m.onDNSEvent(DNSEvent{Name: "good.com", IPs: []netip.Addr{netip.MustParseAddr("8.8.8.8")}, PID: 1})
	if len(sp.lines) != 1 {
		t.Fatalf("non-blocklisted resolution must not emit; got %d", len(sp.lines))
	}
}
```

(Add `"strings"` to the test imports.)

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./edr/netblock/ -run TestManagerDNSReactiveBlock -v`
Expected: FAIL (undefined `onDNSEvent`).

- [ ] **Step 3: Implement**

Add to `Manager` (fields): `dnsFeed DNSFeed`, `reactive *Reactive`. In `NewManager`, after the enforcer is built:

```go
	m.dnsFeed = NewDNSFeed()
	ttl := time.Duration(d.Cfg.Blocklist.DomainBlockTTLMin) * time.Minute
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	m.reactive = NewReactive(blk, d.Cfg.Blocklist.Direction, ttl, time.Now)
```

Add `onDNSEvent`:

```go
// onDNSEvent reactively blocks the IPs a blacklisted domain resolved to and emits
// a branded event. IPs already on the feed blocklist are left to the enforcer
// (not double-managed by Reactive).
func (m *Manager) onDNSEvent(ev DNSEvent) {
	ind, ok := m.store.MatchName(ev.Name)
	if !ok {
		return
	}
	m.recentBlocks.Add(1)
	var toBlock []netip.Addr
	for _, ip := range ev.IPs {
		if _, onList := m.store.Match(ip); !onList {
			toBlock = append(toBlock, ip)
		}
	}
	if m.cfg.Blocklist.Enforce && len(toBlock) > 0 {
		m.reactive.Block(toBlock)
	}
	var p *event.ProcInfo
	if ev.PID != 0 {
		p = &event.ProcInfo{PID: ev.PID, Image: ev.Process}
	}
	firstIP := ""
	if len(ev.IPs) > 0 {
		firstIP = ev.IPs[0].String()
	}
	e := event.NewNetworkEvent(event.ActionBlocked, "outbound", firstIP, 0, ev.Name, ind.Value, p)
	if js, err := e.ToJSON(); err == nil {
		_ = m.spool.Append(js)
	}
	_ = m.cache.RecordNetBlock(cache.NetBlockRecord{
		Direction: "outbound", RemoteIP: firstIP, Domain: ev.Name,
		PID: ev.PID, ProcessPath: ev.Process, Matched: ind.Value, MatchedType: ind.Type,
		Level: ind.Level, Action: event.ActionBlocked,
	})
}
```

Extend `applyLocked` to enforce CIDRs (in addition to exact IPs). After building `desired []netip.Addr`, also build `desiredP []netip.Prefix` from `cidr` indicators (allowlist-filtered against the CIDR's network addr), and call `m.enf.ApplyPrefixes(desiredP)`:

```go
	case "cidr":
		if p, err := netip.ParsePrefix(ind.Value); err == nil {
			if m.allow.Allowed(p.Addr()) { // network addr allowlisted → skip
				dropped++
				continue
			}
			desiredP = append(desiredP, p.Masked())
		}
```

and after the enforce-off early return / before `m.enf.Apply(desired)`:

```go
	if err := m.enf.ApplyPrefixes(desiredP); err != nil {
		logWarn("blocklist enforce apply prefixes: %v", err)
	}
```

(When `!Enforce`, the existing `enf.Clear()` path must also clear prefixes — `Enforcer.Clear` resets `appliedP` and the WFP `Reset` removes prefix filters, so this is covered.)

In `Run`, after the conn-audit consumer, start the DNS feed + a reactive sweep (both gated on `Enforce`):

```go
	if m.cfg.Blocklist.Enforce {
		go func() {
			_ = m.dnsFeed.Run(ctx, m.onDNSEvent)
		}()
		go func() {
			t := time.NewTicker(time.Minute)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					m.reactive.Sweep()
				}
			}
		}()
	}
```

- [ ] **Step 4: Run tests + windows build + race**

Run:
```bash
go test ./edr/netblock/ -race
GOOS=windows GOARCH=amd64 go build ./... && GOOS=windows GOARCH=arm64 go build ./...
```
Expected: host tests PASS (incl. the new DNS test); both windows arches compile.

- [ ] **Step 5: Commit** — SKIP.

---

## Task 10: Config default + full verification + VM acceptance

**Files:**
- Modify: `edr/config/config.go` (IndicatorTypes default)
- Test: `edr/config/config_blocklist_test.go` (update assertion)

- [ ] **Step 1: Update the default indicator types**

In `config.go` `Default()`, change `IndicatorTypes: []string{"ip"}` → `IndicatorTypes: []string{"ip", "domain", "hostname"}`, and the `Load()` overlay's fallback likewise. Update `TestDefaultBlocklist` to assert `len(b.IndicatorTypes) == 3` and it contains `"domain"`.

Run: `go test ./edr/config/ -run TestDefaultBlocklist -v` → PASS.

- [ ] **Step 2: Full cross-build + vet, both arches**

Run from `utmstack-v12/agent/`:
```bash
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr || exit 1
  GOOS=windows GOARCH=$arch go build ./... || exit 1
  GOOS=windows GOARCH=$arch go vet ./edr/... || exit 1
done
```
Expected: all succeed.

- [ ] **Step 3: Host tests + race + labeling audit**

Run:
```bash
go test ./edr/... ./agent/
go test ./edr/netblock/ -race
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*(clam|threatwinds)' edr/ ; echo "audit exit: $? (want 1 = clean)"
```
Expected: all PASS; audit returns nothing.

- [ ] **Step 4: VM acceptance (Windows 10.211.55.12)** — verify final outputs.

The Plan-1 test harness (local mirror + prlctl deploy) applies; the mirror must additionally serve a **domain** feed. Then:

1. **Domain block:** seed a controllable blacklisted domain (e.g. a spare domain you own, or add a test hostname that resolves to a spare public IP) in the mirror's `domain` feed. From the VM, resolve it (`Resolve-DnsName` / a browser). Confirm: (a) a `network_watcher` event with `"domain":"<name>"` + `"indicator":"<listed>"` reaches the spool; (b) a subsequent connection to the resolved IP is **blocked** (the reactive WFP filter); (c) a `NetBlockRecord` with the domain is in `edr.db`.
2. **Subdomain match:** resolve `sub.<listed-domain>` → still blocked + event names the subdomain, indicator names the listed parent.
3. **TTL expiry:** after `DomainBlockTTLMin` with no re-resolution, confirm the reactive filter is swept (the resolved IP reachable again) — test with a short `domain_block_ttl_min` (e.g. 1) via `config set`.
4. **CIDR enforcement:** seed a CIDR (e.g. a `/24` covering a spare test IP) in the `ip` feed; confirm a connection to an IP inside it is **blocked** (`netsh wfp show filters` shows the mask filter), and an IP outside it is reachable.
5. **Daily delta:** update the mirror's daily ndjson to `del` a blocked domain/IP → after a refresh cycle the block lifts; `add` one → it appears.
6. **DoH note:** document that a name resolved over DoH is not caught by the domain path (its IP is still blocked if independently on the IP list).
7. **Process attribution:** confirm domain/IP events show a drive-letter process path (Task 7 normalizer), not a `\device\` path.
8. Confirm `go` cross-build both arches + host tests + labeling audit all green.

- [ ] **Step 5: Record results + commit VM fixes** — leave git to the user; document VM findings and update the `edr-network-blocklist` memory (mirroring the Plan-1 + connaudit VM records).

---

## Self-review — spec coverage

- Spec §4 DNS-Client ETW reactive → Tasks 3 (parse), 8 (ETW feed), 9 (reactive wiring), 4 (TTL). ✅
- Spec §2.1 `dnsfeed_windows.go`/`_other.go` → Task 8. ✅
- Spec §1.2 domain/hostname in scope, DoH limit documented → Tasks 1/9 + Task 10 VM step 6. ✅
- Spec §9 deferrals folded into Plan 2: CIDR enforcement → Task 5; daily-delta application → Task 6. ✅
- Spec §6.2 `Domain` event field → Task 9 (`NewNetworkEvent(..., ev.Name, ...)`). ✅
- Spec §3 reactive short-TTL filters → Task 4 + Task 9 sweep. ✅
- Branding audit extended (`threatwinds`) → Task 10 step 3 + global constraints. ✅
- Config `DomainBlockTTLMin` consumed → Task 9; `IndicatorTypes` domain/hostname default → Task 10. ✅

**Cross-cutting note (for the executor):** connaudit (Plan 1) and the DNS path (Plan 2) both emit `network_watcher` events and both feed WFP filters in the same dynamic session — the manager owns both; the Reactive set is kept disjoint from the enforcer's feed set (`onDNSEvent` skips store IPs) so TTL expiry never removes a feed filter. The WFP net-event collection dependency (connaudit) is unaffected by the DNS path.

**Placeholder scan:** WFP mask code (Task 5) and the DNS event-ID/field-name (Task 8) carry VM-confirmation notes — the project's established Windows-native model, not plan placeholders (real code present). No `TODO`/`TBD` in code steps.

**Known deferrals (post-Plan-2 follow-ups, documented not silent):** decode DNS `PID` → process image path (DNS-Client doesn't carry the image; would need a PID→path lookup); connaudit direction/PID cosmetic items from Plan 1; pre-resolve of blacklisted domains at feed time (spec chose reactive-only).
