package netblock

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeNFTOps struct {
	t                                 *testing.T
	now                               time.Time
	table, owner                      string
	plans                             map[string]*nftPlan
	current                           *nftPlan
	transactions                      []string
	failBefore, failAfter, failDelete int
	foreign                           bool
	corrupt                           bool
}

func newFakeNFT(t *testing.T) (*nftController, *fakeNFTOps) {
	f := &fakeNFTOps{t: t, now: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC), table: "utmstack_edr_test", owner: "utmstack-edr-v1-test", plans: map[string]*nftPlan{}}
	c := &nftController{ops: f, table: f.table, owner: f.owner, now: func() time.Time { return f.now }}
	return c, f
}

func (f *fakeNFTOps) register(p *nftPlan) *nftPlan { f.plans[p.Comment] = p; return p }

func nftPlanJSON(p *nftPlan, table string, now time.Time) []byte {
	objects := []any{map[string]any{"metainfo": map[string]any{"json_schema_version": 1}}, map[string]any{"table": map[string]any{"family": "inet", "name": table, "comment": p.Comment, "handle": 12}}}
	for _, s := range p.Sets {
		typ := "ipv6_addr"
		if s.V4 {
			typ = "ipv4_addr"
		}
		set := map[string]any{"family": "inet", "table": table, "name": s.Name, "type": typ, "flags": []string{"interval", "timeout"}, "handle": 23}
		var elems []any
		if remaining := p.Expires.Sub(now).Milliseconds() - 100; remaining > 0 {
			for _, v := range s.Values {
				var val any = v.First.String()
				if v.Last != v.First {
					val = map[string]any{"range": []string{v.First.String(), v.Last.String()}}
				}
				elems = append(elems, map[string]any{"elem": map[string]any{"val": val, "expires": remaining, "timeout": remaining}})
			}
		}
		if len(elems) > 0 {
			set["elem"] = elems
		}
		objects = append(objects, map[string]any{"set": set})
	}
	for _, chain := range p.Chains {
		objects = append(objects, map[string]any{"chain": map[string]any{"family": "inet", "table": table, "name": chain, "type": "filter", "hook": chain, "prio": 10, "policy": "accept", "handle": 31}})
	}
	for i, r := range p.Rules {
		objects = append(objects, map[string]any{"rule": map[string]any{"family": "inet", "table": table, "chain": r.Chain, "expr": r.Expr, "comment": fmt.Sprintf("edr-%d", i), "handle": 40 + i}})
	}
	b, _ := json.Marshal(map[string]any{"nftables": objects})
	return b
}

