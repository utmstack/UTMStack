package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Engine struct {
	cfg    config.EDRConfig
	cmd    *exec.Cmd
	tuning Tuning
}

func New(cfg config.EDRConfig) *Engine { return &Engine{cfg: cfg} }

// Tuning returns the host-adaptive engine tuning, deriving it (with failsafe)
// on first use so status reporting always has a value.
func (e *Engine) Tuning() Tuning {
	if e.tuning.Tier == "" {
		e.tuning = PlanTuning(e.cfg)
	}
	return e.tuning
}

func (e *Engine) binaryPath() string {
	if e.cfg.ClamdPath != "" {
		return e.cfg.ClamdPath
	}
	return filepath.Join(config.EngineDir, "clamd.exe")
}

// EnsureRunning starts the scan engine child process if it is not already
// answering on ClamdAddr, then blocks until it is healthy or times out.
func (e *Engine) EnsureRunning() error {
	if Ping(e.cfg.ClamdAddr) == nil {
		logger.Info("UTMStack EDR: scan engine already running")
		return nil
	}
	// Derive host-adaptive tuning (with failsafe on profiling failure) and honor
	// the viability gate: on a host too small for a resident daemon, downgrade
	// gracefully rather than risk an OOM (§8.4).
	e.tuning = PlanTuning(e.cfg)
	if !e.tuning.ResidentViable {
		logger.Info("UTMStack EDR: resident engine not viable on this host (%s) — %s; scanning layer downgraded",
			e.tuning.Tier, e.tuning.Reason)
		return fmt.Errorf("resident scan engine not viable on this host")
	}

	bin := e.binaryPath()
	if !fs.Exists(bin) {
		return fmt.Errorf("UTMStack EDR scan engine binary not found at %s", bin)
	}
	// Write the detection-maximizing, host-tuned clamd.conf + freshclam.conf
	// before starting the daemon (ClamAV Configuration Guide §3–§9).
	if err := WriteEngineConfigs(e.cfg, e.tuning); err != nil {
		return fmt.Errorf("writing UTMStack EDR engine config: %w", err)
	}
	logger.Info("UTMStack EDR: engine tier=%s threads=%d maxfile=%dM maxscan=%dM reload=%v",
		e.tuning.Tier, e.tuning.MaxThreads, e.tuning.MaxFileSizeMB, e.tuning.MaxScanSizeMB, e.tuning.ConcurrentReload)

	cmd := exec.Command(bin, "--config-file", filepath.Join(config.EngineDir, "clamd.conf"), "--foreground")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting UTMStack EDR scan engine: %w", err)
	}
	e.cmd = cmd

	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if Ping(e.cfg.ClamdAddr) == nil {
			logger.Info("UTMStack EDR: scan engine healthy")
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("UTMStack EDR scan engine did not become healthy")
}

// Supervise periodically health-checks the scan engine and restarts it if it
// has died, until ctx is cancelled. The engine can be OOM-killed under a scan
// burst (its concurrent-reload spike needs ~2x the signature RAM) or crash;
// without supervision a dead engine leaves the EDR unable to get verdicts —
// file/process scans fail (fail-open) and AMSI stalls. This was observed on the
// test VM: the engine died and never came back until a manual restart.
func (e *Engine) Supervise(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if Ping(e.cfg.ClamdAddr) == nil {
				continue // healthy
			}
			logger.Error("UTMStack EDR: scan engine not responding; restarting it")
			if err := e.EnsureRunning(); err != nil {
				logger.Error("UTMStack EDR: scan engine restart failed: %v", err)
			}
		}
	}
}

func (e *Engine) Stop() error {
	if e.cmd != nil && e.cmd.Process != nil {
		return e.cmd.Process.Kill()
	}
	return nil
}
