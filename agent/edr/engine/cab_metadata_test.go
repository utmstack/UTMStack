package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func referenceCABMetadata(t *testing.T, data []byte, rules map[string]string, want string) []byte {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return nil
	}
	tool, err := os.ReadFile(filepath.Join(ref, "bin", "clamscan"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("unpinned ClamAV 1.5.4 scanner: %v", err)
	}
	dir := t.TempDir()
	db := filepath.Join(dir, "db")
	if err := os.Mkdir(db, 0700); err != nil {
		t.Fatal(err)
	}
	for name, text := range rules {
		if err := os.WriteFile(filepath.Join(db, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(dir, "control.cab")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+db, "--debug", "--no-summary", path)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if want != "" {
		status, ok := err.(*exec.ExitError)
		if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), "control.cab: "+want+".UNOFFICIAL FOUND") {
			t.Fatalf("reference CAB metadata: %v: %s", err, out)
		}
	} else if err != nil || !strings.Contains(string(out), "control.cab: OK") {
		t.Fatalf("reference CAB negative: %v: %s", err, out)
	}
	t.Logf("ClamAV input SHA256=%x signature=%q", sha256.Sum256(data), want)
	return out
}

func TestCABMetadataKnownMiss(t *testing.T) {
	parent := cabQuantumControl(t)
	rules := map[string]string{"meta.cdb": "Metadata.Quantum:CL_TYPE_MSCAB:*:^qtm[.]txt$:0:59:0:2:*:\n"}
	referenceCABMetadata(t, parent, rules, "Metadata.Quantum")
	if sig, err := pipelineState(t, rules).visit(parent, 0); sig != "Metadata.Quantum" || err != nil {
		t.Fatalf("CAB member metadata: signature=%q error=%v", sig, err)
	}
}

func TestCABMetadataFieldsAndControls(t *testing.T) {
	parent := cabQuantumControl(t)
	for _, tc := range []struct{ name, fields, want string }{
		{"all-fields", fmt.Sprintf("CL_TYPE_MSCAB:%d:^qtm[.]txt$:0:59:0:2:*:", len(parent)), "Meta"},
		{"first-index-zero", "CL_TYPE_MSCAB:*:^mszip[.]txt$:0:57:0:0:*:", "Meta"},
		{"middle-index-one", "CL_TYPE_MSCAB:*:^lzx[.]txt$:0:187:0:1:*:", "Meta"},
		{"name-search", "CL_TYPE_MSCAB:*:qtm:0:59:0:2:*:", "Meta"},
		{"wrong-name", "CL_TYPE_MSCAB:*:^qtm[.]exe$:0:59:0:2:*:", ""},
		{"wrong-size", "CL_TYPE_MSCAB:*:^qtm[.]txt$:0:60:0:2:*:", ""},
		{"wrong-compressed-size", "CL_TYPE_MSCAB:*:^qtm[.]txt$:48:59:0:2:*:", ""},
		{"wrong-encryption", "CL_TYPE_MSCAB:*:^qtm[.]txt$:0:59:1:2:*:", ""},
		{"wrong-position", "CL_TYPE_MSCAB:*:^qtm[.]txt$:0:59:0:3:*:", ""},
		{"wrong-container", "CL_TYPE_ZIP:*:^qtm[.]txt$:0:59:0:2:*:", ""},
		{"wrong-archive-size", fmt.Sprintf("CL_TYPE_MSCAB:%d:^qtm[.]txt$:0:59:0:2:*:", len(parent)+1), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules := map[string]string{"meta.cdb": "Meta:" + tc.fields + "\n"}
			referenceCABMetadata(t, parent, rules, tc.want)
			if sig, err := pipelineState(t, rules).visit(parent, 0); sig != tc.want || err != nil {
				t.Fatalf("CAB fields: signature=%q error=%v, want %q", sig, err, tc.want)
			}
		})
	}
}

type cabMetadataFile struct {
	name        string
	folder      uint16
	start, size uint32
}

