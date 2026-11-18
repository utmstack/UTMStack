package engine

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// The trigger is inside a compressed child or a normalized HTML view. Testing
// both the root and child sizes distinguishes a shared root from a per-object
// JSON snapshot; mere JSON activation cannot establish that distinction.
func TestBytecodeJSONAlwaysUsesRootMetadata(t *testing.T) {
	child := []byte(strings.Repeat("harmless padding ", 256) + engineCBCMarker)
	zip := zipFixture(t, child)
	gz := gzipFixture(t, child)
	nested := zipFixture(t, gz)
	for _, b := range [][]byte{zip, gz, nested} {
		if bytes.Contains(b, []byte(engineCBCMarker)) {
			t.Fatal("compressed fixture exposes its child trigger in the parent")
		}
	}
	for _, tc := range []struct {
		name, typ string
		data      []byte
	}{
		{"gzip", "CL_TYPE_GZ", gz},
		{"zip", "CL_TYPE_ZIP", zip},
		{"nested", "CL_TYPE_ZIP", nested},
		{"html", "CL_TYPE_HTML", []byte("<html><body>" + engineCBCMarker + "</body></html>")},
		{"html-retyped", "CL_TYPE_HTML", []byte("prose <html><body>" + engineCBCMarker + "</body></html>")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, q := range []struct {
				mode string
				hit  bool
			}{
				{fmt.Sprintf("json-field-int:FileSize:%d", len(tc.data)), true},
				{fmt.Sprintf("json-field-int:FileSize:%d", len(child)), false},
				{fmt.Sprintf("json-field-length:FileType:%d", len(tc.typ)), true},
				{"json-field-length:FileType:18", false}, // child's CL_TYPE_TEXT_ASCII
				{"json-field-int:ObjectID:0", true},
				{"json-field-length:sha2-256:64", true},
				{"json-field-length:FileName:9", true},
			} {
				t.Run(q.mode, func(t *testing.T) {
					dir := t.TempDir()
					path := filepath.Join(dir, "root.file")
					program := engineCBCProgram(256, q.mode, false)
					if strings.HasPrefix(tc.name, "html") {
						// HTML normalization lowercases the marker. The raw input
						// cannot trigger this rule before a text-to-HTML retype.
						program = bytes.ReplaceAll(program, []byte(hex.EncodeToString([]byte(engineCBCMarker))),
							[]byte(hex.EncodeToString([]byte(strings.ToLower(engineCBCMarker)))))
					}
					cfg := config.Default()
					cfg.CollectMetadata = true
					cfg.SignatureDBDir = t.TempDir()
					e := New(cfg)
					if _, err := e.ActivateGeneration(map[string][]byte{"root.cbc": program}, 1, manifestHash(q.mode)); err != nil {
						t.Fatal(err)
					}
					if !e.MetadataSensitive() {
						t.Fatal("ordinary JSON bytecode must receive the root path and bypass hash-only caching")
					}
					clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), tc.data, path)
					t.Logf("Go clean=%v signature=%q error=%v expected_detection=%v", clean, name, err, q.hit)
					if err != nil || clean == q.hit || (name == engineCBCLabel) != q.hit {
						t.Fatalf("Go: clean=%v name=%q err=%v; detection=%v", clean, name, err, q.hit)
					}
					if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
						if err := os.WriteFile(filepath.Join(dir, "root.cbc"), program, 0o600); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(path, tc.data, 0o600); err != nil {
							t.Fatal(err)
						}
						cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--bytecode-unsigned", "--gen-json",
							"--database="+dir, "--no-summary", path)
						cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc/certs"))
						out, nativeErr := cmd.CombinedOutput()
						checkNativeMetadataVerdict(t, out, nativeErr, q.hit, engineCBCLabel)
					}
				})
			}
		})
	}
}

func TestRootMetadataPhaseAndIdentity(t *testing.T) {
	data := []byte("harmless root data")
	root, err := initialRootMetadata(context.Background(), data, signatures.TypeTextASCII, "/tmp/root.file")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := root["RootFileType"]; ok {
		t.Fatal("RootFileType added before the root metadata phase")
	}
	s := scanState{collectMetadata: true, metadataRoot: root}
	childView := s.objectMetadata()
	normalizedView := s.representationMetadata()
	root["FileType"] = "CL_TYPE_HTML"
	finished, err := finishRootMetadata(context.Background(), data, signatures.TypeTextASCII, root)
	if err != nil || childView["FileSize"] != uint64(len(data)) || normalizedView["RootFileType"] != "CL_TYPE_HTML" || finished["FileName"] != "root.file" {
		t.Fatalf("root views lost identity: child=%v normalized=%v err=%v", childView, normalizedView, err)
	}
}

func TestMetadataRootNameWireFixture(t *testing.T) {
	got, err := os.ReadFile("testdata/metadata_root_name.cbc")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, engineCBCProgram(256, "json-field-length:FileName:9", false)) {
		t.Fatal("scanner regression fixture differs from the native-validated JSON name program")
	}
}

