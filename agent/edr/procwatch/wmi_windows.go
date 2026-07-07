//go:build windows

package procwatch

import (
	"context"
	"runtime"
	"strconv"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/utmstack/UTMStack/shared/logger"
)

type Watcher struct{ h Handler }

func New(h Handler) *Watcher { return &Watcher{h: h} }

func (w *Watcher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := w.subscribe(ctx); err != nil {
			logger.Error("UTMStack EDR process watcher: %v; retrying", err)
			time.Sleep(3 * time.Second)
		}
	}
}

func (w *Watcher) subscribe(ctx context.Context) error {
	// COM is thread-affine: every call must run on the OS thread that called
	// CoInitializeEx. Go migrates goroutines across threads by default, so we
	// MUST pin this goroutine for the lifetime of the COM session — otherwise
	// go-ole calls land on a different thread and cause an access violation
	// (a structured exception that bypasses recover and crashes the service).
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// S_FALSE (already initialized on this thread) is not fatal.
	_ = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return err
	}
	defer unknown.Release()
	locator, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer locator.Release()

	serviceRaw, err := oleutil.CallMethod(locator, "ConnectServer", nil, `root\cimv2`)
	if err != nil {
		return err
	}
	service := serviceRaw.ToIDispatch()
	defer service.Release()

	query := "SELECT * FROM __InstanceCreationEvent WITHIN 1 WHERE TargetInstance ISA 'Win32_Process'"
	eventSourceRaw, err := oleutil.CallMethod(service, "ExecNotificationQuery", query)
	if err != nil {
		return err
	}
	eventSource := eventSourceRaw.ToIDispatch()
	defer eventSource.Release()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		// NextEvent blocks up to the timeout (ms); the short timeout lets ctx cancel.
		evtRaw, err := oleutil.CallMethod(eventSource, "NextEvent", 1000)
		if err != nil {
			continue // timeout / transient
		}
		evt := evtRaw.ToIDispatch()
		targetRaw, err := oleutil.GetProperty(evt, "TargetInstance")
		if err != nil {
			evt.Release()
			continue
		}
		target := targetRaw.ToIDispatch()

		ps := ProcStart{
			PID:     dispInt(target, "ProcessId"),
			PPID:    dispInt(target, "ParentProcessId"),
			Image:   dispStr(target, "ExecutablePath"),
			Cmdline: dispStr(target, "CommandLine"),
		}
		if ps.PID != 0 && ps.Image != "" {
			w.h.OnProcStart(ps)
		}
		// Release the dispatch objects only. ToIDispatch() shares the VARIANT's
		// single COM reference (no AddRef), so calling BOTH Release() and
		// VariantClear() on the same object double-frees it → 0xc0000005.
		target.Release()
		evt.Release()
	}
}

func dispStr(d *ole.IDispatch, prop string) string {
	v, err := oleutil.GetProperty(d, prop)
	if err != nil {
		return ""
	}
	defer v.Clear()
	return v.ToString()
}

func dispInt(d *ole.IDispatch, prop string) int {
	v, err := oleutil.GetProperty(d, prop)
	if err != nil {
		return 0
	}
	defer v.Clear()
	switch val := v.Value().(type) {
	case int32:
		return int(val)
	case int64:
		return int(val)
	case string:
		n, _ := strconv.Atoi(val)
		return n
	default:
		return 0
	}
}
