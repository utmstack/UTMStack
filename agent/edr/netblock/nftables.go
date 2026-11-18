package netblock

// The nftables controller is deliberately independent of Linux syscalls. Tests
// exercise bounded transactions, ownership and readback without a host firewall.
import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	nftMaxEntries        = 262144
	nftMaxPayload        = 24 << 20
	nftMaxOutput         = 64 << 20
	nftLease             = 5 * time.Minute
	nftRefresh           = 30 * time.Second
	nftOperationTimeout  = 12 * time.Second
	MaxIsolationDuration = time.Hour
	MinIsolationDuration = 5 * time.Second
)

var (
	ErrNetworkSafety     = errors.New("network policy requires explicit management reachability")
	ErrNetworkOwnership  = errors.New("nftables table is not owned by this EDR instance")
	ErrNetworkUnverified = errors.New("nftables enforcement could not be verified")
	ErrNetworkBudget     = errors.New("network policy exceeds bounded resource budget")
)

// SafetyPolicy is mandatory before Linux blocklist enforcement. Management must
// contain positively resolved control-plane addresses; NeverBlock is the normal
// blocklist exemption policy, which is deliberately separate from isolation.
type SafetyPolicy struct {
	Management []netip.Addr
	Resolvers  []netip.Addr
	NeverBlock []netip.Prefix
}

// IsolationPolicy permits only exact management addresses, explicit DNS servers
// on TCP/UDP 53, Allow networks, loopback and narrowly scoped DHCP/IPv6 discovery.
// It does not inherit the blocklist's broad private-network exemptions.
type IsolationPolicy struct {
	Management []netip.Addr   `json:"management"`
	Resolvers  []netip.Addr   `json:"resolvers,omitempty"`
	Allow      []netip.Prefix `json:"allow,omitempty"`
	Duration   time.Duration  `json:"duration"`
}

type IsolationStatus struct {
	Supported  bool         `json:"supported"`
	Active     bool         `json:"active"`
	Verified   bool         `json:"verified"`
	ExpiresAt  time.Time    `json:"expires_at,omitempty"`
	Table      string       `json:"table,omitempty"`
	Management []netip.Addr `json:"management,omitempty"`
	Resolvers  []netip.Addr `json:"resolvers,omitempty"`
	// Allow contains the verified canonical intervals, which may merge adjacent
	// configured prefixes; it is coverage, not a copy of the input spelling.
	Allow []string `json:"allow,omitempty"`
}

type nftOperations interface {
	Run(context.Context, []byte, ...string) ([]byte, error)
}

type ipInterval struct{ First, Last netip.Addr }

func normalizedAddr(a netip.Addr) (netip.Addr, error) {
	a = a.Unmap()
	if !a.IsValid() || a.Zone() != "" || a.IsUnspecified() || a.IsMulticast() {
		return netip.Addr{}, fmt.Errorf("invalid network address %q", a)
	}
	return a, nil
}

func normalizedPrefix(p netip.Prefix) (netip.Prefix, error) {
	if !p.IsValid() || p.Addr().Zone() != "" {
		return netip.Prefix{}, errors.New("invalid network prefix")
	}
	if p.Addr().Is4In6() {
		if p.Bits() < 96 {
			return netip.Prefix{}, errors.New("ambiguous IPv4-mapped prefix")
		}
		p = netip.PrefixFrom(p.Addr().Unmap(), p.Bits()-96)
	}
	return p.Masked(), nil
}

func prefixInterval(p netip.Prefix) ipInterval {
	a := p.Masked().Addr()
	if a.Is4() {
		b := a.As4()
		for bit := p.Bits(); bit < 32; bit++ {
			b[bit/8] |= 1 << (7 - bit%8)
		}
		return ipInterval{a, netip.AddrFrom4(b)}
	}
	b := a.As16()
	for bit := p.Bits(); bit < 128; bit++ {
		b[bit/8] |= 1 << (7 - bit%8)
	}
	return ipInterval{a, netip.AddrFrom16(b)}
}

