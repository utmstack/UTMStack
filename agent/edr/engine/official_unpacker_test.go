package engine

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Cisco Talos keeps these benign packed-PE controls XOR-encoded in the
// ClamAV unit-test tree. Set EDR_CLAMAV_PACKED_PE_DIR to a decoded copy to
// run this differential; the ordinary suite has no network dependency.
func TestOfficialClamAVPackedPEControls(t *testing.T) {
	dir := os.Getenv("EDR_CLAMAV_PACKED_PE_DIR")
	if dir == "" {
		t.Skip("set EDR_CLAMAV_PACKED_PE_DIR to decoded ClamAV 1.5.4 unit-test fixtures")
	}
	control, err := os.ReadFile(filepath.Join(dir, "clam.exe"))
	if err != nil {
		t.Fatal(err)
	}
	hash := md5.Sum(control)
	if got := fmt.Sprintf("%x:%d", hash, len(control)); got != "aa15bcf478d165efd2065190eb473bcb:544" {
		t.Fatalf("unexpected ClamAV 1.5.4 control: %s", got)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(control)); got != "71e7b604d18aefd839e51a39c88df8383bb4c071dc31f87f00a2b5df580d4495" {
		t.Fatalf("unexpected ClamAV 1.5.4 control SHA-256: %s", got)
	}
	rule := fmt.Sprintf("%x:%d:Test.ClamAV.Control\n", hash, len(control))
	for _, tc := range []struct{ name, sha256 string }{
		{"clam-aspack.exe", "fa2901d4e97497165ce275bd8a9e94fe5266ea7a0795d96e0e439da0c7645c36"},
		{"clam-fsg.exe", "13f8764444fb9a0ffc9bfe2120e1ad163779846b747163bf14b71519504cb9ae"},
		{"clam-mew.exe", "bfe7eeb1939e8bc16f90cb5d921437056e0e456a00a8ea3b31bd9754f6c89885"},
		{"clam-pespin.exe", "1dba66766f99bbd55c28113775a54a9cc22d93735494be96237a4081be89e82b"},
		{"clam-petite.exe", "f4091b710d78322370e849381cddedc878bbe580563d993f4195b2bcb1ffc5b8"},
		{"clam-upack.exe", "80a03f1b06996e084f54e6218019e1f0e2c3e789c72a9264145c8e0602c84702"},
		{"clam-upx.exe", "d1973ca87229f403ef214905c4a9c2f2a4cca73e1b5b0217eb3f7595e706e16f"},
		{"clam-wwpack.exe", "f8a027d8f09a8943c32e0cdee35fe0d4ce77da174c7773cafde752397d4a75f4"},
		{"clam-yc.exe", "4ebe56972199136a749eca73eb221272cabf725436d6b939b8c1182383c562a9"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, tc.name))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != tc.sha256 {
				t.Fatalf("unexpected ClamAV 1.5.4 fixture SHA-256: %s", got)
			}
			s := pipelineState(t, map[string]string{"control.hdb": rule})
			sig, scanErr := s.visit(data, 0)
			if sig != "Test.ClamAV.Control" || scanErr != nil {
				t.Fatalf("native control detected, Go signature=%q error=%v", sig, scanErr)
			}
		})
	}
}

// This inert PE32 resource follows the authenticated script2exe hook's
// type 2 -> name 129 -> language 0 resource path. The 64-byte data region
// is all zero; the official CBC itself turns it into the expected child.
func script2exeResourceControl() []byte {
	le := binary.LittleEndian
	code := make([]byte, 512)
	copy(code, []byte(`script2exe\Release\mywscript.pdb`))
	rsrc := make([]byte, 0x1000)
	le.PutUint16(rsrc[14:], 1)
	le.PutUint32(rsrc[0x10:], 2)
	le.PutUint32(rsrc[0x14:], 0x80000030)
	le.PutUint16(rsrc[0x30+14:], 1)
	le.PutUint32(rsrc[0x40:], 129)
	le.PutUint32(rsrc[0x44:], 0x80000060)
	le.PutUint16(rsrc[0x60+14:], 1)
	le.PutUint32(rsrc[0x70:], 0)
	le.PutUint32(rsrc[0x74:], 0x90)
	le.PutUint32(rsrc[0x90:], 0x2100)
	le.PutUint32(rsrc[0x94:], 0x100)
	le.PutUint32(rsrc[0x104:], 0x100)
	le.PutUint32(rsrc[0x138:], 64)
	f := simplePE(code)
	f.sections = append(f.sections, peFixtureSection{name: ".rsrc", va: 0x2000, vsize: 0x1000, raw: rsrc})
	pe := f.build()
	le.PutUint32(pe[264:], 0x2000)
	le.PutUint32(pe[268:], 0x1000)
	return pe
}

