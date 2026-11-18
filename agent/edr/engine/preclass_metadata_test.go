package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestRootPreclassMetadataKnownFields(t *testing.T) {
	data := []byte("harmless root input")
	meta, err := rootPreclassMetadata(context.Background(), data, signatures.TypeTextASCII)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]any{
		"Magic": "CLAMJSONv0", "FileSize": uint64(len(data)), "ObjectID": uint64(0),
		"FileType": "CL_TYPE_TEXT_ASCII", "RootFileType": "CL_TYPE_TEXT_ASCII",
		"sha2-256": fmt.Sprintf("%x", sha256.Sum256(data)),
	} {
		if got := meta[key]; got != want {
			t.Fatalf("%s=%v, want %v", key, got, want)
		}
	}
	for _, key := range []string{"FileName", "FilePath", "ContainedObjects", "PDFStats", "md5", "sha1"} {
		if _, ok := meta[key]; ok {
			t.Fatalf("unproved metadata field %q present", key)
		}
	}
	serialized := plainScalarMetadataJSON(map[string]any{
		"Magic": "CLAMJSONv0", "ObjectID": uint64(0),
		"FileSize": uint64(len(data)), "FileType": "CL_TYPE_BINARY_DATA",
		"RootFileType": "CL_TYPE_BINARY_DATA", "sha2-256": "abc",
	})
	wantJSON := fmt.Sprintf("{\n  \"Magic\":\"CLAMJSONv0\",\n  \"FileSize\":%d,\n  \"ObjectID\":0,\n  \"FileType\":\"CL_TYPE_BINARY_DATA\",\n  \"sha2-256\":\"abc\",\n  \"RootFileType\":\"CL_TYPE_BINARY_DATA\"\n}", len(data))
	if string(serialized) != wantJSON {
		t.Fatalf("json-c PRETTY layout: %q", serialized)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := rootPreclassMetadata(ctx, data, signatures.TypeTextASCII); err != context.Canceled {
		t.Fatalf("cancelled hash: %v", err)
	}
}

