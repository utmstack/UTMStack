// edr/internal/feeds/normalize_test.go
package feeds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func gzBytes(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return b.Bytes()
}

func tarGzBytes(t *testing.T, members map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	tw := tar.NewWriter(gw)
	for name, content := range members {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gw.Close()
	return b.Bytes()
}

func gunzipLines(t *testing.T, artifact []byte) string {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(artifact))
	if err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestNormalizeTarGzMergesDedupesSorts(t *testing.T) {
	in := tarGzBytes(t, map[string]string{
		"part1": "9.9.9.9\n1.1.1.1\n# comment\n",
		"part2": "1.1.1.1\n2.2.2.2\n\n",
	})
	artifact, n, err := Normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzipLines(t, artifact); got != "1.1.1.1\n2.2.2.2\n9.9.9.9\n" {
		t.Fatalf("artifact lines = %q", got)
	}
	if n != 3 {
		t.Fatalf("count = %d, want 3", n)
	}
}

func TestNormalizePlainGzipListPassesThrough(t *testing.T) {
	in := gzBytes(t, "b.example.com\na.example.com\na.example.com\n")
	artifact, n, err := Normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzipLines(t, artifact); got != "a.example.com\nb.example.com\n" {
		t.Fatalf("artifact lines = %q", got)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
}

func TestNormalizeRejectsNonGzip(t *testing.T) {
	if _, _, err := Normalize([]byte("not gzip")); err == nil {
		t.Fatal("expected error for non-gzip input")
	}
}

func TestExtractValuesMergesDedupesSorts(t *testing.T) {
	in := tarGzBytes(t, map[string]string{
		"part1": "9.9.9.9\n1.1.1.1\n# comment\n",
		"part2": "1.1.1.1\n2.2.2.2\n\n",
	})
	vals, err := extractValues(in)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"1.1.1.1", "2.2.2.2", "9.9.9.9"}
	if len(vals) != len(want) {
		t.Fatalf("values = %v, want %v", vals, want)
	}
	for i := range want {
		if vals[i] != want[i] {
			t.Fatalf("values = %v, want %v", vals, want)
		}
	}
}

func TestParseGzipValuesRoundTripsGzipValues(t *testing.T) {
	// gzipValues(vals) then parseGzipValues must recover the same values, and
	// the gzip bytes must be exactly what Normalize emits.
	vals := []string{"1.1.1.1", "2.2.2.2", "9.9.9.9"}
	artifact := gzipValues(vals)
	got, err := parseGzipValues(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(vals) {
		t.Fatalf("round-trip = %v, want %v", got, vals)
	}
	for i := range vals {
		if got[i] != vals[i] {
			t.Fatalf("round-trip = %v, want %v", got, vals)
		}
	}
}

func TestParseGzipValuesRejectsNonGzip(t *testing.T) {
	if _, err := parseGzipValues([]byte("not gzip")); err == nil {
		t.Fatal("expected error for non-gzip input")
	}
}
