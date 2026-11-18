package engine

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestPartitionCountControlsLaterDetection(t *testing.T) {
	mbr := mbrOverlapImage(4)
	mbrChild := mbr[4*mbrSectorSize : 6*mbrSectorSize]
	copy(mbrChild[233:], "UTMSTACK-LATER-MBR-PARTITION")
	gpt := gptOverlapImage(6)
	gptChild := gpt[6*mbrSectorSize : 9*mbrSectorSize]
	copy(gptChild[233:], "UTMSTACK-LATER-GPT-PARTITION")
	ebr, ebrChild := mbrLinkedTestImage()
	for _, tc := range []struct {
		name       string
		img, child []byte
	}{
		{"mbr", mbr, mbrChild},
		{"gpt", gpt, gptChild},
		{"ebr", ebr, ebrChild},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hash := md5.Sum(tc.child)
			files := overlapTestDB()
			files["probe.hdb"] = fmt.Sprintf("%x:%d:Later.Partition\n", hash, len(tc.child))
			one := pipelineState(t, files)
			one.maxPartitions = 1
			expectScan(t, one, tc.img, "", "")
			two := pipelineState(t, files)
			two.maxPartitions = 2
			expectScan(t, two, tc.img, "Later.Partition", "")
		})
	}
}

func TestPartitionCountConfigChangesVersion(t *testing.T) {
	img := mbrOverlapImage(4)
	child := img[4*mbrSectorSize : 6*mbrSectorSize]
	copy(child[233:], "UTMSTACK-LATER-MBR-PARTITION")
	hash := md5.Sum(child)
	files := overlapTestDB()
	files["probe.hdb"] = fmt.Sprintf("%x:%d:Later.Partition\n", hash, len(child))
	dir := t.TempDir()
	for name, contents := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Default()
	cfg.SignatureDBDir = dir
	full := New(cfg)
	if err := full.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	cfg.ScanLimits.MaxPartitions = 1
	limited := New(cfg)
	if err := limited.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if full.Version() == limited.Version() {
		t.Fatal("partition limit missing from verdict cache identity")
	}
	if clean, sig, err := full.Scan(context.Background(), img); clean || sig != "Later.Partition" || err != nil {
		t.Fatalf("default limit: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := limited.Scan(context.Background(), img); !clean || sig != "" || err != nil {
		t.Fatalf("one-partition limit: clean=%v sig=%q err=%v", clean, sig, err)
	}
}