func TestLegacyMetadataBytecodeRetainsRootContext(t *testing.T) {
	data := []byte("Harmless plain ASCII without the trigger.\n")
	for _, tc := range []struct {
		mode string
		hit  bool
	}{
		{"json-active", true},
		{fmt.Sprintf("json-field-int:FileSize:%d", len(data)), true},
		{"json-field-int:FileSize:0", false},
		{"json-field-length:FileType:18", true},
		{"json-field-length:FileType:16", false}, // INTERNAL is the JSON layer, not the root
		{"json-field-length:FileName:9", true},
		{"scan-mask:0", true},
		{"scan-mask:268435456", false},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "root.file")
			program := engineCBCProgram(256, tc.mode, false)
			// Only the serialized metadata, never the root content, contains
			// this trigger. The dummy target-13 rule enables that nested pass.
			program = bytes.ReplaceAll(program, []byte(hex.EncodeToString([]byte(engineCBCMarker))),
				[]byte(hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_TEXT_ASCII"`))))
			rules := map[string][]byte{
				"root.cbc":   program,
				"enable.ndb": []byte("Metadata.Never:13:*:4142434445463736353433323130\n"),
			}
			cfg := config.Default()
			cfg.CollectMetadata = true
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			if _, err := e.ActivateGeneration(rules, 1, manifestHash(tc.mode)); err != nil {
				t.Fatal(err)
			}
			clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), data, path)
			t.Logf("Go clean=%v signature=%q error=%v expected_detection=%v", clean, name, err, tc.hit)
			if err != nil || clean == tc.hit || (name == engineCBCLabel) != tc.hit {
				t.Fatalf("Go: clean=%v name=%q err=%v detection=%v", clean, name, err, tc.hit)
			}
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				for file, b := range rules {
					if err := os.WriteFile(filepath.Join(dir, file), b, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.WriteFile(path, data, 0o600); err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--bytecode-unsigned", "--gen-json", "--no-summary", "--database="+dir, path)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc/certs"))
				out, nativeErr := cmd.CombinedOutput()
				checkNativeMetadataVerdict(t, out, nativeErr, tc.hit, engineCBCLabel)
			}
		})
	}
}

func TestLegacyMetadataRetainsRootAncestry(t *testing.T) {
	data := []byte("Harmless plain ASCII without the trigger.\n")
	for _, container := range []string{"CL_TYPE_TEXT_ASCII", "CL_TYPE_ZIP"} {
		t.Run(container, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "root.file")
			needle := hex.EncodeToString([]byte(`"RootFileType":"CL_TYPE_TEXT_ASCII"`))
			rules := []byte("Metadata.Parent;Engine:51-255,Target:13,Container:" + container + ";0;" + needle + "\n")
			cfg := config.Default()
			cfg.CollectMetadata = true
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			if _, err := e.ActivateGeneration(map[string][]byte{"parent.ldb": rules}, 1, manifestHash(container)); err != nil {
				t.Fatal(err)
			}
			clean, name, _, err := e.ScanSnapshotWithPath(context.Background(), data, path)
			hit := container == "CL_TYPE_TEXT_ASCII"
			t.Logf("Go clean=%v signature=%q error=%v expected_detection=%v", clean, name, err, hit)
			if err != nil || clean == hit || (name == "Metadata.Parent") != hit {
				t.Fatalf("Go: clean=%v name=%q err=%v detection=%v", clean, name, err, hit)
			}
			if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
				for file, b := range map[string][]byte{"parent.ldb": rules, "root.file": data} {
					if err := os.WriteFile(filepath.Join(dir, file), b, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--gen-json", "--no-summary", "--database="+dir, path)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc/certs"))
				out, nativeErr := cmd.CombinedOutput()
				checkNativeMetadataVerdict(t, out, nativeErr, hit, "Metadata.Parent.UNOFFICIAL")
			}
		})
	}
}

func TestPreclassOutputStartsAtFirstChildDepth(t *testing.T) {
	child := []byte(engineCBCLabel)
	rules := map[string][]byte{
		"extract.cbc": engineCBCProgram(260, "extract", true),
		"child.hdb":   []byte(fmt.Sprintf("%x:%d:Metadata.Extracted.Child\n", md5.Sum(child), len(child))),
	}
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	cfg.ScanLimits.MaxRecursion = 1
	e := New(cfg)
	if _, err := e.ActivateGeneration(rules, 1, manifestHash("preclass-child-depth")); err != nil {
		t.Fatal(err)
	}
	clean, name, err := e.Scan(context.Background(), []byte("harmless "+engineCBCMarker))
	t.Logf("Go clean=%v signature=%q error=%v expected_detection=true", clean, name, err)
	if err != nil || clean || name != "Metadata.Extracted.Child" {
		t.Fatalf("first preclass child: clean=%v name=%q err=%v", clean, name, err)
	}
	// Confirm independently that this program's native PRECLASS output is
	// scanned. The tight Go resource bound above checks our depth accounting;
	// it does not equate differently defined native and Go recursion limits.
	referenceBytecodeExtractedChildBytes(t, rules["extract.cbc"], []byte("harmless "+engineCBCMarker), child, "--gen-json")
}

func checkNativeMetadataVerdict(t *testing.T, out []byte, err error, hit bool, label string) {
	t.Helper()
	exit := 0
	if err != nil {
		var status *exec.ExitError
		if !errors.As(err, &status) {
			t.Fatalf("ClamAV invocation: %v", err)
		}
		exit = status.ExitCode()
	}
	t.Logf("ClamAV exit=%d output=%s", exit, strings.TrimSpace(string(out)))
	wantExit := 0
	if hit {
		wantExit = 1
	}
	if exit != wantExit || strings.Contains(string(out), " FOUND") != hit || hit && !strings.Contains(string(out), label+" FOUND") {
		t.Fatalf("ClamAV: exit=%d detection=%v label=%q output=%s", exit, hit, label, out)
	}
}
