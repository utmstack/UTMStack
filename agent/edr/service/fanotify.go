package service

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/fanotify"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
	"github.com/utmstack/UTMStack/shared/logger"
)

func nativeFanotifyConfig(cfg config.EDRConfig, maxFileMB int) fanotify.Config {
	f := cfg.LinuxProtection.Effective().Fanotify
	return fanotify.Config{Enabled: f.Enabled, Mounts: f.Mounts, Roots: f.Roots, Mode: f.Mode, FailClosed: f.FailClosed, Deadline: time.Duration(f.DeadlineMS) * time.Millisecond, Workers: f.Workers, MaxPending: f.MaxPending, MaxFileBytes: int64(maxFileMB) << 20}
}

// The kernel-supplied descriptor is the input. Reopening Event.Path would allow
// an attacker to substitute another file between the permission event and scan.
func scanPinned(ctx context.Context, eng *engine.Engine, f *os.File, maxBytes int64) (fanotify.Decision, error) {
	return scanPinnedWithPath(ctx, eng, f, "", maxBytes)
}

func scanPinnedWithPath(ctx context.Context, eng *engine.Engine, f *os.File, path string, maxBytes int64) (fanotify.Decision, error) {
	st, err := f.Stat()
	if err != nil {
		return fanotify.Decision{}, err
	}
	if !st.Mode().IsRegular() || st.Size() < 0 || st.Size() > maxBytes {
		return fanotify.Decision{}, fanotify.ErrFileLimit
	}
	data := make([]byte, int(st.Size())+1)
	n := 0
	for n < len(data) {
		if err := ctx.Err(); err != nil {
			return fanotify.Decision{}, err
		}
		stop := n + (64 << 10)
		if stop > len(data) {
			stop = len(data)
		}
		count, err := f.Read(data[n:stop])
		n += count
		if err == io.EOF {
			break
		}
		if err != nil {
			return fanotify.Decision{}, err
		}
		if count == 0 {
			return fanotify.Decision{}, io.ErrNoProgress
		}
	}
	if int64(n) != st.Size() {
		return fanotify.Decision{}, errors.New("fanotify input changed while reading")
	}
	if path == "" && eng.MetadataSensitive() {
		return fanotify.Decision{}, engine.ErrUnsupported
	}
	clean, sig, generation, err := eng.ScanSnapshotWithPath(ctx, data[:n], path)
	return fanotify.Decision{Clean: clean, Detected: sig != "", Signature: sig, Generation: generation}, err
}

func (p *program) startFilePrevention(ctx context.Context, cfg config.EDRConfig, eng *engine.Engine, sp *event.Spool, orch *orchestrator.Orchestrator, excluded func(string) bool) {
	if runtime.GOOS != "linux" || !cfg.LinuxProtection.Fanotify.Enabled || !cfg.Sensors.FileWatcherOn() {
		return
	}
	fc := nativeFanotifyConfig(cfg, eng.Tuning().MaxFileSizeMB)
	scan := func(ctx context.Context, f *os.File, e fanotify.Event) (fanotify.Decision, error) {
		if excluded != nil && excluded(e.Path) {
			return fanotify.Decision{Clean: true}, nil
		}
		return scanPinnedWithPath(ctx, eng, f, e.Path, fc.MaxFileBytes)
	}
	observe := func(o fanotify.Outcome) {
		if excluded != nil && excluded(o.Event.Path) {
			return
		}
		// Reanalysis through the existing queue is deliberately separate from the
		// pinned permission decision and verifies the current path before quarantine.
		if !o.Event.Permission && o.Err == nil {
			_ = orch.Enqueue(orchestrator.FileEvent{Path: o.Event.Path, Op: "fanotify", Source: event.SourceFileWatcher})
		}
		if !o.Decision.Detected && o.Err == nil {
			return
		}
		ev := event.NewDetection(o.Event.Path, "", o.Decision.Signature, event.SourceFileWatcher)
		if !o.Decision.Detected {
			ev.Action = event.ActionHealth
			ev.Verdict = "unknown"
			ev.Signature = "fanotify_incomplete:" + o.Reason
			ev.Severity = "warning"
		}
		if !o.Allowed {
			ev.Action = event.ActionBlocked
		}
		if o.Event.PID > 0 {
			ev.Process = &event.ProcInfo{PID: o.Event.PID, PPID: o.Event.PPID, Image: o.Event.Image, Cmdline: o.Event.Cmdline}
		}
		if js, err := ev.ToJSON(); err == nil {
			_ = sp.Append(js)
		}
	}
	var m fanotify.Runner
	var err error
	if fc.Mode == fanotify.ModeNotify {
		m, err = fanotify.New(fc, scan, observe)
	} else {
		m, err = fanotify.NewPermissionBroker(fc, scan, observe)
	}
	if err != nil {
		p.sensorHealth["file_prevention"] = fixedHealth{err}
		p.sensorHealth["file_access_monitor"] = fixedHealth{err}
		return
	}
	p.sensorHealth["file_prevention"] = m
	p.sensorHealth["file_access_monitor"] = m
	p.fileAccess = m
	p.goNative("file-prevention", func() {
		if err := m.Run(ctx); err != nil && ctx.Err() == nil {
			logger.Error("UTMStack EDR: file prevention: %v", err)
		}
	})
}

type fixedHealth struct{ err error }

func (h fixedHealth) Health() error { return h.err }
