package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestIdenticalSubsignaturesShareOnePattern(t *testing.T) {
	var ldb strings.Builder
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&ldb, "Shared.%d;Engine:81-255,Target:0;0&1;4d535642564d{2}2e444c4c::i;%x\n", i, fmt.Sprintf("marker%04d", i))
	}
	ldb.WriteString("Other.Offset;Engine:81-255,Target:0;0;1:4d535642564d{2}2e444c4c::i\n") // a different offset is a different pattern
	ldb.WriteString("Other.Case;Engine:81-255,Target:0;0;4d535642564d{2}2e444c4c\n")
	d, err := Compile(map[string][]byte{"a.ldb": []byte(ldb.String()), "b.ndb": []byte("Extended:0:*:4d535642564d{2}2e444c4c\n")})
	if err != nil {
		t.Fatal(err)
	}
	if n := rulePatterns(d); n != 200+3 {
		t.Fatalf("%d compiled rule patterns, want 203", n)
	}
	if d.patternIDs != nil {
		t.Fatal("the compile-time pattern index must be released")
	}
	assertMatch(t, d, []byte("xx msvbvm60.dll marker0007 yy"), "Shared.7")
	// The shared standalone AC body precedes logical-expression evaluation.
	assertMatch(t, d, []byte("MSVBVM60.DLL"), "Extended")
	assertMatch(t, d, []byte(" msvbvm60.dll"), "Other.Offset")
}

func TestPatternKeyStorageAcrossLocalCapacityAndLongBody(t *testing.T) {
	longBody := strings.Repeat("41", 300)
	d := compileTest(t, "ldb", "Five;Engine:51-255,Target:0;0&1&2&3&4;41;42;43;44;45\n"+
		"Long;Engine:51-255,Target:0;0;"+longBody+"\n")
	assertMatch(t, d, []byte("ABCDE"), "Five")
	assertMatch(t, d, bytes.Repeat([]byte("A"), 300), "Long")
}

func formatAnchors(anchors []patternAnchor) string {
	var got []string
	for _, a := range anchors {
		got = append(got, fmt.Sprintf("{%s %v %d %d}", strings.ReplaceAll(string(a.literal), "\x00", `\x00`), a.fold, a.minPre, a.maxPre))
	}
	return "[" + strings.Join(got, " ") + "]"
}

// longestLiteralAnchors applies the longest-literal choice, which patterns a
// state limit can reach keep, to a compiled pattern.
func longestLiteralAnchors(p *bodyPattern) []patternAnchor {
	anchors := anchorsFor(topLevelTokens(p))
	for i, a := range anchors {
		if p.nocase && bytes.ContainsFunc(a.literal, func(r rune) bool { return r < 0x80 && asciiLetter(byte(r)) }) {
			anchors[i].literal, anchors[i].fold = asciiLower(a.literal), true
		}
	}
	return anchors
}

func TestCaseInsensitiveLiteralsAreAnchors(t *testing.T) {
	for body, want := range map[string]string{
		"4d535642564d{2}2e444c4c::i": `[{msvbvm true 0 0}]`, // folded, not dropped
		"2e2e2e{2}41::i":             `[{... false 0 0}]`,   // no letters: a raw anchor
		"41424344::wa":               `[{ABCD false 0 0} {A\x00B\x00C\x00D\x00 false 0 0}]`,
		"41{2-5}4243444546":          `[{BCDEF false 3 6}]`,
		"41*4243444546":              `[{BCDEF false 1 -1}]`,
		"41(4243|444546)":            `[{BC false 1 1} {DEF false 1 1}]`,
		"4142(43|4445)":              `[{AB false 0 0}]`,              // a longer top-level literal wins
		"(41|42??)43":                `[{A false 0 0} {B false 0 0}]`, // ties keep the earlier set
	} {
		p, err := compileBody("*", body)
		if err != nil {
			t.Fatal(err)
		}
		if got := formatAnchors(longestLiteralAnchors(&p)); got != want {
			t.Errorf("%s: longest-literal anchors %v, want %s", body, got, want)
		}
	}
	d := compileTest(t, "ldb", "Wide;Target:0;0;4142::wi\nFold;Target:0;0;4d5a{1}4142::i\nBoth;Target:0;0;787978::wa")
	for data, want := range map[string]string{"a\x00B\x00": "Wide", "mZ.aB": "Fold", "MZ.AB": "Fold", "M Z.AB": "", "x\x00y\x00x\x00": "Both", "xyx": "Both"} {
		assertMatch(t, d, []byte(data), want)
	}
}

