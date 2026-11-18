//go:build linux || edr_nft_readback

package netblock

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"net/netip"
	"testing"
	"time"
)

// Embed the retained native wire capture so standalone guest test binaries do
// not depend on the source checkout or their current working directory.
//
//go:embed testdata/nft-nfproto-1.0.9.json
var nftObservedNumericFamilyJSON []byte

func TestNFTObservedNumericFamilyReadback(t *testing.T) {
	got, err := normalizeNFTReadback(nftObservedNumericFamilyJSON)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Objects []map[string]json.RawMessage `json:"nftables"`
	}
	if err := json.Unmarshal(got, &doc); err != nil {
		t.Fatal(err)
	}
	rules := 0
	for _, o := range doc.Objects {
		if o["rule"] == nil {
			continue
		}
		var r struct {
			Expr []any `json:"expr"`
		}
		if err := json.Unmarshal(o["rule"], &r); err != nil {
			t.Fatal(err)
		}
		family, src, dst := "ipv4", 67, 68
		if rules%2 == 1 {
			family, src, dst = "ipv6", 547, 546
		}
		if rules >= 2 {
			src, dst = dst, src
		}
		want := []any{matchExpr(metaExpr("nfproto"), family), matchExpr(payloadExpr("udp", "sport"), src), matchExpr(payloadExpr("udp", "dport"), dst), verdictExpr("accept")}
		if !bytes.Equal(canonical(r.Expr), canonical(want)) {
			t.Fatalf("observed rule %d not exactly normalized: %s", rules, canonical(r.Expr))
		}
		rules++
	}
	if rules != 4 {
		t.Fatalf("observed rules %d", rules)
	}
}

func TestNFTNumericFamilyRoundtripRetainsStrictPolicy(t *testing.T) {
	now := time.Unix(1700000000, 0)
	p, err := isolationPlan("test-owner", IsolationPolicy{Management: []netip.Addr{netip.MustParseAddr("198.51.100.7")}, Duration: 30 * time.Second}, now)
	if err != nil {
		t.Fatal(err)
	}
	raw := nftPlanWireJSON(t, p, "test-table", now)
	raw = bytes.ReplaceAll(raw, []byte(`"right":"ipv4"`), []byte(`"right":2`))
	raw = bytes.ReplaceAll(raw, []byte(`"right":"ipv6"`), []byte(`"right":10`))
	good, err := normalizeNFTReadback(raw)
	if err != nil {
		t.Fatal(err)
	}
	s, err := readSnapshot(good, "test-table")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.verify(p); err != nil {
		t.Fatal("numeric native policy refused", err)
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"wrong-family":      func(b []byte) []byte { return bytes.Replace(b, []byte(`"right":2`), []byte(`"right":10`), 1) },
		"unknown-family":    func(b []byte) []byte { return bytes.Replace(b, []byte(`"right":2`), []byte(`"right":3`), 1) },
		"fractional-family": func(b []byte) []byte { return bytes.Replace(b, []byte(`"right":2`), []byte(`"right":2.5`), 1) },
		"changed-predicate": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"key":"nfproto"`), []byte(`"key":"l4proto"`), 1)
		},
		"removed-predicate": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"left":{"meta":{"key":"nfproto"}},"op":"==","right":2`), []byte(`"left":{"meta":{"key":"nfproto"}},"op":"!=","right":2`), 1)
		},
	} {
		t.Run(name, func(t *testing.T) {
			bad := mutate(append([]byte(nil), raw...))
			if bytes.Equal(bad, raw) {
				t.Fatal("test mutation did not apply")
			}
			b, e := normalizeNFTReadback(bad)
			if e != nil {
				return
			}
			snapshot, e := readSnapshot(b, "test-table")
			if e == nil && snapshot.verify(p) == nil {
				t.Fatal("modified policy accepted")
			}
		})
	}
}

