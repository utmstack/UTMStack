package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

// Quarantiner contains a malicious file. Injected so the scanner stays testable.
type Quarantiner interface {
	Quarantine(path, sha256, detection string) (string, error)
}

type Scanner struct {
	cfg         config.EDRConfig
	cache       *cache.Cache
	spool       *event.Spool
	quarantiner Quarantiner
	scanBytes   func(addr string, data []byte) (bool, string, error)
	// sigDBVersion returns the signature-DB version a verdict was produced
	// under, stamped onto the cache record so the feed only re-scans an entry
	// when the signatures actually change (not on every update cycle).
	sigDBVersion func() string
}

func New(cfg config.EDRConfig, c *cache.Cache, sp *event.Spool, q Quarantiner) *Scanner {
	return &Scanner{
		cfg: cfg, cache: c, spool: sp, quarantiner: q,
		scanBytes:    engine.ScanBytes,
		sigDBVersion: func() string { return engine.SigDBVersion(cfg.ClamdAddr) },
	}
}

func SHA256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ScanFile hashes the file, checks the cache, scans via the engine on a miss,
// stores the verdict, and — on a malicious verdict (cached OR freshly scanned) —
// quarantines the file and emits a branded detection event.
func (s *Scanner) ScanFile(path, source string) (string, string, error) {
	hash, err := SHA256File(path)
	if err != nil {
		// The image is unreadable — commonly because the file-watcher path just
		// quarantined (moved) it out from under a still-running process. If it was
		// quarantined as malicious moments ago, report that verdict so the caller
		// (the process guard) can still terminate the process. Don't re-quarantine
		// or re-emit here — that already happened on the original detection.
		if os.IsNotExist(err) {
			if rec, ok, qerr := s.cache.RecentQuarantineByPath(path, 5*time.Minute); qerr == nil && ok {
				return cache.VerdictMalicious, rec.Detection, nil
			}
		}
		return "", "", err
	}

	var verdict, sig string
	if rec, found, err := s.cache.Lookup(hash); err != nil {
		return "", "", err
	} else if found && !rec.Stale {
		verdict, sig = rec.Verdict, rec.Signature
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", "", err
		}
		clean, s2, err := s.scanBytes(s.cfg.ClamdAddr, data)
		if err != nil {
			return "", "", err
		}
		sig = s2
		verdict = cache.VerdictClean
		if !clean {
			verdict = cache.VerdictMalicious
		}
		if err := s.cache.Store(cache.VerdictRecord{SHA256: hash, Verdict: verdict, Signature: sig, SigDBVersion: s.sigDBVersion()}); err != nil {
			return "", "", err
		}
	}

	// Act on a malicious verdict for THIS file even when the verdict came from the
	// cache: a re-dropped known-malicious file (same hash as one seen before) must
	// still be quarantined and reported, not silently left in place. The verdict
	// cache exists to skip the engine round-trip, not the response. An already-
	// quarantined file is gone, so its re-scan fails hashing above and never reaches
	// here — no double action.
	if verdict == cache.VerdictMalicious {
		ev := event.NewDetection(path, hash, sig, source)
		if s.quarantiner != nil {
			if _, qerr := s.quarantiner.Quarantine(path, hash, sig); qerr == nil {
				ev.Action = event.ActionQuarantined
			}
			// on quarantine failure, keep Action=detected (still reported)
		}
		js, err := ev.ToJSON()
		if err != nil {
			return verdict, sig, err
		}
		if err := s.spool.Append(js); err != nil {
			return verdict, sig, err
		}
	}
	return verdict, sig, nil
}
