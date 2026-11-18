package engine

import (
	"archive/zip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// vbaOracle reads testdata/vbaproject/oracle.zip: generated Office files
// (docs/2026-10-08-macro-oracle/ in the workspace: mkvba.py, decomp_lab.py)
// and, for each, the md5:size of every macro project dump the reference
// (ClamAV 1.4.3) scanned, from clamscan --debug --leave-temps.
func vbaOracle(t *testing.T) (map[string][]byte, map[string][]string) {
	t.Helper()
	z, err := zip.OpenReader("testdata/vbaproject/oracle.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	files := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(r)
		r.Close()
		if err != nil {
			t.Fatal(err)
		}
		files[f.Name] = b
	}
	var manifest map[string][]string
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	delete(files, "manifest.json")
	return files, manifest
}

func vbaRuleEngine(t *testing.T, name, rules string) *Engine {
	t.Helper()
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, name), []byte(rules), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	return e
}

// vbaUndecidedCodepages are the oracle documents whose conversion this engine
// leaves undecided (vbacodepage.go): the machine's own codepages (0 and 1)
// with non-ASCII text, UTF-16 (1200) and a codepage number it has no table
// for.
var vbaUndecidedCodepages = []string{"p2/cp-0.doc", "p2/cp-1.doc", "p2/cp-1200.doc", "p2/cp-39321.doc"}

// Expectations are the Windows reference's dumps where they were measured
// (all codepage documents and most others), the Debian reference's for the
// rest; the two agree everywhere except in codepage conversion.
// Every dump the reference scanned is produced and scanned byte for byte
// (a hash rule per dump), and a dump is scanned exactly when the reference
// scans one (a rule on the dump's first line).
func TestVBAProjectDumpsAsReference(t *testing.T) {
	files, manifest := vbaOracle(t)
	var hashes strings.Builder
	for _, hs := range manifest {
		for _, h := range hs {
			hashes.WriteString(h + ":Dump." + h[:32] + "\n")
		}
	}
	byHash := vbaRuleEngine(t, "dumps.hdb", hashes.String())
	header := vbaRuleEngine(t, "header.ndb", "Dump.Header:0:0:"+hex.EncodeToString([]byte(vbaDumpHeader[:56]))+"\n")
	names := make([]string, 0, len(manifest))
	for name := range manifest {
		names = append(names, name)
	}
	sort.Strings(names)
	undecided := 0
	for _, name := range names {
		want, data := manifest[name], files[name]
		_, sig, err := byHash.Scan(context.Background(), data)
		_, hsig, herr := header.Scan(context.Background(), data)
		if undecidedCodepage(name) {
			if sig != "" || unknownReasonOf(err) != signatures.UnknownVBA {
				t.Errorf("%s: (%q, %v), want undecided", name, sig, err)
			}
			undecided++
			continue
		}
		if len(want) == 0 {
			if sig != "" || hsig != "" || err != nil || herr != nil {
				t.Errorf("%s: (%q, %v) (%q, %v), want no dump", name, sig, err, hsig, herr)
			}
			continue
		}
		found := false
		for _, h := range want {
			found = found || sig == "Dump."+h[:32]
		}
		if !found || hsig != "Dump.Header" {
			t.Errorf("%s: (%q, %v) (%q, %v), want one of %v", name, sig, err, hsig, herr, want)
		}
	}
	if len(manifest) != 364 || undecided != 4 {
		t.Fatalf("%d documents, %d undecided", len(manifest), undecided)
	}
}

func undecidedCodepage(name string) bool {
	for _, p := range vbaUndecidedCodepages {
		if strings.Contains(name, p) {
			return true
		}
	}
	return false
}

