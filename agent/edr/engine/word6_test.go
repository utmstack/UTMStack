package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const word6Marker = "OPENEDR WORD6 HARMLESS MACRO"

// The fixture contains only XOR-obfuscated text, never WordBasic instructions.
// The 24-byte wire records are assembled independently of the reader.
func word6Control(prefix []byte) []byte {
	b := make([]byte, 0x600)
	le := binary.LittleEndian
	le.PutUint16(b, 0xa5dc)
	le.PutUint32(b[0x118:], 0x140)
	table := append([]byte{0xff}, prefix...)
	table = append(table, 1, 1, 0) // record ID, count=1
	record := make([]byte, 24)
	record[0], record[1] = 7, 0xa5 // ignored version and XOR key
	le.PutUint32(record[12:], uint32(len(word6Marker)))
	le.PutUint32(record[16:], 0x12345678) // ignored macro state
	le.PutUint32(record[20:], 0x400)
	table = append(table, record...)
	le.PutUint32(b[0x11c:], uint32(len(table)))
	copy(b[0x140:], table)
	for i := range word6Marker {
		b[0x400+i] = word6Marker[i] ^ 0xa5
	}
	return b
}

func TestWord6MacrosDecodedAsOLEBuffers(t *testing.T) {
	menu := append([]byte{5, 1, 0}, make([]byte, 12)...)
	firstList := append([]byte{3, 1}, make([]byte, 14)...)
	firstList = append(firstList, 2)
	firstList = append(firstList, make([]byte, 9)...)
	secondList := append([]byte{3, 0, 0, 2, 1}, make([]byte, 5)...)
	ansiName := []byte{0x10, 0x7f, 0, 2, 'x', 'y', 0, 0}
	unicodeName := []byte{0x10, 0xff, 0xff, 0x7f, 0, 1, 'x', 0, 0, 0, 0}
	internalName := []byte{0x11, 1, 0, 0x34, 0x12, 2, 'x', 'y', 0}
	all := append(bytes.Clone(menu), firstList...)
	all = append(all, secondList...)
	all = append(all, ansiName...)
	all = append(all, unicodeName...)
	all = append(all, internalName...)
	for _, tc := range []struct {
		name   string
		prefix []byte
		mutate func([]byte)
		found  bool
	}{
		{name: "direct", found: true},
		{name: "menu", prefix: menu, found: true},
		{name: "first-list", prefix: firstList, found: true},
		{name: "second-list", prefix: secondList, found: true},
		{name: "list-rewind", prefix: []byte{3, 0, 0}, found: true},
		{name: "ansi-name", prefix: ansiName, found: true},
		{name: "unicode-name", prefix: unicodeName, found: true},
		{name: "internal-name", prefix: internalName, found: true},
		{name: "all-record-types", prefix: all, found: true},
		{name: "negative-name-size", prefix: []byte{0x10, 0, 0x80}, found: true},
		{name: "zero-name-size", prefix: []byte{0x10, 0, 0}, found: true},
		{name: "short-table-extent", mutate: func(b []byte) { binary.LittleEndian.PutUint32(b[0x11c:], 2) }, found: true},
		{name: "ignored-start-byte", mutate: func(b []byte) { b[0x140] = 'Q' }, found: true},
		{name: "no-macro-length", mutate: func(b []byte) { clear(b[0x11c:0x120]) }},
		{name: "no-record-dispatch", mutate: func(b []byte) { binary.LittleEndian.PutUint32(b[0x11c:], 1) }},
		{name: "wrapped-end", mutate: func(b []byte) { binary.LittleEndian.PutUint32(b[0x11c:], 0xffffffff) }},
		{name: "end-40", prefix: []byte{0x40}},
		{name: "end-12", prefix: []byte{0x12}},
		{name: "unknown-id", prefix: []byte{0x33}},
		{name: "zero-count", mutate: func(b []byte) { clear(b[0x142:0x144]) }},
		{name: "zero-length", mutate: func(b []byte) { clear(b[0x150:0x154]) }},
		{name: "wrong-key", mutate: func(b []byte) { b[0x145] ^= 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := word6Control(tc.prefix)
			if tc.mutate != nil {
				tc.mutate(stream)
			}
			parent := oleFixture(t, 3, "WordDocument", stream, true, false, false)
			if bytes.Contains(parent, []byte(word6Marker)) {
				t.Fatal("plaintext marker present in input")
			}
			rules := map[string]string{"macro.ndb": "Word6.Decoded:2:0:" + hex.EncodeToString([]byte(word6Marker)) + "\n"}
			want := ""
			if tc.found {
				want = "Word6.Decoded"
			}
			if sig, err := pipelineState(t, rules).visit(parent, 0); sig != want || err != nil {
				t.Fatalf("decoded macro: signature=%q error=%v, want %q", sig, err, want)
			}
			word6Reference(t, parent, rules, want)
		})
	}
}