// Independent stored-folder writer: file records may interleave folder IDs.
func cabMetadataFixture(t *testing.T, bodies [][]byte, methods []uint16, files []cabMetadataFile, flags uint16) []byte {
	t.Helper()
	le := binary.LittleEndian
	b := make([]byte, 36)
	copy(b, "MSCF")
	b[24], b[25] = 3, 1
	le.PutUint16(b[26:], uint16(len(bodies)))
	le.PutUint16(b[28:], uint16(len(files)))
	le.PutUint16(b[30:], flags)
	if flags&1 != 0 {
		b = append(b, []byte("previous.cab\x00\x00")...)
	}
	if flags&2 != 0 {
		b = append(b, []byte("next.cab\x00\x00")...)
	}
	folderStart := len(b)
	b = append(b, make([]byte, 8*len(bodies))...)
	le.PutUint32(b[16:], uint32(len(b)))
	for _, file := range files {
		r := make([]byte, 16)
		le.PutUint32(r, file.size)
		le.PutUint32(r[4:], file.start)
		le.PutUint16(r[8:], file.folder)
		b = append(b, r...)
		b = append(b, file.name...)
		b = append(b, 0)
	}
	for i, body := range bodies {
		h := b[folderStart+8*i:]
		le.PutUint32(h, uint32(len(b)))
		le.PutUint16(h[4:], 1)
		le.PutUint16(h[6:], methods[i])
		block := make([]byte, 8) // checksum omitted
		le.PutUint16(block[4:], uint16(len(body)))
		le.PutUint16(block[6:], uint16(len(body)))
		b = append(b, block...)
		b = append(b, body...)
	}
	le.PutUint32(b[8:], uint32(len(b)))
	return b
}

func TestCABMetadataBeforeFailedExtraction(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]byte)
	}{
		{"unsupported-codec", func(b []byte) { binary.LittleEndian.PutUint16(b[42:], 15) }},
		{"bad-data-offset", func(b []byte) { binary.LittleEndian.PutUint32(b[36:], 0xffffffff) }},
		{"broken-checksum", func(b []byte) { b[len(b)-1] ^= 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := cabFixture(t, 0, [][]byte{[]byte("harmless")}, []cabFixtureFile{{"invoice.exe", 0, 8}}, false)
			tc.mutate(b)
			rules := map[string]string{"meta.cdb": "Before.Decode:CL_TYPE_MSCAB:*:^invoice[.]exe$:0:8:0:0:*:\n"}
			referenceCABMetadata(t, b, rules, "Before.Decode")
			if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Before.Decode" || err != nil {
				t.Fatalf("metadata before extraction: signature=%q error=%v", sig, err)
			}
		})
	}
}

func TestCABMetadataAndChildFileOrder(t *testing.T) {
	a, b := []byte("first harmless child"), []byte("second harmless child")
	parent := cabMetadataFixture(t, [][]byte{a, b}, []uint16{0, 0}, []cabMetadataFile{
		{"first", 1, 0, uint32(len(b))}, {"second", 0, 0, uint32(len(a))},
	}, 0)
	for _, tc := range []struct{ name, meta, hashes, want string }{
		{"earlier-child", fmt.Sprintf("Later.Meta:CL_TYPE_MSCAB:*:^second$:0:%d:0:1:*:\n", len(a)), fmt.Sprintf("%x:%d:Earlier.Child\n", md5.Sum(b), len(b)), "Earlier.Child"},
		{"earlier-meta", fmt.Sprintf("Earlier.Meta:CL_TYPE_MSCAB:*:^first$:0:%d:0:0:*:\n", len(b)), fmt.Sprintf("%x:%d:Later.Child\n", md5.Sum(a), len(a)), "Earlier.Meta"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules := map[string]string{"meta.cdb": tc.meta, "child.hdb": tc.hashes}
			referenceCABMetadata(t, parent, rules, tc.want)
			if sig, err := pipelineState(t, rules).visit(parent, 0); sig != tc.want || err != nil {
				t.Fatalf("member order: signature=%q error=%v, want %q", sig, err, tc.want)
			}
		})
	}
}

func TestCABMetadataInterleavedFolderAccounting(t *testing.T) {
	a, b := []byte("abcdefghABCDEFGH"), []byte("unrelated harmless child")
	parent := cabMetadataFixture(t, [][]byte{a, b}, []uint16{0, 0}, []cabMetadataFile{
		{"one", 0, 0, 8}, {"two", 1, 0, uint32(len(b))}, {"three", 0, 8, 8},
	}, 0)
	rules := map[string]string{"meta.cdb": "No.Match:CL_TYPE_MSCAB:*:^absent$:*:*:*:*:*:\n"}
	referenceCABMetadata(t, parent, rules, "")
	s := pipelineState(t, rules)
	s.remaining, s.maxFiles = int64(len(parent)+len(a)+len(b)), 6
	if sig, err := s.visit(parent, 0); sig != "" || err != nil || s.files != 6 || s.remaining != 0 {
		t.Fatalf("folder reuse accounting: signature=%q error=%v files=%d remaining=%d", sig, err, s.files, s.remaining)
	}
}

