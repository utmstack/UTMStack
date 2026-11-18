package engine

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestHTMLNormalizationAgreesWithReference compares the normalized views of
// testdata/js_oracle.tar.gz with the reference's: 2,247 short pages made by
// the workspace generators docs/2026-09-27-html-oracle/mkhtml.py (tags,
// attributes, entities, comments, data: URIs, styles, small scripts) and
// mkjs.py (random script token soups: names, keywords, unbalanced brackets,
// good and bad escapes, large numbers, comments, unescape and eval calls),
// with the reference's files for each.
func TestHTMLNormalizationAgreesWithReference(t *testing.T) {
	f, err := os.Open("testdata/js_oracle.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(tr)
		path := filepath.Join(dir, filepath.FromSlash(h.Name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	compareHTMLOracle(t, dir)
}

// TestHTMLNormalizationOracle runs the same comparison on a larger folder:
// EDR_HTML_ORACLE names a folder with hin/NNNNN.bin (inputs) and
// hres/NNNNN/ (the reference's nocomment.html, notags.html, javascript and
// rfc2397/ files for that input, kept by its scanner's leave-temps option).
// The workspace script docs/2026-09-27-html-oracle/htmloracle.sh collects
// them.
func TestHTMLNormalizationOracle(t *testing.T) {
	dir := os.Getenv("EDR_HTML_ORACLE")
	if dir == "" {
		t.Skip("EDR_HTML_ORACLE not set")
	}
	compareHTMLOracle(t, dir)
}

// compareHTMLOracle compares every input of dir; pages whose scripts this
// engine reports as not reproduced are counted, not compared.
func compareHTMLOracle(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "hres"))
	if err != nil {
		t.Fatal(err)
	}
	type result struct{ same, differ int }
	views := map[string]*result{}
	var examples []string
	skipped := 0
	for _, e := range entries {
		in, err := os.ReadFile(filepath.Join(dir, "hin", e.Name()+".bin"))
		if err != nil {
			t.Fatal(err)
		}
		// UTF-16 pages (a byte-order mark here) are converted first, as
		// the scan does for objects typed HTML_UTF16.
		if bytes.HasPrefix(in, []byte{0xff, 0xfe}) || bytes.HasPrefix(in, []byte{0xfe, 0xff}) || len(in) > 2 && in[0] == '<' && in[1] == 0 && in[2] != 0 {
			in = htmlUTF16ToASCII(in)
		}
		got := normalizeHTML(in)
		if got.unsupported != "" {
			continue
		}
		js, unsupported := normalizeJS(got.scripts)
		if unsupported {
			skipped++
			continue
		}
		for _, v := range []struct {
			name string
			data []byte
		}{{"nocomment.html", got.noComment}, {"notags.html", got.noTags}, {"javascript", js}} {
			want, err := os.ReadFile(filepath.Join(dir, "hres", e.Name(), v.name))
			if err != nil {
				want = nil
			}
			r := views[v.name]
			if r == nil {
				r = &result{}
				views[v.name] = r
			}
			if bytes.Equal(want, v.data) {
				r.same++
				continue
			}
			r.differ++
			if len(examples) < 12 && (os.Getenv("ONLY") == "" || os.Getenv("ONLY") == v.name) {
				k := 0
				for k < len(want) && k < len(v.data) && want[k] == v.data[k] {
					k++
				}
				lo := max(0, k-40)
				examples = append(examples, e.Name()+" "+v.name+": at "+itoa(k)+"\n    want "+quote(want[lo:min(len(want), k+60)])+"\n    got  "+quote(v.data[lo:min(len(v.data), k+60)]))
			}
		}
	}
	// The data: URI files: the same contents (their names are random).
	rfcSame, rfcDiffer := 0, 0
	for _, e := range entries {
		in, _ := os.ReadFile(filepath.Join(dir, "hin", e.Name()+".bin"))
		if bytes.HasPrefix(in, []byte{0xff, 0xfe}) || bytes.HasPrefix(in, []byte{0xfe, 0xff}) || len(in) > 2 && in[0] == '<' && in[1] == 0 && in[2] != 0 {
			in = htmlUTF16ToASCII(in)
		}
		got := normalizeHTML(in)
		var want []string
		files, _ := os.ReadDir(filepath.Join(dir, "hres", e.Name(), "rfc2397"))
		for _, f := range files {
			b, _ := os.ReadFile(filepath.Join(dir, "hres", e.Name(), "rfc2397", f.Name()))
			want = append(want, string(b))
		}
		var have []string
		for _, r := range got.rfc2397 {
			have = append(have, string(r.compatibility))
		}
		sort.Strings(want)
		sort.Strings(have)
		if strings.Join(want, "\x00|") == strings.Join(have, "\x00|") {
			rfcSame++
		} else {
			rfcDiffer++
			if len(examples) < 20 {
				examples = append(examples, e.Name()+" rfc2397: want "+itoa(len(want))+" got "+itoa(len(have)))
			}
		}
	}
	t.Logf("rfc2397: %d identical, %d different; %d pages with scripts left incomplete", rfcSame, rfcDiffer, skipped)
	if rfcDiffer > 0 {
		t.Fail()
	}
	names := make([]string, 0, len(views))
	for n := range views {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		t.Logf("%s: %d identical, %d different", n, views[n].same, views[n].differ)
	}
	t.Log(strings.Join(examples, "\n"))
	for _, n := range names {
		if views[n].differ > 0 {
			t.Fail()
		}
	}
}

func itoa(n int) string {
	return strings.TrimSpace(strings.Repeat(" ", 0) + string([]byte(fmtInt(n))))
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func quote(b []byte) string {
	var s strings.Builder
	s.WriteByte('"')
	for _, c := range b {
		switch {
		case c == '"' || c == '\\':
			s.WriteByte('\\')
			s.WriteByte(c)
		case c >= 32 && c < 127:
			s.WriteByte(c)
		default:
			s.WriteString("\\x")
			s.WriteByte("0123456789abcdef"[c>>4])
			s.WriteByte("0123456789abcdef"[c&15])
		}
	}
	s.WriteByte('"')
	return s.String()
}