func word6Reference(t *testing.T, parent []byte, rules map[string]string, want string) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned ClamAV 1.5.4 scanner: %v", err)
	}
	dir := t.TempDir()
	for name, text := range rules {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(dir, "word6.doc")
	if err := os.WriteFile(path, parent, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(clamscan, "--database="+dir, "--debug", "--no-summary", path)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if want != "" {
		status, ok := err.(*exec.ExitError)
		if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "word6.doc: "+want+".UNOFFICIAL FOUND") || !strings.Contains(string(out), "Decompress WM project macro:") {
			t.Fatalf("reference Word 6 macro: %v: %s", err, out)
		}
	} else if err != nil || !strings.Contains(string(out), "word6.doc: OK") {
		t.Fatalf("reference Word 6 control: %v: %s", err, out)
	}
	t.Logf("ClamAV input SHA256=%x detected=%t", sha256.Sum256(parent), want != "")
}

func TestWord6MacrosUseBufferRules(t *testing.T) {
	parent := oleFixture(t, 3, "WordDocument", word6Control(nil), true, false, false)
	for _, rules := range []map[string]string{
		{"macro.hdb": fmt.Sprintf("%x:%d:Word6.FileHash\n", md5.Sum([]byte(word6Marker)), len(word6Marker))},
		{"macro.ndb": fmt.Sprintf("Word6.EOF:2:EOF-%d:%x\n", len(word6Marker), word6Marker)},
		{"macro.ndb": fmt.Sprintf("Word6.WrongRoot:7:0:%x\n", word6Marker)},
		{"macro.ndb": fmt.Sprintf("Word6.NoNormalization:2:0:%x\n", strings.ToLower(word6Marker))},
	} {
		if sig, err := pipelineState(t, rules).visit(parent, 0); sig != "" || err != nil {
			t.Fatalf("file-only rule matched a macro buffer: signature=%q error=%v", sig, err)
		}
		word6Reference(t, parent, rules, "")
	}
}

func TestWord6MultipleMacrosAndZeroKey(t *testing.T) {
	rules := map[string]string{"macro.ndb": fmt.Sprintf("Word6.Decoded:2:0:%x\n", word6Marker)}
	for _, tc := range []struct {
		name string
		set  func([]byte)
	}{
		{"later-macro", func(b []byte) {
			binary.LittleEndian.PutUint16(b[0x142:], 2)
			copy(b[0x15c:], b[0x144:0x15c])
			b[0x145] ^= 1 // first record decodes to unrelated text
			binary.LittleEndian.PutUint32(b[0x11c:], 52)
		}},
		{"zero-key", func(b []byte) {
			b[0x145] = 0
			copy(b[0x400:], word6Marker)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := word6Control(nil)
			tc.set(stream)
			parent := oleFixture(t, 3, "WordDocument", stream, true, false, false)
			if sig, err := pipelineState(t, rules).visit(parent, 0); sig != "Word6.Decoded" || err != nil {
				t.Fatalf("macro list: signature=%q error=%v", sig, err)
			}
			word6Reference(t, parent, rules, "Word6.Decoded")
		})
	}
	t.Run("invalid-before-valid", func(t *testing.T) {
		stream := word6Control(nil)
		binary.LittleEndian.PutUint16(stream[0x142:], 2)
		copy(stream[0x15c:], stream[0x144:0x15c])
		binary.LittleEndian.PutUint32(stream[0x158:], 0xffffffff)
		binary.LittleEndian.PutUint32(stream[0x11c:], 52)
		s := pipelineState(t, rules)
		if sig, err := s.word6Macros(stream, 1); sig != "Word6.Decoded" || err != nil {
			t.Fatalf("later detection lost to malformed macro: signature=%q error=%v", sig, err)
		}
	})
	t.Run("decoded-byte-accounting", func(t *testing.T) {
		s := pipelineState(t, rules)
		before, files := s.remaining, s.files
		if sig, err := s.word6Macros(word6Control(nil), 1); sig != "Word6.Decoded" || err != nil || s.remaining != before-int64(len(word6Marker)) || s.files != files+1 {
			t.Fatalf("macro admission: signature=%q error=%v bytes=%d files=%d", sig, err, before-s.remaining, s.files-files)
		}
	})
}

