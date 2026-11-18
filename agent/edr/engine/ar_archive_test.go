package engine

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

type arFixtureMember struct {
	name string
	data []byte
}

// Encode independent fixed-width ASCII records, including their padding. No
// production parsing helper participates in constructing a valid fixture.
func arFixture(t testing.TB, members ...arFixtureMember) []byte {
	t.Helper()
	out := []byte("!<arch>\n")
	for _, member := range members {
		header := fmt.Sprintf("%-16s%-12d%-6d%-6d%-8s%-10d`\n", member.name, 0, 0, 0, "100600", len(member.data))
		if len(header) != 60 {
			t.Fatalf("invalid fixture header length: %d", len(header))
		}
		out = append(out, header...)
		out = append(out, member.data...)
		if len(member.data)&1 != 0 {
			out = append(out, '\n')
		}
	}
	return out
}

func arState(t *testing.T) *scanState {
	t.Helper()
	return pipelineState(t, map[string]string{"ar.ndb": "AR.Safe:0:*:" + hx(formatCoverageMarker) + "\n"})
}

func TestARVariantsAndDebianNestedExtraction(t *testing.T) {
	payload := gzipFixture(t, []byte(formatCoverageMarker))
	long := "fixtures/" + strings.Repeat("long-", 20) + "inert.gz"
	table := []byte(long + "/\nother.gz/\n")
	bsdName := []byte(long + "\x00\x00")
	var tarBuffer bytes.Buffer
	tw := tar.NewWriter(&tarBuffer)
	if err := tw.WriteHeader(&tar.Header{Name: "../../not-extracted.txt", Mode: 0600, Size: int64(len(formatCoverageMarker))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(formatCoverageMarker)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string][]byte{
		"sysv":                   arFixture(t, arFixtureMember{"inert.gz/", payload}),
		"gnu-longname":           arFixture(t, arFixtureMember{"//", table}, arFixtureMember{"/0", payload}),
		"gnu-forward-longname":   arFixture(t, arFixtureMember{"/0", payload}, arFixtureMember{"//", table}),
		"gnu-second-longname":    arFixture(t, arFixtureMember{"//", table}, arFixtureMember{fmt.Sprintf("/%d", len(long)+2), payload}),
		"bsd-longname":           arFixture(t, arFixtureMember{fmt.Sprintf("#1/%d", len(bsdName)), append(bsdName, payload...)}),
		"odd-member-first":       arFixture(t, arFixtureMember{"odd.txt/", []byte("odd")}, arFixtureMember{"inert.gz/", payload}),
		"empty-member-first":     arFixture(t, arFixtureMember{"empty/", nil}, arFixtureMember{"inert.gz/", payload}),
		"symbol-table-first":     arFixture(t, arFixtureMember{"/", []byte("symbol metadata")}, arFixtureMember{"inert.gz/", payload}),
		"bsd-symbol-table-first": arFixture(t, arFixtureMember{"__.SYMDEF/", []byte("symbol metadata")}, arFixtureMember{"inert.gz/", payload}),
		"debian-package": arFixture(t, arFixtureMember{"debian-binary/", []byte("2.0\n")},
			arFixtureMember{"control.tar.gz/", gzipFixture(t, []byte("ordinary harmless control metadata"))},
			arFixtureMember{"data.tar.gz/", gzipFixture(t, tarBuffer.Bytes())}),
	}
	for name, data := range fixtures {
		t.Run(name, func(t *testing.T) {
			if bytes.Contains(data, []byte(formatCoverageMarker)) {
				t.Fatal("fixture exposes its marker without decoding")
			}
			expectScan(t, arState(t), data, "AR.Safe", "")
		})
	}
	for _, content := range [][]byte{[]byte("!<arch>\n"), arFixture(t, arFixtureMember{"inert.txt/", []byte("ordinary harmless text")})} {
		expectScan(t, arState(t), content, "", "")
	}
}

func TestARMetadataRulesSeeDecodedNamesAndActualPayloadSizes(t *testing.T) {
	name := "fixtures/inert.gz"
	payload := []byte("ordinary harmless data")
	for _, data := range [][]byte{
		arFixture(t, arFixtureMember{"//", []byte(name + "/\n")}, arFixtureMember{"/0", payload}),
		arFixture(t, arFixtureMember{fmt.Sprintf("#1/%d", len(name)), append([]byte(name), payload...)}),
	} {
		rules := map[string]string{"member.cdb": fmt.Sprintf("AR.Named:*:*:^fixtures/inert[.]gz$:%d:%d:0:*:*:*\n", len(payload), len(payload))}
		expectScan(t, pipelineState(t, rules), data, "AR.Named", "")
	}
}

func TestARMalformedAndThinRemainIncomplete(t *testing.T) {
	base := arFixture(t, arFixtureMember{"inert.txt/", []byte("ordinary harmless text")})
	for name, change := range map[string]func([]byte) []byte{
		"short-header":   func(b []byte) []byte { return b[:40] },
		"short-body":     func(b []byte) []byte { return b[:len(b)-3] },
		"bad-trailer":    func(b []byte) []byte { b[8+58] = '!'; return b },
		"negative-size":  func(b []byte) []byte { copy(b[8+48:8+58], "-000000001"); return b },
		"overflow-size":  func(b []byte) []byte { copy(b[8+48:8+58], "9999999999"); return b },
		"size-with-junk": func(b []byte) []byte { copy(b[8+48:8+58], "2x        "); return b },
	} {
		t.Run(name, func(t *testing.T) {
			sig, err := arState(t).visit(change(append([]byte(nil), base...)), 0)
			if sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("scan = %q, %v; want malformed", sig, err)
			}
		})
	}
	for name, data := range map[string][]byte{
		"bsd-name-too-large":  arFixture(t, arFixtureMember{"#1/999999", []byte("ordinary harmless text")}),
		"gnu-no-table":        arFixture(t, arFixtureMember{"/0", []byte("ordinary harmless text")}),
		"gnu-offset-midname":  arFixture(t, arFixtureMember{"//", []byte("inert.txt/\n")}, arFixtureMember{"/2", []byte("ordinary harmless text")}),
		"gnu-offset-overflow": arFixture(t, arFixtureMember{"//", []byte("inert.txt/\n")}, arFixtureMember{"/99999999999999", nil}),
		"missing-odd-padding": arFixture(t, arFixtureMember{"odd/", []byte("ordinary!")})[:77],
	} {
		t.Run(name, func(t *testing.T) {
			sig, err := arState(t).visit(data, 0)
			if sig != "" || !errors.Is(err, ErrMalformed) {
				t.Fatalf("scan = %q, %v; want malformed", sig, err)
			}
		})
	}
	thin := []byte("!<thin>\nexternal-file-path-that-is-never-opened")
	if sig, err := arState(t).visit(thin, 0); sig != "" || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("thin archive = %q, %v; want unsupported", sig, err)
	}
}

