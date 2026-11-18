package signatures

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// Oracle archives contain harmless images and ClamAV 1.5.4 hashes.
func loadImageOracleCases(t *testing.T, path string) ([]imageCase, map[string][]byte) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	files := map[string][]byte{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag != tar.TypeReg {
			t.Fatalf("unexpected tar entry %q", h.Name)
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		files[h.Name] = b
	}
	var cases []imageCase
	for _, line := range strings.Split(strings.TrimSpace(string(files["hashes.tsv"])), "\n") {
		name, hash, ok := strings.Cut(line, "\t")
		if !ok || len(files[name]) == 0 {
			t.Fatalf("missing TIFF or reference hash for %q", name)
		}
		cases = append(cases, imageCase{name, files[name], hash})
	}
	return cases, files
}

// The small corpus covers compression, color and byte-order modes with
// exact ClamAV fuzzy-hash matches.
func TestTIFFImageHashAgreesWithReference(t *testing.T) {
	cases, files := loadImageOracleCases(t, "testdata/tiff_oracle.tar.gz")
	if len(cases) != 14 {
		t.Fatalf("got %d TIFF fixtures, want 14", len(cases))
	}
	for _, c := range cases {
		if !fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Errorf("%s was not decoded", c.name)
		}
	}
	checkImageCases(t, cases, 1)
	flat := compileTest(t, "ldb", "TIFF.Flat;Engine:150-255,Target:0;0;fuzzy_img#8000000000000000#0\n")
	for _, name := range []string{"flat.tiff", "flat-lzw.tiff", "flat-deflate.tiff", "flat-packbits.tiff", "gray.tiff"} {
		got, err := matchObject(t, flat, string(files[name]), MatchContext{Type: TypeTIFF})
		if got != "TIFF.Flat" || err != nil {
			t.Errorf("%s: match %q, %v", name, got, err)
		}
	}
}

// The varied corpus guards against an implementation tuned to flat and
// gradient examples. It includes noise, rectangles and different dimensions.
func TestTIFFVariedImageHashes(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/tiff_oracle_varied.tar.gz")
	if len(cases) != 128 {
		t.Fatalf("got %d TIFF fixtures, want 128", len(cases))
	}
	for _, c := range cases {
		if !fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
			t.Errorf("%s was not decoded", c.name)
		}
	}
	checkImageCases(t, cases, 0.9)
}

// The archived hashes are ClamAV 1.5.4 sigtool outputs. Exercise the full
// fuzzy rule decision for every image, including a one-bit nonmatching rule;
// equality of a computed hash alone would not prove a scanner verdict.
func TestTIFFOracleFuzzyRulesDecisive(t *testing.T) {
	var cases []imageCase
	for _, name := range []string{"testdata/tiff_oracle.tar.gz", "testdata/tiff_oracle_varied.tar.gz"} {
		part, _ := loadImageOracleCases(t, name)
		cases = append(cases, part...)
	}
	if len(cases) != 142 {
		t.Fatalf("got %d TIFF cases, want 142", len(cases))
	}
	unique := map[uint64]bool{}
	for _, c := range cases {
		hash, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		ih := fuzzyImageHash(c.data, &matchWork{ctx: context.Background(), limit: 1 << 40})
		if !ih.ok || ih.hash != hash || ih.uncertain != 0 {
			t.Fatalf("%s: hash=%016x uncertain=%016x ok=%t; reference=%016x", c.name, ih.hash, ih.uncertain, ih.ok, hash)
		}
		unique[hash] = true
	}
	hashes := make([]uint64, 0, len(unique))
	for hash := range unique {
		hashes = append(hashes, hash)
	}
	sort.Slice(hashes, func(i, j int) bool { return hashes[i] < hashes[j] })
	labels := map[uint64]string{}
	var positives, negatives strings.Builder
	for i, hash := range hashes {
		if unique[hash^1] {
			t.Fatalf("one-bit negative collides with corpus hash %016x", hash)
		}
		name := fmt.Sprintf("TIFF.Oracle.%d", i)
		labels[hash] = name
		fmt.Fprintf(&positives, "%s;Engine:150-255,Target:0;0;fuzzy_img#%016x#0\n", name, hash)
		fmt.Fprintf(&negatives, "TIFF.Nonmatch.%d;Engine:150-255,Target:0;0;fuzzy_img#%016x#0\n", i, hash^1)
	}
	positive := compileTest(t, "ldb", positives.String())
	negative := compileTest(t, "ldb", negatives.String())
	for _, c := range cases {
		hash, _ := strconv.ParseUint(c.hash, 16, 64)
		got, err := matchObject(t, positive, string(c.data), MatchContext{Type: TypeTIFF})
		if err != nil || got != labels[hash] {
			t.Errorf("%s: positive = %q, %v; want %q", c.name, got, err, labels[hash])
		}
		got, err = matchObject(t, negative, string(c.data), MatchContext{Type: TypeTIFF})
		if err != nil || got != "" {
			t.Errorf("%s: one-bit negative = %q, %v", c.name, got, err)
		}
	}
}
