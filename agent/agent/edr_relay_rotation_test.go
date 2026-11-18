package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/event"
)

func TestEDRRelayRetainsAcceptedRecordsAcrossRotationAndBackpressure(t *testing.T) {
	dir := t.TempDir()
	path, offset := filepath.Join(dir, "events.ndjson"), filepath.Join(dir, ".offset")
	sp, err := event.OpenSpool(path, 128)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	var accepted []string
	for i := 0; i < 12; i++ {
		line := fmt.Sprintf("%02d-%s", i, strings.Repeat("x", 59))
		if err := sp.Append(line); err == nil {
			accepted = append(accepted, line)
		}
		if _, err := drainSpool(path, offset, func(string) bool { return false }); err != nil {
			t.Fatal(err)
		}
	}
	if len(accepted) < 3 {
		t.Fatalf("did not exercise rotation: accepted=%d", len(accepted))
	}
	var got []string
	if _, err := drainSpool(path, offset, func(s string) bool { got = append(got, s); return true }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, accepted) {
		t.Fatalf("accepted records lost across rotation: got %d want %d", len(got), len(accepted))
	}
	got = nil
	if _, err := drainSpool(path, offset, func(s string) bool { got = append(got, s); return true }); err != nil || len(got) != 0 {
		t.Fatalf("already acknowledged records replayed: %d %v", len(got), err)
	}
}

func TestEDRRelayCursorFollowsPartiallyDrainedFileIntoBackup(t *testing.T) {
	dir := t.TempDir()
	path, offset := filepath.Join(dir, "events.ndjson"), filepath.Join(dir, ".offset")
	sp, err := event.OpenSpool(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	for _, s := range []string{"one", "two"} {
		if err := sp.Append(s); err != nil {
			t.Fatal(err)
		}
	}
	var got []string
	if _, err := drainSpool(path, offset, func(s string) bool {
		if s == "two" {
			return false
		}
		got = append(got, s)
		return true
	}); err != nil {
		t.Fatal(err)
	}
	if err := sp.Append("new"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".1"); err != nil {
		t.Fatal("rotation missing", err)
	}
	if _, err := drainSpool(path, offset, func(s string) bool { got = append(got, s); return true }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"one", "two", "new"}) {
		t.Fatalf("cursor lost rotation: %v", got)
	}
}
