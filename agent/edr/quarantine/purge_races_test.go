package quarantine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPurgeClaimRacePreservesReplacement(t *testing.T) {
	s, c, id, _, stored := restoreFixture(t)
	saved := stored + ".original"
	fired := false
	s.responseFault = func(phase string) error {
		if phase == "stored-claim-intent" && !fired {
			fired = true
			if err := os.Rename(stored, saved); err != nil {
				return err
			}
			return os.WriteFile(stored, []byte("benign raced replacement"), 0600)
		}
		return nil
	}
	if err := s.Purge(id); err == nil || !fired {
		t.Fatalf("replacement race was not surfaced: %v", err)
	}
	requirePurgePayload(t, stored, "benign raced replacement")
	requirePurgeRecord(t, c, id, false)
	s.responseFault = nil
	if _, err := s.Recover(); err == nil {
		t.Fatal("unrepaired replacement was silently consumed")
	}
	foreign := filepath.Join(filepath.Dir(s.dir), "foreign-preserved")
	if err := os.Rename(stored, foreign); err != nil {
		t.Fatal(err)
	}
	// Windows claims the pinned original handle; Unix claims then verifies the
	// namespace and rolls a replacement back without deleting it.
	if _, err := os.Stat(saved); err == nil {
		if err := os.Rename(saved, stored); err != nil {
			t.Fatal(err)
		}
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if _, err := s.Recover(); err != nil {
		t.Fatalf("repaired approved operation did not finish: %v", err)
	}
	requirePurgeRecord(t, c, id, true)
	requirePurgePayload(t, foreign, "benign raced replacement")
}

func TestPurgeCaptureReplacementStopsBeforeDeletion(t *testing.T) {
	s, c, id, _, stored := restoreFixture(t)
	r, _, _ := c.GetQuarantine(id)
	saved := r.SourceCaptureDir + ".saved"
	if err := os.Rename(r.SourceCaptureDir, saved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(r.SourceCaptureDir, 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(r.SourceCaptureDir, "source")
	if err := os.WriteFile(foreign, []byte("benign capture replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Purge(id); err == nil {
		t.Fatal("foreign capture was accepted")
	}
	requirePurgePayload(t, stored, "payload")
	requirePurgePayload(t, filepath.Join(saved, "source"), "payload")
	requirePurgePayload(t, foreign, "benign capture replacement")
	requirePurgeRecord(t, c, id, false)
}

func TestPendingPurgeSerializesOtherStoresAndRecovery(t *testing.T) {
	s, c, id, _, _ := restoreFixture(t)
	other, err := New(s.dir, c)
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	s.responseFault = func(phase string) error {
		if phase == "stored-claimed" {
			close(entered)
			<-release
		}
		return nil
	}
	finished := make(chan error, 1)
	go func() { finished <- s.Purge(id) }()
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	select {
	case <-entered:
	case err := <-finished:
		t.Fatalf("purge did not reach its private claim: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("purge stalled")
	}
	for _, call := range []func(string) error{other.Restore, other.Purge} {
		if err := call(id); !errors.Is(err, ErrQuarantineBusy) {
			t.Errorf("competing management operation ignored item lock: %v", err)
		}
	}
	if results, err := other.Recover(); err != nil || len(results) != 0 {
		t.Fatalf("recovery took over live work: %+v %v", results, err)
	}
	close(release)
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("purge failed to release its item lock")
	}
	requirePurgeRecord(t, c, id, true)
}
