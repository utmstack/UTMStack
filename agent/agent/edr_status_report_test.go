package agent

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
)

// edrStatusTestSender implements resultSender and records what was sent,
// thread-safe (the reporter runs in its own goroutine).
type edrStatusTestSender struct {
	mu   sync.Mutex
	sent []*BidirectionalStream
}

func (s *edrStatusTestSender) Send(m *BidirectionalStream) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, m)
	return nil
}

func (s *edrStatusTestSender) reports() []*EdrStatusReport {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*EdrStatusReport, 0, len(s.sent))
	for _, m := range s.sent {
		if r := m.GetEdrStatusReport(); r != nil {
			out = append(out, r)
		}
	}
	return out
}

// flakyEdrSender fails exactly one Send, to prove a delivery error does not
// kill the reporter goroutine.
type flakyEdrSender struct {
	mu       sync.Mutex
	failOnce bool
	sent     []*BidirectionalStream
}

func (f *flakyEdrSender) Send(m *BidirectionalStream) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failOnce {
		f.failOnce = false
		return errors.New("stream broken")
	}
	f.sent = append(f.sent, m)
	return nil
}

func (f *flakyEdrSender) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// stubEdrStatusPaths points the reporter's file and interval seams at test
// values and restores them on cleanup.
func stubEdrStatusPaths(t *testing.T, file string, poll, report time.Duration) {
	t.Helper()
	oldFile := edrStatusFile
	oldPoll := edrStatusPollInterval
	oldReport := edrStatusReportInterval
	edrStatusFile = file
	edrStatusPollInterval = poll
	edrStatusReportInterval = report
	t.Cleanup(func() {
		edrStatusFile = oldFile
		edrStatusPollInterval = oldPoll
		edrStatusReportInterval = oldReport
	})
}

// stubPolicyVersion redirects the config seam so the report carries a known
// policy version without touching the real edr.json.
func stubPolicyVersion(t *testing.T, version string) {
	t.Helper()
	old := edrCfgLoad
	edrCfgLoad = func() (edrconfig.EDRConfig, error) {
		return edrconfig.EDRConfig{PolicyVersion: version}, nil
	}
	t.Cleanup(func() { edrCfgLoad = old })
}

// startEdrReporter runs the reporter in a goroutine and returns a stop func.
// Poll and report intervals come from stubEdrStatusPaths.
func startEdrReporter(t *testing.T, sender resultSender) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	go sendEdrStatusReports(ctx, sender, &config.Config{AgentID: 42})
	return func() {
		cancel()
	}
}

// waitEdrReports polls until the sender has captured n reports.
func waitEdrReports(t *testing.T, s *edrStatusTestSender, n int) []*EdrStatusReport {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if got := s.reports(); len(got) >= n {
			return got[:n]
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d reports, got %d", n, len(s.reports()))
	return nil
}

func writeStatusFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEdrStatusReportSendsOnFirstTick(t *testing.T) {
	path := writeStatusFile(t, t.TempDir(), "status.json", `{"engine":"clamav","version":"1.2.3"}`)
	stubEdrStatusPaths(t, path, 10*time.Millisecond, time.Hour)
	stubPolicyVersion(t, "v9")

	sender := &edrStatusTestSender{}
	stop := startEdrReporter(t, sender)
	defer stop()

	reports := waitEdrReports(t, sender, 1)
	r := reports[0]
	if r.GetStatusJson() != `{"engine":"clamav","version":"1.2.3"}` {
		t.Fatalf("StatusJson=%q want the file content", r.GetStatusJson())
	}
	if r.GetPolicyVersion() != "v9" {
		t.Fatalf("PolicyVersion=%q want v9 from the config", r.GetPolicyVersion())
	}
	if r.GetAgentId() != "42" {
		t.Fatalf("AgentId=%q want 42", r.GetAgentId())
	}
	if r.GetReportedAt() == nil {
		t.Fatal("ReportedAt is nil")
	}
}

// Unchanged content within the report interval is not re-sent.
func TestEdrStatusReportNoResendUnchanged(t *testing.T) {
	path := writeStatusFile(t, t.TempDir(), "status.json", `{"a":1}`)
	stubEdrStatusPaths(t, path, 10*time.Millisecond, time.Hour)
	stubPolicyVersion(t, "")

	sender := &edrStatusTestSender{}
	stop := startEdrReporter(t, sender)
	defer stop()

	waitEdrReports(t, sender, 1)
	// Several more poll ticks at the same content: nothing new may arrive.
	time.Sleep(100 * time.Millisecond)
	if got := len(sender.reports()); got != 1 {
		t.Fatalf("sent %d reports for unchanged content, want 1", got)
	}
}

// A change to the file triggers a new report with the new content.
func TestEdrStatusReportSendsOnChange(t *testing.T) {
	dir := t.TempDir()
	path := writeStatusFile(t, dir, "status.json", `{"a":1}`)
	stubEdrStatusPaths(t, path, 10*time.Millisecond, time.Hour)
	stubPolicyVersion(t, "")

	sender := &edrStatusTestSender{}
	stop := startEdrReporter(t, sender)
	defer stop()

	waitEdrReports(t, sender, 1)
	writeStatusFile(t, dir, "status.json", `{"a":2}`)
	reports := waitEdrReports(t, sender, 2)
	if reports[1].GetStatusJson() != `{"a":2}` {
		t.Fatalf("second report StatusJson=%q want the new content", reports[1].GetStatusJson())
	}
}

// No status file means the module is absent: nothing is sent.
func TestEdrStatusReportNoFile(t *testing.T) {
	stubEdrStatusPaths(t, filepath.Join(t.TempDir(), "does-not-exist.json"), 10*time.Millisecond, time.Hour)
	stubPolicyVersion(t, "")

	sender := &edrStatusTestSender{}
	stop := startEdrReporter(t, sender)
	defer stop()

	time.Sleep(100 * time.Millisecond)
	if got := len(sender.reports()); got != 0 {
		t.Fatalf("sent %d reports with no status file, want 0", got)
	}
}

// Even unchanged content is re-sent once the report interval has elapsed.
func TestEdrStatusReportCadence(t *testing.T) {
	path := writeStatusFile(t, t.TempDir(), "status.json", `{"a":1}`)
	stubEdrStatusPaths(t, path, 10*time.Millisecond, 50*time.Millisecond)
	stubPolicyVersion(t, "")

	sender := &edrStatusTestSender{}
	stop := startEdrReporter(t, sender)
	defer stop()

	reports := waitEdrReports(t, sender, 2)
	if reports[1].GetStatusJson() != `{"a":1}` {
		t.Fatalf("cadence re-send StatusJson=%q want the unchanged content", reports[1].GetStatusJson())
	}
}

// A failed send must not kill the goroutine: the next tick retries and
// succeeds.
func TestEdrStatusReportSendFailureSurvives(t *testing.T) {
	path := writeStatusFile(t, t.TempDir(), "status.json", `{"a":1}`)
	stubEdrStatusPaths(t, path, 10*time.Millisecond, time.Hour)
	stubPolicyVersion(t, "")

	sender := &flakyEdrSender{failOnce: true}
	stop := startEdrReporter(t, sender)
	defer stop()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sender.count() >= 1 {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("reporter did not retry after a failed send — the goroutine died")
}