// This positive control exercises an actual authenticated feed program's
// output branch, not a synthetic stand-in for ClamBC. The local hash rule
// exists only for the child bytes; the parent never contains that child.
func TestOfficialScript2ExeExtractedChild(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	program, err := os.ReadFile(filepath.Join(dir, "bytecode-6754169.cbc"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != "ae3cade7e964a53e6ef465df673b0ea695102f62a1e630f070d9d79c8d5742c4" {
		t.Fatalf("official CBC SHA-256 = %s", got)
	}
	parent := script2exeResourceControl()
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != "3cb6e99b372f7edbcd6bc5b274e9a03b5ad5a35bf3754a0f988d58cc668fff13" {
		t.Fatalf("inert PE SHA-256 = %s", got)
	}
	child, err := hex.DecodeString("482e924071a2e41d5224af186976fb59ee33e53335d80092a26e244e8b075bb914deb2612312b65c6d5d821e4627ad79d804ed0ffe2c87f02a353af4a0f7ec33")
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "9e051a669bb7e1ed0fef6f463f52eb3a99bbf7fc9d0820f408861abdc3f29537" || bytes.Contains(parent, child) {
		t.Fatalf("child SHA-256 = %s; embedded in parent = %t", got, bytes.Contains(parent, child))
	}
	rule := fmt.Sprintf("%x:%d:Child.Official.Script2Exe\n", md5.Sum(child), len(child))
	s := pipelineState(t, map[string]string{"hook.cbc": string(program), "child.hdb": rule})
	if sig, err := s.visit(parent, 0); sig != "Child.Official.Script2Exe" || err != nil {
		t.Fatalf("official CBC child: signature=%q err=%v", sig, err)
	}
	withoutHook := pipelineState(t, map[string]string{"child.hdb": rule})
	if sig, err := withoutHook.visit(parent, 0); sig != "" || err != nil {
		t.Fatalf("parent without CBC: signature=%q err=%v", sig, err)
	}
	broken := bytes.Clone(parent)
	binary.LittleEndian.PutUint32(broken[0x440:], 128) // no name 129
	if sig, err := s.visit(broken, 0); sig != "" || err != nil {
		t.Fatalf("broken resource path: signature=%q err=%v", sig, err)
	}
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	toolBytes, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("reference clamscan is not the pinned 1.5.4 binary: %v", err)
	}
	refDir := t.TempDir()
	for name, data := range map[string][]byte{"hook.cbc": program, "child.hdb": []byte(rule)} {
		if err := os.WriteFile(filepath.Join(refDir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		found bool
	}{{"positive.exe", parent, true}, {"broken.exe", broken, false}} {
		path := filepath.Join(refDir, tc.name)
		if err := os.WriteFile(path, tc.data, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(clamscan, "--database="+refDir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if tc.found {
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 1 ||
				!strings.Contains(string(out), tc.name+": Child.Official.Script2Exe.UNOFFICIAL FOUND") ||
				!strings.Contains(string(out), "previous tempfile had 64 bytes") {
				t.Fatalf("ClamAV official output: %v, %s", err, out)
			}
		} else if err != nil || !strings.Contains(string(out), tc.name+": OK") {
			t.Fatalf("ClamAV broken resource: %v, %s", err, out)
		}
	}
	input := filepath.Join(refDir, "positive.exe")
	cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode=no", "--no-summary", input)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "positive.exe: OK") {
		t.Fatalf("ClamAV parent without bytecode: %v, %s", err, out)
	}
}

// The authenticated hta2exe hook uses a related resource tree with type 23,
// name 130, and language 0. It writes the bitwise inverse of the resource
// data, so the expected child is derived from the input independently of
// the bytecode interpreter.
func hta2exeResourceControl() []byte {
	pe := script2exeResourceControl()
	for i := 0x200; i < 0x400; i++ {
		pe[i] = 0
	}
	copy(pe[0x200:], []byte(`script2exe\Release\hta2exe.pdb`))
	binary.LittleEndian.PutUint32(pe[0x410:], 23)
	binary.LittleEndian.PutUint32(pe[0x440:], 130)
	return pe
}

func TestOfficialHTA2ExeExtractedChild(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	program, err := os.ReadFile(filepath.Join(dir, "bytecode-6754171.cbc"))
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != "fbaf59bd7f832bb9691c6ddb27e638627923eb709d641afb0ff6a0ad689f7b32" {
		t.Fatalf("official CBC SHA-256 = %s", got)
	}
	parent := hta2exeResourceControl()
	if got := fmt.Sprintf("%x", sha256.Sum256(parent)); got != "4122be868ee2f778e068b9f2abbd9cad469c8f4b554a39c28713e93566b66623" {
		t.Fatalf("inert PE SHA-256 = %s", got)
	}
	child := make([]byte, 256)
	for i := range child {
		child[i] = ^parent[0x500+i]
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(child)); got != "577325583f2cc73d41dfa4c68fe67e5aa3c10677c2f65a48922650161464f9bf" || bytes.Contains(parent, child) {
		t.Fatalf("child SHA-256 = %s; embedded in parent = %t", got, bytes.Contains(parent, child))
	}
	rule := fmt.Sprintf("%x:%d:Child.Official.HTA2Exe\n", md5.Sum(child), len(child))
	s := pipelineState(t, map[string]string{"hook.cbc": string(program), "child.hdb": rule})
	if sig, err := s.visit(parent, 0); sig != "Child.Official.HTA2Exe" || err != nil {
		t.Fatalf("official CBC child: signature=%q err=%v", sig, err)
	}
	withoutHook := pipelineState(t, map[string]string{"child.hdb": rule})
	if sig, err := withoutHook.visit(parent, 0); sig != "" || err != nil {
		t.Fatalf("parent without CBC: signature=%q err=%v", sig, err)
	}
	broken := bytes.Clone(parent)
	binary.LittleEndian.PutUint32(broken[0x440:], 129) // no name 130
	if sig, err := s.visit(broken, 0); sig != "" || err != nil {
		t.Fatalf("broken resource path: signature=%q err=%v", sig, err)
	}
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	clamscan := filepath.Join(ref, "bin", "clamscan")
	toolBytes, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(toolBytes)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("reference clamscan is not the pinned 1.5.4 binary: %v", err)
	}
	refDir := t.TempDir()
	for name, data := range map[string][]byte{"hook.cbc": program, "child.hdb": []byte(rule)} {
		if err := os.WriteFile(filepath.Join(refDir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		found bool
	}{{"positive.exe", parent, true}, {"broken.exe", broken, false}} {
		path := filepath.Join(refDir, tc.name)
		if err := os.WriteFile(path, tc.data, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(clamscan, "--database="+refDir,
			"--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", path)
		cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
			"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
		out, err := cmd.CombinedOutput()
		if tc.found {
			status, ok := err.(*exec.ExitError)
			if !ok || status.ExitCode() != 1 ||
				!strings.Contains(string(out), tc.name+": Child.Official.HTA2Exe.UNOFFICIAL FOUND") ||
				!strings.Contains(string(out), "previous tempfile had 256 bytes") ||
				!strings.Contains(string(out), "fmap_get_hash: calculated sha2-256 hash: 577325583f2cc73d41dfa4c68fe67e5aa3c10677c2f65a48922650161464f9bf") {
				t.Fatalf("ClamAV official output: %v, %s", err, out)
			}
		} else if err != nil || !strings.Contains(string(out), tc.name+": OK") {
			t.Fatalf("ClamAV broken resource: %v, %s", err, out)
		}
	}
	cmd := exec.Command(clamscan, "--database="+refDir, "--bytecode=no", "--no-summary", filepath.Join(refDir, "positive.exe"))
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	if out, err := cmd.CombinedOutput(); err != nil || !strings.Contains(string(out), "positive.exe: OK") {
		t.Fatalf("ClamAV parent without bytecode: %v, %s", err, out)
	}
}

// These are authenticated bytecode.cvd programs, not local approximations.
// Their logical signatures can be reached on a harmless PE without
// supplying a packed payload. These are negative output controls; the
// separate script2exe test above validates one positive output path.
func TestOfficialPEBytecodeTriggerControls(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_BYTECODE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_BYTECODE_DIR to expanded authenticated bytecode.cvd members")
	}
	for _, tc := range []struct {
		id, digest string
		trigger    []byte
	}{
		{"5588995", "c52182036629a6ea8cea53e52113a8e380b1a4c3763bc79072d0ebaff2a37b0f", []byte{0xe8, 0, 0, 0, 0, 0xe9, 0, 0, 0xff, 0xff}},
		{"6754169", "ae3cade7e964a53e6ef465df673b0ea695102f62a1e630f070d9d79c8d5742c4", []byte(`script2exe\Release\mywscript.pdb`)},
		{"6754171", "fbaf59bd7f832bb9691c6ddb27e638627923eb709d641afb0ff6a0ad689f7b32", []byte(`script2exe\Release\hta2exe.pdb`)},
	} {
		t.Run(tc.id, func(t *testing.T) {
			program, err := os.ReadFile(filepath.Join(dir, "bytecode-"+tc.id+".cbc"))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(program)); got != tc.digest {
				t.Fatalf("official CBC SHA-256 %s, want %s", got, tc.digest)
			}
			code := make([]byte, 512)
			copy(code, tc.trigger)
			pe := simplePE(code).build()
			s := pipelineState(t, map[string]string{"hook.cbc": string(program)})
			if sig, err := s.visit(pe, 0); sig != "" || err != nil {
				t.Fatalf("native official hook: signature=%q error=%v", sig, err)
			}
			ref := os.Getenv("EDR_CLAMAV_PORTABLE")
			if ref == "" {
				return
			}
			refDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(refDir, "hook.cbc"), program, 0o600); err != nil {
				t.Fatal(err)
			}
			input := filepath.Join(refDir, "input.exe")
			if err := os.WriteFile(input, pe, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+refDir,
				"--bytecode-unsigned=yes", "--bytecode=yes", "--debug", "--no-summary", input)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), "input.exe: OK") ||
				!strings.Contains(string(out), "Bytecode: executed 1 bytecodes for this hook") {
				t.Fatalf("ClamAV official hook not reached and clean: error=%v output=%s", err, out)
			}
		})
	}
}
