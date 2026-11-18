package quarantine

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAdmissionSourceSyncFailurePreservesRecovery(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		name := "before-intent"
		if failAt == 2 {
			name = "after-capture"
		}
		t.Run(name, func(t *testing.T) {
			s, c, path, _ := admissionFixture(t)
			original, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			rawID, err := admissionFileID(original)
			original.Close()
			if err != nil {
				t.Fatal(err)
			}
			failure := errors.New("owned simulated source flush I/O failure")
			calls := 0
			s.admissionSync = func(f *os.File) error {
				calls++
				got, err := admissionFileID(f)
				if err != nil || got != rawID {
					t.Fatal("source barrier lost the pinned identity", got, err)
				}
				if calls == failAt {
					return failure
				}
				return admissionSyncSource(f)
			}
			id, err := s.Quarantine(path, payloadHash("malicious marker"), "Safe.Source.Sync.Control")
			if !errors.Is(err, failure) || calls != failAt {
				t.Fatal("source flush failure was hidden", id, err, calls)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatal("failed flush claimed committed quarantine", rows, err)
			}
			if failAt == 1 {
				var contained *ContainmentError
				if id != "" || errors.As(err, &contained) {
					t.Fatal("failed initial flush claimed a capture", id, err)
				}
				requirePurgePayload(t, path, "malicious marker")
				entries, err := os.ReadDir(filepath.Join(s.dir, ".admissions"))
				if !os.IsNotExist(err) && (err != nil || len(entries) != 0) {
					t.Fatal("failed initial flush wrote an admission operation", entries, err)
				}
				return
			}
			var contained *ContainmentError
			if id == "" || !errors.As(err, &contained) || contained.QuarantineID != id {
				t.Fatal("captured flush failure omitted physical containment", id, err)
			}
			journal := filepath.Join(s.dir, ".admissions", id, "manifest.json")
			before, err := os.ReadFile(journal)
			if err != nil {
				t.Fatal(err)
			}
			var m admissionManifest
			if err = json.Unmarshal(before, &m); err != nil || m.Phase != "ready" || m.Record.SourceFileID != rawID || m.SnapshotID == rawID {
				t.Fatal("failed flush advanced the admission journal", m, err)
			}
			raw := filepath.Join(m.Record.SourceCaptureDir, "source")
			requirePurgePayload(t, raw, "malicious marker")
			requirePurgePayload(t, filepath.Join(filepath.Dir(journal), "snapshot"), "malicious marker")
			s.admissionSync = func(*os.File) error { return failure }
			results, err := s.Recover()
			if err == nil || len(results) != 1 || results[0].Status != "needs_attention" {
				t.Fatal("recovery ignored source flush failure", results, err)
			}
			after, err := os.ReadFile(journal)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatal("failed recovery flush changed the retained journal", err)
			}
			if rows, err := c.ListQuarantine(); err != nil || len(rows) != 0 {
				t.Fatal("failed recovery flush claimed completion", rows, err)
			}
			s.admissionSync = nil
			if results, err = s.Recover(); err != nil {
				t.Fatal("source flush retry could not recover", results, err)
			}
			rec, found, err := c.GetQuarantine(id)
			if err != nil || !found || !rec.Restorable || rec.Restored || rec.Purged || rec.SourceFileID != rawID {
				t.Fatal("successful retry recorded the wrong result", rec, err)
			}
			if err = s.Restore(id); err != nil {
				t.Fatal(err)
			}
			requirePurgePayload(t, path, "malicious marker")
		})
	}
}
