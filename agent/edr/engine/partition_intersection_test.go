package engine

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func overlapTestDB() map[string]string {
	return map[string]string{
		"probe.ftm": "1:510:55aa:Disk Image - Master Boot Record:CL_TYPE_ANY:CL_TYPE_MBR:77\n",
		"probe.ndb": "Inert.Pattern:0:*:55544d535441434b2d4e4f2d4d41544348\n",
	}
}

func mbrOverlapImage(secondFirst uint32) []byte {
	img := make([]byte, 8*mbrSectorSize)
	for i, first := range [...]uint32{1, secondFirst} {
		entry := img[446+i*16:]
		entry[4] = 0x83
		binary.LittleEndian.PutUint32(entry[8:], first)
		binary.LittleEndian.PutUint32(entry[12:], [...]uint32{3, 2}[i])
	}
	img[510], img[511] = 0x55, 0xaa
	return img
}

func ebrOverlapImage(link uint32) []byte {
	img := make([]byte, 12*mbrSectorSize)
	img[450] = 0x05
	binary.LittleEndian.PutUint32(img[454:], 1)
	binary.LittleEndian.PutUint32(img[458:], 10)
	img[510], img[511] = 0x55, 0xaa
	first := img[mbrSectorSize : 2*mbrSectorSize]
	first[450] = 0x83
	binary.LittleEndian.PutUint32(first[454:], 1)
	binary.LittleEndian.PutUint32(first[458:], 3)
	first[466] = 0x05
	binary.LittleEndian.PutUint32(first[470:], link)
	binary.LittleEndian.PutUint32(first[474:], 6)
	first[510], first[511] = 0x55, 0xaa
	second := img[(1+int(link))*mbrSectorSize:]
	second[450] = 0x83
	binary.LittleEndian.PutUint32(second[454:], 1)
	binary.LittleEndian.PutUint32(second[458:], 2)
	second[510], second[511] = 0x55, 0xaa
	return img
}

func gptOverlapImage(secondFirst uint64) []byte {
	img, _ := gptTestImage()
	for _, tableAt := range [...]int{2, 62} {
		table := img[tableAt*512 : (tableAt+1)*512]
		entry := table[128:256]
		entry[0], entry[16] = 0xa2, 2
		binary.LittleEndian.PutUint64(entry[32:], secondFirst)
		binary.LittleEndian.PutUint64(entry[40:], 8)
		crc := crc32.ChecksumIEEE(table)
		headerAt := 1
		if tableAt == 62 {
			headerAt = 63
		}
		header := img[headerAt*512 : headerAt*512+gptHeaderSize]
		binary.LittleEndian.PutUint32(header[88:], crc)
		binary.LittleEndian.PutUint32(header[16:], 0)
		binary.LittleEndian.PutUint32(header[16:], crc32.ChecksumIEEE(header))
	}
	return img
}

func TestPartitionIntersectionAlertMatchesReference(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		image      []byte
	}{
		{"mbr-overlap", "Heuristics.MBRPartitionnIntersect", mbrOverlapImage(3)},
		{"mbr-adjacent", "", mbrOverlapImage(4)},
		{"ebr-overlap", "Heuristics.MBRPartitionnIntersect", ebrOverlapImage(2)},
		{"ebr-adjacent", "", ebrOverlapImage(3)},
		{"gpt-overlap", "Heuristics.GPTPartitionIntersection", gptOverlapImage(5)},
		{"gpt-adjacent", "", gptOverlapImage(6)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := pipelineState(t, overlapTestDB())
			expectScan(t, s, tc.image, "", "") // ClamAV default: alert off.
			s = pipelineState(t, overlapTestDB())
			s.partitionIntersectionHeuristicOn = true
			expectScan(t, s, tc.image, tc.want, "")
		})
	}
}

func TestPartitionIntersectionPolicyChangesVersion(t *testing.T) {
	dir := t.TempDir()
	for name, contents := range overlapTestDB() {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.SignatureDBDir = dir
	disabled := New(cfg)
	if err := disabled.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	on := true
	cfg.ScanHeuristics.PartitionIntersection = &on
	enabled := New(cfg)
	if err := enabled.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if disabled.Version() == enabled.Version() {
		t.Fatal("different heuristic policies share a verdict cache identity")
	}
	img := mbrOverlapImage(3)
	if clean, sig, err := disabled.Scan(context.Background(), img); !clean || sig != "" || err != nil {
		t.Fatalf("default policy: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := enabled.Scan(context.Background(), img); clean || sig != "Heuristics.MBRPartitionnIntersect" || err != nil {
		t.Fatalf("enabled policy: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

func TestPartitionIntersectionHonorsPartitionLimit(t *testing.T) {
	for _, img := range [][]byte{mbrOverlapImage(3), ebrOverlapImage(2), gptOverlapImage(5)} {
		s := pipelineState(t, overlapTestDB())
		s.partitionIntersectionHeuristicOn = true
		s.maxPartitions = 1
		expectScan(t, s, img, "", "")
	}
}
