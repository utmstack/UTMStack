package signatures

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestELFAndMachOUnpackerHookDispatch(t *testing.T) {
	elf := elf64(binary.LittleEndian, 1, 0, nil, 1, 64)
	copy(elf[len(elf)-2:], "AB")
	mach := macho64(0x01000007)
	copy(mach[len(mach)-2:], "AB")
	noTrigger := append([]byte(nil), elf...)
	copy(noTrigger[len(noTrigger)-2:], "XY")
	elfHook := callbackWithLabel([]byte("Hook.ELF"), true)
	elfHook.kind = 261
	elfHook.logical = "ELF.Hook;Engine:51-255,Target:6;0;4142"
	machHook := callbackWithLabel([]byte("Hook.MachO"), true)
	machHook.kind = 262
	machHook.logical = "MachO.Hook;Engine:51-255,Target:9;0;4142"
	d, err := Compile(map[string][]byte{"elf.cbc": elfHook.encode(), "macho.cbc": machHook.encode()})
	if err != nil {
		t.Fatal(err)
	}
	if d.Report().Unknown[UnknownBytecode] != 0 {
		t.Fatalf("executable hooks remain unknown: %v", d.Report().Unknown)
	}
	for _, tc := range []struct {
		data []byte
		typ  FileType
		want string
	}{
		{elf, TypeELF, "Hook.ELF"},
		{mach, TypeMachO, "Hook.MachO"},
		{noTrigger, TypeELF, ""},
	} {
		res, err := d.MatchObject(nil2ctx(), tc.data, MatchContext{Type: tc.typ})
		if err != nil || res.Name != tc.want {
			t.Errorf("type %s on %x: (%q, %v), want %q", tc.typ, tc.data[:4], res.Name, err, tc.want)
		}
	}
}

func TestELFAndMachOUnpackerHooksScanExtractedChild(t *testing.T) {
	child := []byte("HARMLESS-EXTRACTED-HOOK-CHILD")
	for _, tc := range []struct {
		kind   uint64
		target string
		typ    FileType
		data   []byte
	}{
		{261, "6", TypeELF, elf64(binary.LittleEndian, 1, 0, nil, 1, 64)},
		{262, "9", TypeMachO, macho64(0x01000007)},
	} {
		copy(tc.data[len(tc.data)-2:], "AB")
		f := callbackWithLabel(child, false)
		f.kind = tc.kind
		f.logical = "Output.Hook;Engine:51-255,Target:" + tc.target + ";0;4142"
		f.apiMax, f.apis = 3, []string{callbackAPI(3, 70, "write")}
		f.blocks = [][]string{{
			callbackCall(0, 3, callbackGlobalRef(2), callbackConstant(uint64(len(child)), 32)),
			callbackReturn(32, 0),
		}}
		d, err := Compile(map[string][]byte{"hook.cbc": f.encode(), "child.hdb": hashLine(child, "Child.Hook.Hash")})
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		res, err := d.MatchObject(nil2ctx(), tc.data, MatchContext{Type: tc.typ,
			ScanExtracted: func(output []byte, _ int32) (string, error) {
				calls++
				if !bytes.Equal(output, child) {
					t.Errorf("kind %d: output %q, want %q", tc.kind, output, child)
				}
				return "Child.Hook.Hash", nil
			}})
		if err != nil || res.Name != "Child.Hook.Hash" || calls != 1 {
			t.Errorf("kind %d: result=%+v err=%v child calls=%d", tc.kind, res, err, calls)
		}
		if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
			dir := t.TempDir()
			for name, data := range map[string][]byte{"hook.cbc": f.encode(), "child.hdb": hashLine(child, "Child.Hook.Hash"), "object.bin": tc.data} {
				if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
				"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", filepath.Join(dir, "object.bin"))
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "Child.Hook.Hash.UNOFFICIAL FOUND") {
				t.Errorf("kind %d reference child scan: error=%v output=%s", tc.kind, err, out)
			}
			t.Logf("kind %d reference child scan: %s", tc.kind, out)
		}
	}
}

func TestExecutableUnconditionalHooks(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	for _, tc := range []struct {
		name string
		kind uint64
		typ  FileType
		data []byte
	}{
		{"pe-all", 259, TypeMSEXE, hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})},
		{"pe-unpacker", 257, TypeMSEXE, hookPE(0x800, 0x1000, [4]uint32{0x400, 0x1000, 0x400, 0x400})},
		{"elf", 261, TypeELF, elf64(binary.LittleEndian, 1, 0, nil, 1, 64)},
		{"macho", 262, TypeMachO, macho64(0x01000007)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := callbackWithLabel([]byte("Bare.Hook.Label"), true)
			f.kind, f.logical = tc.kind, ""
			d, err := Compile(map[string][]byte{"hook.cbc": f.encode()})
			if err != nil {
				t.Fatal(err)
			}
			res, err := d.MatchObject(nil2ctx(), tc.data, MatchContext{Type: tc.typ})
			if err != nil || res.Name != "Bare.Hook.Label" || d.Report().Unknown[UnknownBytecode] != 0 {
				t.Fatalf("Go bare hook: result=%+v error=%v report=%+v", res, err, d.Report())
			}
			if ref == "" {
				return
			}
			dir := t.TempDir()
			for name, data := range map[string][]byte{"hook.cbc": f.encode(), "object.bin": tc.data} {
				if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
				"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", filepath.Join(dir, "object.bin"))
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "Bare.Hook.Label FOUND") {
				t.Fatalf("reference bare hook: error=%v output=%s", err, out)
			}
			t.Logf("reference bare hook: %s", out)
		})
	}
}

// Set EDR_CLAMAV_PORTABLE to the pinned portable reference root for an
// independent native ClamAV check of the harmless hook programs.
func TestELFAndMachOUnpackerHookClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 portable root")
	}
	for _, tc := range []struct {
		name, logical, signature string
		kind                     uint64
		data                     []byte
	}{
		{"elf", "ELF.Hook;Engine:51-255,Target:6;0;4142", "Hook.ELF", 261, elf64(binary.LittleEndian, 1, 0, nil, 1, 64)},
		{"macho", "MachO.Hook;Engine:51-255,Target:9;0;4142", "Hook.MachO", 262, macho64(0x01000007)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy(tc.data[len(tc.data)-2:], "AB")
			dir := t.TempDir()
			f := callbackWithLabel([]byte(tc.signature), true)
			f.kind, f.logical = tc.kind, tc.logical
			if err := os.WriteFile(filepath.Join(dir, "hook.cbc"), f.encode(), 0o600); err != nil {
				t.Fatal(err)
			}
			object := filepath.Join(dir, "object.bin")
			if err := os.WriteFile(object, tc.data, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
				"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", object)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), tc.signature+" FOUND") {
				t.Fatalf("reference scan: error=%v output=%s", err, out)
			}
			t.Logf("reference detected: %s", out)
			copy(tc.data[len(tc.data)-2:], "XY")
			if err := os.WriteFile(object, tc.data, 0o600); err != nil {
				t.Fatal(err)
			}
			cmd = exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir,
				"--bytecode-unsigned=yes", "--bytecode=yes", "--no-summary", object)
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err = cmd.CombinedOutput()
			if err != nil || !strings.Contains(string(out), object+": OK") {
				t.Fatalf("reference no-trigger scan: error=%v output=%s", err, out)
			}
			t.Logf("reference no-trigger: %s", out)
		})
	}
}
