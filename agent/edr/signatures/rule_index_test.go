package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

// brokenPE is an MZ object with an unmapped entry point. The reference rejects
// its executable header, so executable offsets have no matches.
func brokenPE() []byte {
	b := append(append([]byte(nil), testPE()[:800]...), bytes.Repeat([]byte{'.'}, 100)...)
	copy(b[512:], "....")
	binary.LittleEndian.PutUint32(b[168:], 0x3000)
	return b
}

// brokenPEWith is brokenPE followed by extra bytes.
func brokenPEWith(extra string) []byte { return append(brokenPE(), extra...) }

// A rule none of whose subsignatures occurs is still evaluated when its
// expression holds with every count zero, exactly as before: a count
// comparison below a bound, an absence test, and a bytecode trigger. A pattern
// that occurs nowhere counts 0 even at an executable offset on an unusable
// header. Rule order still decides which detection is reported.
func TestRulesTrueWithoutTheirSubsignaturesStillApply(t *testing.T) {
	pe := testPE() // "ABCD" at the entry point
	for _, tt := range []struct {
		ext, rules string
		data       []byte
		want       string
		incomplete bool
	}{
		{"ldb", "Z;Target:0;0<3;41424344", []byte("xyz"), "Z", false},
		{"ldb", "Z;Target:0;0<3;41424344", []byte("ABCDABCDABCD"), "", false},
		{"ldb", "Z;Target:0;0|1=0;41424344;45464748", []byte("xyz"), "Z", false},
		{"ldb", "Z;Target:0;(0|1)<1;41424344;45464748", []byte("xyz"), "Z", false},
		{"ldb", "Z;Target:0;(0|1)<1;41424344;45464748", []byte("xEFGH"), "", false},
		{"ldb", "Z;Target:0;0=0&1=0;41424344;45464748", []byte(""), "Z", false},
		{"ldb", "Z;Target:0;0=0&1;41424344;45464748", []byte("xyz"), "", false},
		{"ldb", "Z;Engine:51-255,Target:0,FileSize:5-10;0<1;41424344", []byte("xyz"), "", false},
		{"ldb", "Z;Engine:51-255,Target:0,FileSize:1-10;0<1;41424344", []byte("xyz"), "Z", false},
		// Target 1: skipped on non-MZ objects, evaluated on every MZ object.
		{"ldb", "Z;Target:1;0=0;EP+0:57575757", []byte("xyz"), "", false},
		{"ldb", "Z;Target:1;0=0;EP+0:57575757", pe, "Z", false},
		{"ldb", "Z;Target:1;0=0;EP+0:41424344", pe, "", false},
		{"ldb", "Z;Target:1;0=0;EP+0:57575757", brokenPE(), "Z", false},
		{"ldb", "Z;Target:1;0=0;EP+0:57575757", brokenPEWith("WWWW"), "Z", false},
		{"ldb", "Z;Target:1;0<1;57575757", brokenPE(), "Z", false},
		{"ldb", "Z;Engine:51-255,Target:1,NumberOfSections:1-1;0<1;57575757", pe, "Z", false},
		{"ldb", "Z;Engine:51-255,Target:1,NumberOfSections:2-3;0<1;57575757", pe, "", false},
		{"ldb", "Z;Engine:51-255,Target:1,NumberOfSections:1-1;0<1;57575757", brokenPE(), "", false},
		{"ndb", "N:1:EP+0:57575757", brokenPE(), "", false},
		{"ndb", "N:1:EP+0:57575757", brokenPEWith("WWWW"), "", false},
		{"ndb", "N:1:EP+0:57575757", pe, "", false},
		{"ndb", "N:1:EP+0:57575757", []byte("xyz"), "", false},
		// Order between rules reached through their subsignatures and rules
		// true without them.
		{"ldb", "First;Target:0;0;78797a\nSecond;Target:0;0<1;41424344", []byte("xyz"), "First", false},
		{"ldb", "First;Target:0;0<1;41424344\nSecond;Target:0;0;78797a", []byte("xyz"), "First", false},
		{"ldb", "Unknown;Target:1;0;EP+0:57575757\nLater;Target:1;0<1;41424344", brokenPE(), "Later", false},
		{"ldb", "Unknown;Target:1;0;EP+0:57575757\nLater;Target:1;0<1;2e2e2e2e", brokenPEWith("WWWW"), "", false},
		{"ldb", "Unknown;Target:1;0;EP+0:57575757\nLater;Target:1;0<1;2e2e2e2e", brokenPE(), "", false},
		// A rejected header makes header-relative patterns count zero even
		// when their bytes occur elsewhere in the file.
		{"ldb", "A;Engine:51-255,Target:1,FileSize:1-10;0;EP+0:57575757", brokenPE(), "", false},
		{"ldb", "A;Engine:51-255,Target:1,FileSize:1-10;0;EP+0:57575757\nB;Engine:51-255,Target:1,FileSize:1-1000;0;EP+0:58585858", brokenPEWith("WWWWXXXX"), "", false},
		{"ldb", "A;Engine:51-255,Target:1,FileSize:1-10;0|1;EP+0:57575757;2e2e2e2e\nB;Engine:51-255,Target:1,FileSize:1-1000;0;EP+0:58585858", brokenPEWith("XXXX"), "", false},
		{"ldb", "A;Engine:51-255,Target:1;0|1;EP+0:57575757;2e2e2e2e", brokenPE(), "A", false},
		// A failed header keeps its executable offsets absent regardless of
		// whether the bytes appear outside those offsets.
		{"ldb", "A;Engine:51-255,Target:1;0&1<1;EP+0:57575757;2e2e2e2e\nB;Engine:51-255,Target:1;0;EP+0:58585858", brokenPEWith("XXXX"), "", false},
		{"ldb", "A;Engine:51-255,Target:1;0&1<1;EP+0:57575757;2e2e2e2e\nB;Engine:51-255,Target:1;0;EP+0:58585858", brokenPE(), "", false},
		{"ldb", "A;Engine:51-255,Target:1;0&1<1;EP+0:57575757;2e2e2e2e", brokenPE(), "", false},
	} {
		d := compileTest(t, tt.ext, tt.rules)
		name, _, err := d.matchCounted(context.Background(), tt.data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), tt.data, 1<<40)
		if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%s on %q: Match = %q, %v; previous matcher %q, %v", tt.rules, tt.data, name, err, oldName, oldErr)
		}
		if name != tt.want || errors.Is(err, ErrRuleIncomplete) != tt.incomplete || (err != nil) != tt.incomplete {
			t.Fatalf("%s on %q: Match = %q, %v; want %q, incomplete %v", tt.rules, tt.data, name, err, tt.want, tt.incomplete)
		}
	}
	// A bytecode trigger that holds with no occurrence runs its program.
	f := callbackWithLabel([]byte("UTMStack.Zero.Trigger"), true)
	f.logical = "Callback.Trigger;Engine:51-255,Target:0;0<1;4142"
	d := compileCallback(t, f)
	for data, want := range map[string]string{"xyz": "UTMStack.Zero.Trigger", "AB": ""} {
		name, _, err := d.matchCounted(context.Background(), []byte(data), 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), []byte(data), 1<<40)
		if name != want || err != nil || oldName != want || oldErr != nil {
			t.Fatalf("bytecode trigger on %q: Match = %q, %v; previous %q, %v; want %q", data, name, err, oldName, oldErr, want)
		}
	}
}

