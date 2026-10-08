//go:build windows || linux

package ransomware

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestIsRansomNote(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"README_RESTORE.txt", true},
		{"How_To_Decrypt.html", true},
		{"recovery.md", true},
		{"README.bin", true}, // README_ prefix is explicit, any extension
		{"BITCOIN_WALLET.txt", true},
		{"notes.txt", false},
		{"report.docx", false},
		{"deleted_files.txt", true},
		{"deleted_files.pdf", false}, // weak token + binary extension
		{"README_how_to_restore.TXT", true},
		{"RESTORE.YOUR.DATA.md", true},
		{"how_to_decrypt.bmp", false}, // how_to weak + binary; no HOW_TO_ prefix
		{"encryptor.tmp", false},
		{"payment_information.txt", true},
	}
	for _, c := range cases {
		if got := IsRansomNote(c.name); got != c.want {
			t.Errorf("IsRansomNote(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestRansomNote_RecordFiresOncePerPath(t *testing.T) {
	base := time.Unix(1000, 0)
	r := NewRansomNoteSensor(config.Default(), func() time.Time { return base })

	ev := r.Record(FileEvent{PID: 5, Path: `/data/README_RESTORE.txt`, Op: OpCreate})
	if ev == nil || ev.Kind != KindRansomNote || ev.Detail != "README_RESTORE.txt" {
		t.Fatalf("first note write did not fire: %+v", ev)
	}
	// Same PID+path again → deduped.
	if ev := r.Record(FileEvent{PID: 5, Path: `/data/README_RESTORE.txt`, Op: OpWrite}); ev != nil {
		t.Fatalf("repeat note write fired again: %+v", ev)
	}
	// A different path is a new signal.
	if ev := r.Record(FileEvent{PID: 5, Path: `/data/RECOVER.md`, Op: OpCreate}); ev == nil {
		t.Fatal("second note path did not fire")
	}
	// Non-mutating ops never fire.
	if ev := r.Record(FileEvent{PID: 5, Path: `/data/README_RESTORE.txt`, Op: OpDelete}); ev != nil {
		t.Fatalf("delete fired: %+v", ev)
	}
	// A non-note file never fires.
	if ev := r.Record(FileEvent{PID: 5, Path: `/data/report.docx`, Op: OpWrite}); ev != nil {
		t.Fatalf("normal doc fired: %+v", ev)
	}
}