func TestFoldedAnchorKeyVariants(t *testing.T) {
	for n := 1; n <= 4; n++ {
		literal := []byte("ABCD"[:n])
		d := compileTest(t, "ldb", fmt.Sprintf("Fold;Target:0;0;%s::i\n", hex.EncodeToString(literal)))
		table := d.index.long
		if n < 4 {
			table = d.index.short[n-1]
		}
		for mask := 0; mask < 1<<n; mask++ {
			spelling := bytes.Clone(literal)
			for i := range spelling {
				if mask&(1<<i) != 0 {
					spelling[i] += 'a' - 'A'
				}
			}
			if table[packKey(spelling)] == nil {
				t.Errorf("%d-byte folded key %q was not indexed", n, spelling)
			}
			assertMatch(t, d, spelling, "Fold")
		}
	}
}

// A pattern no state limit can reach is indexed by the literal expected to
// occur least, weighed by the starts each occurrence leaves: padding runs and
// common opcode bytes lose to rarer literals. A pattern the limit can reach
// keeps its longest literal.
func TestAnchorsPreferRareLiterals(t *testing.T) {
	for body, want := range map[string]string{
		"(41|42??)43": `[{C false 1 2}]`,
		"e8??000000{-100}8602(f9|fc|66|eb){-15}8602c3": `[{\x86\x02\xc3 false 8 123}]`,
		"00000000000000000000{2}8b45fc":                `[{\x8bE\xfc false 12 12}]`,
		"ffffffffffffffff{4}4d5a9000":                  `[{MZ\x90\x00 false 12 12}]`,
		"9090909090909090{-20}7061796d656e74::i":       `[{payment true 8 28}]`,
		"4d535642564d{2}2e444c4c::i":                   `[{msvbvm true 0 0}]`,
		"0000000000000000*8b45fc":                      `[{\x00\x00\x00\x00\x00\x00\x00\x00 false 0 0}]`, // limitable
	} {
		p, err := compileBody("*", body)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.NewReplacer("\x86", `\x86`, "\x02", `\x02`, "\xc3", `\xc3`, "\x8b", `\x8b`, "\xfc", `\xfc`, "\x90", `\x90`).Replace(formatAnchors(p.graph.anchors))
		if got != want {
			t.Errorf("%s: anchors %s, want %s", body, got, want)
		}
	}
	if at := keyWindow([]byte("\x00\x00\x00\x00\x8b\x45\xfc"), false); at != 3 {
		t.Errorf("key window of a zero run and a rare tail starts at %d, want 3", at)
	}
	d := compileTest(t, "ldb", "Rare;Target:0;0;00000000000000000000{2}8b45fc\nFolded;Target:0;0;9090909090909090{-20}7061796d656e74::i")
	data := make([]byte, 4096)
	assertMatch(t, d, data, "")
	copy(data[1000:], "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00..\x8b\x45\xfc")
	assertMatch(t, d, data, "Rare")
	copy(data[1000:], "\x90\x90\x90\x90\x90\x90\x90\x90 xx PayMent")
	assertMatch(t, d, data, "Folded")
}