func TestCABMetadataUnknownDoesNotHideChild(t *testing.T) {
	child := []byte("harmless decoded child")
	b := cabFixture(t, 1, [][]byte{child}, []cabFixtureFile{{"invoice", 0, uint32(len(child))}}, false)
	rules := map[string]string{"meta.cdb": "Unproved:CL_TYPE_MSCAB:*:\\<invoice\\>:*:*:*:*:*:\n"}
	if sig, err := pipelineState(t, rules).visit(b, 0); sig != "" || unknownReasonOf(err) != "container-name-regex" {
		t.Fatalf("undecided CAB metadata: signature=%q error=%v", sig, err)
	}
	rules["child.hdb"] = fmt.Sprintf("%x:%d:Known.Child\n", md5.Sum(child), len(child))
	if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Known.Child" || err != nil {
		t.Fatalf("unknown metadata hid child: signature=%q error=%v", sig, err)
	}
}

func TestCABMetadataLimitsAndCancellation(t *testing.T) {
	b := cabQuantumControl(t)
	rules := map[string]string{"meta.cdb": "Metadata.Quantum:CL_TYPE_MSCAB:*:^qtm[.]txt$:0:59:0:2:*:\n"}
	s := pipelineState(t, rules)
	s.remaining = int64(len(b) + 1)
	if sig, err := s.visit(b, 0); sig != "Metadata.Quantum" || err != nil {
		t.Fatalf("decode budget hid later metadata: signature=%q error=%v", sig, err)
	}
	s = pipelineState(t, rules)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.ctx = ctx
	if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled metadata: signature=%q error=%v", sig, err)
	}
	if !bytes.Equal(b, cabQuantumControl(t)) {
		t.Fatal("metadata scan changed its input")
	}
}

func TestCABMetadataSetAndSalvageEntries(t *testing.T) {
	body := []byte("harmless member")
	for _, tc := range []struct {
		name   string
		folder uint16
		flags  uint16
	}{
		{"previous-set", 0xfffd, 1},
		{"next-set", 0xfffe, 2},
		{"both-sets", 0xffff, 3},
		{"set-header-local-file", 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := cabMetadataFixture(t, [][]byte{body}, []uint16{0}, []cabMetadataFile{{"continued", tc.folder, 0, uint32(len(body))}}, tc.flags)
			rules := map[string]string{"meta.cdb": fmt.Sprintf("Set.Meta:CL_TYPE_MSCAB:*:^continued$:0:%d:0:0:*:\n", len(body))}
			referenceCABMetadata(t, b, rules, "Set.Meta")
			if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Set.Meta" || err != nil {
				t.Fatalf("metadata before unresolved set: signature=%q error=%v", sig, err)
			}
		})
	}
	t.Run("skip-invalid-folder", func(t *testing.T) {
		b := cabMetadataFixture(t, [][]byte{body}, []uint16{0}, []cabMetadataFile{{"discarded", 7, 0, 8}, {"kept", 0, 0, uint32(len(body))}}, 0)
		rules := map[string]string{"meta.cdb": fmt.Sprintf("Salvaged.Meta:CL_TYPE_MSCAB:*:^kept$:0:%d:0:0:*:\n", len(body))}
		referenceCABMetadata(t, b, rules, "Salvaged.Meta")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Salvaged.Meta" || err != nil {
			t.Fatalf("salvaged position: signature=%q error=%v", sig, err)
		}
	})
	t.Run("valid-prefix-truncated-table", func(t *testing.T) {
		b := cabMetadataFixture(t, [][]byte{body}, []uint16{0}, []cabMetadataFile{{"kept", 0, 0, uint32(len(body))}}, 0)
		b = b[:len(b)-len(body)-8]
		binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
		binary.LittleEndian.PutUint16(b[28:], 2) // only the first file record is present
		rules := map[string]string{"meta.cdb": fmt.Sprintf("Prefix.Meta:CL_TYPE_MSCAB:*:^kept$:0:%d:0:0:*:\n", len(body))}
		referenceCABMetadata(t, b, rules, "")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "" || !errors.Is(err, ErrMalformed) {
			t.Fatalf("truncated table must not dispatch its prefix: signature=%q error=%v", sig, err)
		}
	})
	t.Run("atypical-header-offset", func(t *testing.T) {
		b := cabMetadataFixture(t, [][]byte{body}, []uint16{0}, []cabMetadataFile{{"kept", 0, 0, uint32(len(body))}}, 0)
		binary.LittleEndian.PutUint32(b[16:], uint32(len(b)-1))
		rules := map[string]string{"meta.cdb": fmt.Sprintf("Nominal.Meta:CL_TYPE_MSCAB:*:^kept$:0:%d:0:0:*:\n", len(body))}
		referenceCABMetadata(t, b, rules, "Nominal.Meta")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Nominal.Meta" || err != nil {
			t.Fatalf("nominal file table: signature=%q error=%v", sig, err)
		}
	})
	t.Run("skip-invalid-name", func(t *testing.T) {
		b := cabMetadataFixture(t, [][]byte{body}, []uint16{0}, []cabMetadataFile{{strings.Repeat("x", 256), 0, 0, uint32(len(body))}, {"kept", 0, 0, uint32(len(body))}}, 0)
		// Remove the terminator beyond the invalid 256-byte first name. The
		// reference resumes at the next record after that failed bounded read.
		nameEnd := 36 + 8 + 16 + 256
		b = append(b[:nameEnd], b[nameEnd+1:]...)
		binary.LittleEndian.PutUint32(b[8:], uint32(len(b)))
		rules := map[string]string{"meta.cdb": fmt.Sprintf("NameSalvage.Meta:CL_TYPE_MSCAB:*:^kept$:0:%d:0:0:*:\n", len(body))}
		referenceCABMetadata(t, b, rules, "NameSalvage.Meta")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "NameSalvage.Meta" || err != nil {
			t.Fatalf("name salvage position: signature=%q error=%v", sig, err)
		}
	})
}

