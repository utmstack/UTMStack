package guard

import (
	"os"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
	"github.com/utmstack/UTMStack/shared/logger"
)

type ImageScanner interface {
	ScanFile(path, source string) (string, string, error)
}

type TreeResponder interface {
	KillTree(rootPID int, signature string) ([]int, error)
	Suspend(pid int) error
	Resume(pid int) error
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
	err          error
}

// OnStart implements the run-before-scan race handling (spec §6.2):
// optionally suspend the new process, expedite the scan of its image, and kill
// the tree on a malicious verdict, otherwise resume. The scan uses
// source=process_watcher so the detection event is attributed correctly.
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

	suspended := false
	if g.cfg.SuspendOnLaunch && !isSystemImage(p.Image) {
		if err := g.r.Suspend(p.PID); err == nil {
			suspended = true
		}
	}

	// Run the (potentially slow) image scan off to the side so a suspend can be
	// time-bounded.
	done := make(chan scanResult, 1)
	go func() {
		v, s, e := g.sc.ScanFile(p.Image, event.SourceProcessWatcher)
		done <- scanResult{verdict: v, sig: s, err: e}
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
			_ = g.r.Resume(p.PID)
			suspended = false
			logger.Error("UTMStack EDR: image scan of %s exceeded suspend timeout; resumed (fail-open)", p.Image)
			res = <-done
		}
	} else {
		res = <-done
	}

	if res.err != nil {
		logger.Error("UTMStack EDR: expedited scan of %s failed: %v", p.Image, res.err)
		if suspended {
			_ = g.r.Resume(p.PID) // fail-open: don't hang a launch on scan error
		}
		return
	}

	if res.verdict == cache.VerdictMalicious {
		// Kill the tree even if we already resumed on timeout — a malicious
		// process must be terminated regardless.
		if _, err := g.r.KillTree(p.PID, res.sig); err != nil {
			logger.Error("UTMStack EDR: kill tree %d failed: %v", p.PID, err)
		}
		return
	}
	if suspended {
		_ = g.r.Resume(p.PID)
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
	p := normImage(image)
	return p == systemRoot || strings.HasPrefix(p, systemRoot+"/")
}
