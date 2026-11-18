package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestDrainSpoolRetriesRejectedLineWithoutAdvancingPastIt(t *testing.T) {
	dir := t.TempDir()
	spool, offset := filepath.Join(dir, "events.ndjson"), filepath.Join(dir, ".offset")
	if err := os.WriteFile(spool, []byte("first\nsecond\nthird\npartial"), 0600); err != nil {
		t.Fatal(err)
	}
	var accepted []string
	pos, err := drainSpool(spool, offset, func(raw string) bool {
		if raw == "second" {
			return false // queue is full, so nothing accepted this record
		}
		accepted = append(accepted, raw)
		return true
	})
	if err != nil || pos != int64(len("first\n")) || readOffset(offset) != pos || !reflect.DeepEqual(accepted, []string{"first"}) {
		t.Fatalf("first drain: pos=%d persisted=%d accepted=%v err=%v", pos, readOffset(offset), accepted, err)
	}
	pos, err = drainSpool(spool, offset, func(raw string) bool {
		accepted = append(accepted, raw)
		return true
	})
	if err != nil || pos != int64(len("first\nsecond\nthird\n")) || !reflect.DeepEqual(accepted, []string{"first", "second", "third"}) {
		t.Fatalf("retry: pos=%d accepted=%v err=%v", pos, accepted, err)
	}
	// A partial final record remains unread until its newline arrives.
	f, err := os.OpenFile(spool, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString("\n"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = drainSpool(spool, offset, func(raw string) bool {
		accepted = append(accepted, raw)
		return true
	})
	if err != nil || !reflect.DeepEqual(accepted, []string{"first", "second", "third", "partial"}) {
		t.Fatalf("completed line: accepted=%v err=%v", accepted, err)
	}
}

func readOffset(path string) int64 {
	raw, _ := os.ReadFile(path)
	var cursor struct {
		Offset int64 `json:"offset"`
	}
	if json.Unmarshal(raw, &cursor) == nil {
		return cursor.Offset
	}
	n, _ := strconv.ParseInt(strings.TrimSpace(string(raw)), 10, 64)
	return n
}
