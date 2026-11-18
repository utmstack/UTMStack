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
	"strconv"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

const (
	engineCBCMarker = "UTMSTACK-BYTECODE-HARMLESS-MARKER"
	engineCBCLabel  = "Bytecode.Native.Detected"
)

// These synthetic programs are independently encoded from the ClamBC wire
// format. A detected HDB child proves the reference actually scanned bytes
// written by the VM; the plaintext child is absent from the input object.
func referenceBytecodeExtractedChild(t *testing.T, program, input []byte) {
	referenceBytecodeExtractedChildBytes(t, program, input, []byte(engineCBCLabel))
}

func referenceBytecodeExtractedChildBytes(t *testing.T, program, input, child []byte, nativeArgs ...string) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	dir := t.TempDir()
	for name, data := range map[string][]byte{
		"logical.cbc": program,
		"child.hdb":   []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
	} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	inputPath := filepath.Join(dir, "input.bin")
	if err := os.WriteFile(inputPath, input, 0o600); err != nil {
		t.Fatal(err)
	}
	args := []string{"--database=" + dir, "--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary"}
	args = append(args, nativeArgs...)
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append(args, inputPath)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	var status *exec.ExitError
	if !errors.As(err, &status) || status.ExitCode() != 1 ||
		!strings.Contains(string(out), "input.bin: Child.Bytecode.Hash.UNOFFICIAL FOUND") {
		t.Fatalf("ClamAV did not detect extracted child: error=%v output=%s", err, out)
	}
}

// These fixtures independently encode the public ClamBC wire format. They use
// only harmless markers and do not include upstream programs or malware.
func engineCBCNumber(value uint64) string {
	var digits []byte
	for value != 0 {
		digits = append(digits, 0x60|byte(value&15))
		value >>= 4
	}
	return string(append([]byte{0x60 + byte(len(digits))}, digits...))
}

func engineCBCFixed(value uint64, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 0x60 | byte(value&15)
		value >>= 4
	}
	return string(b)
}

func engineCBCConstant(value uint64, bits int) string {
	b := []byte(engineCBCNumber(value))
	b[0] &^= 0x20
	return string(b) + engineCBCFixed(uint64(bits/8), 1)
}

func engineCBCString(value string) string {
	s := "|" + engineCBCNumber(uint64(len(value)+1))
	for _, b := range append([]byte(value), 0) {
		s += engineCBCFixed(uint64(b), 2)
	}
	return s
}

func engineCBCGlobal(typ int, values ...uint64) string {
	s := engineCBCNumber(uint64(typ))
	for _, value := range values {
		b := []byte(engineCBCNumber(value))
		b[0] &^= 0x20
		s += string(b)
	}
	return s + "`"
}