// The reference feed repeats one case-insensitive subsignature about 107,000
// times. Shared and anchored, it costs one scan per object, not one per rule.
func TestRepeatedCaseInsensitiveSubsignatureIsCountedOnce(t *testing.T) {
	if testing.Short() {
		t.Skip("large rule set")
	}
	const rules = 107425
	r := rand.New(rand.NewSource(7))
	markers := make([][]byte, rules)
	var ldb strings.Builder
	for i := range markers {
		markers[i] = make([]byte, 16)
		r.Read(markers[i])
		markers[i][0] = 0xfe // absent from the scanned data, so no marker is a candidate
		fmt.Fprintf(&ldb, "Win.Trojan.VB-%d;Engine:81-255,Target:1;0&1;4d535642564d{2}2e444c4c::i;%x\n", i, markers[i])
	}
	d, err := CompileWithLimits(map[string][]byte{"daily.ldb": []byte(ldb.String())}, Limits{MaxMatchSteps: 20_000_000})
	if err != nil {
		t.Fatal(err)
	}
	if n := rulePatterns(d); n != rules+1 {
		t.Fatalf("%d rule patterns, want %d", n, rules+1)
	}
	pe := make([]byte, 2<<20)
	r.Read(pe)
	for i := range pe {
		if pe[i] == 'm' || pe[i] == 'M' || pe[i] == 0xfe {
			pe[i] = '.' // many occurrences of the ".dll" run, none of "msvbvm"
		}
	}
	copy(pe, testPE())
	if got, err := d.Match(context.Background(), pe); err != nil || got != "" {
		t.Fatalf("Match = %q, %v", got, err)
	}
	copy(pe[4096:], "MSVBVM60.DLL")
	copy(pe[8192:], markers[42])
	if got, err := d.Match(context.Background(), pe); err != nil || got != "Win.Trojan.VB-42" {
		t.Fatalf("Match = %q, %v", got, err)
	}
}

// oracleMatch evaluates logical rules the way Match did before patterns were
// shared or case-insensitive literals indexed: every rule counts each of its
// subsignatures from scratch, at every start position, with its own cap.
func oracleMatch(lines []string, data []byte) (string, error) {
	w := &matchWork{ctx: context.Background(), limit: 1 << 40}
	var pending error
	for _, line := range lines {
		p := strings.Split(line, ";")
		expr, err := parseExpression(p[2], len(p)-3)
		if err != nil {
			return "", err
		}
		counts := make([]int, len(p)-3)
		for i, sub := range p[3:] {
			off, body := "*", sub
			if j := strings.IndexByte(sub, ':'); j >= 0 && (j+1 == len(sub) || sub[j+1] != ':') {
				off, body = sub[:j], sub[j+1:]
			}
			pat, err := compileBody(off, body)
			if err != nil {
				return "", err
			}
			limit := expr.countLimit(i)
			lo, hi, ok := pat.offset.bounds(len(data), nil)
			hi = min(hi, len(data)-pat.minWidth())
			for start := lo; ok && start <= hi && counts[i] < limit; start++ {
				matched, err := pat.matchesAt(data, start, w)
				if errors.Is(err, errPatternStates) {
					counts[i] = countStateLimited
					break
				}
				if matched {
					counts[i]++
				}
			}
		}
		switch expr.evaluate(counts).state {
		case logicTrue:
			return p[0], nil
		case logicUnknown:
			if pending == nil {
				pending = ErrRuleIncomplete
			}
		}
	}
	return "", pending
}

