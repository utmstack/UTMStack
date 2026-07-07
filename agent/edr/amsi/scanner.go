package amsi

import (
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Scanner struct {
	cfg       config.EDRConfig
	spool     *event.Spool
	scanBytes func(addr string, data []byte) (bool, string, error)
}

func NewScanner(cfg config.EDRConfig, sp *event.Spool) *Scanner {
	return &Scanner{cfg: cfg, spool: sp, scanBytes: engine.ScanBytes}
}

// ScanBuffer scans dynamic content submitted via AMSI and returns whether the
// host should block execution. It emits a branded amsi event on a block and
// applies the configured fail policy when the engine is unreachable.
func (s *Scanner) ScanBuffer(appName string, data []byte) bool {
	clean, sig, err := s.scanBytes(s.cfg.ClamdAddr, data)
	if err != nil {
		failClosed := s.cfg.FailMode == "closed"
		logger.Error("UTMStack EDR AMSI: scan engine unreachable, fail-%s", modeName(failClosed))
		s.emit(event.NewAMSIEvent(event.ActionHealth, "unknown", "", appName))
		return failClosed
	}
	if !clean {
		s.emit(event.NewAMSIEvent(event.ActionBlocked, "malicious", sig, appName))
		return true
	}
	// Clean content is allowed silently (emitting per-buffer would be far too noisy).
	return false
}

func (s *Scanner) emit(ev event.Event) {
	if s.spool == nil {
		return
	}
	if js, err := ev.ToJSON(); err == nil {
		_ = s.spool.Append(js)
	}
}

func modeName(closed bool) string {
	if closed {
		return "closed"
	}
	return "open"
}
