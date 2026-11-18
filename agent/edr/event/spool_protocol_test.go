package event

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func pairedSpool(t *testing.T, limit int64) (*Spool, string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.ndjson")
	offset := filepath.Join(filepath.Dir(path), ".offset")
	s, err := OpenSpool(path, limit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	if _, err := DrainSpool(context.Background(), path, offset, func(string) bool { return false }); err != nil {
		t.Fatal(err)
	}
	return s, path, offset
}
func drainAll(t *testing.T, path, offset string) []string {
	t.Helper()
	var got []string
	for i := 0; i < 8; i++ {
		before := len(got)
		_, err := DrainSpool(context.Background(), path, offset, func(s string) bool { got = append(got, s); return true })
		if err != nil {
			t.Fatal(err)
		}
		if before == len(got) {
			return got
		}
	}
	t.Fatal("did not drain bounded fixture")
	return nil
}
func TestSpoolLegacyProducerWaitsForUpgradedRelay(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.ndjson")
	offset := filepath.Join(filepath.Dir(path), ".offset")
	s, err := OpenSpool(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, v := range []string{"one", "two"} {
		if err = s.Append(v); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.Append("new"); !errors.Is(err, ErrSpoolFull) {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "one\ntwo\n" {
		t.Fatalf("older relay no longer sees plain events: %q", raw)
	}
	// An older unbound cursor cannot prove which retained segment was read.
	if err = os.WriteFile(offset, []byte("8"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatal(got)
	}
	if err = s.Append("new"); err != nil {
		t.Fatal(err)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"new"}) {
		t.Fatal(got)
	}
	if s.Snapshot().LegacyMode {
		t.Fatal("producer did not observe upgraded relay")
	}
}
func TestSpoolLegacyBackupMigrationReplaysWithoutDiscardingBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.ndjson")
	offset := filepath.Join(filepath.Dir(path), ".offset")
	for p, data := range map[string]string{path: "new\n", path + ".1": "old\n", offset: "4"} {
		if err := os.WriteFile(p, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	s, err := OpenSpool(path, 8)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"old", "new"}) {
		t.Fatal(got)
	}
	if _, err = os.Stat(path + ".1"); !os.IsNotExist(err) {
		t.Fatal("acknowledged backup retained", err)
	}
}
func TestSpoolInvalidRecordsDoNotChangeAcceptedData(t *testing.T) {
	s, path, offset := pairedSpool(t, 8)
	for _, bad := range []string{"", "a\nb", "a\rb", "12345678", strings.Repeat("x", MaxSpoolRecordBytes)} {
		if err := s.Append(bad); !errors.Is(err, ErrSpoolRecord) {
			t.Fatal(err)
		}
	}
	if got := drainAll(t, path, offset); len(got) != 0 {
		t.Fatal(got)
	}
	if stat := s.Snapshot(); stat.Accepted != 0 || stat.InvalidRejections != 5 || !stat.Degraded {
		t.Fatal(stat)
	}
}
func TestSpoolFailedRotationPreservesCurrentAndCanRetry(t *testing.T) {
	s, path, offset := pairedSpool(t, 4)
	if err := s.Append("one"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path+".1", 0700); err != nil {
		t.Fatal(err)
	}
	if err := s.Append("two"); !errors.Is(err, ErrSpoolIntegrity) {
		t.Fatal(err)
	}
	if err := os.Remove(path + ".1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Append("two"); err != nil {
		t.Fatal(err)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"one", "two"}) {
		t.Fatal(got)
	}
}
func TestSpoolCursorFailureReplaysAcceptedBatchAndKeepsBackup(t *testing.T) {
	s, path, offset := pairedSpool(t, 4)
	if err := s.Append("one"); err != nil {
		t.Fatal(err)
	}
	if err := s.Append("two"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(offset+".next", 0700); err != nil {
		t.Fatal(err)
	}
	var got []string
	_, err := DrainSpool(context.Background(), path, offset, func(v string) bool { got = append(got, v); return true })
	if err == nil || !reflect.DeepEqual(got, []string{"one"}) {
		t.Fatalf("got %v error %v", got, err)
	}
	if _, err = os.Stat(path + ".1"); err != nil {
		t.Fatal("unacknowledged backup lost", err)
	}
	if err = os.Remove(offset + ".next"); err != nil {
		t.Fatal(err)
	}
	got = append(got, drainAll(t, path, offset)...)
	if !reflect.DeepEqual(got, []string{"one", "one", "two"}) {
		t.Fatal(got)
	}
}
func TestSpoolLockContentionCancellationAndRetry(t *testing.T) {
	s, path, offset := pairedSpool(t, 64)
	unlock, err := lockSpool(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err = s.Append("busy")
	elapsed := time.Since(start)
	unlock()
	if !errors.Is(err, ErrSpoolBusy) || elapsed > time.Second {
		t.Fatalf("error=%v elapsed=%s", err, elapsed)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = DrainSpool(ctx, path, offset, func(string) bool { t.Error("emitted after cancellation"); return true }); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = s.Append("after"); err != nil {
		t.Fatal(err)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"after"}) {
		t.Fatal(got)
	}
	if stat := s.Snapshot(); stat.BusyRejections != 1 || stat.Accepted != 1 || !stat.Degraded {
		t.Fatal(stat)
	}
}
func TestSpoolPartialTailAndUnalignedCursorAreNotSkipped(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.ndjson")
	offset := filepath.Join(filepath.Dir(path), ".offset")
	if err := os.WriteFile(path, []byte("good\npartial"), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := OpenSpool(path, 64)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	var initial []string
	_, err = DrainSpool(context.Background(), path, offset, func(v string) bool { initial = append(initial, v); return true })
	if !errors.Is(err, ErrSpoolIntegrity) || !reflect.DeepEqual(initial, []string{"good"}) {
		t.Fatalf("partial tail: %v %v", initial, err)
	}
	if err = s.Append("new"); !errors.Is(err, ErrSpoolIntegrity) {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.WriteString("\n")
	if e := f.Close(); err != nil || e != nil {
		t.Fatal(err, e)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"partial"}) {
		t.Fatal(got)
	}
	cursor, err := readSpoolCursor(offset)
	if err != nil {
		t.Fatal(err)
	}
	cursor.Offset = 1
	if err = writeSpoolCursor(offset, cursor); err != nil {
		t.Fatal(err)
	}
	_, err = DrainSpool(context.Background(), path, offset, func(string) bool { t.Error("emitted from unaligned cursor"); return true })
	if !errors.Is(err, ErrSpoolIntegrity) {
		t.Fatal(err)
	}
}
func TestSpoolConcurrentProducersResumeAfterBackpressure(t *testing.T) {
	first, path, offset := pairedSpool(t, 64)
	second, err := OpenSpool(path, 64)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var mu sync.Mutex
	accepted := map[string]bool{}
	var got []string
	pressure := make(chan struct{})
	var pressureOnce sync.Once
	var writers sync.WaitGroup
	for p, s := range []*Spool{first, second} {
		writers.Add(1)
		go func(p int, s *Spool) {
			defer writers.Done()
			for i := 0; i < 64; i++ {
				v := fmt.Sprintf("p%d-%03d", p, i)
				for {
					err := s.Append(v)
					if err == nil {
						mu.Lock()
						accepted[v] = true
						mu.Unlock()
						break
					}
					if !errors.Is(err, ErrSpoolFull) && !errors.Is(err, ErrSpoolBusy) {
						t.Error(err)
						return
					}
					pressureOnce.Do(func() { close(pressure) })
					select {
					case <-ctx.Done():
						t.Error("producer deadline")
						return
					case <-time.After(time.Millisecond):
					}
				}
			}
		}(p, s)
	}
	select {
	case <-pressure:
	case <-ctx.Done():
		t.Fatal("did not exercise backpressure")
	}
	writersDone := make(chan struct{})
	go func() { writers.Wait(); close(writersDone) }()
	for {
		_, err := DrainSpool(ctx, path, offset, func(v string) bool { got = append(got, v); return true })
		if err != nil && !errors.Is(err, ErrSpoolBusy) {
			t.Fatal(err)
		}
		select {
		case <-writersDone:
			got = append(got, drainAll(t, path, offset)...)
			goto complete
		case <-ctx.Done():
			t.Fatal("drain deadline")
		case <-time.After(time.Millisecond):
		}
	}
complete:
	mu.Lock()
	defer mu.Unlock()
	if len(accepted) != 128 || len(got) != 128 {
		t.Fatalf("accepted=%d delivered=%d", len(accepted), len(got))
	}
	for _, v := range got {
		if !accepted[v] {
			t.Fatalf("duplicate or unaccepted %q", v)
		}
		delete(accepted, v)
	}
}
func TestSpoolCrashHelper(t *testing.T) {
	role := os.Getenv("OPENEDR_SPOOL_CRASH_ROLE")
	if role == "" {
		t.Skip("subprocess helper")
	}
	path := os.Getenv("OPENEDR_SPOOL_CRASH_PATH")
	offset := filepath.Join(filepath.Dir(path), ".offset")
	marker, err := os.ReadFile(filepath.Join(filepath.Dir(path), "owned-fixture"))
	if err != nil || string(marker) != "OPENEDR_HARMLESS_SPOOL_FIXTURE" {
		t.Fatal("not an owned fixture")
	}
	switch role {
	case "producer":
		s, err := OpenSpool(path, 64)
		if err != nil {
			t.Fatal(err)
		}
		for _, v := range []string{"one", "two", "three"} {
			if err = s.Append(v); err != nil {
				t.Fatal(err)
			}
		}
		os.Exit(97)
	case "relay":
		_, err := DrainSpool(context.Background(), path, offset, func(v string) bool {
			f, err := os.OpenFile(filepath.Join(filepath.Dir(path), "accepted-before-crash"), os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				t.Fatal(err)
			}
			_, err = f.WriteString(v)
			if err != nil {
				t.Fatal(err)
			}
			if err = f.Sync(); err != nil {
				t.Fatal(err)
			}
			f.Close()
			os.Exit(97)
			return true
		})
		t.Fatal("relay failed to reach crash", err)
	default:
		t.Fatal("invalid fixture role")
	}
}
func TestSpoolProcessCrashPreservesRecordsBeforeAcknowledgement(t *testing.T) {
	_, path, offset := pairedSpool(t, 64)
	if err := os.WriteFile(filepath.Join(filepath.Dir(path), "owned-fixture"), []byte("OPENEDR_HARMLESS_SPOOL_FIXTURE"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"producer", "relay"} {
		cmd := exec.Command(os.Args[0], "-test.run=^TestSpoolCrashHelper$", "-test.timeout=10s")
		cmd.Env = append(os.Environ(), "OPENEDR_SPOOL_CRASH_ROLE="+role, "OPENEDR_SPOOL_CRASH_PATH="+path)
		out, err := cmd.CombinedOutput()
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 97 {
			t.Fatalf("%s exit=%v output=%s", role, err, out)
		}
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(path), "accepted-before-crash"))
	if err != nil || string(raw) != "one" {
		t.Fatalf("receipt %q %v", raw, err)
	}
	if got := drainAll(t, path, offset); !reflect.DeepEqual(got, []string{"one", "two", "three"}) {
		t.Fatal(got)
	}
}