// A subsignature without an anchor literal is a candidate on every object, so
// a rule reached only through it is still evaluated everywhere.
func TestRulesReachedOnlyThroughUnanchoredSubsignatures(t *testing.T) {
	d := compileTest(t, "ldb", "Nibble;Target:0;0;f?f?1?\nNibbles;Target:0;0>1;f?f?1?\nNibblePE;Target:1;0;e?e?2?")
	if len(d.unanchored) != 2 {
		t.Fatalf("%d unanchored subsignatures, want 2", len(d.unanchored))
	}
	for _, tt := range []struct {
		data []byte
		want string
	}{
		{[]byte("xx\xf1\xff\x10yy"), "Nibble"},
		{append(testPE(), "\xe3\xee\x21"...), "NibblePE"},
		{[]byte("xx\xf1\xff\x20yy"), ""},
		{brokenPE(), ""},
	} {
		name, _, err := d.matchCounted(context.Background(), tt.data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), tt.data, 1<<40)
		if name != tt.want || err != nil || oldName != name || oldErr != nil {
			t.Fatalf("%q: Match = %q, %v; previous matcher %q, %v; want %q", tt.data, name, err, oldName, oldErr, tt.want)
		}
	}
	d = compileTest(t, "ldb", "Nibble;Target:0;0;f?f?1?\nNibbles;Target:0;0>1;f?f?1?")
	if name, err := d.Match(context.Background(), []byte("\xf1\xff\x10\xf2\xfe\x11")); name != "Nibble" || err != nil {
		t.Fatalf("Match = %q, %v; want Nibble", name, err)
	}
	d = compileTest(t, "ldb", "Nibbles;Target:0;0>1;f?f?1?\nNibble;Target:0;0;f?f?1?")
	if name, err := d.Match(context.Background(), []byte("\xf1\xff\x10\xf2\xfe\x11")); name != "Nibbles" || err != nil {
		t.Fatalf("Match = %q, %v; want Nibbles", name, err)
	}
}

