package engine

import (
	"archive/zip"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
)

// workbookStreams returns the streams named Workbook of a compound file
// (an embedded workbook has its own).
func workbookStreams(t *testing.T, data []byte) [][]byte {
	t.Helper()
	s := oleTestState(t)
	f, err := readOLE(s, data)
	if err != nil {
		return nil
	}
	var out [][]byte
	for _, st := range f.streams {
		b, err := f.read(st)
		if err == nil && st.name == "workbook" {
			out = append(out, b)
		}
	}
	return out
}

// testdata/vbaproject/pictures.zip holds generated workbooks
// (docs/2026-10-08-macro-oracle/mkxls.py) with the sizes of the pictures the
// reference extracted and whether its image-target rule matched one.
func TestWorkbookPicturesAsReference(t *testing.T) {
	z, err := zip.OpenReader("testdata/vbaproject/pictures.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	files := map[string][]byte{}
	for _, f := range z.File {
		r, _ := f.Open()
		files[f.Name], _ = io.ReadAll(r)
		r.Close()
	}
	var manifest map[string]struct {
		Sizes     []int `json:"sizes"`
		ImageRule bool  `json:"image_rule"`
	}
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	e := vbaRuleEngine(t, "p.ndb", "Pic.T5:5:*:"+hex.EncodeToString([]byte("UTMSTACK_PICTURE_MARKER"))+"\n")
	for name, want := range manifest {
		var sizes []int
		for _, wb := range workbookStreams(t, files[name]) {
			pictures, decided := biffPictures(wb)
			if !decided {
				t.Errorf("%s: undecided", name)
			}
			for _, p := range pictures {
				sizes = append(sizes, len(p))
			}
		}
		if !slices.Equal(sizes, want.Sizes) && !(len(sizes) == 0 && len(want.Sizes) == 0) {
			t.Errorf("%s: pictures %v, reference %v", name, sizes, want.Sizes)
		}
		_, sig, _ := e.Scan(context.Background(), files[name])
		if (sig == "Pic.T5") != want.ImageRule {
			t.Errorf("%s: image rule %q, reference matched=%v", name, sig, want.ImageRule)
		}
	}
	if len(manifest) != 48 {
		t.Fatal(len(manifest))
	}
}

// TestWorkbookPicturesCorpus compares the pictures extracted from real
// workbooks with the sizes in the reference's debug logs (EDR_VBA_ORACLE,
// EDR_VBA_ORACLE_ROOT as for TestVBAOracleCorpus).
func TestWorkbookPicturesCorpus(t *testing.T) {
	out, root := os.Getenv("EDR_VBA_ORACLE"), os.Getenv("EDR_VBA_ORACLE_ROOT")
	if out == "" {
		t.Skip("EDR_VBA_ORACLE not set")
	}
	index, err := os.ReadFile(filepath.Join(out, "index"))
	if err != nil {
		t.Fatal(err)
	}
	sizeRE := regexp.MustCompile(`Scanning extracted image of size (\d+)`)
	same, differ, undecided := 0, 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(index)), "\n") {
		f := strings.Split(line, "\t")
		if len(f) != 3 || !strings.HasSuffix(strings.ToLower(f[1]), ".xls") {
			continue
		}
		data, err := os.ReadFile(root + f[1])
		if err != nil {
			continue
		}
		log, _ := os.ReadFile(filepath.Join(out, f[0]+".log"))
		var want []string
		for _, m := range sizeRE.FindAllStringSubmatch(string(log), -1) {
			want = append(want, m[1])
		}
		streams := workbookStreams(t, data)
		if streams == nil {
			continue
		}
		var got []string
		decided := true
		for _, wb := range streams {
			pictures, ok := biffPictures(wb)
			decided = decided && ok
			for _, p := range pictures {
				got = append(got, itoa(len(p)))
			}
		}
		// An embedded workbook can reach the reference twice (as a storage
		// and as an extracted object): compare distinct sizes.
		sort.Strings(got)
		sort.Strings(want)
		got, want = slices.Compact(got), slices.Compact(want)
		switch {
		case !decided:
			undecided++
		case strings.Join(got, ",") == strings.Join(want, ","):
			same++
		default:
			differ++
			t.Errorf("%s: pictures %v, reference %v", f[1], got, want)
		}
	}
	t.Logf("workbooks: %d same, %d differ, %d undecided", same, differ, undecided)
}