func TestCABMetadataNamesEmptyMembersAndRuleOrder(t *testing.T) {
	body := []byte("harmless")
	for _, tc := range []struct{ name, pattern string }{
		{"docs/invoice.exe", "^docs/invoice[.]exe$"},
		{`docs\invoice.exe`, `^docs[\\]invoice[.]exe$`},
		{strings.Repeat("x", 255), "^x{255}$"},
	} {
		t.Run(fmt.Sprintf("name-%d-%x", len(tc.name), md5.Sum([]byte(tc.name))), func(t *testing.T) {
			b := cabFixture(t, 0, [][]byte{body}, []cabFixtureFile{{tc.name, 0, uint32(len(body))}}, true)
			rules := map[string]string{"meta.cdb": fmt.Sprintf("Name.Meta:CL_TYPE_MSCAB:*:%s:0:%d:0:0:*:\n", tc.pattern, len(body))}
			referenceCABMetadata(t, b, rules, "Name.Meta")
			if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Name.Meta" || err != nil {
				t.Fatalf("name metadata: signature=%q error=%v", sig, err)
			}
		})
	}
	t.Run("empty-member", func(t *testing.T) {
		b := cabFixture(t, 0, [][]byte{body}, []cabFixtureFile{{"empty", 0, 0}}, false)
		rules := map[string]string{"meta.cdb": "Empty.Meta:CL_TYPE_MSCAB:*:^empty$:0:0:0:0:*:\n"}
		referenceCABMetadata(t, b, rules, "Empty.Meta")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Empty.Meta" || err != nil {
			t.Fatalf("empty member metadata: signature=%q error=%v", sig, err)
		}
	})
	t.Run("last-rule-first", func(t *testing.T) {
		b := cabFixture(t, 0, [][]byte{body}, []cabFixtureFile{{"one", 0, 8}}, false)
		rules := map[string]string{"meta.cdb": "First:CL_TYPE_MSCAB:*:^one$:0:8:0:0:*:\nSecond:CL_TYPE_MSCAB:*:^one$:0:8:0:0:*:\n"}
		referenceCABMetadata(t, b, rules, "Second")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "Second" || err != nil {
			t.Fatalf("rule priority: signature=%q error=%v", sig, err)
		}
	})
}