func TestSharedPatternsMatchLikeIndependentRules(t *testing.T) {
	r := rand.New(rand.NewSource(20260923))
	alphabet := []byte("aAbB.1\x00mMzZ")
	literal := func(n int) string {
		var b strings.Builder
		for i := 0; i < n; i++ {
			b.WriteString(hex.EncodeToString([]byte{alphabet[r.Intn(len(alphabet))]}))
		}
		return b.String()
	}
	subsignature := func() string {
		body := literal(1 + r.Intn(3))
		switch r.Intn(9) {
		case 0:
			body += fmt.Sprintf("{%d}", r.Intn(3)) + literal(1+r.Intn(2))
		case 1:
			body += fmt.Sprintf("{%d-%d}", r.Intn(2), 2+r.Intn(3)) + literal(1)
		case 2:
			body += "??" + literal(1)
		case 3:
			body = "(" + literal(1) + "|" + literal(2) + ")" + body
		case 4:
			body += "*" + literal(1)
		case 5: // the anchor follows a variable gap
			body = literal(1) + fmt.Sprintf("{%d-%d}", r.Intn(3), 3+r.Intn(4)) + literal(3+r.Intn(2))
		case 6: // the anchor follows an unbounded gap
			body = literal(1) + "*" + literal(3+r.Intn(2))
		case 7: // every branch of an alternative supplies an anchor
			body = literal(1) + "(" + literal(2) + "|" + literal(3) + "{1}" + literal(2) + ")" + literal(1)
		case 8:
			body = "(" + literal(2) + "|" + literal(2) + ")" + "{0-2}" + literal(2)
		}
		body += []string{"", "", "::i", "::w", "::wi", "::a", "::f", "::if", "::wa", "::wia"}[r.Intn(10)]
		if strings.Contains(body, "::w") && strings.ContainsAny(body, "{?(*") {
			body = strings.SplitN(body, "::", 2)[0]
		}
		return []string{"", "", "", "0:", "2,4:", "EOF-3:", "1:"}[r.Intn(7)] + body
	}
	expressions := [][]string{
		{"0", "0>1", "0=2", "0<2", "0=0", "0>3"},
		{"0&1", "0|1", "(0|1)>2", "(0|1)>1,2", "0>1&1", "0|1=0", "(0|1)=3"},
		{"0&1&2", "0|1&2", "(0&1)|2", "(0|1|2)>2,2", "0&(1|2)", "(0|1)>1&2<2"},
	}
	checked, detected, multi, ranged := 0, 0, 0, 0
	for round := 0; round < 400; round++ {
		pool := make([]string, 6)
		for i := range pool {
			pool[i] = subsignature()
		}
		var lines []string
		for i := 0; i < 12; i++ {
			n := 1 + r.Intn(3)
			line := fmt.Sprintf("R%d;Target:0;%s", i, expressions[n-1][r.Intn(len(expressions[n-1]))])
			for j := 0; j < n; j++ {
				line += ";" + pool[r.Intn(len(pool))]
			}
			if _, err := Compile(map[string][]byte{"one.ldb": []byte(line)}); err == nil {
				lines = append(lines, line)
			}
		}
		if len(lines) == 0 {
			continue
		}
		d, err := Compile(map[string][]byte{"all.ldb": []byte(strings.Join(lines, "\n"))})
		if err != nil {
			t.Fatal(err)
		}
		for sample := 0; sample < 20; sample++ {
			data := make([]byte, r.Intn([]int{16, 48, 160}[sample%3]))
			for i := range data {
				data[i] = alphabet[r.Intn(len(alphabet))]
			}
			got, gotErr := d.Match(context.Background(), data)
			want, wantErr := oracleMatch(lines, data)
			if got != want || errors.Is(gotErr, ErrRuleIncomplete) != errors.Is(wantErr, ErrRuleIncomplete) {
				t.Fatalf("rules\n%s\ndata %q: shared %q %v, independent %q %v", strings.Join(lines, "\n"), data, got, gotErr, want, wantErr)
			}
			checked++
			if got != "" {
				detected++
			}
		}
		for _, anchors := range indexedPatternAnchors(d) {
			if len(anchors) > 1 {
				multi++
			}
			if len(anchors) == 1 && anchors[0].maxPre != anchors[0].minPre {
				ranged++
			}
		}
	}
	// The comparison must exercise detections, alternative anchors and anchors
	// at variable distances from the match start.
	if checked < 6000 || detected < checked/10 || multi < 100 || ranged < 100 {
		t.Fatalf("weak comparison: %d checks, %d detections, %d multi-anchor and %d ranged patterns", checked, detected, multi, ranged)
	}
}

// A pattern anchored on several alternative literals may start at an
// occurrence of any of them, so its start positions are the union of what
// every anchor allows, whichever literal the index happens to visit last.
func TestAlternativeAnchorsBoundStartsTogether(t *testing.T) {
	d := compileTest(t, "ldb", "Alt;Target:0;0;(4141414141|4242424242)4343\nBoth;Target:0;0>1;58595a::wa")
	for _, tt := range []struct{ data, want string }{
		{"AAAAACC....BBBBB", "Alt"},
		{"BBBBB....AAAAACC", "Alt"},
		{"BBBBBCC....AAAAA", "Alt"},
		{"AAAAA....BBBBBCC", "Alt"},
		{"AAAAA....BBBBB.CC", ""},
		{"XYZ....X\x00Y\x00Z\x00", "Both"}, // one plain and one wide match
		{"X\x00Y\x00Z\x00....XYZ", "Both"},
		{"XYZ....XY", ""},
	} {
		// The index visits literals in map order; repeat so every order is seen.
		for i := 0; i < 32; i++ {
			assertMatch(t, d, []byte(tt.data), tt.want)
		}
	}
}

