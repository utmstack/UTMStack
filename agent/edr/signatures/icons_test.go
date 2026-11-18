package signatures

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteIconLab writes the generated icon programs for the reference to
// hash (EDR_ICON_LAB_DIR); see docs/edr/native-engine/parity/icons.md.
func TestWriteIconLab(t *testing.T) {
	dir := os.Getenv("EDR_ICON_LAB_DIR")
	if dir == "" {
		t.Skip("EDR_ICON_LAB_DIR not set")
	}
	for i, p := range labPrograms() {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("lab_%02d.exe", i)), p, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// one program per icon, for threshold experiments
	for i, icons := range labIconSets() {
		for j, ic := range icons {
			name := filepath.Join(dir, fmt.Sprintf("one_%02d_%02d.exe", i, j))
			if err := os.WriteFile(name, labProgram([]labIcon{ic}), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// labIcons parses a generated program's icons as a scan would.
func labIcons(t *testing.T, program []byte) []programIcon {
	t.Helper()
	meta := parsePE(program)
	if !meta.valid || !meta.hasRes {
		t.Fatal("generated program has no resources")
	}
	icons, ok := meta.programIcons(program)
	if !ok {
		t.Fatal("generated program's icons are uncertain")
	}
	return icons
}

// oracleLines reads "program value..." lines of a testdata file.
func oracleLines(t *testing.T, name string) [][]string {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" && line[0] != '#' {
			out = append(out, strings.Fields(line))
		}
	}
	return out
}

// TestIconHashesAgreeWithReference compares the generated programs' icon
// hashes with the reference's (testdata/icon_hashes.txt: program number and
// hash of every icon the reference parsed, in order): every measure this
// engine computes must be equal, and the icons it parses must be the ones
// the reference parsed.
func TestIconHashesAgreeWithReference(t *testing.T) {
	want := map[int][]iconHash{}
	for _, f := range oracleLines(t, "icon_hashes.txt") {
		var p int
		fmt.Sscan(f[0], &p)
		h, err := decodeIconHash(f[1]) // widths 18 and 30 too
		if err != nil {
			t.Fatal(err)
		}
		want[p] = append(want[p], h)
	}
	gray, colored := 0, 0
	for i, program := range labPrograms() {
		icons := labIcons(t, program)
		if len(icons) != len(want[i]) {
			t.Fatalf("program %d: %d icons parsed, the reference parsed %d", i, len(icons), len(want[i]))
		}
		for j, icon := range icons {
			g, w := icon.hash, want[i][j]
			if g != w {
				t.Errorf("program %d icon %d: got %+v, want %+v", i, j, g, w)
				continue
			}
			if icon.grayMode {
				gray++
				continue
			}
			colored++
		}
	}
	if gray < 20 || colored < 100 {
		t.Errorf("too few icons compared: %d gray, %d colored", gray, colored)
	}
}

// TestIconMatchesAgreeWithReference checks the icon condition on entries
// derived from the reference's hashes of the generated icons (exact copies
// and perturbations; testdata/icon_matches.txt: program number, entry hash
// and whether the reference matched the program): a decided condition must
// be the reference's.
func TestIconMatchesAgreeWithReference(t *testing.T) {
	programs := labPrograms()
	icons := make([][]programIcon, len(programs))
	for i, p := range programs {
		icons[i] = labIcons(t, p)
	}
	decided, outcomes := map[bool]int{}, map[bool]int{}
	for _, f := range oracleLines(t, "icon_matches.txt") {
		var p int
		fmt.Sscan(f[0], &p)
		e, err := parseIconHash(f[1])
		if err != nil {
			t.Fatal(err)
		}
		reference := f[2] == "FOUND"
		outcomes[reference]++
		state := logicFalse
		for _, icon := range icons[p] {
			if iconComparisonWidth(icon.hash.width) != e.width {
				continue
			}
			switch iconConfidence(icon, e) {
			case logicTrue:
				state = logicTrue
			case logicUnknown:
				if state == logicFalse {
					state = logicUnknown
				}
			}
		}
		if state == logicUnknown {
			continue
		}
		if (state == logicTrue) != reference {
			t.Errorf("program %d entry %s: decided %v, the reference %v", p, f[1], state == logicTrue, reference)
		}
		decided[reference]++
	}
	t.Logf("matches: %d of %d decided; misses: %d of %d decided", decided[true], outcomes[true], decided[false], outcomes[false])
	for _, r := range []bool{true, false} {
		if decided[r]*2 < outcomes[r] {
			t.Errorf("fewer than half the reference's outcomes %v decided: %d of %d", r, decided[r], outcomes[r])
		}
	}
}

func TestIconDatabaseRejectsWhatTheReferenceRejects(t *testing.T) {
	valid := "10" + strings.Repeat("0", 122)
	set := func(at int, digits string) string { return valid[:at] + digits + valid[at+len(digits):] }
	for _, c := range []struct {
		hash string
		ok   bool
	}{
		{valid, true},
		{"18" + valid[2:], true},
		{"20" + valid[2:], true},
		{"0c" + valid[2:], false},        // width
		{"30" + valid[2:], false},        // width
		{set(2, "fe8"), true},            // color value 4072
		{set(2, "fe9"), false},           // 4073
		{set(5, "0e"), true},             // color corner 14 at width 16
		{set(5, "0f"), false},            // 15
		{"20" + set(5, "1c")[2:], true},  // 28 at width 32
		{"20" + set(5, "1d")[2:], false}, // 29
		{set(30, "fe9"), false},          // a gray value
		{set(44, "ff"), true},            // 8-bit values are not checked
		{set(46, "0f"), false},           // a bright corner
		{set(100, "0f"), false},          // a noedge corner
		{set(116, "67"), true},           // hue figures summing to 103
		{set(116, "68"), false},          // 104
		{set(116, "222223"), true},       // 34+34+35
		{set(116, "222323"), false},      // 34+35+35
		{set(122, "64"), true},           // colored share 100
		{set(122, "65"), false},          // 101
	} {
		_, err := Compile(map[string][]byte{"a.idb": []byte("Test.Icon:A:B:" + c.hash + "\n"), "a.ndb": []byte("Hit:0:*:414243\n")})
		if (err == nil) != c.ok {
			t.Errorf("%s: loaded %v, want %v (%v)", c.hash, err == nil, c.ok, err)
		}
	}
}

func TestIconDatabaseHoldsAtMost256NamesPerGroup(t *testing.T) {
	load := func(n int, second bool) error {
		var b strings.Builder
		for i := 0; i < n; i++ {
			if second {
				fmt.Fprintf(&b, "E%d:X:G%d:%s\n", i, i, iconMetrics)
			} else {
				fmt.Fprintf(&b, "E%d:G%d:X:%s\n", i, i, iconMetrics)
			}
		}
		_, err := Compile(map[string][]byte{"a.idb": []byte(b.String()), "a.ndb": []byte("Hit:0:*:414243\n")})
		return err
	}
	for _, second := range []bool{false, true} {
		if err := load(256, second); err != nil {
			t.Errorf("256 names (second %v): %v", second, err)
		}
		if load(257, second) == nil {
			t.Errorf("257 names (second %v) loaded", second)
		}
	}
}

func TestGrayAndColoredIconsScoreNoColor(t *testing.T) {
	colored := programIcon{hash: iconHash{width: 32, spread: [4]int{30, 30, 30, 100}}}
	gray := programIcon{hash: iconHash{width: 32}, grayMode: true}
	// A colored icon and a gray entry: no color or gray, and a spread of 0.
	if s := iconConfidence(colored, gray.hash); s != logicFalse {
		t.Errorf("colored icon, gray entry: %v", s)
	}
	// Gray is a zero colored share, whatever the hue figures: an entry
	// with hues and no colored share uses the six-area comparison.
	if s := iconConfidence(gray, iconHash{width: 32, spread: [4]int{9, 0, 0, 0}}); s != logicTrue {
		t.Errorf("gray icon, entry with hues and no colored share: %v", s)
	}
	// A gray icon and a colored entry differing by 1 in the colored share:
	// the zero area values match, and the reference scores this above 64.
	if s := iconConfidence(gray, iconHash{width: 32, spread: [4]int{0, 0, 0, 1}}); s != logicTrue {
		t.Errorf("gray icon, colored entry: %v", s)
	}
	// Exact copies of both kinds score 100.
	colored16 := programIcon{hash: iconHash{width: 16, spread: [4]int{30, 30, 30, 100}}}
	if s := iconConfidence(colored16, colored16.hash); s != logicTrue {
		t.Errorf("colored copy: %v", s)
	}
	gray16 := programIcon{hash: iconHash{width: 16}, grayMode: true}
	if s := iconConfidence(gray16, gray16.hash); s != logicTrue {
		t.Errorf("gray copy: %v", s)
	}
}