// Only the first cause of an incomplete result is reported. PE-relative rules
// on a rejected header no longer contribute a cause before an undecidable
// state-limited pattern.
func TestFirstIncompleteCauseFollowsRuleOrder(t *testing.T) {
	defer func(states, forward int) { maxPatternStates, maxForwardSteps = states, forward }(maxPatternStates, maxForwardSteps)
	maxPatternStates, maxForwardSteps = 1, 1
	for _, tt := range []struct {
		rules string
		want  error
	}{
		{"A;Engine:51-255,Target:1;0;EP+0:57575757\nB;Target:1;0;2e*2e*58\nC;Engine:51-255,Target:1;0;EP+0:59595959", errPatternStates},
		{"B;Target:1;0;2e*2e*58\nA;Engine:51-255,Target:1;0;EP+0:57575757\nC;Engine:51-255,Target:1;0;EP+0:59595959", errPatternStates},
		// The typed root is evaluated before root 0, whatever the load order.
		{"B;Target:0;0;2e*2e*58\nA;Engine:51-255,Target:1;0;EP+0:57575757\nC;Engine:51-255,Target:1;0;EP+0:59595959", errPatternStates},
	} {
		d := compileTest(t, "ldb", tt.rules)
		data := brokenPEWith("WWWWYYYY")
		name, _, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, _, oldErr := newLegacyMatcher(d).match(context.Background(), data, 1<<40)
		if name != "" || !errors.Is(err, ErrRuleIncomplete) || !errors.Is(err, tt.want) || oldName != "" || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%s: Match = %q, %v; previous matcher %q, %v; want cause %v", tt.rules, name, err, oldName, oldErr, tt.want)
		}
	}
}

// A small object must not pay for every rule of a large database: only rules
// with a subsignature that occurs in it, and rules true without any, are
// evaluated (and charged).
func TestSmallObjectEvaluatesOnlyReachableRules(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	var ndb, ldb strings.Builder
	const rules = 30000
	for i := 0; i < rules; i++ {
		b := make([]byte, 6)
		r.Read(b)
		fmt.Fprintf(&ndb, "Literal.%d:0:*:%s\n", i, hex.EncodeToString(b))
		fmt.Fprintf(&ldb, "Pair.%d;Target:1;0&1;%s;EP+0:%s\n", i, hex.EncodeToString(b[:4]), hex.EncodeToString(b[2:]))
		// Unknown on an MZ object whose header cannot be modeled.
		fmt.Fprintf(&ndb, "Entry.%d:1:EP+0:%s\n", i, hex.EncodeToString(b[1:5]))
	}
	ldb.WriteString("Zero.Late;Target:0;0<2;7a7a7a7a7a7a\n")
	d, err := Compile(map[string][]byte{"a.ndb": []byte(ndb.String()), "b.ldb": []byte(ldb.String())})
	if err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{[]byte("a small clean object"), testPE(), brokenPE()} {
		name, steps, err := d.matchCounted(context.Background(), data, 1<<40)
		oldName, oldSteps, oldErr := newLegacyMatcher(d).match(context.Background(), data, 1<<40)
		if name != oldName || fmt.Sprint(err) != fmt.Sprint(oldErr) {
			t.Fatalf("%q: Match = %q, %v; previous matcher %q, %v", data[:8], name, err, oldName, oldErr)
		}
		t.Logf("%d bytes: %d steps, previous matcher %d", len(data), steps, oldSteps)
		if name != "Zero.Late" || err != nil {
			t.Fatalf("%q: Match = %q, %v; want Zero.Late", data[:8], name, err)
		}
		// The anchor pass costs about one step per byte here.
		if steps > 2000+4*int64(len(data)) {
			t.Fatalf("%d-byte object: %d steps for %d rules", len(data), steps, 3*rules+1)
		}
	}
}

// BenchmarkMatchSmallObject measures the fixed per-object cost with a large
// rule set: a 100-byte object contains almost none of its subsignatures.
func BenchmarkMatchSmallObject(b *testing.B) {
	d, _ := compileSyntheticFeed(b)
	object := []byte(strings.Repeat("small text object ", 6)[:100])
	b.ResetTimer()
	var steps int64
	for i := 0; i < b.N; i++ {
		_, n, err := d.matchCounted(context.Background(), object, 1<<40)
		if err != nil {
			b.Fatal(err)
		}
		steps = n
	}
	b.ReportMetric(float64(steps), "steps/op")
}