// An anchor's distance from the match start can exceed 32 bits: 256 gaps of
// up to 2^24 bytes allow 2^32 + 1. That bound must not wrap into a small one
// that excludes the real start.
func TestAnchorDistanceBeyond32BitsKeepsMatches(t *testing.T) {
	body := "41" + strings.Repeat("{0-16777216}", 256) + "4343434343"
	d := compileTest(t, "ndb", "Far:0:*:"+body)
	assertMatch(t, d, []byte("A..........CCCCC"), "Far")
	assertMatch(t, d, []byte("..........CCCCC"), "")
}

// Anchors sharing long prefixes (wide paths, zero runs) land in crowded
// buckets that are split level by level; every anchor must still be found.
func TestCrowdedAnchorBucketsFindEveryAnchor(t *testing.T) {
	wide := func(s string) string {
		var b strings.Builder
		for _, c := range []byte(s) {
			fmt.Fprintf(&b, "%02x00", c)
		}
		return b.String()
	}
	var lines []string
	for i := 0; i < 60; i++ {
		lines = append(lines, fmt.Sprintf("Path.%d;Target:0;0;%s", i, wide(fmt.Sprintf("C:\\Windows\\dir%02d\\x", i))))
		lines = append(lines, fmt.Sprintf("Zero.%d;Target:0;0;%s%02x", i, strings.Repeat("00", 4+i%13), i+1))
		lines = append(lines, fmt.Sprintf("Short.%d;Target:0;0;%s%02x", i, strings.Repeat("00", 4), i+1))
		lines = append(lines, fmt.Sprintf("Fold.%d;Target:0;0;%s::i", i, hex.EncodeToString([]byte(fmt.Sprintf("C:\\Program Files\\App%02d", i)))))
		// Every four-byte window of a zero run is the same key.
		lines = append(lines, fmt.Sprintf("Run.%d;Target:0;0;%s", i, strings.Repeat("00", 21+i)))
	}
	d, err := Compile(map[string][]byte{"crowd.ldb": []byte(strings.Join(lines, "\n"))})
	if err != nil {
		t.Fatal(err)
	}
	crowded := false
	for _, bucket := range d.index.long {
		crowded = crowded || bucket.chunk != nil || bucket.single != nil
	}
	if !crowded {
		t.Fatal("the fixture does not produce a crowded bucket")
	}
	r := rand.New(rand.NewSource(3))
	wideBytes := func(s string) []byte {
		b, _ := hex.DecodeString(wide(s))
		return b
	}
	for round := 0; round < 300; round++ {
		var data []byte
		for len(data) < 400 {
			switch r.Intn(5) {
			case 0:
				data = append(data, wideBytes(fmt.Sprintf("C:\\Windows\\dir%02d\\x", r.Intn(70)))...)
			case 1:
				data = append(data, make([]byte, r.Intn([]int{20, 90}[r.Intn(2)]))...)
				data = append(data, byte(r.Intn(70)))
			case 2:
				data = append(data, []byte(fmt.Sprintf("c:\\PROGRAM files\\app%02d", r.Intn(70)))...)
			default:
				data = append(data, byte(r.Intn(256)))
			}
		}
		got, gotErr := d.Match(context.Background(), data)
		want, wantErr := oracleMatch(lines, data)
		if got != want || (gotErr == nil) != (wantErr == nil) {
			t.Fatalf("data %q: indexed %q %v, independent %q %v", data, got, gotErr, want, wantErr)
		}
	}
}

// rulePatterns counts the compiled patterns rules use, leaving out the type
// patterns of the file-type table.
func rulePatterns(d *Database) int {
	used := map[int]bool{}
	for _, r := range d.rules {
		for _, id := range r.patterns {
			used[id] = true
		}
	}
	return len(used)
}