func (f *fakeNFTOps) Run(ctx context.Context, input []byte, args ...string) ([]byte, error) {
	f.t.Helper()
	if _, ok := ctx.Deadline(); !ok {
		f.t.Fatal("firewall operation has no finite deadline")
	}
	switch strings.Join(args, " ") {
	case "-j -n list tables":
		if f.current == nil && !f.foreign {
			return []byte(`{"nftables":[]}`), nil
		}
		return []byte(`{"nftables":[{"table":{"family":"inet","name":"` + f.table + `"}}]}`), nil
	case "-j -n list table inet " + f.table:
		if f.foreign {
			return []byte(`{"nftables":[{"table":{"family":"inet","name":"` + f.table + `","comment":"someone else"}}]}`), nil
		}
		if f.current == nil {
			return nil, errors.New("missing table")
		}
		data := nftPlanJSON(f.current, f.table, f.now)
		if f.corrupt {
			data = []byte(strings.Replace(string(data), `"policy":"accept"`, `"policy":"drop"`, 1))
		}
		return data, nil
	case "-f -":
		f.transactions = append(f.transactions, string(input))
		if f.failBefore > 0 {
			f.failBefore--
			return nil, errors.New("injected transaction rejected")
		}
		if !strings.Contains(string(input), "table inet "+f.table+" {") {
			if f.failDelete > 0 {
				f.failDelete--
				return nil, errors.New("injected delete failure")
			}
			f.current = nil
			return nil, nil
		}
		var selected *nftPlan
		for comment, p := range f.plans {
			if strings.Contains(string(input), `comment "`+comment+`"`) {
				selected = p
				break
			}
		}
		if selected == nil {
			return nil, errors.New("unregistered fake nft policy")
		}
		f.current = selected
		if f.failAfter > 0 {
			f.failAfter--
			return nil, context.DeadlineExceeded
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected operation %v", args)
	}
}

func testSafety() SafetyPolicy {
	return SafetyPolicy{Management: []netip.Addr{netip.MustParseAddr("198.51.100.7")}, Resolvers: []netip.Addr{netip.MustParseAddr("192.0.2.53")}}
}
func testIsolation() IsolationPolicy {
	return IsolationPolicy{Management: []netip.Addr{netip.MustParseAddr("198.51.100.7")}, Resolvers: []netip.Addr{netip.MustParseAddr("192.0.2.53")}, Allow: []netip.Prefix{netip.MustParsePrefix("203.0.113.0/28")}, Duration: 30 * time.Second}
}

func TestNetworkPolicySafetyAndBounds(t *testing.T) {
	now := time.Now()
	for _, mgmt := range [][]netip.Addr{nil, {netip.MustParseAddr("127.0.0.1")}, {netip.MustParseAddr("0.0.0.0")}, {netip.MustParseAddr("ff02::1")}, {netip.MustParseAddr("fe80::1")}} {
		if _, err := blockPlan("test", SafetyPolicy{Management: mgmt}, nil, nil, now); !errors.Is(err, ErrNetworkSafety) {
			t.Fatalf("unsafe management accepted %v: %v", mgmt, err)
		}
	}
	for _, duration := range []time.Duration{0, time.Second, MaxIsolationDuration + time.Second} {
		p := testIsolation()
		p.Duration = duration
		if _, err := isolationPlan("test", p, now); err == nil {
			t.Fatalf("unsafe duration accepted: %s", duration)
		}
	}
	for _, network := range []string{"0.0.0.0/0", "::/0"} {
		p := testIsolation()
		p.Allow = []netip.Prefix{netip.MustParsePrefix(network)}
		if _, err := isolationPlan("test", p, now); err == nil {
			t.Fatal("default-route exemption accepted")
		}
	}
	for _, networks := range [][]string{{"0.0.0.0/1", "128.0.0.0/1"}, {"::/1", "8000::/1"}} {
		p := testIsolation()
		p.Allow = nil
		for _, network := range networks {
			p.Allow = append(p.Allow, netip.MustParsePrefix(network))
		}
		if _, err := isolationPlan("test", p, now); err == nil {
			t.Fatalf("split default-route bypass accepted %v", networks)
		}
	}
	p := testIsolation()
	p.Allow = make([]netip.Prefix, 129)
	if _, err := isolationPlan("test", p, now); !errors.Is(err, ErrNetworkBudget) {
		t.Fatal(err)
	}
	if _, err := intervals(make([]netip.Addr, nftMaxEntries+1), nil); !errors.Is(err, ErrNetworkBudget) {
		t.Fatal(err)
	}
}

func TestNetworkIntervalCanonicalization(t *testing.T) {
	got, err := intervals([]netip.Addr{netip.MustParseAddr("::ffff:192.0.2.1"), netip.MustParseAddr("192.0.2.2"), netip.MustParseAddr("2001:db8::1")}, []netip.Prefix{netip.MustParsePrefix("192.0.2.3/32"), netip.MustParsePrefix("2001:db8::2/127")})
	if err != nil {
		t.Fatal(err)
	}
	want := []ipInterval{{netip.MustParseAddr("192.0.2.1"), netip.MustParseAddr("192.0.2.3")}, {netip.MustParseAddr("2001:db8::1"), netip.MustParseAddr("2001:db8::3")}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if _, err := normalizedPrefix(netip.MustParsePrefix("::ffff:192.0.2.0/80")); err == nil {
		t.Fatal("ambiguous mapped prefix accepted")
	}
	if v := prefixInterval(netip.MustParsePrefix("::/0")); v.Last.String() != "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff" {
		t.Fatal(v)
	}
}

func TestIsolationPlanHasFiniteDeniesAndNarrowExceptions(t *testing.T) {
	p, err := isolationPlan("test", testIsolation(), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	script, err := p.script("utmstack_edr_test", false, p.Expires.Add(-30*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, required := range []string{"policy accept", "flags interval,timeout", "timeout 30s", "@management4 accept", "@management6 accept", "@resolver4 udp dport 53 accept", "@resolver6 tcp sport 53 accept", "@isolate4 counter drop", "@isolate6 counter drop", "ip6 hoplimit 255 icmpv6 type 135 accept"} {
		if !strings.Contains(text, required) {
			t.Fatalf("missing %s", required)
		}
	}
	for _, forbidden := range []string{"policy drop", "ct state", "flush ruleset", "192.168.0.0", "10.0.0.0/8", "169.254.0.0/16"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("broad isolation bypass: %s", forbidden)
		}
	}
	if strings.Index(text, "@management4 accept") > strings.Index(text, "@isolate4 counter drop") {
		t.Fatal("management exception follows deny")
	}
}

func TestNFTReadbackRejectsChangedRulesExpiredMembershipAndMissingTimeouts(t *testing.T) {
	c, f := newFakeNFT(t)
	p, _ := isolationPlan(c.owner, testIsolation(), f.now)
	good := nftPlanJSON(p, c.table, f.now)
	s, err := readSnapshot(good, c.table)
	if err != nil || s.verify(p) != nil {
		t.Fatalf("valid readback failed %v", err)
	}
	for _, data := range [][]byte{[]byte(strings.Replace(string(good), `"policy":"accept"`, `"policy":"drop"`, 1)), []byte(strings.Replace(string(good), `"drop":null`, `"accept":null`, 1)), []byte(strings.Replace(string(good), `"interval","timeout"`, `"interval"`, 1)), []byte(strings.ReplaceAll(string(good), `"expires":29900`, `"expires":0`))} {
		s, err := readSnapshot(data, c.table)
		if err == nil && s.verify(p) == nil {
			t.Fatal("unverified enforcement accepted")
		}
	}
}

func TestNFTConntrackNumericDirectionEquivalenceIsNarrow(t *testing.T) {
	left := map[string]any{"ct": map[string]any{"key": "direction"}}
	for n, direction := range []string{"original", "reply"} {
		if !bytes.Equal(canonicalRuleExpr([]any{matchExpr(left, n)}), canonicalRuleExpr([]any{matchExpr(left, direction)})) {
			t.Fatal("valid numeric conntrack direction rejected")
		}
	}
	if bytes.Equal(canonicalRuleExpr([]any{matchExpr(left, 2)}), canonicalRuleExpr([]any{matchExpr(left, "reply")})) {
		t.Fatal("unknown direction treated as reply")
	}
}

func TestNFTTransactionsRollbackAndPreserveForeignTables(t *testing.T) {
	for _, kind := range []string{"before", "after"} {
		t.Run(kind, func(t *testing.T) {
			c, f := newFakeNFT(t)
			p, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{netip.MustParseAddr("192.0.2.4"): "out"}, nil, f.now)
			f.register(p)
			if err := c.apply(context.Background(), p); err != nil {
				t.Fatal(err)
			}
			q, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{netip.MustParseAddr("192.0.2.5"): "both"}, nil, f.now)
			f.register(q)
			if kind == "before" {
				f.failBefore = 1
			} else {
				f.failAfter = 1
			}
			if err := c.apply(context.Background(), q); err == nil {
				t.Fatal("failed transaction claimed success")
			}
			if f.current.Comment != p.Comment || c.count() != p.Count || c.health() == nil {
				t.Fatal("previous verified policy or error lost")
			}
			if err := c.apply(context.Background(), q); err != nil || c.health() != nil {
				t.Fatalf("retry failed %v", err)
			}
			for _, script := range f.transactions {
				if strings.Contains(script, "flush ruleset") || strings.Contains(script, "delete table inet unrelated") {
					t.Fatal("collateral firewall mutation")
				}
			}
		})
	}
	c, f := newFakeNFT(t)
	f.foreign = true
	p, _ := isolationPlan(c.owner, testIsolation(), f.now)
	if err := c.apply(context.Background(), p); !errors.Is(err, ErrNetworkOwnership) || len(f.transactions) != 0 {
		t.Fatalf("foreign table touched: %v %v", err, f.transactions)
	}
}

func TestNFTFreshIsolatorRestoresPriorLeaseAndStatus(t *testing.T) {
	c, f := newFakeNFT(t)
	p, _ := isolationPlan(c.owner, testIsolation(), f.now)
	f.register(p)
	f.current = p
	h := &HostIsolator{controller: c}
	status, err := h.Status(context.Background())
	if err != nil || !status.Active || !status.Verified || len(status.Management) != 1 {
		t.Fatalf("status %+v %v", status, err)
	}
	qPolicy := testIsolation()
	qPolicy.Management = []netip.Addr{netip.MustParseAddr("198.51.100.8")}
	q, _ := isolationPlan(c.owner, qPolicy, f.now)
	f.register(q)
	f.failAfter = 1
	if _, err = h.Isolate(context.Background(), qPolicy); err == nil {
		t.Fatal("failed replacement claimed success")
	}
	if f.current.Comment != p.Comment || c.last == nil || c.last.Expires.After(p.Expires) {
		t.Fatal("old isolation was lost or its deadline renewed")
	}
	if err = h.Release(context.Background()); err != nil || f.current != nil {
		t.Fatalf("release failed %v", err)
	}
}

func TestNFTCleanupFailureIsRetainedAndRetryable(t *testing.T) {
	c, f := newFakeNFT(t)
	p, _ := isolationPlan(c.owner, testIsolation(), f.now)
	f.register(p)
	if err := c.apply(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	f.failDelete = 1
	if err := c.apply(context.Background(), nil); err == nil || c.last == nil || f.current == nil || c.health() == nil {
		t.Fatal("failed cleanup erased ownership/error")
	}
	if err := c.apply(context.Background(), nil); err != nil || c.last != nil || f.current != nil || c.health() != nil {
		t.Fatalf("cleanup retry failed %v", err)
	}
}

func TestNFTLeaseExpiryCannotReportActiveFilters(t *testing.T) {
	c, f := newFakeNFT(t)
	p, _ := blockPlan(c.owner, testSafety(), map[netip.Addr]string{netip.MustParseAddr("192.0.2.4"): "both"}, nil, f.now)
	f.register(p)
	if err := c.apply(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(nftLease + time.Second)
	if c.count() != 0 || c.health() == nil {
		t.Fatal("expired kernel lease claimed healthy active enforcement")
	}
	c2, g := newFakeNFT(t)
	iso, _ := isolationPlan(c2.owner, testIsolation(), g.now)
	g.register(iso)
	g.current = iso
	g.now = g.now.Add(time.Minute)
	status, err := (&HostIsolator{controller: c2}).Status(context.Background())
	if err != nil || status.Active || !status.Verified {
		t.Fatalf("expired isolation status %+v %v", status, err)
	}
}

type batchFakeBlocker struct {
	*fakeBlocker
	calls int
	fail  bool
}

func (b *batchFakeBlocker) ApplyDelta(add, remove []netip.Addr, addP, removeP []netip.Prefix, dir string) error {
	b.calls++
	if b.fail {
		return errors.New("injected batch failure")
	}
	for _, a := range add {
		b.set[a] = dir
	}
	for _, a := range remove {
		delete(b.set, a)
	}
	for _, p := range addP {
		b.pset[p] = dir
	}
	for _, p := range removeP {
		delete(b.pset, p)
	}
	return nil
}

func TestEnforcerBulkFeedHasOneTransactionAndRetainsReactiveOwnership(t *testing.T) {
	b := &batchFakeBlocker{fakeBlocker: newFake()}
	e := NewEnforcer(b, "both")
	var ips []netip.Addr
	for i := 0; i < 10000; i++ {
		ips = append(ips, netip.AddrFrom4([4]byte{192, byte(i / 256), byte(i % 256), 1}))
	}
	if err := e.ApplyPolicy(ips, []netip.Prefix{netip.MustParsePrefix("2001:db8::/32")}); err != nil || b.calls != 1 {
		t.Fatalf("bulk feed %d %v", b.calls, err)
	}
	retained := ips[0]
	e.retainIP = func(a netip.Addr) bool { return a == retained }
	if err := e.ApplyPolicy(ips[1:], nil); err != nil {
		t.Fatal(err)
	}
	if _, ok := b.set[retained]; !ok || e.OwnsIP(retained) {
		t.Fatal("reactive handoff removed its filter or retained feed claim")
	}
	b.fail = true
	if err := e.ApplyPolicy(nil, nil); err == nil || e.OwnsIP(ips[1]) {
		t.Fatal("failed bulk update claimed certain ownership")
	}
}