func TestCABMetadataArchiveExtentAndEmbedding(t *testing.T) {
	base := cabQuantumControl(t)
	for _, tc := range []struct {
		name string
		data []byte
		size int
	}{
		{"trailing", append(bytes.Clone(base), []byte("outer trailer")...), len(base) + 13},
		{"embedded", append(peValidationFixture(), base...), len(base)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rules := map[string]string{"meta.cdb": fmt.Sprintf("Extent.Meta:CL_TYPE_MSCAB:%d:^qtm[.]txt$:0:59:0:2:*:\n", tc.size)}
			referenceCABMetadata(t, tc.data, rules, "Extent.Meta")
			if sig, err := pipelineState(t, rules).visit(tc.data, 0); sig != "Extent.Meta" || err != nil {
				t.Fatalf("archive extent: signature=%q error=%v", sig, err)
			}
		})
	}
}

func TestCABMetadataUnusedFolderAndRepeatedViews(t *testing.T) {
	body := []byte("harmless member")
	t.Run("unused-unsupported-folder", func(t *testing.T) {
		b := cabMetadataFixture(t, [][]byte{body, body}, []uint16{0, 15}, []cabMetadataFile{{"one", 0, 0, uint32(len(body))}}, 0)
		rules := map[string]string{"meta.cdb": "No.Match:CL_TYPE_MSCAB:*:^absent$:*:*:*:*:*:\n"}
		referenceCABMetadata(t, b, rules, "")
		if sig, err := pipelineState(t, rules).visit(b, 0); sig != "" || err != nil {
			t.Fatalf("unused folder affected a file: signature=%q error=%v", sig, err)
		}
	})
	t.Run("repeated-view-budget", func(t *testing.T) {
		b := cabFixture(t, 0, [][]byte{body}, []cabFixtureFile{{"one", 0, uint32(len(body))}, {"two", 0, uint32(len(body))}}, false)
		s := pipelineState(t, map[string]string{"meta.cdb": "No.Match:CL_TYPE_MSCAB:*:^absent$:*:*:*:*:*:\n"})
		s.remaining = int64(len(b) + len(body))
		if sig, err := s.visit(b, 0); sig != "" || !errors.Is(err, ErrLimit) {
			t.Fatalf("unpaid repeated view: signature=%q error=%v", sig, err)
		}
	})
}

func FuzzCABMetadata(f *testing.F) {
	seed, err := os.ReadFile("testdata/libmspack_mszip_lzx_qtm.cab")
	if err != nil {
		f.Fatal(err)
	}
	db, err := signatures.Compile(map[string][]byte{"meta.cdb": []byte("Metadata.Fuzz:CL_TYPE_MSCAB:*:^never$:*:*:*:*:*:\n")})
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Add([]byte{})
	child := bytes.Repeat([]byte("bounded CAB salvage input "), 16)
	stored := cabFixture(f, 0, [][]byte{child}, []cabFixtureFile{{"first", 0, uint32(len(child))}}, false)
	f.Add(stored)
	prefix := bytes.Clone(stored)
	binary.LittleEndian.PutUint16(prefix[40:], 2)
	prefix = append(prefix, 1)
	binary.LittleEndian.PutUint32(prefix[8:], uint32(len(prefix)))
	f.Add(prefix)
	f.Add(cabFixture(f, 1, [][]byte{child}, []cabFixtureFile{{"first", 0, uint32(len(child))}}, false))
	f.Add(cabLZXFixture(f, child, 18, false))
	for _, size := range []uint32{0, 35, 0xffffffff} {
		advisory := bytes.Clone(stored)
		binary.LittleEndian.PutUint32(advisory[8:], size)
		advisory[12], advisory[24], advisory[25] = 0xff, 0xff, 0xff
		f.Add(advisory)
	}
	f.Add(cabHeaderSet(stored, 2, 0xfffe))
	f.Add(cabHeaderSet(stored, 3, 0xffff))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		s := scanState{ctx: context.Background(), db: db, maxFile: 65536, remaining: 131072, maxFiles: 128, maxDepth: 3}
		s.layers.Push(signatures.Layer{Type: signatures.TypeMSCAB, Size: int64(len(data))})
		s.cab(data, 0)
		if s.remaining < 0 || s.files > s.maxFiles {
			t.Fatalf("CAB budget escaped: files=%d remaining=%d", s.files, s.remaining)
		}
	})
}
