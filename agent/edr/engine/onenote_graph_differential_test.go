package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The pinned corpus changes one byte at each recorded offset of the official
// web export. It checks the graph, attachment traversal, and final scanner
// verdict against ClamAV 1.5.4 rather than just parser acceptance.
func TestOneNoteGraphMutationDifferential(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_DIFF_DIR")
	fixtureDir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" || fixtureDir == "" {
		t.Skip("set fixture and differential directories")
	}
	oneNoteMutationDifferential(t, dir, fixtureDir, "manifest.json", "reference.txt", 100, 51, 51, 12, 0)
}

func TestOneNoteExtendedMutationDifferential(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_EXTENDED_DIFF_DIR")
	fixtureDir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" || fixtureDir == "" {
		t.Skip("set fixture and extended differential directories")
	}
	oneNoteMutationDifferential(t, dir, fixtureDir, "extended-manifest.json", "extended-reference.txt", 400, 218, 218, 65, 13)
}

func TestOneNoteWholeFileMutationDifferential(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_WHOLE_DIFF_DIR")
	fixtureDir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" || fixtureDir == "" {
		t.Skip("set fixture and whole-file differential directories")
	}
	oneNoteMutationDifferential(t, dir, fixtureDir, "manifest.json", "reference.txt", 1000, 846, 846, 44, 11)
}

func TestOneNoteWebReferenceAndFramingRecovery(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_ONENOTE_DIR")
	if dir == "" {
		t.Skip("set official OneNote fixture directory")
	}
	orig, err := os.ReadFile(filepath.Join(dir, "webapp.one"))
	if err != nil {
		t.Fatal(err)
	}
	base := pipelineState(t, map[string]string{
		"child.hdb": "aa15bcf478d165efd2065190eb473bcb:544:OneNote.Child\n",
	})
	base.pePermissive = true
	for _, offset := range []int{28678, 32928, 34876, 35611, 36383, 43714} {
		t.Run(strconv.Itoa(offset), func(t *testing.T) {
			mutated := bytes.Clone(orig)
			mutated[offset] ^= 1
			scan := *base
			sig, err := scan.visit(mutated, 0)
			if sig != "OneNote.Child" || err != nil {
				t.Fatalf("bounded OneStore recovery: %q, %v", sig, err)
			}
		})
	}
	mutated := bytes.Clone(orig)
	mutated[39117] ^= 1 // odd-length CachedTitleString in a page metadata object
	scan := *base
	if sig, err := scan.visit(mutated, 0); sig != "" || err == nil {
		t.Fatalf("malformed page title produced a detection or clean verdict: %q, %v", sig, err)
	}
}

