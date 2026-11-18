package service

import (
	"context"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/feed"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
)

func TestDefaultTrustedProcessesMatchesFPWorkloads(t *testing.T) {
	tr := orchestrator.NewExcluder(defaultTrustedProcesses())
	trusted := []string{
		`C:\Program Files\Veeam\Backup\Veeam.Backup.Manager.exe`, // subtree
		`C:\Program Files (x86)\Acronis\TrueImage\service.exe`,
		`C:\Windows\System32\vssvc.exe`, // OS VSS service (full path)
		`C:\Program Files\Microsoft OneDrive\OneDrive.exe`,
	}
	for _, p := range trusted {
		if !tr.Excluded(p) {
			t.Errorf("expected built-in trust to cover %q", p)
		}
	}
	// Must NOT trust the T1490 tells themselves, or arbitrary user binaries.
	for _, p := range []string{
		`C:\Windows\System32\vssadmin.exe`,
		`C:\Windows\System32\wbadmin.exe`,
		`C:\Users\victim\Downloads\enc.exe`,
	} {
		if tr.Excluded(p) {
			t.Errorf("built-in trust must NOT cover %q", p)
		}
	}
}

type activeStore struct{}

func (activeStore) FeedState() (uint64, string) { return 2, "digest" }
func (activeStore) Version() string             { return "active" }
func (activeStore) ActivateGeneration(map[string][]byte, uint64, string) (string, error) {
	return "", nil
}

// status.json reports when the serving generation was activated as the last
// update when this run has not refreshed yet, and a failed refresh while that
// generation serves as refresh_failed, not as unavailable updates.
func TestSignatureStatusAfterRestartWithFailedRefresh(t *testing.T) {
	activated := time.Date(2026, 9, 24, 2, 45, 47, 0, time.UTC)

	var disabled statusDoc
	signatureStatus(&disabled, nil, activated)
	if disabled.SignatureLastUpdate != "2026-09-24T02:45:47Z" || disabled.SignatureGenerationActivated != "2026-09-24T02:45:47Z" || disabled.SignatureUpdateState != "" {
		t.Fatalf("module disabled: %+v", disabled)
	}

	sf := feed.New(config.Default(), nil, activeStore{})
	if _, err := sf.UpdateOnce(context.Background()); err == nil {
		t.Fatal("refresh without a mirror succeeded")
	}
	var doc statusDoc
	signatureStatus(&doc, sf, activated)
	if doc.SignatureLastUpdate != "2026-09-24T02:45:47Z" {
		t.Fatalf("last update = %q, want the activation time", doc.SignatureLastUpdate)
	}
	if doc.SignatureUpdateState != feed.UpdateRefreshFailed || doc.SignatureUpdateError == "" {
		t.Fatalf("state = %q, error = %q", doc.SignatureUpdateState, doc.SignatureUpdateError)
	}

	var none statusDoc
	signatureStatus(&none, nil, time.Time{})
	if none.SignatureLastUpdate != "" || none.SignatureGenerationActivated != "" {
		t.Fatalf("no generation: %+v", none)
	}
}
