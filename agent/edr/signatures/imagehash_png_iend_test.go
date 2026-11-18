package signatures

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPNGTerminalIENDCRCMatchesReference(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	checked := 0
	ruleChecked := false
	for _, c := range cases {
		if !strings.HasSuffix(c.name, ".png") {
			continue
		}
		if len(c.data) < 20 || !bytes.Equal(c.data[len(c.data)-12:len(c.data)-4],
			[]byte{0, 0, 0, 0, 'I', 'E', 'N', 'D'}) {
			t.Fatalf("%s has no terminal IEND", c.name)
		}
		bad := bytes.Clone(c.data)
		bad[len(bad)-4] ^= 1
		if _, err := png.Decode(bytes.NewReader(bad)); err == nil {
			t.Fatalf("%s was accepted by Go's ordinary PNG decoder", c.name)
		}
		got := fuzzyImageHash(bad, &matchWork{ctx: context.Background(), limit: 1 << 40})
		want, err := strconv.ParseUint(c.hash, 16, 64)
		if err != nil {
			t.Fatal(err)
		}
		if !got.ok || got.hash != want {
			t.Errorf("%s: repaired IEND hash %016x ok=%t; want %016x", c.name, got.hash, got.ok, want)
		}
		for kept := 0; kept < 4; kept++ {
			short := c.data[:len(c.data)-4+kept]
			if _, err := png.Decode(bytes.NewReader(short)); err == nil {
				t.Fatalf("%s IEND CRC prefix %d was accepted by ordinary PNG decoder", c.name, kept)
			}
			gotShort := fuzzyImageHash(short, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !gotShort.ok || gotShort.hash != want {
				t.Errorf("%s IEND CRC prefix %d: hash %016x ok=%t; want %016x", c.name, kept, gotShort.hash, gotShort.ok, want)
			}
		}
		for kept := 0; kept <= 4; kept++ {
			other := append(append([]byte(nil), c.data[:len(c.data)-8]...), []byte("XXXX")...)
			other = append(other, c.data[len(c.data)-4:len(c.data)-4+kept]...)
			if _, err := png.Decode(bytes.NewReader(other)); err == nil {
				t.Fatalf("%s unexpected terminal type with CRC prefix %d was accepted by Go's decoder", c.name, kept)
			}
			gotOther := fuzzyImageHash(other, &matchWork{ctx: context.Background(), limit: 1 << 40})
			if !gotOther.ok || gotOther.hash != want {
				t.Errorf("%s unexpected terminal type CRC prefix %d: hash %016x ok=%t; want %016x", c.name, kept, gotOther.hash, gotOther.ok, want)
			}
		}
		if !ruleChecked && got.ok && got.uncertain == 0 {
			rule := compileTest(t, "ldb", fmt.Sprintf("PNG.IEND;Engine:150-255,Target:0;0;fuzzy_img#%016x#0\n", want))
			if name, err := matchObject(t, rule, string(bad), MatchContext{Type: TypePNG}); name != "PNG.IEND" || err != nil {
				t.Fatalf("fuzzy rule on %s: %q, %v", c.name, name, err)
			}
			ruleChecked = true
		}
		checked++
	}
	if checked != 191 || !ruleChecked {
		t.Fatalf("checked %d PNGs, tested certain rule=%t; want 191 and true", checked, ruleChecked)
	}
}

func TestPNGTerminalIENDRepairDoesNotConcealOtherCorruption(t *testing.T) {
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	var data []byte
	for _, c := range cases {
		if strings.HasSuffix(c.name, ".png") {
			data = c.data
			break
		}
	}
	if data == nil {
		t.Fatal("no PNG oracle fixture")
	}
	badLength := bytes.Clone(data)
	badLength[len(badLength)-12] = 1
	if fuzzyImageHash(badLength, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
		t.Fatal("nonempty IEND was decoded")
	}
	if fuzzyImageHash(data[:len(data)-12], &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
		t.Fatal("missing IEND chunk was decoded")
	}
	emptyIDAT := append(append([]byte(nil), data[:len(data)-8]...), []byte("IDAT")...)
	emptyIDAT = append(emptyIDAT, data[len(data)-4:]...)
	if fuzzyImageHash(emptyIDAT, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
		t.Fatal("terminal empty IDAT was decoded")
	}
	badIDAT := bytes.Clone(data)
	for pos := 8; pos+12 <= len(badIDAT); {
		n := int(binary.BigEndian.Uint32(badIDAT[pos:]))
		if n > len(badIDAT)-pos-12 {
			t.Fatal("invalid PNG fixture")
		}
		if bytes.Equal(badIDAT[pos+4:pos+8], []byte("IDAT")) {
			badIDAT[pos+8+n] ^= 1
			break
		}
		pos += 12 + n
	}
	if fuzzyImageHash(badIDAT, &matchWork{ctx: context.Background(), limit: 1 << 40}).ok {
		t.Fatal("bad IDAT CRC was decoded")
	}
}

func TestPNGTerminalIENDClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV portable root")
	}
	cases, _ := loadImageOracleCases(t, "testdata/image_oracle_154.tar.gz")
	var sample imageCase
	for _, c := range cases {
		if strings.HasSuffix(c.name, ".png") {
			sample = c
			break
		}
	}
	if len(sample.data) == 0 {
		t.Fatal("no PNG oracle fixture")
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "local.ldb"),
		[]byte(fmt.Sprintf("PNG.IEND.Test;Engine:150-255,Target:0;0;fuzzy_img#%s#0\n", sample.hash)), 0o600); err != nil {
		t.Fatal(err)
	}
	bad := bytes.Clone(sample.data)
	bad[len(bad)-4] ^= 1
	badIDAT := bytes.Clone(sample.data)
	foundIDAT := false
	for pos := 8; pos+12 <= len(badIDAT); {
		n := int(binary.BigEndian.Uint32(badIDAT[pos:]))
		if n > len(badIDAT)-pos-12 {
			t.Fatal("invalid PNG fixture")
		}
		if bytes.Equal(badIDAT[pos+4:pos+8], []byte("IDAT")) {
			badIDAT[pos+8+n] ^= 1
			foundIDAT = true
			break
		}
		pos += 12 + n
	}
	if !foundIDAT {
		t.Fatal("no IDAT chunk")
	}
	emptyIDAT := append(append([]byte(nil), sample.data[:len(sample.data)-8]...), []byte("IDAT")...)
	emptyIDAT = append(emptyIDAT, sample.data[len(sample.data)-4:]...)
	variants := map[string][]byte{"valid": sample.data, "bad-iend": bad, "bad-idat": badIDAT,
		"no-iend": sample.data[:len(sample.data)-12], "empty-terminal-idat": emptyIDAT}
	for kept := 0; kept < 4; kept++ {
		variants[fmt.Sprintf("short-iend-%d", kept)] = sample.data[:len(sample.data)-4+kept]
	}
	for kept := 0; kept <= 4; kept++ {
		other := append(append([]byte(nil), sample.data[:len(sample.data)-8]...), []byte("XXXX")...)
		variants[fmt.Sprintf("other-terminal-%d", kept)] = append(other, sample.data[len(sample.data)-4:len(sample.data)-4+kept]...)
	}
	paths := make([]string, 0, len(variants))
	for name, data := range variants {
		path := filepath.Join(root, name+".png")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append([]string{"--database=" + root, "--no-summary"}, paths...)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference scan exit: %v output=%s", err, out)
	}
	for _, name := range []string{"valid", "bad-iend", "short-iend-0", "short-iend-1", "short-iend-2", "short-iend-3",
		"other-terminal-0", "other-terminal-1", "other-terminal-2", "other-terminal-3", "other-terminal-4"} {
		path := filepath.Join(root, name+".png")
		want := path + ": PNG.IEND.Test.UNOFFICIAL FOUND"
		if !strings.Contains(string(out), want) {
			t.Fatalf("reference missing %q in %s", want, out)
		}
	}
	if want := filepath.Join(root, "bad-idat.png") + ": OK"; !strings.Contains(string(out), want) {
		t.Fatalf("reference missing %q in %s", want, out)
	}
	if want := filepath.Join(root, "no-iend.png") + ": OK"; !strings.Contains(string(out), want) {
		t.Fatalf("reference missing %q in %s", want, out)
	}
	if want := filepath.Join(root, "empty-terminal-idat.png") + ": OK"; !strings.Contains(string(out), want) {
		t.Fatalf("reference missing %q in %s", want, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		t.Log(line)
	}
}
