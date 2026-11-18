// edr/internal/sigs/sigs_test.go
package sigs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

func writeFakeCVD(t *testing.T, dir, name, version string) {
	t.Helper()
	header := "ClamAV-VDB:09 Jul 2026 10-00 -0400:" + version + ":100:63:...:sig:builder:1720000000"
	pad := make([]byte, 512-len(header))
	if err := os.WriteFile(filepath.Join(dir, name), append([]byte(header), pad...), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadCVDVersion(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	v, err := ReadCVDVersion(filepath.Join(dir, "daily.cvd"))
	if err != nil || v != "27461" {
		t.Fatalf("version = %q, err %v", v, err)
	}
}

func TestDatabaseVersionsListsCVDs(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	writeFakeCVD(t, dir, "main.cvd", "62")
	_ = os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644)
	dbs := DatabaseVersions(dir)
	if dbs["daily.cvd"] != "27461" || dbs["main.cvd"] != "62" || len(dbs) != 2 {
		t.Fatalf("dbs = %v", dbs)
	}
}

func TestSyncOnceRunsCvdUpdateAndRecords(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	var calls [][]string
	s := &Syncer{
		DBDir: dir,
		Rec:   mirror.NewRecorder(t.TempDir()),
		Run: func(name string, args ...string) (string, error) {
			calls = append(calls, append([]string{name}, args...))
			return "ok", nil
		},
	}
	if err := s.EnsureConfig(); err != nil {
		t.Fatal(err)
	}
	if err := s.SyncOnce(); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0][0] != "cvd" || calls[1][1] != "update" {
		t.Fatalf("calls = %v", calls)
	}
}
