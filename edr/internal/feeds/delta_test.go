// edr/internal/feeds/delta_test.go
package feeds

import (
	"bufio"
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestDiffAddsOnly(t *testing.T) {
	adds, dels := Diff([]string{"a"}, []string{"a", "b"})
	if !reflect.DeepEqual(adds, []string{"b"}) {
		t.Fatalf("adds = %v, want [b]", adds)
	}
	if len(dels) != 0 {
		t.Fatalf("dels = %v, want empty", dels)
	}
}

func TestDiffDelsOnly(t *testing.T) {
	adds, dels := Diff([]string{"a", "b"}, []string{"a"})
	if len(adds) != 0 {
		t.Fatalf("adds = %v, want empty", adds)
	}
	if !reflect.DeepEqual(dels, []string{"b"}) {
		t.Fatalf("dels = %v, want [b]", dels)
	}
}

func TestDiffBothSorted(t *testing.T) {
	// Unsorted inputs must yield sorted outputs.
	adds, dels := Diff([]string{"m", "z", "a"}, []string{"a", "y", "b"})
	if !reflect.DeepEqual(adds, []string{"b", "y"}) {
		t.Fatalf("adds = %v, want [b y]", adds)
	}
	if !reflect.DeepEqual(dels, []string{"m", "z"}) {
		t.Fatalf("dels = %v, want [m z]", dels)
	}
}

func TestDiffEmpty(t *testing.T) {
	adds, dels := Diff(nil, nil)
	if len(adds) != 0 || len(dels) != 0 {
		t.Fatalf("adds=%v dels=%v, want empty", adds, dels)
	}
}

type dailyLine struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Op    string `json:"op"`
}

func parseDaily(t *testing.T, b []byte) []dailyLine {
	t.Helper()
	var out []dailyLine
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var rec dailyLine
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("line %q is not valid JSON: %v", line, err)
		}
		out = append(out, rec)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestRenderDailyAddsThenDelsDeterministicOrder(t *testing.T) {
	// Pass adds/dels intentionally unsorted; RenderDaily must emit adds first
	// (sorted) then dels (sorted).
	b := RenderDaily([]string{"2.2.2.2", "1.1.1.1"}, []string{"9.9.9.9", "5.5.5.5"}, "ip")
	recs := parseDaily(t, b)
	want := []dailyLine{
		{"1.1.1.1", "ip", "add"},
		{"2.2.2.2", "ip", "add"},
		{"5.5.5.5", "ip", "del"},
		{"9.9.9.9", "ip", "del"},
	}
	if !reflect.DeepEqual(recs, want) {
		t.Fatalf("records = %+v, want %+v", recs, want)
	}
}

func TestRenderDailyIPvsCIDRType(t *testing.T) {
	b := RenderDaily([]string{"10.0.0.0/8", "1.1.1.1"}, nil, "ip")
	recs := parseDaily(t, b)
	byVal := map[string]string{}
	for _, r := range recs {
		byVal[r.Value] = r.Type
	}
	if byVal["1.1.1.1"] != "ip" {
		t.Fatalf("1.1.1.1 type = %q, want ip", byVal["1.1.1.1"])
	}
	if byVal["10.0.0.0/8"] != "cidr" {
		t.Fatalf("10.0.0.0/8 type = %q, want cidr", byVal["10.0.0.0/8"])
	}
}

func TestRenderDailyDomainAndHostnameType(t *testing.T) {
	d := parseDaily(t, RenderDaily([]string{"a.example.com"}, nil, "domain"))
	if len(d) != 1 || d[0].Type != "domain" {
		t.Fatalf("domain type = %+v, want type domain", d)
	}
	h := parseDaily(t, RenderDaily([]string{"host.local"}, nil, "hostname"))
	if len(h) != 1 || h[0].Type != "hostname" {
		t.Fatalf("hostname type = %+v, want type hostname", h)
	}
}

func TestRenderDailyEmptyProducesNoLines(t *testing.T) {
	b := RenderDaily(nil, nil, "ip")
	if len(parseDaily(t, b)) != 0 {
		t.Fatalf("empty diff must produce no lines, got %q", b)
	}
}