func TestARBudgetsCancellationAndNameBounds(t *testing.T) {
	data := arFixture(t, arFixtureMember{"one/", []byte("ordinary harmless member")}, arFixtureMember{"two/", []byte("ordinary harmless second member")})
	for _, limit := range []string{"file-size", "aggregate", "depth", "objects", "header-count"} {
		t.Run(limit, func(t *testing.T) {
			s := arState(t)
			switch limit {
			case "file-size":
				s.maxFile = int64(len(data) - 1)
			case "aggregate":
				s.remaining = int64(len(data) + 10)
			case "depth":
				s.maxDepth = 0
			case "objects":
				s.maxFiles = 1
			case "header-count":
				s.maxFiles = 2
				data = arFixture(t, arFixtureMember{"/", nil}, arFixtureMember{"//", nil}, arFixtureMember{"empty/", nil})
			}
			if sig, err := s.visit(data, 0); sig != "" || !errors.Is(err, ErrLimit) {
				t.Fatalf("scan = %q, %v; want limit", sig, err)
			}
		})
	}
	s := arState(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if sig, err := s.visit(data, 0); sig != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled = %q, %v", sig, err)
	}
	long := strings.Repeat("a", arMaxName+1)
	for _, data := range [][]byte{
		arFixture(t, arFixtureMember{"//", []byte(long + "/\n")}, arFixtureMember{"/0", []byte("ordinary harmless content")}),
		arFixture(t, arFixtureMember{fmt.Sprintf("#1/%d", len(long)), []byte(long + "ordinary harmless content")}),
	} {
		if sig, err := arState(t).visit(data, 0); sig != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("oversized name = %q, %v; want limit", sig, err)
		}
	}
}

func TestARDetectionSurvivesMalformedLaterMemberAndName(t *testing.T) {
	payload := gzipFixture(t, []byte(formatCoverageMarker))
	for _, data := range [][]byte{
		append(arFixture(t, arFixtureMember{"inert.gz/", payload}), []byte("truncated later header")...),
		arFixture(t, arFixtureMember{"/0", payload}), // invalid name must not hide known payload
	} {
		expectScan(t, arState(t), data, "AR.Safe", "")
	}
}

func TestARNumericParsingRejectsOverflowAndNonDecimal(t *testing.T) {
	for _, field := range []string{"", " ", "-1", "+1", "0x10", "12 34", strings.Repeat("9", 100)} {
		if value, ok := arNumber([]byte(field)); ok {
			t.Errorf("accepted %q as %d", field, value)
		}
	}
	if value, ok := arNumber([]byte("  12345  ")); !ok || value != 12345 {
		t.Fatalf("decimal = %d, %v", value, ok)
	}
}

func FuzzARArchiveBounded(f *testing.F) {
	f.Add([]byte("!<arch>\n"))
	f.Add(arFixture(f, arFixtureMember{"inert.txt/", []byte("ordinary harmless text")}))
	f.Add(arFixture(f, arFixtureMember{"//", []byte("long/inert.txt/\n")}, arFixtureMember{"/0", []byte("ordinary harmless text")}))
	f.Add([]byte("!<thin>\n"))
	db, err := signatures.Compile(map[string][]byte{"ar.ndb": []byte("AR.Safe:0:*:" + hx(formatCoverageMarker) + "\n")})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128<<10 {
			t.Skip()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		s := scanState{ctx: ctx, db: db, maxFile: 256 << 10, remaining: 512 << 10, maxFiles: 64, maxDepth: 4}
		_, _ = s.visit(data, 0)
	})
}