func oneNoteMutationDifferential(t *testing.T, dir, fixtureDir, manifestName, referenceName string, expectedTotal, expectedDetected, expectedGraphDetected, expectedGraphClean, expectedSemanticClean int) {
	t.Helper()
	var manifest struct {
		Files []struct {
			Path   string `json:"path"`
			SHA256 string `json:"sha256"`
		} `json:"files"`
	}
	raw, err := os.ReadFile(filepath.Join(dir, manifestName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	ref, err := os.ReadFile(filepath.Join(dir, referenceName))
	if err != nil {
		t.Fatal(err)
	}
	states := map[string]bool{}
	for _, line := range strings.Split(string(ref), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			t.Fatal(line)
		}
		states[filepath.Base(parts[0])] = strings.Contains(parts[1], "FOUND")
	}
	orig, err := os.ReadFile(filepath.Join(fixtureDir, "webapp.one"))
	if err != nil {
		t.Fatal(err)
	}
	base := pipelineState(t, map[string]string{
		"child.hdb": "aa15bcf478d165efd2065190eb473bcb:544:OneNote.Child\n",
	})
	base.pePermissive = true // reference clamscan did not enable --alert-broken
	var graphPositive, graphClean, semanticPositive, semanticClean, scannerPositive, scannerClean int
	var mismatches []string
	for _, file := range manifest.Files {
		name := strings.TrimSuffix(strings.TrimPrefix(file.Path, "flip-"), ".one")
		offset, err := strconv.Atoi(name)
		if err != nil || offset < 0 || offset >= len(orig) {
			t.Fatalf("invalid mutation offset %q: %v", file.Path, err)
		}
		b := bytes.Clone(orig)
		b[offset] ^= 1
		if fmt.Sprintf("%x", sha256.Sum256(b)) != file.SHA256 {
			t.Fatalf("hash drift %s", file.Path)
		}
		want, ok := states[file.Path]
		if !ok {
			t.Fatalf("missing reference verdict for %s", file.Path)
		}
		graphOK, semanticOK := false, false
		pos, ok := oneNoteFSSHTTPB(b)
		if ok {
			if tree, err := fssPackageTree(context.Background(), b, pos); err == nil {
				if graph, err := parseFSSGraph(tree); err == nil {
					if _, err := graph.reachableBlobs(); err == nil {
						graphOK = true
						if want {
							graphPositive++
						} else {
							graphClean++
						}
						if candidates, err := oneNoteWebCandidates(context.Background(), graph); err == nil && len(candidates) > 0 {
							semanticOK = true
							if want {
								semanticPositive++
							} else {
								semanticClean++
							}
						}
					}
				}
			}
		}
		scan := *base // scan-wide limits and ancestry reset for each corpus file
		sig, scanErr := scan.visit(b, 0)
		if sig != "" && sig != "OneNote.Child" {
			mismatches = append(mismatches, file.Path)
			t.Logf("unexpected scanner signature %s: %q reference-detected=%v graph=%v semantic=%v error=%v", file.Path, sig, want, graphOK, semanticOK, scanErr)
			t.Logf("mutation location %s: %s", file.Path, oneNoteMutationLocation(b, offset))
			continue
		}
		if sig != "" {
			if want {
				scannerPositive++
			} else {
				scannerClean++
			}
		}
		if (sig != "") != want {
			mismatches = append(mismatches, file.Path)
			t.Logf("scanner/reference mismatch %s: signature=%q reference-detected=%v graph=%v semantic=%v error=%v", file.Path, sig, want, graphOK, semanticOK, scanErr)
			t.Logf("mutation location %s: %s", file.Path, oneNoteMutationLocation(b, offset))
		}
	}
	t.Logf("OneNote differential: graph detected/clean %d/%d, semantic %d/%d, scanner %d/%d", graphPositive, graphClean, semanticPositive, semanticClean, scannerPositive, scannerClean)
	if len(mismatches) != 0 || len(manifest.Files) != expectedTotal || len(states) != expectedTotal || expectedGraphDetected >= 0 && graphPositive != expectedGraphDetected || expectedGraphClean >= 0 && graphClean != expectedGraphClean || semanticPositive != expectedDetected || expectedSemanticClean >= 0 && semanticClean != expectedSemanticClean || scannerPositive != expectedDetected || scannerClean != 0 {
		t.Fatalf("OneNote differential: graph detected/clean %d/%d, semantic %d/%d, scanner %d/%d, corpus %d/%d", graphPositive, graphClean, semanticPositive, semanticClean, scannerPositive, scannerClean, len(manifest.Files), len(states))
	}
}

func oneNoteMutationLocation(data []byte, offset int) string {
	pos, ok := oneNoteFSSHTTPB(data)
	if !ok {
		return "web envelope"
	}
	tree, err := fssPackageTree(context.Background(), data, pos)
	if err != nil {
		return fmt.Sprintf("package parse: %v", err)
	}
	for ei, elem := range tree.children {
		w := fssWire{data: elem.body}
		_, valid := w.id()
		if !valid || !w.serial() {
			continue
		}
		kind, valid := w.compact()
		if !valid || kind != 5 {
			continue
		}
		group, err := parseFSSGroup(elem)
		if err != nil {
			return fmt.Sprintf("element=%d object group: %v", ei, err)
		}
		dataNode := elem.children[len(elem.children)-1]
		for pi, child := range dataNode.children {
			if offset < child.start || offset >= child.end {
				continue
			}
			part := group.parts[pi]
			var jcid uint32
			for _, p := range group.parts {
				if p.id == part.id && p.partition == 4 && len(p.data) >= 4 {
					jcid = binary.LittleEndian.Uint32(p.data)
				}
			}
			return fmt.Sprintf("element=%d part=%d partition=%d JCID=%08x bodyOffset=%d dataOffset=%d data=%x", ei, pi, part.partition, jcid, offset-child.start, offset-part.dataStart, part.data[:min(len(part.data), 120)])
		}
		continue
	}
	return "outside parsed elements"
}