func engineCBCProgram(kind uint64, mode string, declareWrite bool) []byte {
	// Dynamic type 69 is the initialized label array. Type 70 is the API
	// prototype i32(i8*, i32). Type 71, when needed, is i32(i32) for
	// extract_new. Four predefined pointer types precede them;
	// the serialized count includes one reserved sentinel.
	declareExtract := mode == "extract-explicit" || mode == "extract-multi"
	declareOptionEx := strings.HasPrefix(mode, "option-ex:")
	optionName := strings.TrimPrefix(mode, "option-ex:")
	declareRootType := mode == "json-root-type" || mode == "json-hwpml-object" || strings.HasPrefix(mode, "json-field-")
	jsonKey := "FileType"
	jsonValue, jsonValueAPI := uint64(len("CL_TYPE_TEXT_ASCII")), uint64(97)
	if mode == "json-hwpml-object" {
		jsonKey, jsonValue, jsonValueAPI = "HWPML", 4, 94
	}
	if strings.HasPrefix(mode, "json-field-") {
		parts := strings.Split(mode, ":")
		if len(parts) != 3 {
			panic("invalid JSON fixture mode")
		}
		jsonKey = parts[1]
		var err error
		jsonValue, err = strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			panic(err)
		}
		if parts[0] == "json-field-int" {
			jsonValueAPI = 100
		}
	}
	declareScanMask := strings.HasPrefix(mode, "scan-mask:")
	declareJSON := mode == "json-active" || declareRootType || declareScanMask
	zeroAPI, zeroExpected, zeroName := uint64(92), uint64(1), "json_is_active"
	if declareScanMask {
		zeroAPI, zeroName = 68, "engine_scan_options"
		var err error
		zeroExpected, err = strconv.ParseUint(strings.TrimPrefix(mode, "scan-mask:"), 10, 32)
		if err != nil {
			panic(err)
		}
	}
	typeCount := uint64(7)
	if declareExtract {
		typeCount++
	}
	if declareJSON {
		typeCount++
	}
	if declareOptionEx {
		typeCount++
	}
	if declareRootType {
		typeCount++
		typeCount++
	}
	header := "ClamBC" + engineCBCNumber(7) + engineCBCNumber(0) + "|" + engineCBCNumber(0) +
		engineCBCNumber(0) + engineCBCNumber(kind) + strings.Repeat(engineCBCNumber(0), 3) + "|" + engineCBCNumber(0) +
		engineCBCNumber(typeCount) + engineCBCNumber(1) + engineCBCNumber(0x53e5493e9f3d1c30) + engineCBCFixed(42, 2) + ":4096"
	trigger := "Bytecode.Native.{Detected};Target:0;0;" + hex.EncodeToString([]byte(engineCBCMarker))
	types := "T" + engineCBCFixed(69, 2) +
		engineCBCFixed(4, 1) + engineCBCNumber(uint64(len(engineCBCLabel)+1)) + engineCBCNumber(8) +
		engineCBCFixed(1, 1) + engineCBCNumber(3) + engineCBCNumber(32) + engineCBCNumber(65) + engineCBCNumber(32)
	if declareExtract {
		types += engineCBCFixed(1, 1) + engineCBCNumber(2) + engineCBCNumber(32) + engineCBCNumber(32)
	}
	if declareJSON {
		if declareRootType {
			types += engineCBCFixed(1, 1) + engineCBCNumber(4) + engineCBCNumber(32) + engineCBCNumber(65) + engineCBCNumber(32) + engineCBCNumber(32)
		} else {
			types += engineCBCFixed(1, 1) + engineCBCNumber(1) + engineCBCNumber(32)
		}
	}
	if declareRootType {
		types += engineCBCFixed(4, 1) + engineCBCNumber(uint64(len(jsonKey)+1)) + engineCBCNumber(8)
		types += engineCBCFixed(1, 1) + engineCBCNumber(2) + engineCBCNumber(32) + engineCBCNumber(32)
	}
	if declareOptionEx {
		types += engineCBCFixed(4, 1) + engineCBCNumber(uint64(len(optionName)+1)) + engineCBCNumber(8)
	}
	apis := engineCBCNumber(5) + engineCBCNumber(70) + engineCBCString("setvirusname")
	apiCount := uint64(1)
	if declareWrite {
		// Merely declaring write must prevent activation. Its call is never
		// present, so this covers rejection of unused extraction capability.
		apis += engineCBCNumber(3) + engineCBCNumber(70) + engineCBCString("write")
		apiCount++
	}
	apiMax := uint64(5)
	if declareExtract {
		apis += engineCBCNumber(22) + engineCBCNumber(71) + engineCBCString("extract_new")
		apiCount++
		apiMax = 22
	}
	if declareJSON {
		if declareRootType {
			apis += engineCBCNumber(93) + engineCBCNumber(71) + engineCBCString("json_get_object")
			apiName := "json_get_string_length"
			if jsonValueAPI == 94 {
				apiName = "json_get_type"
			} else if jsonValueAPI == 100 {
				apiName = "json_get_int"
			}
			apis += engineCBCNumber(jsonValueAPI) + engineCBCNumber(73) + engineCBCString(apiName)
			apiCount++
			apiMax = jsonValueAPI
		} else {
			apis += engineCBCNumber(zeroAPI) + engineCBCNumber(71) + engineCBCString(zeroName)
			apiMax = zeroAPI
		}
		apiCount++
	}
	if declareOptionEx {
		apis += engineCBCNumber(101) + engineCBCNumber(70) + engineCBCString("engine_scan_options_ex")
		apiCount++
		apiMax = 101
	}
	var label []uint64
	for _, b := range append([]byte(engineCBCLabel), 0) {
		label = append(label, uint64(b))
	}
	globalCount := uint64(3)
	if declareRootType || declareOptionEx {
		globalCount = 5
	}
	globals := "G" + engineCBCNumber(0) + engineCBCNumber(globalCount) + engineCBCGlobal(0, 0) +
		engineCBCGlobal(69, label...) + engineCBCGlobal(65, 0, 1)
	if declareRootType {
		var key []uint64
		for _, b := range []byte(jsonKey + "\x00") {
			key = append(key, uint64(b))
		}
		globals += engineCBCGlobal(72, key...) + engineCBCGlobal(65, 0, 3)
	}
	if declareOptionEx {
		var key []uint64
		for _, b := range []byte(optionName + "\x00") {
			key = append(key, uint64(b))
		}
		globals += engineCBCGlobal(71, key...) + engineCBCGlobal(65, 0, 3)
	}
	// Global 2 is a virtual i8* into label global 1. API 5 receives the
	// complete detection name and its byte length without the trailing NUL.
	setName := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
		engineCBCFixed(2, 1) + engineCBCNumber(5) + engineCBCConstant(2, 0) + engineCBCConstant(uint64(len(engineCBCLabel)), 32)
	writeChild := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
		engineCBCFixed(2, 1) + engineCBCNumber(3) + engineCBCConstant(2, 0) + engineCBCConstant(uint64(len(engineCBCLabel)), 32)
	ret := "T" + engineCBCFixed(19, 2) + engineCBCNumber(32) + engineCBCConstant(0, 32)
	blocks := [][]string{{setName, ret}}
	instructionMode := mode
	if strings.HasPrefix(mode, "json-field-") {
		instructionMode = "json-root-type"
	} else if declareScanMask {
		instructionMode = "json-active"
	}
	switch instructionMode {
	case "loop":
		blocks = [][]string{{setName, "T" + engineCBCFixed(18, 2) + engineCBCNumber(1)}, {"T" + engineCBCFixed(18, 2) + engineCBCNumber(1)}}
	case "abort":
		blocks = [][]string{{setName, "T" + engineCBCFixed(45, 2)}}
	case "no-label":
		blocks = [][]string{{"T" + engineCBCFixed(19, 2) + engineCBCNumber(32) + engineCBCConstant(99, 32)}}
	case "extract":
		blocks = [][]string{{writeChild, ret}}
	case "extract-fault":
		blocks = [][]string{{writeChild, "T" + engineCBCFixed(45, 2)}}
	case "extract-explicit":
		extract := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(1, 1) + engineCBCNumber(22) + engineCBCConstant(7, 32)
		blocks = [][]string{{writeChild, extract, ret}}
	case "extract-multi":
		writePrefix := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(2, 1) + engineCBCNumber(3) + engineCBCConstant(2, 0) + engineCBCConstant(9, 32)
		extract := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(1, 1) + engineCBCNumber(22) + engineCBCConstant(7, 32)
		blocks = [][]string{{writePrefix, extract, writeChild, ret}}
	case "json-active":
		query := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(0, 1) + engineCBCNumber(zeroAPI)
		active := engineCBCNumber(1) + engineCBCNumber(1) + engineCBCFixed(21, 2) +
			engineCBCNumber(32) + engineCBCNumber(0) + engineCBCConstant(zeroExpected, 32)
		branch := "T" + engineCBCFixed(17, 2) + engineCBCNumber(1) + engineCBCNumber(1) + engineCBCNumber(2)
		queries := []string{query}
		if declareScanMask {
			// Isolate FILE_PROPERTIES from unrelated native module options.
			queries = append(queries, engineCBCNumber(32)+engineCBCNumber(0)+engineCBCFixed(11, 2)+
				engineCBCNumber(0)+engineCBCConstant(0x10000000, 32))
		}
		blocks = [][]string{append(queries, active, branch), {setName, ret}, {ret}}
	case "json-hwpml-object":
		query := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(3, 1) + engineCBCNumber(93) + engineCBCConstant(4, 0) +
			engineCBCConstant(uint64(len(jsonKey)+1), 32) + engineCBCConstant(0, 32)
		present := engineCBCNumber(1) + engineCBCNumber(2) + engineCBCFixed(21, 2) +
			engineCBCNumber(32) + engineCBCNumber(0) + engineCBCConstant(1, 32)
		branchPresent := "T" + engineCBCFixed(17, 2) + engineCBCNumber(2) + engineCBCNumber(1) + engineCBCNumber(3)
		kind := engineCBCNumber(32) + engineCBCNumber(1) + engineCBCFixed(33, 2) +
			engineCBCFixed(1, 1) + engineCBCNumber(94) + engineCBCNumber(0)
		object := engineCBCNumber(1) + engineCBCNumber(2) + engineCBCFixed(21, 2) +
			engineCBCNumber(32) + engineCBCNumber(1) + engineCBCConstant(4, 32)
		branchObject := "T" + engineCBCFixed(17, 2) + engineCBCNumber(2) + engineCBCNumber(2) + engineCBCNumber(3)
		blocks = [][]string{{query, present, branchPresent}, {kind, object, branchObject}, {setName, ret}, {ret}}
	case "json-root-type":
		query := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(3, 1) + engineCBCNumber(93) + engineCBCConstant(4, 0) +
			engineCBCConstant(uint64(len(jsonKey)+1), 32) + engineCBCConstant(0, 32)
		length := engineCBCNumber(32) + engineCBCNumber(1) + engineCBCFixed(33, 2) +
			engineCBCFixed(1, 1) + engineCBCNumber(jsonValueAPI) + engineCBCNumber(0)
		expected := engineCBCNumber(1) + engineCBCNumber(2) + engineCBCFixed(21, 2) +
			engineCBCNumber(32) + engineCBCNumber(1) + engineCBCConstant(jsonValue, 32)
		branch := "T" + engineCBCFixed(17, 2) + engineCBCNumber(2) + engineCBCNumber(1) + engineCBCNumber(2)
		blocks = [][]string{{query, length, expected, branch}, {setName, ret}, {ret}}
	}
	if declareOptionEx {
		query := engineCBCNumber(32) + engineCBCNumber(0) + engineCBCFixed(33, 2) +
			engineCBCFixed(2, 1) + engineCBCNumber(101) + engineCBCConstant(4, 0) +
			engineCBCConstant(uint64(len(optionName)), 32)
		one := engineCBCNumber(1) + engineCBCNumber(1) + engineCBCFixed(21, 2) +
			engineCBCNumber(32) + engineCBCNumber(0) + engineCBCConstant(1, 32)
		branch := "T" + engineCBCFixed(17, 2) + engineCBCNumber(1) + engineCBCNumber(1) + engineCBCNumber(2)
		blocks = [][]string{{query, one, branch}, {setName, ret}, {ret}}
	}
	count := 0
	for _, block := range blocks {
		count += len(block)
	}
	declaration := "A" + engineCBCFixed(0, 1) + engineCBCNumber(32) + "L" + engineCBCNumber(1) +
		engineCBCNumber(32) + engineCBCFixed(0, 1)
	if declareJSON {
		if declareRootType {
			declaration = "A" + engineCBCFixed(0, 1) + engineCBCNumber(32) + "L" + engineCBCNumber(3) +
				engineCBCNumber(32) + engineCBCFixed(0, 1) + engineCBCNumber(32) + engineCBCFixed(0, 1) + engineCBCNumber(1) + engineCBCFixed(0, 1)
		} else {
			declaration = "A" + engineCBCFixed(0, 1) + engineCBCNumber(32) + "L" + engineCBCNumber(2) +
				engineCBCNumber(32) + engineCBCFixed(0, 1) + engineCBCNumber(1) + engineCBCFixed(0, 1)
		}
	}
	if declareOptionEx {
		declaration = "A" + engineCBCFixed(0, 1) + engineCBCNumber(32) + "L" + engineCBCNumber(2) +
			engineCBCNumber(32) + engineCBCFixed(0, 1) + engineCBCNumber(1) + engineCBCFixed(0, 1)
	}
	declaration += "F" + engineCBCNumber(uint64(count)) + engineCBCNumber(uint64(len(blocks)))
	lines := []string{header, trigger, types, "E" + engineCBCNumber(apiMax) + engineCBCNumber(apiCount) + apis, globals, declaration}
	for i, block := range blocks {
		line := "B" + strings.Join(block, "")
		if i == len(blocks)-1 {
			line += "E"
		}
		lines = append(lines, line)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func engineCBCActivate(t *testing.T, program []byte) (*Engine, string) {
	t.Helper()
	cfg := config.Default()
	cfg.CollectMetadata = true
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	version, err := e.ActivateGeneration(map[string][]byte{"logical.cbc": program}, 10, manifestHash("bytecode-generation-10"))
	if err != nil {
		t.Fatalf("activate logical bytecode: %v", err)
	}
	return e, version
}

func engineCBCCheckState(t *testing.T, e *Engine, version string) {
	t.Helper()
	if e.Version() != version || e.SignatureCount() != 1 {
		t.Fatalf("generation changed: version %q, signatures %d", e.Version(), e.SignatureCount())
	}
	if seq, hash := e.FeedState(); seq != 10 || hash != manifestHash("bytecode-generation-10") {
		t.Fatalf("feed state = %d, %q", seq, hash)
	}
}

func TestLogicalBytecodeEnginePersistenceAndRecursiveScan(t *testing.T) {
	program := engineCBCProgram(256, "detect", false)
	e, version := engineCBCActivate(t, program)
	if !strings.HasPrefix(version, ImplementationVersion+":") {
		t.Fatalf("engine version has no implementation identity: %q", version)
	}
	engineCBCCheckState(t, e, version)
	current, err := os.ReadFile(filepath.Join(e.DatabaseDir(), "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := os.ReadFile(filepath.Join(e.DatabaseDir(), "generations", strings.TrimSpace(string(current)), "signatures", "logical.cbc"))
	if err != nil || !bytes.Equal(persisted, program) {
		t.Fatalf("persisted CBC differs: %v", err)
	}
	if err := e.Stop(); err != nil {
		t.Fatal(err)
	}
	restarted := New(e.cfg)
	if err := restarted.EnsureRunning(); err != nil {
		t.Fatalf("reload persisted CBC: %v", err)
	}
	engineCBCCheckState(t, restarted, version)
	payload := []byte("ordinary prefix " + engineCBCMarker + " ordinary suffix")
	for name, input := range map[string][]byte{
		"raw": payload, "zip": zipFixture(t, payload), "gzip": gzipFixture(t, payload),
		"zip-gzip": zipFixture(t, gzipFixture(t, payload)), "gzip-zip": gzipFixture(t, zipFixture(t, payload)),
	} {
		t.Run(name, func(t *testing.T) {
			if name != "raw" && bytes.Contains(input, []byte(engineCBCMarker)) {
				t.Fatal("archive contains plain trigger; recursive detection would not be demonstrated")
			}
			clean, label, snapshot, err := restarted.ScanSnapshot(context.Background(), input)
			if err != nil || clean || label != engineCBCLabel || snapshot != version {
				t.Fatalf("scan clean=%v label=%q version=%q err=%v", clean, label, snapshot, err)
			}
		})
	}
	for name, input := range map[string][]byte{
		"raw":    []byte("ordinary benign content"),
		"nested": zipFixture(t, gzipFixture(t, []byte("ordinary benign content"))),
	} {
		t.Run("benign-"+name, func(t *testing.T) {
			clean, label, snapshot, err := restarted.ScanSnapshot(context.Background(), input)
			if err != nil || !clean || label != "" || snapshot != version {
				t.Fatalf("benign scan clean=%v label=%q version=%q err=%v", clean, label, snapshot, err)
			}
		})
	}
}

func TestDefaultScanDoesNotActivateJSONBytecodeForOfficeXML(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	rules := map[string][]byte{
		"logical.cbc": engineCBCProgram(256, "json-active", false),
		"word.ftm":    []byte("0:0:3c3f786d6c:MS XML:CL_TYPE_ANY:CL_TYPE_XML_WORD\n"),
	}
	if _, err := e.ActivateGeneration(rules, 10, manifestHash("json-default-off")); err != nil {
		t.Fatal(err)
	}
	input := []byte(`<?xml version="1.0"?><wordDocument>` + engineCBCMarker + `</wordDocument>`)
	_, label, err := e.Scan(context.Background(), input)
	if label != "" {
		t.Fatalf("default scan activated JSON bytecode: %q, %v", label, err)
	}
	if err != nil && IncompleteReason(err) == "Unsupported.bytecode" {
		t.Fatalf("inactive JSON bytecode left rule undecided: %v", err)
	}
}

func TestLogicalJSONActivationMatchesCollectionMode(t *testing.T) {
	program := engineCBCProgram(256, "json-active", false)
	input := []byte("Harmless content with " + engineCBCMarker + " here.\n")
	activate := func(collect bool) *Engine {
		cfg := config.Default()
		cfg.CollectMetadata = collect
		cfg.SignatureDBDir = t.TempDir()
		e := New(cfg)
		if _, err := e.ActivateGeneration(map[string][]byte{"logical.cbc": program}, 1, manifestHash("logical-json-mode")); err != nil {
			t.Fatal(err)
		}
		return e
	}
	if clean, name, err := activate(false).Scan(context.Background(), input); !clean || name != "" || err != nil {
		t.Fatalf("default JSON-off logical hook: clean=%v name=%q err=%v", clean, name, err)
	}
	if clean, name, err := activate(true).Scan(context.Background(), input); clean || name != engineCBCLabel || err != nil {
		t.Fatalf("opt-in JSON logical hook: clean=%v name=%q err=%v", clean, name, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		dir := t.TempDir()
		for name, contents := range map[string][]byte{"logical.cbc": program, "input.txt": input} {
			if err := os.WriteFile(filepath.Join(dir, name), contents, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		for _, tc := range []struct {
			collect bool
			want    string
		}{
			{false, "input.txt: OK"}, {true, "input.txt: " + engineCBCLabel + " FOUND"},
		} {
			args := []string{"--bytecode-unsigned", "--database=" + dir, "--no-summary"}
			if tc.collect {
				args = append(args, "--gen-json")
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append(args, filepath.Join(dir, "input.txt"))...)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if (err != nil) != tc.collect || !strings.Contains(string(out), tc.want) {
				t.Fatalf("native logical JSON collect=%v: err=%v output=%s", tc.collect, err, out)
			}
		}
	}
}

func TestScanOptionsExPinnedNativeParity(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	input := []byte("Harmless content with " + engineCBCMarker + " here.\n")
	for _, option := range []string{"general collect metadata", "parse archive", "general heuristics", "unknown option", "detect-control"} {
		for _, enabled := range []bool{false, true} {
			name := strings.ReplaceAll(option, " ", "-")
			t.Run(fmt.Sprintf("%s/metadata=%t", name, enabled), func(t *testing.T) {
				dir := t.TempDir()
				mode := "option-ex:" + option
				if option == "detect-control" {
					mode = "detect"
				}
				program := engineCBCProgram(256, mode, false)
				cfg := config.Default()
				cfg.CollectMetadata = enabled
				cfg.SignatureDBDir = t.TempDir()
				engine := New(cfg)
				if _, err := engine.ActivateGeneration(map[string][]byte{"option.cbc": program}, 1, manifestHash("option-ex-"+mode)); err != nil {
					t.Fatal(err)
				}
				clean, name, err := engine.Scan(context.Background(), input)
				if wantDetect := option == "detect-control"; err != nil || clean == wantDetect || (name == engineCBCLabel) != wantDetect {
					t.Fatalf("Go option=%q metadata=%t clean=%t name=%q err=%v", option, enabled, clean, name, err)
				}
				if ref == "" {
					return
				}
				if err := os.WriteFile(filepath.Join(dir, "option.cbc"), program, 0o600); err != nil {
					t.Fatal(err)
				}
				inputPath := filepath.Join(dir, "input.txt")
				if err := os.WriteFile(inputPath, input, 0o600); err != nil {
					t.Fatal(err)
				}
				args := []string{"--bytecode-unsigned", "--database=" + dir, "--no-summary"}
				if enabled {
					args = append(args, "--gen-json")
				}
				cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append(args, inputPath)...)
				cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
				out, err := cmd.CombinedOutput()
				if wantDetect := option == "detect-control"; (err != nil) != wantDetect || strings.Contains(string(out), " FOUND") != wantDetect {
					t.Fatalf("native option=%q metadata=%t err=%v output=%s", option, enabled, err, out)
				}
			})
		}
	}
}

func TestPreclassHookRunsAfterRootScanOnly(t *testing.T) {
	program := engineCBCProgram(260, "json-active", false)
	e, _ := engineCBCActivate(t, program)
	root := []byte("ordinary prefix " + engineCBCMarker + " ordinary suffix")
	clean, name, err := e.Scan(context.Background(), root)
	if err != nil || clean || name != engineCBCLabel {
		t.Fatalf("root preclass: clean=%v name=%q err=%v", clean, name, err)
	}
	// A match in a child layer does not flag its hook in the root layer.
	child := zipFixture(t, root)
	if bytes.Contains(child, []byte(engineCBCMarker)) {
		t.Fatal("fixture exposes child marker in raw ZIP bytes")
	}
	clean, name, err = e.Scan(context.Background(), child)
	if err != nil || !clean || name != "" {
		t.Fatalf("child trigger escaped root layer: clean=%v name=%q err=%v", clean, name, err)
	}
}

func TestBytecodeOutputUsesBoundedRecursiveScanner(t *testing.T) {
	program := engineCBCProgram(256, "extract", true)
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	child := []byte(engineCBCLabel)
	rules := map[string][]byte{
		"logical.cbc": program,
		"child.hdb":   []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
	}
	if _, err := e.ActivateGeneration(rules, 10, manifestHash("bytecode-extracted-child")); err != nil {
		t.Fatal(err)
	}
	parent := []byte("ordinary prefix " + engineCBCMarker + " ordinary suffix")
	if bytes.Contains(parent, child) {
		t.Fatal("parent already contains the extracted child")
	}
	for name, input := range map[string][]byte{"raw": parent, "zip": zipFixture(t, parent)} {
		t.Run(name, func(t *testing.T) {
			clean, label, err := e.Scan(context.Background(), input)
			if clean || label != "Child.Bytecode.Hash" || err != nil {
				t.Fatalf("extracted child: clean=%t label=%q err=%v", clean, label, err)
			}
		})
	}
	referenceBytecodeExtractedChild(t, program, parent)
	if clean, label, err := e.Scan(context.Background(), []byte("ordinary benign input")); !clean || label != "" || err != nil {
		t.Fatalf("false trigger: clean=%t label=%q err=%v", clean, label, err)
	}
	// The generated child cannot bypass depth, file-count or aggregate-byte
	// limits merely because it came from a bytecode program.
	for _, tc := range []struct {
		name         string
		depth, files int
		remaining    int64
	}{
		{"depth", 0, 10, 1 << 20},
		{"file-count", 10, 1, 1 << 20},
		{"aggregate-bytes", 10, 10, int64(len(parent) + len(child) - 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := scanState{ctx: context.Background(), db: e.current.Load().db, maxFile: 1 << 20, remaining: tc.remaining, maxFiles: tc.files, maxDepth: tc.depth}
			if label, err := s.visit(parent, 0); label != "" || !errors.Is(err, ErrLimit) {
				t.Fatalf("limited child: label=%q err=%v", label, err)
			}
		})
	}
}

func TestBytecodeExplicitExtractNewScansChild(t *testing.T) {
	program := engineCBCProgram(256, "extract-explicit", true)
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	child := []byte(engineCBCLabel)
	rules := map[string][]byte{
		"logical.cbc": program,
		"child.hdb":   []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
	}
	if _, err := e.ActivateGeneration(rules, 10, manifestHash("bytecode-explicit-child")); err != nil {
		t.Fatal(err)
	}
	if clean, label, err := e.Scan(context.Background(), []byte(engineCBCMarker)); clean || label != "Child.Bytecode.Hash" || err != nil {
		t.Fatalf("explicit extract_new: clean=%t label=%q err=%v", clean, label, err)
	}
	referenceBytecodeExtractedChild(t, program, []byte(engineCBCMarker))
}

func TestBytecodeMultipleExtractedChildren(t *testing.T) {
	program := engineCBCProgram(256, "extract-multi", true)
	parent := []byte("ordinary prefix " + engineCBCMarker + " ordinary suffix")
	children := [][]byte{[]byte(engineCBCLabel[:9]), []byte(engineCBCLabel)}
	for i, child := range children {
		t.Run(fmt.Sprintf("child-%d", i+1), func(t *testing.T) {
			if bytes.Contains(parent, child) {
				t.Fatal("parent contains child marker")
			}
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			e := New(cfg)
			rules := map[string][]byte{
				"logical.cbc": program,
				"child.hdb":   []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
			}
			if _, err := e.ActivateGeneration(rules, 10, manifestHash(fmt.Sprintf("bytecode-multi-child-%d", i))); err != nil {
				t.Fatal(err)
			}
			if clean, label, err := e.Scan(context.Background(), parent); clean || label != "Child.Bytecode.Hash" || err != nil {
				t.Fatalf("child %d: clean=%t label=%q err=%v", i+1, clean, label, err)
			}
			referenceBytecodeExtractedChildBytes(t, program, parent, child)
		})
	}
}

func TestBytecodePendingOutputScannedAfterVMFault(t *testing.T) {
	program := engineCBCProgram(256, "extract-fault", true)
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	child := []byte(engineCBCLabel)
	rules := map[string][]byte{
		"logical.cbc": program,
		"child.hdb":   []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
	}
	if _, err := e.ActivateGeneration(rules, 10, manifestHash("bytecode-fault-child")); err != nil {
		t.Fatal(err)
	}
	if clean, label, err := e.Scan(context.Background(), []byte(engineCBCMarker)); clean || label != "Child.Bytecode.Hash" || err != nil {
		t.Fatalf("pending output after VM fault: clean=%t label=%q err=%v", clean, label, err)
	}
	withoutChildRule, _ := engineCBCActivate(t, program)
	if clean, label, err := withoutChildRule.Scan(context.Background(), []byte(engineCBCMarker)); clean || label != "" || !errors.Is(err, bytecode.ErrRuntime) {
		t.Fatalf("fault without child detection became clean: clean=%t label=%q err=%v", clean, label, err)
	}
}

// This optional exporter writes only synthetic harmless bytes, for an
// independent local ClamAV oracle run with --bytecode-unsigned. The flag must
// never be used with a program from an untrusted source.
func TestExportHarmlessBytecodeExtractionOracle(t *testing.T) {
	dir := os.Getenv("EDR_EXTRACT_FIXTURE_DIR")
	if dir == "" {
		t.Skip("EDR_EXTRACT_FIXTURE_DIR not set")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	child := []byte(engineCBCLabel)
	files := map[string][]byte{
		"logical.cbc":  engineCBCProgram(256, "extract", true),
		"explicit.cbc": engineCBCProgram(256, "extract-explicit", true),
		"fault.cbc":    engineCBCProgram(256, "extract-fault", true),
		"child.hdb":    []byte(fmt.Sprintf("%x:%d:Child.Bytecode.Hash\n", md5.Sum(child), len(child))),
		"input.bin":    []byte("ordinary prefix " + engineCBCMarker + " ordinary suffix"),
		"benign.bin":   []byte("ordinary benign input"),
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("wrote %d harmless extraction oracle files to %s", len(files), dir)
}

func TestLogicalBytecodeEngineRuntimeFailureIsIncomplete(t *testing.T) {
	for _, tc := range []struct {
		mode string
		want error
	}{
		{"loop", bytecode.ErrLimit}, {"abort", bytecode.ErrRuntime},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			e, version := engineCBCActivate(t, engineCBCProgram(256, tc.mode, false))
			for name, input := range map[string][]byte{
				"raw": []byte(engineCBCMarker), "nested": zipFixture(t, gzipFixture(t, []byte(engineCBCMarker))),
			} {
				t.Run(name, func(t *testing.T) {
					// The program sets a label before failing. Neither that
					// partial label nor a clean verdict may escape to callers.
					clean, label, snapshot, err := e.ScanSnapshot(context.Background(), input)
					if clean || label != "" || snapshot != version || !errors.Is(err, tc.want) {
						t.Fatalf("failed scan clean=%v label=%q version=%q err=%v", clean, label, snapshot, err)
					}
				})
			}
			if clean, label, err := e.Scan(context.Background(), []byte("no logical trigger here")); !clean || label != "" || err != nil {
				t.Fatalf("untriggered callback executed: %v %q %v", clean, label, err)
			}
			engineCBCCheckState(t, e, version)
		})
	}
}

func TestLogicalBytecodeEngineReturnValueIsNotDetection(t *testing.T) {
	e, version := engineCBCActivate(t, engineCBCProgram(256, "no-label", false))
	clean, label, snapshot, err := e.ScanSnapshot(context.Background(), []byte(engineCBCMarker))
	if err != nil || !clean || label != "" || snapshot != version {
		t.Fatalf("integer return became detection: clean=%v label=%q version=%q err=%v", clean, label, snapshot, err)
	}
}

func TestLogicalBytecodeRejectedGenerationPreservesDurableState(t *testing.T) {
	good := engineCBCProgram(256, "detect", false)
	e, version := engineCBCActivate(t, good)
	current, err := os.ReadFile(filepath.Join(e.DatabaseDir(), "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	// A trigger naming a subsignature it does not have is malformed; the
	// reference rejects such a database too.
	malformed := bytes.Replace(good, []byte(";Target:0;0;"), []byte(";Target:0;1;"), 1)
	for _, tc := range []struct {
		name, diagnostic string
		program          []byte
	}{
		{"malformed-trigger", "names more subsignatures", malformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fresh := New(e.cfg)
			files := map[string][]byte{"a-good.cbc": good, "z-unsupported.cbc": tc.program}
			if _, err := fresh.ActivateGeneration(files, 11, manifestHash("rejected-"+tc.name)); err == nil || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("wrong rejection: %v", err)
			}
			engineCBCCheckState(t, e, version)
			engineCBCCheckState(t, fresh, version)
			after, err := os.ReadFile(filepath.Join(e.DatabaseDir(), "CURRENT"))
			if err != nil || !bytes.Equal(current, after) {
				t.Fatalf("rejection changed durable pointer: %q, %v", after, err)
			}
			restarted := New(e.cfg)
			if err := restarted.EnsureRunning(); err != nil {
				t.Fatal(err)
			}
			engineCBCCheckState(t, restarted, version)
			if clean, label, err := restarted.Scan(context.Background(), []byte(engineCBCMarker)); clean || label != engineCBCLabel || err != nil {
				t.Fatalf("rejection disabled prior rule: %v %q %v", clean, label, err)
			}
			if _, err := restarted.ActivateGeneration(map[string][]byte{"logical.cbc": good}, 9, manifestHash("rollback-after-rejection")); err == nil {
				t.Fatal("failed CBC update erased durable replay protection")
			}
		})
	}
}

// Programs this engine cannot run but the reference would (a generic kind)
// load as unknown when their trigger holds: the generation activates, and
// objects holding the trigger are incomplete.
func TestUnsupportedLogicalBytecodeActivatesAsUnknownWhenReached(t *testing.T) {
	for name, program := range map[string][]byte{
		"generic": engineCBCProgram(0, "detect", false),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := bytecode.Parse(program); err != nil {
				t.Fatalf("invalid fixture: %v", err)
			}
			e, _ := engineCBCActivate(t, program)
			clean, label, err := e.Scan(context.Background(), []byte(engineCBCMarker))
			var unknown *signatures.UnknownError
			if clean || label != "" || !errors.As(err, &unknown) || unknown.Reason != signatures.UnknownBytecode {
				t.Fatalf("trigger holds: clean %v %q %v", clean, label, err)
			}
			if clean, label, err := e.Scan(context.Background(), []byte("an ordinary object")); !clean || label != "" || err != nil {
				t.Fatalf("trigger false: %v %q %v", clean, label, err)
			}
		})
	}
}

// A logical program that declares write but does not call it runs like any
// other: only a run that writes output is undecided.
func TestLogicalBytecodeDeclaringWriteRuns(t *testing.T) {
	program := engineCBCProgram(256, "detect", true)
	e, _ := engineCBCActivate(t, program)
	clean, label, err := e.Scan(context.Background(), []byte(engineCBCMarker))
	if clean || label != "Bytecode.Native.Detected" || err != nil {
		t.Fatalf("trigger holds: clean %v %q %v", clean, label, err)
	}
}
