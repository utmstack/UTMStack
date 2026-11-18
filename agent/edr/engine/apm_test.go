package engine

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const apmFTM = "0:0:4552:Apple Partition Map:CL_TYPE_ANY:CL_TYPE_APM:77\n"

func apmTestImage(blockSize int, oldSchool bool) ([]byte, []byte) {
	img := make([]byte, 8*blockSize)
	copy(img, "ER")
	binary.BigEndian.PutUint16(img[2:], uint16(blockSize))
	binary.BigEndian.PutUint32(img[4:], 8)
	sector := blockSize
	if oldSchool {
		sector = 512
	}
	entry := func(index int, kind string, start, count uint32) {
		e := img[index*sector : index*sector+apmEntrySize]
		copy(e, "PM")
		binary.BigEndian.PutUint32(e[4:], 4)
		binary.BigEndian.PutUint32(e[8:], start)
		binary.BigEndian.PutUint32(e[12:], count)
		copy(e[16:], fmt.Sprintf("inert-%d", index))
		copy(e[48:], kind)
	}
	entry(1, "Apple_Partition_Map", 1, 1)
	entry(2, "Apple_Free", 0, 0)
	entry(3, "Apple_Free", 0, 0)
	start, size, kind, index := uint32(5), uint32(2), "Apple_HFS", 2
	if oldSchool {
		start, size, kind, index = 16, 1, "Apple_Driver43", 4
	}
	entry(index, kind, start, size)
	childSize := int(size) * sector
	if oldSchool {
		childSize = int(size) * 2048
	}
	part := img[int(start)*sector : int(start)*sector+childSize]
	copy(part[211:], "UTMSTACK-APM-INERT-CHILD-CONTENT")
	return img, part
}

func TestAPMPartitionHash(t *testing.T) {
	for _, tc := range []struct {
		name      string
		blockSize int
		oldSchool bool
	}{
		{"modern-512", 512, false},
		{"modern-2048", 2048, false},
		{"old-school-driver", 2048, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			img, part := apmTestImage(tc.blockSize, tc.oldSchool)
			sum := md5.Sum(part)
			files := map[string]string{
				"probe.ftm": apmFTM,
				"probe.hdb": fmt.Sprintf("%x:%d:APM.Child\n", sum, len(part)),
			}
			expectDetected(t, pipelineState(t, files), img, "APM.Child")
			control := append([]byte(nil), img...)
			start := 5 * tc.blockSize
			if tc.oldSchool {
				start = 16 * 512
			}
			control[start+211] ^= 1
			expectScan(t, pipelineState(t, files), control, "", "")
		})
	}
}

func TestAPMPartitionOverlapAndMalformed(t *testing.T) {
	img, _ := apmTestImage(512, false)
	files := map[string]string{"probe.ftm": apmFTM, "inert.ndb": "Inert.Pattern:0:*:55544d535441434b2d4e4f2d4d41544348\n"}
	state := pipelineState(t, files)
	expectScan(t, state, img, "", "")
	state = pipelineState(t, files)
	state.partitionIntersectionHeuristicOn = true
	expectScan(t, state, img, "", "")
	overlap := append([]byte(nil), img...)
	// Entry 3 overlaps the map entry at block 1. ClamAV checks the whole
	// table before scanning entry 2, including map and empty entries.
	binary.BigEndian.PutUint32(overlap[3*512+8:], 1)
	binary.BigEndian.PutUint32(overlap[3*512+12:], 1)
	state = pipelineState(t, files)
	state.partitionIntersectionHeuristicOn = true
	expectScan(t, state, overlap, "Heuristics.APMPartitionIntersection", "")
	bad := append([]byte(nil), img...)
	binary.BigEndian.PutUint32(bad[4:], 9)
	if sig, err := pipelineState(t, files).visit(bad, 0); sig != "" || !errors.Is(err, ErrMalformed) {
		t.Fatalf("invalid DDM length: sig=%q err=%v", sig, err)
	}
}

func TestAPMFallbackTyping(t *testing.T) {
	img, part := apmTestImage(512, false)
	sum := md5.Sum(part)
	expectDetected(t, pipelineState(t, map[string]string{
		"probe.hdb": fmt.Sprintf("%x:%d:APM.Fallback\n", sum, len(part)),
	}), img, "APM.Fallback")
}

func TestAPMPortableClamAVOracle(t *testing.T) {
	ref := os.Getenv("EDR_CLAMAV_PORTABLE")
	if ref == "" {
		t.Skip("EDR_CLAMAV_PORTABLE not set")
	}
	dir := t.TempDir()
	var paths []string
	for _, tc := range []struct {
		name      string
		blockSize int
		oldSchool bool
	}{
		{"modern-512", 512, false},
		{"modern-2048", 2048, false},
		{"old-school-driver", 2048, true},
	} {
		img, part := apmTestImage(tc.blockSize, tc.oldSchool)
		if err := os.WriteFile(filepath.Join(dir, tc.name+".img"), img, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, filepath.Join(dir, tc.name+".img"))
		control := append([]byte(nil), img...)
		start := 5 * tc.blockSize
		if tc.oldSchool {
			start = 16 * 512
		}
		control[start+211] ^= 1
		if err := os.WriteFile(filepath.Join(dir, tc.name+"-control.img"), control, 0o600); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, filepath.Join(dir, tc.name+"-control.img"))
		sum := md5.Sum(part)
		if err := os.WriteFile(filepath.Join(dir, tc.name+".hdb"), []byte(fmt.Sprintf("%x:%d:APM.%s\n", sum, len(part), tc.name)), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "probe.ftm"), []byte(apmFTM), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), append([]string{"--database=" + dir, "--no-summary"}, paths...)...)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err := cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok {
		t.Fatalf("reference exit: %v output=%s", err, out)
	}
	for _, name := range []string{"modern-512", "modern-2048", "old-school-driver"} {
		if !strings.Contains(string(out), name+".img: APM."+name+".UNOFFICIAL FOUND") {
			t.Fatalf("reference missed %s: %s", name, out)
		}
		if !strings.Contains(string(out), name+"-control.img: OK") {
			t.Fatalf("reference misclassified %s control: %s", name, out)
		}
	}
	img, _ := apmTestImage(512, false)
	binary.BigEndian.PutUint32(img[3*512+8:], 1)
	binary.BigEndian.PutUint32(img[3*512+12:], 1)
	overlap := filepath.Join(dir, "overlap.img")
	if err := os.WriteFile(overlap, img, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command(filepath.Join(ref, "bin", "clamscan"), "--database="+dir, "--no-summary", "--alert-partition-intersection=yes", "--heuristic-scan-precedence=yes", overlap)
	cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"), "CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
	out, err = cmd.CombinedOutput()
	if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "overlap.img: Heuristics.APMPartitionIntersection FOUND") {
		t.Fatalf("reference overlap: exit=%v output=%s", err, out)
	}
}
