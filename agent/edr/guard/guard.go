package guard

import (
	"errors"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

// ImageScanner analyses a launched image. On a malicious verdict it defers the
// response to contain, which the guard calls after killing the tree (a running
// image usually cannot be moved); contain quarantines the image, retrying while
// the process exits, and writes the one detection event with the action that
// happened. contain is nil when no response is due. Concurrent scans of one
// image share one analysis and one contain.
type ImageScanner interface {
	ScanImage(path, source string) (verdict, sig string, contain func() string, err error)
}

type TreeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
}

// IdentityResponder binds an asynchronous image verdict to the launch actually
// observed. A numeric PID can name another process by the time analysis ends.
type IdentityResponder interface {
	KillTreeIdentity(int, int64, string) ([]int, error)
	SuspendIdentity(int, int64) error
	ResumeIdentity(int, int64) error
}

type Guard struct {
	cfg config.EDRConfig
	sc  ImageScanner
	r   TreeResponder
}

func New(cfg config.EDRConfig, sc ImageScanner, r TreeResponder) *Guard {
	return &Guard{cfg: cfg, sc: sc, r: r}
}

type scanResult struct {
	verdict, sig string
	contain      func() string
	err          error
}

// OnStart implements the run-before-scan race handling (spec §6.2):
// optionally suspend the new process, expedite the scan of its image, and kill
// the tree and then quarantine the image on a malicious verdict, otherwise
// resume. The scan uses source=process_watcher so the detection event is
// attributed correctly.
//
// Suspend-on-launch is powerful but dangerous — suspending every launch can
// freeze the whole system. Two safety rails apply whenever it is enabled:
//   - OS/system processes (under %SystemRoot%) are NEVER suspended, so the
//     platform keeps running; they are still scanned and killed if malicious.
//   - No process is ever held suspended longer than SuspendTimeoutMs. If the
//     engine is slow or backed up, the process is resumed (fail-open) and the
//     verdict is still acted on when it arrives — a scan backlog can never
//     freeze the host.
func (g *Guard) OnStart(p proctable.Proc) {
	if p.Image == "" {
		return
	}
	identified, hasIdentity := g.r.(IdentityResponder)
	suspend := func() error {
		if p.OSIdentity > 0 || p.IdentityRequired {
			if !hasIdentity || p.OSIdentity <= 0 {
				return errors.New("identity-bound suspension unavailable")
			}
			return identified.SuspendIdentity(p.PID, p.OSIdentity)
		}
		return g.r.Suspend(p.PID)
	}
	resume := func() error {
		if p.OSIdentity > 0 || p.IdentityRequired {
			if !hasIdentity || p.OSIdentity <= 0 {
				return errors.New("identity-bound resume unavailable")
			}
			return identified.ResumeIdentity(p.PID, p.OSIdentity)
		}
		return g.r.Resume(p.PID)
	}
	kill := func(sig string) ([]int, error) {
		if p.OSIdentity > 0 || p.IdentityRequired {
			if !hasIdentity || p.OSIdentity <= 0 {
				return nil, errors.New("identity-bound response unavailable")
			}
			return identified.KillTreeIdentity(p.PID, p.OSIdentity, sig)
		}
		return g.r.KillTree(p.PID, sig)
	}

	suspended := false
	if g.cfg.SuspendOnLaunch && !isSystemImage(p.Image) {
		if err := suspend(); err == nil {
			suspended = true
		}
	}

	// Run the (potentially slow) image scan off to the side so a suspend can be
	// time-bounded.
	done := make(chan scanResult, 1)
	go func() {
		v, s, c, e := g.sc.ScanImage(p.Image, event.SourceProcessWatcher)
		done <- scanResult{verdict: v, sig: s, contain: c, err: e}
	}()

	var res scanResult
	if suspended {
		timeout := time.Duration(g.cfg.SuspendTimeoutMs) * time.Millisecond
		if timeout <= 0 {
			timeout = 5 * time.Second
		}
		select {
		case res = <-done:
		case <-time.After(timeout):
			// Fail-open: never hold a process past the bound; let the scan finish
			// in the background and still act on its verdict below.
			_ = resume()
			suspended = false
			logger.Error("UTMStack EDR: image scan of %s exceeded suspend timeout; resumed (fail-open)", p.Image)
			res = <-done
		}
	} else {
		res = <-done
	}

	if res.err != nil && res.verdict != cache.VerdictMalicious {
		logger.Error("UTMStack EDR: expedited scan of %s failed: %v", p.Image, res.err)
		if suspended {
			_ = resume() // fail-open: don't hang a launch on scan error
		}
		return
	}

	if res.verdict == cache.VerdictMalicious {
		if res.err != nil {
			// The image is confirmed malicious; a secondary failure such as
			// an unwritten event must not leave the process running.
			logger.Error("UTMStack EDR: expedited scan of %s detected %s with an error: %v", p.Image, res.sig, res.err)
		}
		// Kill the tree even if we already resumed on timeout — a malicious
		// process must be terminated regardless.
		if _, err := kill(res.sig); err != nil {
			if suspended {
				_ = resume()
				suspended = false
			}
			logger.Error("UTMStack EDR: kill tree %d failed: %v", p.PID, err)
		}
		// Only now contain the image: before the kill the move fails for an image
		// on another volume or one the process holds open.
		if res.contain != nil {
			res.contain()
		}
		return
	}
	if suspended {
		_ = resume()
	}
}

// systemRoot is the lower-cased, forward-slashed OS directory (e.g.
// "c:/windows"); processes under it are never suspended.
var systemRoot = func() string {
	r := os.Getenv("SystemRoot")
	if r == "" {
		r = os.Getenv("windir")
	}
	if r == "" {
		r = `C:\Windows`
	}
	return normImage(r)
}()

func normImage(s string) string {
	return strings.TrimRight(strings.ReplaceAll(strings.ToLower(s), `\`, "/"), "/")
}

// isSystemImage reports whether an image lives under the OS directory. Suspending
// core OS processes (services, session/window managers, RPC hosts, …) can
// deadlock the machine, so they are never frozen — only scanned and, if
// malicious, killed.
func isSystemImage(image string) bool {
	if runtime.GOOS == "linux" && strings.HasPrefix(image, "/") {
		return linuxSystemImage(image)
	}
	p := normImage(image)
	return p == systemRoot || strings.HasPrefix(p, systemRoot+"/")
}

// Preserve the no-freeze rail for shared Linux OS/service binaries. This is a
// suspension exemption only; confirmed malware can still be terminated.
func linuxSystemImage(image string) bool {
	p := path.Clean(image)
	for _, root := range []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/lib", "/lib64", "/usr/lib", "/usr/lib64"} {
		if p == root || strings.HasPrefix(p, root+"/") {
			return true
		}
	}
	return false
}