func TestRootMetadataCollectionIsOptIn(t *testing.T) {
	program := engineCBCProgram(260, "json-active", false)
	data := []byte("Harmless plain text with " + engineCBCMarker + " in it.\n")
	activate := func(collect bool) *Engine {
		cfg := config.Default()
		cfg.CollectMetadata = collect
		cfg.SignatureDBDir = t.TempDir()
		e := New(cfg)
		if _, err := e.ActivateGeneration(map[string][]byte{"preclass.cbc": program}, 1, manifestHash("metadata-opt-in")); err != nil {
			t.Fatal(err)
		}
		return e
	}
	off, on := activate(false), activate(true)
	if off.MetadataSensitive() || !on.MetadataSensitive() || off.Version() == on.Version() {
		t.Fatalf("metadata policy or cache identity: off=%q on=%q", off.Version(), on.Version())
	}
	if clean, name, err := off.Scan(context.Background(), data); !clean || name != "" || err != nil {
		t.Fatalf("default metadata-off: clean=%v name=%q err=%v", clean, name, err)
	}
	if clean, name, err := on.Scan(context.Background(), data); clean || name != engineCBCLabel || err != nil {
		t.Fatalf("opt-in PRECLASS: clean=%v name=%q err=%v", clean, name, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, contents := range map[string][]byte{
			"preclass.cbc": program, "input.txt": data,
		} {
			if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		run := func(collect bool) (string, error) {
			args := []string{"--bytecode-unsigned", "--database=" + dir, "--no-summary"}
			if collect {
				args = append(args, "--gen-json")
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append(args, filepath.Join(dir, "input.txt"))...)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			return string(out), err
		}
		if out, err := run(false); err != nil || !strings.Contains(out, "input.txt: OK") {
			t.Fatalf("native default metadata-off: err=%v output=%s", err, out)
		}
		if out, err := run(true); err == nil || !strings.Contains(out, "input.txt: "+engineCBCLabel+" FOUND") {
			t.Fatalf("native metadata opt-in: err=%v output=%s", err, out)
		}
	}
}

func TestRootPreclassMetadataFilePathOrder(t *testing.T) {
	data := []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb}
	path := "/tmp/edr-metadata/trigger.bin"
	meta, err := rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, path)
	if err != nil {
		t.Fatal(err)
	}
	if meta["FileName"] != "trigger.bin" || meta["FilePath"] != path {
		t.Fatalf("path metadata: %#v", meta)
	}
	want := "{\n  \"Magic\":\"CLAMJSONv0\",\n  \"FileName\":\"trigger.bin\",\n  \"FilePath\":\"" + path + "\",\n  \"FileSize\":8,\n  \"ObjectID\":0,\n  \"FileType\":\"CL_TYPE_BINARY_DATA\",\n  \"sha2-256\":\"" + fmt.Sprintf("%x", sha256.Sum256(data)) + "\",\n  \"RootFileType\":\"CL_TYPE_BINARY_DATA\"\n}"
	if got := string(plainScalarMetadataJSON(meta)); got != want {
		t.Fatalf("metadata JSON = %q, want %q", got, want)
	}
	textMeta, err := rootPreclassMetadataForPath(context.Background(), []byte("Harmless plain text.\n"), signatures.TypeTextASCII, "/tmp/plain.txt")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(plainScalarMetadataJSON(textMeta)); !strings.Contains(got, `"FileType":"CL_TYPE_TEXT_ASCII"`) || !strings.Contains(got, `"RootFileType":"CL_TYPE_TEXT_ASCII"`) || strings.Contains(got, "CL_TYPE_BINARY_DATA") {
		t.Fatalf("ASCII text metadata: %q", got)
	}
	textMeta["UnmodelledParserField"] = true
	if got := plainScalarMetadataJSON(textMeta); got != nil {
		t.Fatalf("unmodelled metadata field was dropped: %q", got)
	}
	meta, err = rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, "/tmp/a<&>.bin")
	if err != nil {
		t.Fatal(err)
	}
	if got := string(plainScalarMetadataJSON(meta)); !strings.Contains(got, `"FileName":"a<&>.bin"`) || strings.Contains(got, `\u003c`) {
		t.Fatalf("json-c must leave HTML characters in file names unescaped: %q", got)
	}
	if _, err := rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, "/tmp/\xff.bin"); IncompleteReason(err) != "Unsupported.metadata-json" {
		t.Fatalf("invalid path encoding must be incomplete, got %v", err)
	}
	backslashPath := "/tmp/prefix\\alias.txt"
	meta, err = rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, backslashPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := meta["FileName"]; got != "alias.txt" {
		t.Fatalf("reference backslash basename = %q", got)
	}
	if got := string(plainScalarMetadataJSON(meta)); !strings.Contains(got, `"FileName":"alias.txt"`) || !strings.Contains(got, `"FilePath":"/tmp/prefix\\alias.txt"`) {
		t.Fatalf("reference backslash path JSON = %q", got)
	}
	unicodePath := "/tmp/line\u2028sep.txt"
	meta, err = rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, unicodePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(plainScalarMetadataJSON(meta)); !strings.Contains(got, "\"FileName\":\"line\u2028sep.txt\"") || strings.Contains(got, `\u2028`) {
		t.Fatalf("reference raw Unicode path JSON = %q", got)
	}
	if got, ok := jsonCString(`literal\u2028`); !ok || string(got) != `"literal\\u2028"` {
		t.Fatalf("literal backslash Unicode sequence = %q, %v", got, ok)
	}
	if _, err := rootPreclassMetadataForPath(context.Background(), data, signatures.TypeBinaryData, "/tmp/nul\x00name"); IncompleteReason(err) != "Unsupported.metadata-json" {
		t.Fatalf("NUL path must remain incomplete, got %v", err)
	}
}

