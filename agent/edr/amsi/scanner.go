package amsi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

const (
	maxIncompleteReports = 1024
	incompleteWindow     = 10 * time.Minute
)

type Scanner struct {
	cfg          config.EDRConfig
	spool        *event.Spool
	scanSnapshot func(context.Context, []byte) (bool, string, string, error)
	reportMu     sync.Mutex
	reported     map[[32]byte]time.Time
}

// NewScanner requires the engine owned and supervised by the service or CLI.
// The returned generation belongs to the actual scan, including during reload.
func NewScanner(cfg config.EDRConfig, sp *event.Spool, eng *engine.Engine) *Scanner {
	if eng == nil {
		panic("amsi.NewScanner requires a shared native engine")
	}
	return &Scanner{cfg: cfg, spool: sp, reported: make(map[[32]byte]time.Time), scanSnapshot: eng.ScanSnapshot}
}

// ScanBuffer scans dynamic content submitted via AMSI and returns whether the
// host should block execution. It emits a branded amsi event on a block and
// applies the configured fail policy when analysis is incomplete.
func (s *Scanner) ScanBuffer(appName string, data []byte) bool {
	return s.ScanBufferContext(context.Background(), appName, data)
}

func (s *Scanner) ScanBufferContext(ctx context.Context, appName string, data []byte) bool {
	block, _ := s.scanBufferDecision(ctx, appName, data)
	return block
}

func (s *Scanner) scanBufferDecision(ctx context.Context, appName string, data []byte) (block, incomplete bool) {
	clean, sig, generation, err := s.scanSnapshot(ctx, data)
	if err != nil {
		failClosed := s.cfg.FailMode == "closed"
		s.noteIncomplete(appName, data, err, failClosed, generation)
		return failClosed, true
	}
	if !clean {
		s.emit(event.NewAMSIEvent(event.ActionBlocked, "malicious", sig, appName))
		return true, false
	}
	// Clean content is allowed silently (emitting per-buffer would be far too noisy).
	return false, false
}

// noteIncomplete reports an incomplete analysis at most once per window for the
// same content, generation and reason, so a script host cannot flood the log or
// the spool (whose rotation would drop unsent detections). A down engine is one
// engine-wide condition and is keyed without the content. The fail policy is
// applied by the caller for every buffer regardless.
func (s *Scanner) noteIncomplete(appName string, data []byte, err error, failClosed bool, version string) {
	reason := incompleteReason(err)
	hash := ""
	if reason != "Unavailable" && data != nil {
		sum := sha256.Sum256(data)
		hash = hex.EncodeToString(sum[:])
	}
	key := sha256.Sum256([]byte(hash + "\x00" + version + "\x00" + reason))
	now := time.Now()
	s.reportMu.Lock()
	if last, ok := s.reported[key]; ok && now.Sub(last) < incompleteWindow {
		s.reportMu.Unlock()
		return
	}
	if len(s.reported) >= maxIncompleteReports {
		var oldestKey [32]byte
		var oldest time.Time
		for k, t := range s.reported {
			if oldest.IsZero() || t.Before(oldest) {
				oldestKey, oldest = k, t
			}
		}
		delete(s.reported, oldestKey)
	}
	s.reported[key] = now
	s.reportMu.Unlock()

	logger.Error("UTMStack EDR AMSI: analysis incomplete (%s), fail-%s", reason, modeName(failClosed))
	ev := event.NewAMSIEvent(event.ActionHealth, "unknown", "Analysis.Incomplete."+reason, appName)
	ev.FileHash, ev.Severity = hash, "medium"
	if !s.emit(ev) && s.spool != nil {
		// A failed spool write is not a delivered report; allow a future retry.
		s.reportMu.Lock()
		if s.reported[key] == now {
			delete(s.reported, key)
		}
		s.reportMu.Unlock()
	}
}

// incompleteReason classifies an analysis error for operators, in the same
// vocabulary as the file scanner.
func incompleteReason(err error) string { return engine.IncompleteReason(err) }

func (s *Scanner) emit(ev event.Event) bool {
	if s.spool == nil {
		return false
	}
	js, err := ev.ToJSON()
	return err == nil && s.spool.Append(js) == nil
}

func modeName(closed bool) string {
	if closed {
		return "closed"
	}
	return "open"
}