// TestVBAOracleCorpus checks real Office files against the reference's
// outputs (docs/2026-10-08-macro-oracle/dumpsb.py): EDR_VBA_ORACLE is its
// output folder (index, N.log, N.dumpK, N.modK) and EDR_VBA_ORACLE_ROOT the
// prefix of the scanned paths. Every dump the reference scanned must be
// produced byte for byte, and every module text its module reader scanned
// must be matched as module text (a target-2 rule on a stretch of it that
// holds upper-case letters, which the lower-cased dump does not).
func TestVBAOracleCorpus(t *testing.T) {
	out, root := os.Getenv("EDR_VBA_ORACLE"), os.Getenv("EDR_VBA_ORACLE_ROOT")
	if out == "" {
		t.Skip("EDR_VBA_ORACLE not set")
	}
	index, err := os.ReadFile(filepath.Join(out, "index"))
	if err != nil {
		t.Fatal(err)
	}
	type entry struct {
		path    string
		dumps   []string
		modules []string
	}
	var entries []entry
	var hashes strings.Builder
	moduleRules := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(index)), "\n") {
		f := strings.Split(line, "\t")
		if len(f) != 3 {
			continue
		}
		n := f[0]
		e := entry{path: root + f[1]}
		if st, err := os.Stat(e.path); err != nil || !st.Mode().IsRegular() {
			continue
		}
		log, _ := os.ReadFile(filepath.Join(out, n+".log"))
		failed, k := map[int]bool{}, 0
		for _, l := range strings.Split(string(log), "\n") {
			switch {
			case strings.Contains(l, "Dumping VBA project from dir"):
				k++
			case strings.Contains(l, "Failed to read dir from") && k > 0:
				failed[k] = true
			}
		}
		for i := 1; ; i++ {
			b, err := os.ReadFile(filepath.Join(out, n+".dump"+itoa(i)))
			if err != nil {
				break
			}
			if !failed[i] {
				h := md5hex(b) + ":" + itoa(len(b))
				e.dumps = append(e.dumps, h)
				hashes.WriteString(h + ":Dump." + h[:32] + "\n")
			}
		}
		for i := 1; ; i++ {
			b, err := os.ReadFile(filepath.Join(out, n+".mod"+itoa(i)))
			if err != nil {
				break
			}
			if s := upperStretch(b); s != nil {
				name := "Mod." + n + "." + itoa(i)
				e.modules = append(e.modules, name)
				moduleRules[name] = name + ":2:*:" + hex.EncodeToString(s) + "\n"
			}
		}
		entries = append(entries, e)
	}
	check := func(when string) {
		for j := range entries {
			if !strings.HasPrefix(entries[j].path, root) {
				t.Fatalf("memory changed %s: entry %d of %d %q", when, j, len(entries), entries[j].path[:min(40, len(entries[j].path))])
			}
		}
	}
	check("after reading")
	if len(entries) > 444 {
		t.Logf("entry 444: %q", entries[444].path)
	}
	byHash := vbaRuleEngine(t, "dumps.hdb", hashes.String())
	check("after compiling")
	dumpsOK, dumpsBad, modsOK, modsBad, unknown, malformed := 0, 0, 0, 0, 0, 0
	for i, e := range entries {
		for j := range entries {
			if !strings.HasPrefix(entries[j].path, root) {
				t.Fatalf("memory changed before entry %d (%s): entry %d", i, entries[i-1].path, j)
			}
		}
		data, err := os.ReadFile(e.path)
		if err != nil {
			t.Fatal(err)
		}
		if len(e.dumps) > 0 {
			_, sig, err := byHash.Scan(context.Background(), data)
			switch {
			case strings.HasPrefix(sig, "Dump."):
				dumpsOK++
			case unknownReasonOf(err) == signatures.UnknownVBA:
				unknown++
				t.Logf("%s: dump undecided", e.path)
			case errors.Is(err, ErrMalformed):
				// A container this engine rejects (incomplete, never clean).
				malformed++
				t.Logf("%s: container rejected", e.path)
			default:
				dumpsBad++
				t.Errorf("%s: (%q, %v), want one of %v", e.path, sig, err, e.dumps)
			}
		}
		if len(e.modules) > 0 {
			// Each module separately: the scan stops at the first match.
			for _, m := range e.modules {
				e1 := vbaRuleEngine(t, "m.ndb", moduleRules[m])
				_, sig, err := e1.Scan(context.Background(), data)
				if sig == m {
					modsOK++
				} else {
					modsBad++
					t.Logf("%s: module %s not matched (%q, %v)", e.path, m, sig, err)
				}
			}
		}
	}
	// Module texts the reference's reader took from other offsets than the
	// directory's (one module of 60279.doc) are reported, not failed.
	t.Logf("dumps: %d same, %d differ, %d undecided, %d in rejected containers; module texts: %d matched, %d not", dumpsOK, dumpsBad, unknown, malformed, modsOK, modsBad)
}

