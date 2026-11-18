package engine

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const windowMarker = "EDR-WINDOW-EMBEDDED-SAFE-MARKER-20261101"

func windowZIP(t *testing.T) []byte {
	t.Helper()
	payload := bytes.Repeat([]byte(windowMarker), 10)
	return append([]byte{0, 1, 2, 3}, formatRawZipFixture(t, 8, deflated(t, payload), payload)...)
}

func referenceWindowDetection(t *testing.T, rules map[string]string, input []byte) {
	t.Helper()
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		return
	}
	dir := t.TempDir()
	for name, content := range rules {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "input.fixture")
	if err := os.WriteFile(path, input, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", path)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 ||
		!strings.Contains(string(out), "Safe.Window.Marker.UNOFFICIAL FOUND") {
		t.Fatalf("ClamAV did not detect the extracted member: %v: %s", err, out)
	}
}

// ClamAV scans CPIO member windows with ordinary type recognition enabled.
// The old-CPIO parent disables that recognition, so only the member can find
// and extract this interior ZIP. Its marker is absent from the raw archive.
func TestOldCPIOMemberFindsEmbeddedZIP(t *testing.T) {
	member := windowZIP(t)
	archive := cpioArchive(cpioBinaryLE, []string{"window.bin"}, [][]byte{member}, true)
	if bytes.Contains(archive, []byte(windowMarker)) {
		t.Fatal("the marker is visible in the parent")
	}
	rules := map[string]string{
		"type.ftm": "0:0:c771:old CPIO:CL_TYPE_ANY:CL_TYPE_CPIO_OLD\n" +
			"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n",
		"mark.ndb": "Safe.Window.Marker:0:*:" + hx(windowMarker) + "\n",
	}
	s := pipelineState(t, rules)
	expectScan(t, s, archive, "Safe.Window.Marker", "")
	referenceWindowDetection(t, rules, archive)
}

// Fat Mach-O slices also use ordinary nested maps in ClamAV 1.5.4. The
// scanner must inspect a slice's own embedded types, even when it is a
// window into the universal file rather than an independently copied file.
func TestMachOUniversalSliceFindsEmbeddedZIP(t *testing.T) {
	member := windowZIP(t)
	fat := make([]byte, 32+len(member))
	copy(fat, []byte{0xca, 0xfe, 0xba, 0xbe})
	binary.BigEndian.PutUint32(fat[4:], 1)
	binary.BigEndian.PutUint32(fat[8:], 7)
	binary.BigEndian.PutUint32(fat[12:], 3)
	binary.BigEndian.PutUint32(fat[16:], 32)
	binary.BigEndian.PutUint32(fat[20:], uint32(len(member)))
	copy(fat[32:], member)
	if bytes.Contains(fat, []byte(windowMarker)) {
		t.Fatal("the marker is visible in the parent")
	}
	rules := map[string]string{
		"type.ftm": "0:0:cafebabe:fat Mach-O:CL_TYPE_ANY:CL_TYPE_MACHO_UNIBIN\n" +
			"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n",
		"mark.ndb": "Safe.Window.Marker:0:*:" + hx(windowMarker) + "\n",
	}
	s := pipelineState(t, rules)
	if sig, err := s.machoSlices(fat, 0); sig != "Safe.Window.Marker" || err != nil {
		t.Fatalf("Mach-O slice: signature=%q error=%v", sig, err)
	}
	referenceWindowDetection(t, rules, fat)
}

func partitionWindowRules() map[string]string {
	return map[string]string{
		"type.ftm": "1:510:55aa:MBR:CL_TYPE_ANY:CL_TYPE_MBR:77\n" +
			"1:*:504b0304:zip self-extractor:CL_TYPE_ANY:CL_TYPE_ZIPSFX\n",
		"mark.ndb": "Safe.Window.Marker:0:*:" + hx(windowMarker) + "\n",
	}
}

// MBR and GPT partitions are nested file-map windows in ClamAV 1.5.4.
// The parent disk image can encounter the same bytes in its raw pass, so
// invoke each partition reader directly to prove the nested path works.
func TestDiskPartitionWindowsFindEmbeddedZIP(t *testing.T) {
	member := windowZIP(t)
	for _, tc := range []struct {
		name  string
		image []byte
		scan  func(*scanState, []byte) (string, error)
	}{
		{"mbr-boot", func() []byte {
			image := make([]byte, 2*mbrSectorSize)
			copy(image[4:], member[4:])
			image[510], image[511] = 0x55, 0xaa
			return image
		}(), func(s *scanState, data []byte) (string, error) { return s.mbr(data, 0) }},
		{"mbr", func() []byte {
			image := make([]byte, 9*mbrSectorSize)
			image[450] = 0x83
			binary.LittleEndian.PutUint32(image[454:], 1)
			binary.LittleEndian.PutUint32(image[458:], 8)
			image[510], image[511] = 0x55, 0xaa
			copy(image[mbrSectorSize+4:], member[4:])
			return image
		}(), func(s *scanState, data []byte) (string, error) { return s.mbr(data, 0) }},
		{"gpt", func() []byte {
			image, part := gptTestImage()
			copy(part[4:], member[4:])
			return image
		}(), func(s *scanState, data []byte) (string, error) { return s.gpt(data, 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Contains(tc.image, []byte(windowMarker)) {
				t.Fatal("the marker is visible in the disk image")
			}
			rules := partitionWindowRules()
			referenceWindowDetection(t, rules, tc.image)
			s := pipelineState(t, rules)
			if sig, err := tc.scan(s, tc.image); sig != "Safe.Window.Marker" || err != nil {
				t.Fatalf("partition scan: signature=%q error=%v", sig, err)
			}
		})
	}
}