func TestLegacyTarget13UsesPathWithoutHashOnlyReuse(t *testing.T) {
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	rules := map[string][]byte{
		"metadata.ndb": []byte("Native.Metadata.Filename:13:*:" + hex.EncodeToString([]byte(`"FileName":"trigger.bin"`)) + "\n"),
		"preclass.cbc": engineCBCProgram(260, "json-active", false),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("target13-path")); err != nil {
		t.Fatal(err)
	}
	if !e.MetadataSensitive() {
		t.Fatal("active metadata rules must disable hash-only cache reuse")
	}
	data := []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb}
	for _, tc := range []struct {
		path, want string
	}{
		{"/tmp/edr-metadata/safe.bin", ""},
		{"/tmp/edr-metadata/trigger.bin", "Native.Metadata.Filename"},
		{"/tmp/edr-metadata/safe.bin", ""},
	} {
		clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), data, tc.path)
		if err != nil || clean != (tc.want == "") || name != tc.want {
			t.Fatalf("%s: clean=%v name=%q err=%v, want %q", tc.path, clean, name, err, tc.want)
		}
	}
}

func TestLegacyTarget13PlainASCIITextMatchesReference(t *testing.T) {
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	rules := map[string][]byte{
		"text.ndb":     []byte("Native.Metadata.Text:13:*:" + hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_TEXT_ASCII"`)) + "\n"),
		"preclass.cbc": engineCBCProgram(260, "json-active", false),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("target13-text-on")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path string
		data []byte
	}{
		{"/tmp/plain.txt", []byte("This is harmless plain ASCII content, with no markup.\n")},
		{"/tmp/utf8.txt", []byte("Harmless caf\u00e9 text.\n")},
	} {
		clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), tc.data, tc.path)
		if err != nil || clean || name != "Native.Metadata.Text" {
			typ, _ := e.current.Load().db.DetermineType(tc.data)
			t.Fatalf("%s type=%s: clean=%v name=%q err=%v", tc.path, typ, clean, name, err)
		}
	}
	// HTML can add parser and contained-object fields to the native metadata
	// tree. A matching target-13 rule cannot decide the result from our
	// scalar-only representation, even though this oracle's HTML is clean.
	clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), []byte("<head></head><body>harmless body</body>"), "/tmp/html.txt")
	if clean || name != "" || IncompleteReason(err) != "Unsupported.metadata-json" {
		t.Fatalf("parser-bearing HTML metadata: clean=%v name=%q err=%v", clean, name, err)
	}
	delete(rules, "preclass.cbc")
	if _, err := e.ActivateGeneration(rules, 2, manifestHash("target13-text-only")); err != nil {
		t.Fatal(err)
	}
	if !e.MetadataSensitive() {
		t.Fatal("opt-in target-13 rule must keep metadata collection active")
	}
	clean, name, _, err = e.ScanSnapshotWithPath(context.Background(), []byte("This is harmless plain ASCII content, with no markup.\n"), "/tmp/plain.txt")
	if err != nil || clean || name != "Native.Metadata.Text" {
		t.Fatalf("opt-in target-13 only: clean=%v name=%q err=%v", clean, name, err)
	}
	offCfg := config.Default()
	offCfg.SignatureDBDir = t.TempDir()
	off := New(offCfg)
	if _, err := off.ActivateGeneration(rules, 1, manifestHash("target13-text-default-off")); err != nil {
		t.Fatal(err)
	}
	if off.MetadataSensitive() {
		t.Fatal("default metadata-off rule must not vary by path")
	}
	clean, name, _, err = off.ScanSnapshotWithPath(context.Background(), []byte("This is harmless plain ASCII content, with no markup.\n"), "/tmp/plain.txt")
	if err != nil || !clean || name != "" {
		t.Fatalf("default metadata-off target-13: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestPreclassBytecodeReadsRootFileType(t *testing.T) {
	e, _ := engineCBCActivate(t, engineCBCProgram(260, "json-root-type", false))
	input := []byte("harmless " + engineCBCMarker + " content")
	clean, name, err := e.Scan(context.Background(), input)
	if err != nil || clean || name != engineCBCLabel {
		t.Fatalf("preclass root FileType: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestLegacyTarget13ScansCompleteBinaryRootOnly(t *testing.T) {
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	// The real daily.ftm uses this json-c PRETTY prefix to type legacy
	// metadata as INTERNAL. A preclass program activates metadata collection
	// but its trigger does not match either harmless root below.
	ftm := "0:0:7b0a2020224d61676963223a22434c414d4a534f4e763022:Internal properties:CL_TYPE_ANY:CL_TYPE_INTERNAL\n"
	marker := []byte(`"RootFileType":"CL_TYPE_BINARY_DATA"`)
	rules := map[string][]byte{
		"legacy.ftm":   []byte(ftm),
		"legacy.ndb":   []byte("Legacy.Metadata.Binary:13:*:" + hex.EncodeToString(marker) + "\n"),
		"preclass.cbc": engineCBCProgram(260, "json-active", false),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("legacy-target-13")); err != nil {
		t.Fatal(err)
	}
	plain := []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb}
	clean, name, err := e.Scan(context.Background(), plain)
	if err != nil || clean || name != "Legacy.Metadata.Binary" {
		typ, _ := e.current.Load().db.DetermineType(plain)
		t.Fatalf("plain binary root: type=%v clean=%v name=%q err=%v", typ, clean, name, err)
	}
	child := zipFixture(t, []byte("harmless child content"))
	clean, name, err = e.Scan(context.Background(), child)
	if clean || name != "" || IncompleteReason(err) != "Unsupported.metadata-json" {
		t.Fatalf("partial container metadata: clean=%v name=%q err=%v", clean, name, err)
	}
	// A complete scalar tree with no matching target-13 rule is decidable.
	rules["legacy.ndb"] = []byte("Legacy.Other:13:*:" + hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_PDF"`)) + "\n")
	if _, err := e.ActivateGeneration(rules, 2, manifestHash("legacy-target-13-negative")); err != nil {
		t.Fatal(err)
	}
	clean, name, err = e.Scan(context.Background(), plain)
	if err != nil || !clean || name != "" {
		t.Fatalf("negative complete metadata: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestLegacyTarget13DoesNotActivateDefaultMetadata(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	ftm := "0:0:7b0a2020224d61676963223a22434c414d4a534f4e763022:Internal properties:CL_TYPE_ANY:CL_TYPE_INTERNAL\n"
	rules := map[string][]byte{
		"legacy.ftm": []byte(ftm),
		"legacy.ndb": []byte("Legacy.Metadata.Binary:13:*:" + hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_BINARY_DATA"`)) + "\n"),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("legacy-target-13-off")); err != nil {
		t.Fatal(err)
	}
	if e.MetadataSensitive() {
		t.Fatal("inactive target-13 rules must not disable hash-only cache reuse")
	}
	clean, name, err := e.Scan(context.Background(), []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb})
	if err != nil || !clean || name != "" {
		t.Fatalf("default metadata-off: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestLegacyTarget13UsesBuiltInInternalType(t *testing.T) {
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	rules := map[string][]byte{
		"legacy.ndb":   []byte("Legacy.Metadata.Binary:13:*:" + hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_BINARY_DATA"`)) + "\n"),
		"preclass.cbc": engineCBCProgram(260, "json-active", false),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("legacy-target-13-builtin")); err != nil {
		t.Fatal(err)
	}
	clean, name, err := e.Scan(context.Background(), []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb})
	if err != nil || clean || name != "Legacy.Metadata.Binary" {
		t.Fatalf("built-in INTERNAL type: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestLegacyTarget13NestedPassHonorsFileBudget(t *testing.T) {
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	cfg.ScanLimits.MaxFiles = 1
	e := New(cfg)
	rules := map[string][]byte{
		"legacy.ftm":   []byte("0:0:7b0a2020224d61676963223a22434c414d4a534f4e763022:Internal properties:CL_TYPE_ANY:CL_TYPE_INTERNAL\n"),
		"legacy.ndb":   []byte("Legacy.Metadata.Binary:13:*:" + hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_BINARY_DATA"`)) + "\n"),
		"preclass.cbc": engineCBCProgram(260, "json-active", false),
	}
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("legacy-target-13-budget")); err != nil {
		t.Fatal(err)
	}
	clean, name, err := e.Scan(context.Background(), []byte{0x7f, 0x80, 0, 1, 0xff, 0xfe, 0xaa, 0xbb})
	if clean || name != "" || IncompleteReason(err) != "ResourceLimit" {
		t.Fatalf("target-13 file budget: clean=%v name=%q err=%v", clean, name, err)
	}
}