// upperStretch is a 40-byte stretch of b with an upper-case letter and no
// line end, or nil.
func upperStretch(b []byte) []byte {
	for i := 0; i+40 <= len(b); i++ {
		s := b[i : i+40]
		if strings.ContainsAny(string(s), "\r\n") {
			continue
		}
		for _, c := range s {
			if 'A' <= c && c <= 'Z' {
				return s
			}
		}
	}
	return nil
}

func md5hex(b []byte) string {
	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}

// Decompression work the reference discards (a window overwritten by
// copies) is charged like kept output.
func TestVBAInflateChargesDiscardedWork(t *testing.T) {
	// One literal, then copies of 4098 bytes that never flush a window.
	c := []byte{1, 0, 0xb0, 0xfe, 'A'}
	for i := 0; i < 7; i++ {
		c = append(c, 0xff, 0x0f)
	}
	budget := int64(1 << 20)
	out, err := vbaInflate(context.Background(), c, &budget)
	if err != nil || len(out) > 4096 || budget > 1<<20-7*4098 {
		t.Fatalf("out %d err %v budget %d", len(out), err, budget)
	}
	budget = 4000
	if _, err := vbaInflate(context.Background(), c, &budget); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzVBAProjectDump(f *testing.F) {
	files, _ := vbaOracleFiles()
	for name, data := range files {
		if strings.HasPrefix(name, "p3/") {
			f.Add(data)
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		streams := map[string][]byte{"A": data, "B": data}
		budget := int64(1 << 20)
		_, _ = vbaProjectDump(context.Background(), data, streams, &budget)
		budget = 1 << 20
		_, _ = vbaInflate(context.Background(), data, &budget)
		_, _ = biffPictures(data)
		_ = vbaNormalize(data)
	})
}

// vbaOracleFiles reads the oracle archive without a test (for fuzz seeds).
func vbaOracleFiles() (map[string][]byte, error) {
	z, err := zip.OpenReader("testdata/vbaproject/oracle.zip")
	if err != nil {
		return nil, err
	}
	defer z.Close()
	files := map[string][]byte{}
	for _, f := range z.File {
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		files[f.Name], _ = io.ReadAll(r)
		r.Close()
	}
	return files, nil
}

// Values of the Windows reference's conversion (its dumps of the codepage
// probes): bytes without a mapping take best-fit characters instead of
// failing, as they do on Debian.
func TestVBAConvertAsWindowsReference(t *testing.T) {
	cases := []struct {
		cp    uint16
		in    string
		want  string
		known bool
	}{
		{1252, "\x80\x81", "€\u0081", true},
		{10000, "\xc6", "∆", true},
		{28593, "\xa5", "", true},
		{932, "\x81\x60\xa1", "～｡", true},
		{932, "\x81\x7f", "・", true},
		{949, "\x81\x40", "?", true},
		{932, "\x81", "", false},     // a lead byte without its trail
		{932, "\x81\x22", "", false}, // a trail outside 0x40-0xFE
		{0, "abc", "abc", true},      // the machine's codepage: ASCII only
		{0, "\x80", "", false},
		{1200, "ab", "", false},       // not decided here
		{65001, "\xff", "\xff", true}, // copied unchanged
		{20127, "\x80", "\x80", true},
	}
	for _, c := range cases {
		out, ok, known := vbaConvert([]byte(c.in), c.cp)
		if known != c.known || known && (!ok || string(out) != c.want) {
			t.Errorf("%d %q: %q ok=%v known=%v", c.cp, c.in, out, ok, known)
		}
	}
}