func mergeIntervals(in []ipInterval) []ipInterval {
	if len(in) == 0 {
		return nil
	}
	sort.Slice(in, func(i, j int) bool {
		if c := in[i].First.Compare(in[j].First); c != 0 {
			return c < 0
		}
		return in[i].Last.Compare(in[j].Last) < 0
	})
	out := make([]ipInterval, 0, len(in))
	for _, v := range in {
		if len(out) > 0 {
			prev := &out[len(out)-1]
			if prev.First.Is4() == v.First.Is4() && (v.First.Compare(prev.Last) <= 0 || prev.Last.Next() == v.First) {
				if v.Last.Compare(prev.Last) > 0 {
					prev.Last = v.Last
				}
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func intervals(addrs []netip.Addr, prefixes []netip.Prefix) ([]ipInterval, error) {
	if len(addrs)+len(prefixes) > nftMaxEntries {
		return nil, ErrNetworkBudget
	}
	all := make([]ipInterval, 0, len(addrs)+len(prefixes))
	for _, a := range addrs {
		a, err := normalizedAddr(a)
		if err != nil {
			return nil, err
		}
		all = append(all, ipInterval{a, a})
	}
	for _, p := range prefixes {
		p, err := normalizedPrefix(p)
		if err != nil {
			return nil, err
		}
		all = append(all, prefixInterval(p))
	}
	return mergeIntervals(all), nil
}

func managementValid(a []netip.Addr) error {
	if len(a) == 0 || len(a) > 64 {
		return ErrNetworkSafety
	}
	for _, ip := range a {
		ip, err := normalizedAddr(ip)
		if err != nil || !ip.IsGlobalUnicast() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return ErrNetworkSafety
		}
	}
	return nil
}

type nftSet struct {
	Name   string
	V4     bool
	Values []ipInterval
}
type nftRule struct {
	Chain, Text string
	Expr        []any
}
type nftPlan struct {
	Sets    []nftSet
	Rules   []nftRule
	Chains  []string
	Expires time.Time
	Comment string
	Count   int
}

func (p *nftPlan) addSets(name string, values []ipInterval) {
	for _, v4 := range []bool{true, false} {
		s := nftSet{Name: name + "6", V4: v4}
		if v4 {
			s.Name = name + "4"
		}
		for _, v := range values {
			if v.First.Is4() == v4 {
				s.Values = append(s.Values, v)
			}
		}
		p.Sets = append(p.Sets, s)
	}
}

func matchExpr(left any, right any) any {
	return map[string]any{"match": map[string]any{"op": "==", "left": left, "right": right}}
}
func payloadExpr(protocol, field string) any {
	return map[string]any{"payload": map[string]any{"protocol": protocol, "field": field}}
}
func metaExpr(key string) any  { return map[string]any{"meta": map[string]any{"key": key}} }
func verdictExpr(v string) any { return map[string]any{v: nil} }
func (p *nftPlan) addRule(chain, text string, expr ...any) {
	p.Rules = append(p.Rules, nftRule{chain, text, expr})
}

func (p *nftPlan) setRule(chain, base, verdict string, portProtocol string) {
	field := "daddr"
	port := "dport"
	if chain == "input" {
		field = "saddr"
		port = "sport"
	}
	for _, proto := range []string{"ip", "ip6"} {
		suffix := "4"
		if proto == "ip6" {
			suffix = "6"
		}
		t := proto + " " + field + " @" + base + suffix
		e := []any{matchExpr(payloadExpr(proto, field), "@"+base+suffix)}
		if portProtocol != "" {
			t += " " + portProtocol + " " + port + " 53"
			e = append(e, matchExpr(payloadExpr(portProtocol, port), 53))
		}
		if verdict == "drop" {
			t += " counter"
			e = append(e, map[string]any{"counter": nil})
		}
		p.addRule(chain, t+" "+verdict, append(e, verdictExpr(verdict))...)
	}
}

func (p *nftPlan) connectionDeny(chain, base, direction string) {
	field := "daddr"
	if chain == "input" {
		field = "saddr"
	}
	for _, proto := range []string{"ip", "ip6"} {
		suffix := "4"
		if proto == "ip6" {
			suffix = "6"
		}
		p.addRule(chain, "ct direction "+direction+" "+proto+" "+field+" @"+base+suffix+" counter drop", matchExpr(map[string]any{"ct": map[string]any{"key": "direction"}}, direction), matchExpr(payloadExpr(proto, field), "@"+base+suffix), map[string]any{"counter": nil}, verdictExpr("drop"))
	}
}

func (p *nftPlan) blockRules() {
	for _, chain := range p.Chains {
		p.setRule(chain, "safe", "accept", "")
		p.setRule(chain, "denyany", "drop", "")
		original, reply := "denyin", "denyout"
		if chain == "output" {
			original, reply = reply, original
		}
		p.connectionDeny(chain, original, "original")
		p.connectionDeny(chain, reply, "reply")
	}
}

func finishPlan(p *nftPlan, owner string) *nftPlan {
	// Identity binds actual desired objects, never a user-supplied nft expression.
	b, _ := json.Marshal(struct {
		Sets   []nftSet
		Rules  []nftRule
		Chains []string
	}{p.Sets, p.Rules, p.Chains})
	h := sha256.Sum256(b)
	p.Comment = owner + ":" + hex.EncodeToString(h[:16])
	return p
}

func blockPlan(owner string, safety SafetyPolicy, ips map[netip.Addr]string, prefixes map[netip.Prefix]string, now time.Time) (*nftPlan, error) {
	if err := managementValid(safety.Management); err != nil {
		return nil, err
	}
	if len(ips)+len(prefixes) > nftMaxEntries || len(safety.Resolvers) > 64 {
		return nil, ErrNetworkBudget
	}
	allow, err := intervals(append(append([]netip.Addr{}, safety.Management...), safety.Resolvers...), safety.NeverBlock)
	if err != nil {
		return nil, err
	}
	p := &nftPlan{Expires: now.Add(nftLease)}
	p.addSets("safe", allow)
	for _, dir := range []string{"in", "out", "both"} {
		var aa []netip.Addr
		var pp []netip.Prefix
		for a, d := range ips {
			if d == dir {
				aa = append(aa, a)
			}
		}
		for a, d := range prefixes {
			if d == dir {
				pp = append(pp, a)
			}
		}
		ranges, err := intervals(aa, pp)
		if err != nil {
			return nil, err
		}
		base := "deny" + dir
		if dir == "both" {
			base = "denyany"
		}
		p.addSets(base, ranges)
		if len(ranges) == 0 {
			continue
		}
		p.Count += len(ranges)
		if dir == "both" {
			p.Count += len(ranges)
		}
	}
	if p.Count > 0 {
		p.Chains = []string{"input", "output"}
		p.blockRules()
	}
	return finishPlan(p, owner), nil
}

func isolationPlan(owner string, policy IsolationPolicy, now time.Time) (*nftPlan, error) {
	if err := managementValid(policy.Management); err != nil {
		return nil, err
	}
	if policy.Duration < MinIsolationDuration || policy.Duration > MaxIsolationDuration {
		return nil, fmt.Errorf("isolation duration must be between %s and %s", MinIsolationDuration, MaxIsolationDuration)
	}
	if len(policy.Resolvers) > 64 || len(policy.Allow) > 128 {
		return nil, ErrNetworkBudget
	}
	for _, v := range policy.Allow {
		if !v.IsValid() || v.Bits() == 0 {
			return nil, fmt.Errorf("isolation does not permit a default-route exemption")
		}
	}
	mgmt, err := intervals(policy.Management, nil)
	if err != nil {
		return nil, err
	}
	resolver, err := intervals(policy.Resolvers, nil)
	if err != nil {
		return nil, err
	}
	allow, err := intervals(nil, policy.Allow)
	if err != nil {
		return nil, err
	}
	for _, v := range allow {
		if v == prefixInterval(netip.PrefixFrom(v.First, 0)) {
			return nil, errors.New("isolation exemptions cover an entire address family")
		}
	}
	p := &nftPlan{Expires: now.Add(policy.Duration), Chains: []string{"input", "output"}, Count: 4}
	p.addSets("management", mgmt)
	p.addSets("resolver", resolver)
	p.addSets("allow", allow)
	p.addSets("isolate", []ipInterval{prefixInterval(netip.MustParsePrefix("0.0.0.0/0")), prefixInterval(netip.MustParsePrefix("::/0"))})
	for _, chain := range p.Chains {
		iface := "iifname"
		if chain == "output" {
			iface = "oifname"
		}
		p.addRule(chain, iface+` "lo" accept`, matchExpr(metaExpr(iface), "lo"), verdictExpr("accept"))
		p.setRule(chain, "management", "accept", "")
		p.setRule(chain, "allow", "accept", "")
		p.setRule(chain, "resolver", "accept", "udp")
		p.setRule(chain, "resolver", "accept", "tcp")
		// ARP is outside inet; DHCP and IPv6 neighbor discovery keep the selected
		// control-plane routes usable without exempting a private subnet.
		// Input is server -> client; output is client -> server. DHCPv6
		// clients listen on 546 and servers on 547 (RFC 8415 section 7.2).
		// Pin each pair to its family so these ports do not exempt unrelated
		// IPv4/IPv6 datagrams from isolation.
		for _, dhcp := range []struct {
			family         string
			server, client int
		}{{"ipv4", 67, 68}, {"ipv6", 547, 546}} {
			src, dst := dhcp.server, dhcp.client
			if chain == "output" {
				src, dst = dst, src
			}
			p.addRule(chain, fmt.Sprintf("meta nfproto %s udp sport %d udp dport %d accept", dhcp.family, src, dst), matchExpr(metaExpr("nfproto"), dhcp.family), matchExpr(payloadExpr("udp", "sport"), src), matchExpr(payloadExpr("udp", "dport"), dst), verdictExpr("accept"))
		}
		for _, typ := range []int{133, 134, 135, 136} {
			p.addRule(chain, fmt.Sprintf("ip6 hoplimit 255 icmpv6 type %d accept", typ), matchExpr(payloadExpr("ip6", "hoplimit"), 255), matchExpr(payloadExpr("icmpv6", "type"), typ), verdictExpr("accept"))
		}
		p.setRule(chain, "isolate", "drop", "")
	}
	return finishPlan(p, owner), nil
}

// Text is generated exclusively from fixed identifiers, canonical numeric IPs,
// durations and fixed expressions. No shell or administrator expression parser
// is involved. Every deny membership has a kernel timeout; chain policy accepts.
func (p *nftPlan) script(table string, replace bool, now time.Time) ([]byte, error) {
	remaining := p.Expires.Sub(now)
	if remaining <= 0 {
		return nil, errors.New("network policy lease expired")
	}
	seconds := int64(remaining / time.Second)
	if seconds < 1 {
		return nil, errors.New("network policy lease nearly expired")
	}
	var b strings.Builder
	if replace {
		fmt.Fprintf(&b, "delete table inet %s\n", table)
	}
	fmt.Fprintf(&b, "table inet %s {\n comment %q\n", table, p.Comment)
	for _, s := range p.Sets {
		typ := "ipv6_addr"
		if s.V4 {
			typ = "ipv4_addr"
		}
		fmt.Fprintf(&b, " set %s { type %s; flags interval,timeout; timeout %ds; size %d;", s.Name, typ, seconds, nftMaxEntries)
		if len(s.Values) > 0 {
			b.WriteString(" elements = { ")
			for i, v := range s.Values {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(v.First.String())
				if v.Last != v.First {
					b.WriteByte('-')
					b.WriteString(v.Last.String())
				}
				fmt.Fprintf(&b, " timeout %ds", seconds)
			}
			b.WriteString(" };")
		}
		b.WriteString(" }\n")
	}
	for _, c := range p.Chains {
		fmt.Fprintf(&b, " chain %s { type filter hook %s priority 10; policy accept;\n", c, c)
		for i, r := range p.Rules {
			if r.Chain == c {
				fmt.Fprintf(&b, " %s comment \"edr-%d\"\n", r.Text, i)
			}
		}
		b.WriteString(" }\n")
	}
	b.WriteString("}\n")
	if b.Len() > nftMaxPayload {
		return nil, ErrNetworkBudget
	}
	return []byte(b.String()), nil
}

type nftSnapshot struct {
	Exists  bool
	Comment string
	Sets    map[string]nftReadSet
	Chains  map[string]map[string]any
	Rules   []map[string]any
}
type nftReadSet struct {
	V4         bool
	Values     []ipInterval
	MinExpires time.Duration
}

func decodeNFT(data []byte) ([]map[string]json.RawMessage, error) {
	if len(data) > nftMaxOutput {
		return nil, ErrNetworkBudget
	}
	var doc struct {
		Objects []map[string]json.RawMessage `json:"nftables"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("invalid nftables readback: %w", err)
	}
	if doc.Objects == nil {
		return nil, errors.New("missing nftables readback objects")
	}
	return doc.Objects, nil
}

func valueInterval(v any) (ipInterval, time.Duration, error) {
	var expiry time.Duration
	if m, ok := v.(map[string]any); ok {
		if e, ok := m["elem"].(map[string]any); ok {
			// libnftables emits per-element expires and timeout in milliseconds.
			ms, ok := e["expires"].(float64)
			if !ok || ms <= 0 || ms > float64((MaxIsolationDuration+nftLease)/time.Millisecond) {
				return ipInterval{}, 0, ErrNetworkUnverified
			}
			expiry = time.Duration(ms) * time.Millisecond
			v = e["val"]
		}
	}
	if s, ok := v.(string); ok {
		if a, err := netip.ParseAddr(s); err == nil {
			return ipInterval{a.Unmap(), a.Unmap()}, expiry, nil
		}
		if p, err := netip.ParsePrefix(s); err == nil {
			p, err = normalizedPrefix(p)
			if err == nil {
				return prefixInterval(p), expiry, nil
			}
		}
	}
	if m, ok := v.(map[string]any); ok {
		if p, ok := m["prefix"].(map[string]any); ok {
			a, e := netip.ParseAddr(fmt.Sprint(p["addr"]))
			n, ok := p["len"].(float64)
			if e == nil && ok && n == float64(int(n)) {
				if p, e := normalizedPrefix(netip.PrefixFrom(a, int(n))); e == nil {
					return prefixInterval(p), expiry, nil
				}
			}
		}
		if r, ok := m["range"].([]any); ok && len(r) == 2 {
			a, e1 := netip.ParseAddr(fmt.Sprint(r[0]))
			b, e2 := netip.ParseAddr(fmt.Sprint(r[1]))
			a = a.Unmap()
			b = b.Unmap()
			if e1 == nil && e2 == nil && a.Is4() == b.Is4() && a.Compare(b) <= 0 {
				return ipInterval{a, b}, expiry, nil
			}
		}
	}
	return ipInterval{}, 0, errors.New("unrecognized nftables set element")
}

func readSnapshot(data []byte, table string) (nftSnapshot, error) {
	s := nftSnapshot{Sets: map[string]nftReadSet{}, Chains: map[string]map[string]any{}}
	objects, err := decodeNFT(data)
	if err != nil {
		return s, err
	}
	for _, o := range objects {
		for kind, raw := range o {
			if kind == "metainfo" {
				continue
			}
			var v map[string]any
			if err := json.Unmarshal(raw, &v); err != nil {
				return s, err
			}
			if v["family"] != "inet" {
				return s, ErrNetworkUnverified
			}
			if kind == "table" {
				if v["name"] != table || s.Exists {
					return s, ErrNetworkUnverified
				}
				s.Exists = true
				s.Comment, _ = v["comment"].(string)
				continue
			}
			if v["table"] != table {
				return s, ErrNetworkUnverified
			}
			name, _ := v["name"].(string)
			switch kind {
			case "chain":
				if _, dup := s.Chains[name]; dup {
					return s, ErrNetworkUnverified
				}
				s.Chains[name] = v
			case "rule":
				s.Rules = append(s.Rules, v)
			case "set":
				if _, dup := s.Sets[name]; dup {
					return s, ErrNetworkUnverified
				}
				typ, _ := v["type"].(string)
				if typ != "ipv4_addr" && typ != "ipv6_addr" {
					return s, ErrNetworkUnverified
				}
				flags, _ := v["flags"].([]any)
				hasTimeout, hasInterval := false, false
				for _, f := range flags {
					hasTimeout = hasTimeout || f == "timeout"
					hasInterval = hasInterval || f == "interval"
				}
				if !hasTimeout || !hasInterval {
					return s, ErrNetworkUnverified
				}
				r := nftReadSet{V4: typ == "ipv4_addr"}
				vv, ok := v["elem"].([]any)
				if v["elem"] != nil && !ok {
					return s, ErrNetworkUnverified
				}
				if len(vv) > nftMaxEntries {
					return s, ErrNetworkBudget
				}
				for _, e := range vv {
					iv, expiry, err := valueInterval(e)
					if err != nil || expiry <= 0 || iv.First.Is4() != r.V4 {
						return s, ErrNetworkUnverified
					}
					r.Values = append(r.Values, iv)
					if r.MinExpires == 0 || expiry < r.MinExpires {
						r.MinExpires = expiry
					}
				}
				r.Values = mergeIntervals(r.Values)
				s.Sets[name] = r
			default:
				return s, ErrNetworkUnverified
			}
		}
	}
	return s, nil
}

func canonical(v any) []byte { b, _ := json.Marshal(v); return b }

func canonicalRuleExpr(v any) []byte {
	b := canonical(v)
	var expr []map[string]any
	if json.Unmarshal(b, &expr) != nil {
		return nil
	}
	for _, statement := range expr {
		if _, ok := statement["counter"]; ok {
			statement["counter"] = nil
		}
		if match, ok := statement["match"].(map[string]any); ok {
			if left, ok := match["left"].(map[string]any); ok {
				if ct, ok := left["ct"].(map[string]any); ok && ct["key"] == "direction" {
					// Linux UAPI IP_CT_DIR_ORIGINAL=0 / IP_CT_DIR_REPLY=1;
					// numeric readback and symbolic expressions mean the same.
					switch match["right"] {
					case float64(0):
						match["right"] = "original"
					case float64(1):
						match["right"] = "reply"
					}
				}
			}
		}
	}
	return canonical(expr)
}

func (s nftSnapshot) verify(p *nftPlan) error {
	if !s.Exists || s.Comment != p.Comment || len(s.Sets) != len(p.Sets) || len(s.Chains) != len(p.Chains) || len(s.Rules) != len(p.Rules) {
		return fmt.Errorf("%w: table/object identity mismatch (sets %d/%d chains %d/%d rules %d/%d)", ErrNetworkUnverified, len(s.Sets), len(p.Sets), len(s.Chains), len(p.Chains), len(s.Rules), len(p.Rules))
	}
	for _, want := range p.Sets {
		got, ok := s.Sets[want.Name]
		if !ok || got.V4 != want.V4 || !reflect.DeepEqual(got.Values, want.Values) {
			return fmt.Errorf("%w: set %s coverage mismatch", ErrNetworkUnverified, want.Name)
		}
	}
	for _, name := range p.Chains {
		c, ok := s.Chains[name]
		if !ok || c["type"] != "filter" || c["hook"] != name || c["policy"] != "accept" || c["prio"] != float64(10) {
			return fmt.Errorf("%w: chain %s configuration mismatch", ErrNetworkUnverified, name)
		}
	}
	for i, want := range p.Rules {
		got := s.Rules[i]
		if got["chain"] != want.Chain || got["comment"] != "edr-"+strconv.Itoa(i) || !bytes.Equal(canonicalRuleExpr(got["expr"]), canonicalRuleExpr(want.Expr)) {
			return fmt.Errorf("%w: ordered rule %d expression mismatch", ErrNetworkUnverified, i)
		}
	}
	return nil
}

// recoverPlan permits a fresh CLI/service process to roll back an already owned
// policy without renewing its original deadline. Unknown rule structures are
// never deleted or silently adopted merely because their name matches ours.
func recoverPlan(s nftSnapshot, owner string, now time.Time) (*nftPlan, error) {
	if !s.Exists {
		return nil, nil
	}
	var p *nftPlan
	if _, isolation := s.Sets["isolate4"]; isolation {
		p, _ = isolationPlan(owner, IsolationPolicy{Management: []netip.Addr{netip.MustParseAddr("192.0.2.1")}, Duration: MinIsolationDuration}, now)
	} else {
		p = &nftPlan{}
		for _, base := range []string{"safe", "denyin", "denyout", "denyany"} {
			p.addSets(base, nil)
		}
		for _, chain := range []string{"input", "output"} {
			if _, ok := s.Chains[chain]; ok {
				p.Chains = append(p.Chains, chain)
			}
		}
		p.blockRules()
	}
	min := time.Duration(0)
	anyValues := false
	for i := range p.Sets {
		got, ok := s.Sets[p.Sets[i].Name]
		if !ok {
			return nil, ErrNetworkUnverified
		}
		p.Sets[i].Values = got.Values
		if len(got.Values) > 0 {
			anyValues = true
			if min == 0 || got.MinExpires < min {
				min = got.MinExpires
			}
		}
		if strings.HasPrefix(p.Sets[i].Name, "deny") {
			p.Count += len(got.Values)
			if strings.HasPrefix(p.Sets[i].Name, "denyany") {
				p.Count += len(got.Values)
			}
		}
	}
	finishPlan(p, owner)
	if !anyValues {
		// Leases may all expire while the process is absent. Empty timed sets
		// plus a verified policy-accept rule structure are safe to remove.
		p.Comment = s.Comment
		if err := s.verify(p); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err := s.verify(p); err != nil {
		return nil, err
	}
	if min < time.Second {
		return nil, errors.New("existing network lease is nearly expired")
	}
	p.Expires = now.Add(min)
	return p, nil
}

type nftController struct {
	ops           nftOperations
	table, owner  string
	now           func() time.Time
	last          *nftPlan
	verifiedUntil time.Time
	err           error
	lock          func(context.Context) (func(), error)
}

func (c *nftController) snapshot(ctx context.Context) (nftSnapshot, error) {
	// Listing tables first distinguishes a missing owned table from permissions,
	// protocol failure or an nft executable error; those must never mean success.
	data, err := c.ops.Run(ctx, nil, "-j", "-n", "list", "tables")
	if err != nil {
		return nftSnapshot{}, err
	}
	objects, err := decodeNFT(data)
	if err != nil {
		return nftSnapshot{}, err
	}
	found := false
	for _, o := range objects {
		if raw, ok := o["table"]; ok {
			var t struct{ Family, Name string }
			if json.Unmarshal(raw, &t) != nil {
				return nftSnapshot{}, ErrNetworkUnverified
			}
			if t.Family == "inet" && t.Name == c.table {
				found = true
			}
		}
	}
	if !found {
		return nftSnapshot{}, nil
	}
	data, err = c.ops.Run(ctx, nil, "-j", "-n", "list", "table", "inet", c.table)
	if err != nil {
		return nftSnapshot{}, err
	}
	s, err := readSnapshot(data, c.table)
	if err != nil {
		return s, err
	}
	if !strings.HasPrefix(s.Comment, c.owner+":") {
		return s, ErrNetworkOwnership
	}
	return s, nil
}

func (c *nftController) transaction(ctx context.Context, p *nftPlan, exists bool) error {
	var script []byte
	var err error
	if p == nil {
		if !exists {
			return nil
		}
		script = []byte("delete table inet " + c.table + "\n")
	} else {
		script, err = p.script(c.table, exists, c.now())
		if err != nil {
			return err
		}
	}
	_, err = c.ops.Run(ctx, script, "-f", "-")
	if err != nil {
		return err
	}
	s, err := c.snapshot(ctx)
	if err != nil {
		return err
	}
	if p == nil {
		if s.Exists {
			return ErrNetworkUnverified
		}
		return nil
	}
	if err = s.verify(p); err != nil {
		return err
	}
	// Refuse a malformed/unsupported timeout unit instead of claiming a finite
	// lease that the kernel interpreted differently. A small allowance covers IO.
	max := p.Expires.Sub(c.now()) + 2*time.Second
	for _, set := range s.Sets {
		if len(set.Values) > 0 && (set.MinExpires <= 0 || set.MinExpires > max) {
			return ErrNetworkUnverified
		}
	}
	return nil
}

func (c *nftController) apply(ctx context.Context, p *nftPlan) error {
	ctx, cancel := context.WithTimeout(ctx, nftOperationTimeout)
	defer cancel()
	if c.lock != nil {
		unlock, err := c.lock(ctx)
		if err != nil {
			c.err = err
			return err
		}
		defer unlock()
	}
	s, err := c.snapshot(ctx)
	if err != nil {
		c.err = err
		c.verifiedUntil = time.Time{}
		return err
	}
	previous := c.last
	if previous == nil && s.Exists && p != nil {
		previous, err = recoverPlan(s, c.owner, c.now())
		if err != nil {
			c.err = err
			c.verifiedUntil = time.Time{}
			return err
		}
	}
	err = c.transaction(ctx, p, s.Exists)
	if err == nil {
		c.last = p
		c.err = nil
		c.verifiedUntil = time.Time{}
		if p != nil {
			c.verifiedUntil = p.Expires.Add(-time.Second)
		}
		return nil
	}
	// A subprocess deadline can happen after kernel commit. Inspect and restore
	// only this owned table; never assume a failed command did not take effect.
	rollbackCtx, done := context.WithTimeout(context.Background(), nftOperationTimeout)
	defer done()
	current, readErr := c.snapshot(rollbackCtx)
	if readErr == nil {
		if previous != nil && !previous.Expires.After(c.now().Add(time.Second)) {
			previous = nil
		}
		if restoreErr := c.transaction(rollbackCtx, previous, current.Exists); restoreErr != nil {
			readErr = restoreErr
		}
	}
	c.err = errors.Join(err, readErr)
	c.verifiedUntil = time.Time{}
	if readErr == nil {
		c.last = previous
		if previous != nil {
			c.verifiedUntil = previous.Expires.Add(-time.Second)
		}
	}
	return c.err
}

func (c *nftController) count() int {
	if c.last == nil || !c.now().Before(c.verifiedUntil) {
		return 0
	}
	return c.last.Count
}
func (c *nftController) health() error {
	if c.err != nil {
		return c.err
	}
	if c.last != nil && !c.now().Before(c.verifiedUntil) {
		return errors.New("nftables policy lease expired or has not been verified")
	}
	return nil
}

// HostIsolator shares no table with the continuously refreshed blocklist. Its
// finite kernel lease survives a CLI exit but cannot become permanent.
type HostIsolator struct {
	mu         sync.Mutex
	controller *nftController
}

func (h *HostIsolator) Isolate(ctx context.Context, policy IsolationPolicy) (IsolationStatus, error) {
	if h == nil || h.controller == nil {
		return IsolationStatus{}, ErrUnsupported
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	p, err := isolationPlan(h.controller.owner, policy, h.controller.now())
	if err != nil {
		return IsolationStatus{}, err
	}
	if err = h.controller.apply(ctx, p); err != nil {
		return IsolationStatus{}, err
	}
	return h.statusLocked(ctx)
}

func (h *HostIsolator) Release(ctx context.Context) error {
	if h == nil || h.controller == nil {
		return ErrUnsupported
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.controller.apply(ctx, nil)
}

func (h *HostIsolator) Status(ctx context.Context) (IsolationStatus, error) {
	if h == nil || h.controller == nil {
		return IsolationStatus{}, ErrUnsupported
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.statusLocked(ctx)
}

func (h *HostIsolator) statusLocked(ctx context.Context) (IsolationStatus, error) {
	c := h.controller
	result := IsolationStatus{Supported: true, Table: c.table}
	ctx, cancel := context.WithTimeout(ctx, nftOperationTimeout)
	defer cancel()
	if c.lock != nil {
		unlock, err := c.lock(ctx)
		if err != nil {
			return result, err
		}
		defer unlock()
	}
	s, err := c.snapshot(ctx)
	if err != nil {
		return result, err
	}
	if !s.Exists {
		result.Verified = true
		return result, nil
	}
	// Reconstruct expected expressions from verified set coverage, so a fresh CLI
	// process checks the actual rules rather than trusting a table comment.
	mgmt, ok4 := s.Sets["management4"]
	mgmt6, ok6 := s.Sets["management6"]
	deny, d4 := s.Sets["isolate4"]
	deny6, d6 := s.Sets["isolate6"]
	if !ok4 || !ok6 || !d4 || !d6 {
		return result, ErrNetworkUnverified
	}
	for _, v := range append(append([]ipInterval{}, mgmt.Values...), mgmt6.Values...) {
		// Adjacent management addresses may be merged into a range by nft; retain
		// the small bounded address list, never expand an arbitrary range.
		for a, n := v.First, 0; ; a, n = a.Next(), n+1 {
			if n >= 64 {
				return result, ErrNetworkBudget
			}
			result.Management = append(result.Management, a)
			if a == v.Last {
				break
			}
		}
	}
	min := time.Duration(0)
	for _, v := range []nftReadSet{deny, deny6} {
		if len(v.Values) > 0 && (min == 0 || v.MinExpires < min) {
			min = v.MinExpires
		}
	}
	if len(deny.Values) == 0 && len(deny6.Values) == 0 {
		// Expired deny memberships are harmless only with the original policy-
		// accept chain/rule structure. Verify it before reporting inactive.
		result.Active = false
	} else {
		if len(result.Management) == 0 || !reflect.DeepEqual(deny.Values, []ipInterval{prefixInterval(netip.MustParsePrefix("0.0.0.0/0"))}) || !reflect.DeepEqual(deny6.Values, []ipInterval{prefixInterval(netip.MustParsePrefix("::/0"))}) || min <= 0 || min > MaxIsolationDuration+time.Second {
			return result, ErrNetworkUnverified
		}
		result.Active = true
		result.ExpiresAt = c.now().Add(min)
	}
	// Rule structure is independent of the lease contents; regenerate with the
	// original set intervals to validate every current chain and ordered rule.
	probePolicy := IsolationPolicy{Management: []netip.Addr{netip.MustParseAddr("192.0.2.1")}, Duration: MinIsolationDuration}
	p, _ := isolationPlan(c.owner, probePolicy, c.now())
	for i := range p.Sets {
		got, ok := s.Sets[p.Sets[i].Name]
		if !ok {
			return result, ErrNetworkUnverified
		}
		p.Sets[i].Values = got.Values
	}
	if result.Active {
		finishPlan(p, c.owner)
		if p.Comment != s.Comment {
			return result, ErrNetworkUnverified
		}
	} else {
		p.Comment = s.Comment
	}
	if err = s.verify(p); err != nil {
		return result, err
	}
	for _, name := range []string{"resolver4", "resolver6"} {
		for _, v := range s.Sets[name].Values {
			for a, n := v.First, 0; ; a, n = a.Next(), n+1 {
				if n >= 64 {
					return result, ErrNetworkBudget
				}
				result.Resolvers = append(result.Resolvers, a)
				if a == v.Last {
					break
				}
			}
		}
	}
	for _, name := range []string{"allow4", "allow6"} {
		for _, v := range s.Sets[name].Values {
			text := v.First.String()
			if v.Last != v.First {
				text += "-" + v.Last.String()
			}
			result.Allow = append(result.Allow, text)
		}
	}
	result.Verified = true
	return result, nil
}