func TestWord6MacrosLimitsMalformedAndCancellation(t *testing.T) {
	rules := map[string]string{"macro.ndb": fmt.Sprintf("Word6.Decoded:2:0:%x\n", word6Marker)}
	for _, tc := range []struct {
		name string
		set  func(*scanState)
	}{
		{"file-size", func(s *scanState) { s.maxFile = int64(len(word6Marker) - 1) }},
		{"aggregate", func(s *scanState) { s.remaining = int64(len(word6Marker) - 1) }},
		{"file-count", func(s *scanState) { s.files = s.maxFiles }},
		{"depth", func(s *scanState) { s.maxDepth = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := pipelineState(t, rules)
			tc.set(s)
			if sig, err := s.word6Macros(word6Control(nil), 1); sig != "" || !errors.Is(err, ErrLimit) {
				t.Fatalf("macro limit: signature=%q error=%v", sig, err)
			}
		})
	}
	s := pipelineState(t, rules)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if sig, err := s.word6Macros(word6Control(nil), 1); sig != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled macro: signature=%q error=%v", sig, err)
	}
	for _, tc := range []struct {
		name string
		set  func([]byte)
		want error
	}{
		{"count", func(b []byte) { binary.LittleEndian.PutUint16(b[0x142:], 65535) }, nil},
		{"length", func(b []byte) { binary.LittleEndian.PutUint32(b[0x150:], 0xffffffff) }, ErrMalformed},
		{"offset", func(b []byte) { binary.LittleEndian.PutUint32(b[0x158:], 0xffffffff) }, ErrMalformed},
		{"table", func(b []byte) { binary.LittleEndian.PutUint32(b[0x118:], uint32(len(b))) }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stream := word6Control(nil)
			tc.set(stream)
			if sig, err := pipelineState(t, rules).word6Macros(stream, 1); sig != "" || !errors.Is(err, tc.want) {
				t.Fatalf("malformed macro: signature=%q error=%v", sig, err)
			}
		})
	}
}

// The upstream ClamAV 1.5.4 regression document is suspicious-looking but
// benign, as documented by unit_tests/input/xor_testfile.py. Decode its .xor
// member with that tool and supply the resulting path. It is never executed.
// It has no readable Word 6 directory: its legacy fields point outside the
// WordDocument stream. This is a negative calibration control, not evidence
// of positive macro extraction from an independently authored document.
func TestWord6OfficialClamAVNegativeControl(t *testing.T) {
	path := os.Getenv("EDR_CLAMAV_WORD6_FILE")
	if path == "" {
		t.Skip("set EDR_CLAMAV_WORD6_FILE to decoded ClamAV 1.5.4 clam.ole.doc")
	}
	parent, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != "5ecdc3d43a7632d400bf3c6fd9422f31d67467da0feca98f2f414f7dde803c39" {
		t.Fatalf("upstream Word 6 control SHA-256 = %s", got)
	}
	s := oleTestState(t)
	f, err := readOLE(s, parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range f.streams {
		if st.name != "worddocument" {
			continue
		}
		data, err := f.read(st)
		if err != nil {
			t.Fatal(err)
		}
		if records, err := word6Directory(context.Background(), data); len(records) != 0 || err != nil {
			t.Fatalf("unreadable legacy directory: records=%d error=%v", len(records)/24, err)
		}
		rules := map[string]string{"macro.ndb": fmt.Sprintf("Word6.Decoded:2:0:%x\n", word6Marker)}
		if sig, err := pipelineState(t, rules).visit(parent, 0); sig != "" || err != nil {
			t.Fatalf("upstream negative control: signature=%q error=%v", sig, err)
		}
		word6Reference(t, parent, rules, "")
		return
	}
	t.Fatal("upstream WordDocument stream missing")
}

func FuzzWord6Directory(f *testing.F) {
	f.Add(word6Control(nil))
	f.Add(word6Control([]byte{0x10, 0xff, 0xff, 1, 0, 0}))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		records, _ := word6Directory(context.Background(), b)
		if len(records)%24 != 0 {
			t.Fatalf("partial macro record: %d bytes", len(records))
		}
	})
}
