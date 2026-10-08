package ransomware

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// flatBytes returns a 16 KiB buffer covering all 256 byte values uniformly
// (~8 bits/byte of Shannon entropy).
func flatBytes() []byte {
	buf := make([]byte, 16*1024)
	for i := range buf {
		buf[i] = byte(i)
	}
	return buf
}

func TestEntropy_FlatContentFires(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "victim.bin")
	if err := os.WriteFile(p, flatBytes(), 0600); err != nil {
		t.Fatal(err)
	}
	base := time.Unix(1000, 0)
	e := NewEntropySensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return base }, nil)
	ev := e.Record(FileEvent{PID: 1, Path: p, Op: OpWrite})
	if ev == nil || ev.Kind != KindEntropy {
		t.Fatalf("flat content did not fire: %+v", ev)
	}
	if ev.Weight != 40 {
		t.Fatalf("weight = %v, want default 40", ev.Weight)
	}
	// Per-path cooldown: an immediate re-record of the same path is silent.
	if ev2 := e.Record(FileEvent{PID: 1, Path: p, Op: OpWrite}); ev2 != nil {
		t.Fatalf("re-fired inside cooldown: %+v", ev2)
	}
}

func TestEntropy_TextContentDoesNotFire(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "notes.txt")
	text := make([]byte, 8192)
	for i := range text {
		text[i] = 'a' + byte(i%26)
	}
	if err := os.WriteFile(p, text, 0600); err != nil {
		t.Fatal(err)
	}
	base := time.Unix(1000, 0)
	e := NewEntropySensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return base }, nil)
	if ev := e.Record(FileEvent{PID: 2, Path: p, Op: OpWrite}); ev != nil {
		t.Fatalf("text content fired: %+v", ev)
	}
}

func TestEntropy_ShortFileNoPanic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "tiny.bin")
	// 100 flat distinct bytes → ~6.64 bits/byte (below the default 7.5), so
	// lower this test's threshold to 5.0: it proves the clamp-to-one-read
	// path samples and fires without panicking on a file smaller than the
	// sample size.
	if err := os.WriteFile(p, flatBytes()[:100], 0600); err != nil {
		t.Fatal(err)
	}
	b := config.DefaultBehaviorTuning()
	b.EntropyBitsPerByte = 5.0
	base := time.Unix(1000, 0)
	e := NewEntropySensor(cfgWithBehavior(b), func() time.Time { return base }, nil)
	ev := e.Record(FileEvent{PID: 3, Path: p, Op: OpWrite})
	if ev == nil {
		t.Fatal("short flat file did not fire under the lowered threshold")
	}
}

func TestEntropy_ReaderIOPerCallBounded(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "big.bin")
	if err := os.WriteFile(p, flatBytes(), 0600); err != nil {
		t.Fatal(err)
	}
	// Fake reader: every call is capped at the requested size, and we assert
	// each individual ReadAt is bounded by the sample size — the sensor never
	// reads whole files.
	calls := 0
	fake := func(path string, offset int64, size int) ([]byte, error) {
		calls++
		if size > 4096 {
			t.Errorf("requested %d bytes in one read, want ≤ 4096", size)
		}
		if offset < 0 || offset+int64(size) > 16*1024 {
			t.Errorf("read [%d:+%d) out of file bounds", offset, size)
		}
		return flatBytes()[offset : offset+int64(size)], nil
	}
	base := time.Unix(1000, 0)
	e := NewEntropySensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return base }, fake)
	ev := e.Record(FileEvent{PID: 4, Path: p, Op: OpWrite})
	if ev == nil {
		t.Fatal("flat content did not fire with fake reader")
	}
	if calls == 0 {
		t.Fatal("reader was never called")
	}
}

func TestEntropy_IgnoresNonWrites(t *testing.T) {
	base := time.Unix(1000, 0)
	e := NewEntropySensor(cfgWithBehavior(config.DefaultBehaviorTuning()), func() time.Time { return base }, nil)
	if ev := e.Record(FileEvent{PID: 5, Path: "/data/x", Op: OpRename}); ev != nil {
		t.Fatalf("rename sampled: %+v", ev)
	}
}