// Existing controller fakes deliberately use internal milliseconds. Only this
// wire fixture converts them to the observed upstream schema-1 seconds.
func nftPlanWireJSON(t *testing.T, p *nftPlan, table string, now time.Time) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(nftPlanJSON(p, table, now), &doc); err != nil {
		t.Fatal(err)
	}
	for _, object := range doc["nftables"].([]any) {
		set, ok := object.(map[string]any)["set"].(map[string]any)
		if !ok {
			continue
		}
		for _, element := range func() []any { e, _ := set["elem"].([]any); return e }() {
			e := element.(map[string]any)["elem"].(map[string]any)
			for _, key := range []string{"timeout", "expires"} {
				e[key] = int(e[key].(float64) / 1000)
			}
		}
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestNFTSchemaOneLeaseSecondsAndStrictWireBoundary(t *testing.T) {
	now := time.Unix(1700000000, 0)
	p, err := isolationPlan("test-owner", testIsolation(), now)
	if err != nil {
		t.Fatal(err)
	}
	raw := nftPlanWireJSON(t, p, "test-table", now)
	data, err := normalizeNFTReadback(raw)
	if err != nil {
		t.Fatal(err)
	}
	s, err := readSnapshot(data, "test-table")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.verify(p); err != nil {
		t.Fatal(err)
	}
	for name, set := range s.Sets {
		if len(set.Values) > 0 && set.MinExpires != 29*time.Second {
			t.Fatalf("%s deadline %s, want29s", name, set.MinExpires)
		}
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"unknown-schema": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"json_schema_version":1`), []byte(`"json_schema_version":2`), 1)
		},
		"missing-schema": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"json_schema_version":1`), []byte(`"unknown_schema":1`), 1)
		},
		"fractional-expiry": func(b []byte) []byte { return bytes.Replace(b, []byte(`"expires":29`), []byte(`"expires":29.5`), 1) },
		"negative-expiry":   func(b []byte) []byte { return bytes.Replace(b, []byte(`"expires":29`), []byte(`"expires":-1`), 1) },
		"overflow-expiry": func(b []byte) []byte {
			return bytes.Replace(b, []byte(`"expires":29`), []byte(`"expires":9223372036854775807`), 1)
		},
		"string-expiry": func(b []byte) []byte { return bytes.Replace(b, []byte(`"expires":29`), []byte(`"expires":"29"`), 1) },
		"trailing-json": func(b []byte) []byte { return append(b, []byte(`{}`)...) },
	} {
		t.Run(name, func(t *testing.T) {
			bad := mutate(append([]byte(nil), raw...))
			if bytes.Equal(bad, raw) {
				t.Fatal("mutation failed")
			}
			if _, err := normalizeNFTReadback(bad); err == nil {
				t.Fatal("malformed wire value accepted")
			}
		})
	}
	zero := bytes.ReplaceAll(raw, []byte(`"expires":29`), []byte(`"expires":0`))
	data, err = normalizeNFTReadback(zero)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = readSnapshot(data, "test-table"); err == nil {
		t.Fatal("expired lease accepted")
	}
}

func TestNFTWirePreservesUnknownFieldsAndExactNumbers(t *testing.T) {
	raw := []byte(`{"nftables":[{"metainfo":{"json_schema_version":1,"future":18446744073709551615}},{"set":{"timeout":30,"unknown":{"timeout":7},"elem":[{"elem":{"val":"198.51.100.7","timeout":30,"expires":29,"future":18446744073709551614}}]}},{"element":{"elem":[{"elem":{"timeout":30,"expires":28}}]}},{"rule":{"expr":[{"match":{"left":{"meta":{"key":"nfproto"}},"op":"==","right":2}},{"future":{"timeout":8}}]}}],"future":{"expires":9}}`)
	data, err := normalizeNFTReadback(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`"future":18446744073709551615`, `"future":18446744073709551614`, `"unknown":{"timeout":7}`, `"future":{"timeout":8}`, `"future":{"expires":9}`, `"timeout":30000`, `"expires":29000`, `"expires":28000`, `"right":"ipv4"`} {
		if !bytes.Contains(data, []byte(fragment)) {
			t.Fatalf("missing unchanged/normalized fragment %s in%s", fragment, data)
		}
	}
}
