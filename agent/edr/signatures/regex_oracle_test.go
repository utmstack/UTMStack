package signatures

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
)

// testdata/regex_oracle.txt.gz holds random regex cases and, for each, what
// the regular expression library the reference links (PCRE2 10.47, 8-bit,
// without Unicode, the options of the subsignature flags) returns for one
// match call, with the smallest match and depth limits that still give that
// result. The cases were generated at random and run through the library by
// a small C program; both are kept outside the repository.
//
// Per line, tab-separated: pattern (hex), flags, subject (a chunk in hex, a
// repeat count and a tail in hex, joined by ':'), start offset, anchored
// (0 or 1), then the library's status (M match, N no match, C compile error,
// L over 20 million calls or frames, E other error), match start, match end,
// smallest match limit and smallest depth limit.
//
// EDR_REGEX_ORACLE names more case files in the same format (plain or gzip),
// such as cases built from the official feed's patterns.
func TestRegexAgreesWithReferenceLibrary(t *testing.T) {
	files := []string{"testdata/regex_oracle.txt.gz"}
	if extra := os.Getenv("EDR_REGEX_ORACLE"); extra != "" {
		files = append(files, strings.Split(extra, ",")...)
	}
	for _, name := range files {
		t.Run(name, func(t *testing.T) { checkRegexOracle(t, name) })
	}
}

type regexOracleStats struct {
	cases, matches, noMatches, rejected, unsupported, skipped int
	// maxRatio is the largest limit the library needed over the counts
	// this engine made (it must not exceed 1).
	maxCallRatio, maxDepthRatio float64
	// wrong counts results that differ; undercounts matches whose counts
	// stay below the library's.
	wrong, undercounts int
	// overLimit counts cases this engine gave up on within the test's
	// own large limits.
	overLimit int
	// the largest shortfalls, and the largest ratios where the library
	// needed at least 1,000 calls or a depth of 100.
	callGap, depthGap           int
	bigCallRatio, bigDepthRatio float64
}

func checkRegexOracle(t *testing.T, name string) {
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var in io.Reader = f
	if strings.HasSuffix(name, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		in = gz
	}
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	var st regexOracleStats
	failures := 0
	fail := func(format string, args ...any) {
		t.Helper()
		if failures++; failures <= 20 || strings.Contains(format, "this engine [") || strings.Contains(format, "library no match") || strings.Contains(format, "rejects") {
			t.Errorf(format, args...)
		}
	}
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) != 10 {
			t.Fatalf("bad line %q", sc.Text())
		}
		pat, err1 := hex.DecodeString(f[0])
		data, err2 := oracleSubject(f[2])
		start, err3 := strconv.Atoi(f[3])
		ms, _ := strconv.Atoi(f[6])
		me, _ := strconv.Atoi(f[7])
		minCalls, _ := strconv.Atoi(f[8])
		minDepth, _ := strconv.Atoi(f[9])
		if err1 != nil || err2 != nil || err3 != nil {
			t.Fatalf("bad line %q", sc.Text())
		}
		pattern, flags, status := string(pat), f[1], f[5]
		st.cases++
		r, err := compileRegexSub("*", pattern, flags, dconfPCREGlobal)
		switch status {
		case "C":
			st.rejected++
			// Unsupported is enough: such a rule stays unknown when reached.
			if err == nil {
				fail("/%s/%s: the library rejects it; this engine accepts it", pattern, flags)
			}
			continue
		case "L", "E":
			st.skipped++
			continue
		}
		if err != nil {
			var u *errRegexUnsupported
			if !errors.As(err, &u) {
				fail("/%s/%s: %v; the library compiles it", pattern, flags, err)
			}
			st.unsupported++
			continue
		}
		w := &matchWork{ctx: context.Background(), limit: 1 << 50}
		m := newRxMatcher(r.pattern, data, 0, len(data), w)
		// Far above the library's limits, but bounded: this engine lacks
		// the library's optimizations, so some searches the library cuts
		// short are exponential here (they stay unknown in a scan).
		m.callLimit, m.depthLimit = 300_000, 100_000
		anchored := f[4] == "1" || r.pattern.anchored
		gotS, gotE, found, err := r.find(m, start, len(data), anchored)
		if err == errRegexNoMatch {
			found, err = false, nil
		}
		if err == errRegexLimit {
			st.overLimit++
			continue
		}
		if err != nil {
			fail("/%s/%s on %q from %d: %v", pattern, flags, data, start, err)
			continue
		}
		switch {
		case status == "M" && (!found || gotS != ms || gotE != me):
			st.wrong++
			fail("/%s/%s on %q from %d (anchored %v): library [%d,%d), this engine [%d,%d) found %v", pattern, flags, data, start, anchored, ms, me, gotS, gotE, found)
			continue
		case status == "N" && found:
			st.wrong++
			fail("/%s/%s on %q from %d (anchored %v): library no match, this engine [%d,%d)", pattern, flags, data, start, anchored, gotS, gotE)
			continue
		case status == "N":
			st.noMatches++
			continue
		}
		st.matches++
		// The engine's counts, with its margins, must reach the library's:
		// then a match within this engine's limits is one the library
		// reports too.
		if m.peakCalls < minCalls || m.peakDepth < minDepth {
			st.undercounts++
		}
		if m.peakCalls+rxCallSlack < minCalls || m.peakDepth+rxDepthSlack < minDepth {
			fail("/%s/%s on %d bytes from %d: library needs %d calls and depth %d, this engine counted %d and %d", pattern, flags, len(data), start, minCalls, minDepth, m.peakCalls, m.peakDepth)
		}
		st.callGap = max(st.callGap, minCalls-m.peakCalls)
		st.depthGap = max(st.depthGap, minDepth-m.peakDepth)
		if minCalls >= 1000 {
			st.bigCallRatio = max(st.bigCallRatio, float64(minCalls)/float64(max(m.peakCalls, 1)))
		}
		if minDepth >= 100 {
			st.bigDepthRatio = max(st.bigDepthRatio, float64(minDepth)/float64(max(m.peakDepth, 1)))
		}
		st.maxCallRatio = max(st.maxCallRatio, float64(minCalls)/float64(max(m.peakCalls, 1)))
		st.maxDepthRatio = max(st.maxDepthRatio, float64(minDepth)/float64(max(m.peakDepth, 1)))
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d cases: %d matches, %d without a match, %d the library rejects, %d unsupported here, %d skipped; %d different results, %d matches counted below the library, %d over this test's limits here; largest library/engine ratio: calls %.2f, depth %.2f",
		st.cases, st.matches, st.noMatches, st.rejected, st.unsupported, st.skipped, st.wrong, st.undercounts, st.overLimit, st.maxCallRatio, st.maxDepthRatio)
	t.Logf("largest shortfall: %d calls, depth %d; largest ratio where the library needed 1,000 calls: %.2f, depth 100: %.2f", st.callGap, st.depthGap, st.bigCallRatio, st.bigDepthRatio)
	if failures > 20 {
		t.Errorf("%d failures in all", failures)
	}
	if name == "testdata/regex_oracle.txt.gz" && (st.matches < 5000 || st.noMatches < 5000 || st.unsupported > st.cases/20) {
		t.Errorf("weak comparison")
	}
}

// oracleSubject expands chunk:repeat:tail.
func oracleSubject(s string) ([]byte, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return nil, errors.New("bad subject")
	}
	chunk, err := hex.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	tail, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	var out []byte
	for i := 0; i < n; i++ {
		out = append(out, chunk...)
	}
	return append(out, tail...), nil
}
