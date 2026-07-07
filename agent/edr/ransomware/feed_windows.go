//go:build windows

package ransomware

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

// Microsoft-Windows-Kernel-File is the ETW provider that emits per-process file
// operations. Each op carries the PID of the acting process in the event
// header — the attribution the USN journal cannot provide, which is why the
// ransomware guard uses ETW for this stream.
const (
	kernelFileProviderName = "Microsoft-Windows-Kernel-File"
	kernelFileProviderGUID = "{EDD08927-9CC4-4E65-B970-C2560FB5C289}"
)

// opForEventID maps Kernel-File event IDs to a FileOp, keeping only mutating
// ops (reads/closes are dropped to keep the stream small). IDs vary slightly
// by OS build; this mapping is a defensible best-effort and is confirmed on the
// VM (Task 12).
//   12 = Create, 30 = SetInformation, 26 = DeletePath, 27 = RenamePath,
//   32 = Write.
func opForEventID(id uint16) (FileOp, bool) {
	switch id {
	case 32:
		return OpWrite, true
	case 30:
		return OpSetInfo, true
	case 26:
		return OpDelete, true
	case 27:
		return OpRename, true
	case 12:
		return OpCreate, true
	default:
		return 0, false
	}
}

type etwFeed struct{}

// NewFeed returns the Windows ETW file-activity feed.
func NewFeed() FileActivityFeed { return &etwFeed{} }

// Run opens a real-time ETW session on Microsoft-Windows-Kernel-File, forwards
// each mutating file op to sink as a FileEvent, and blocks until ctx is
// cancelled.
func (f *etwFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	session := etw.NewRealTimeSession("UTMStackEDR-RansomFile")
	defer func() { _ = session.Stop() }()

	// Build the provider directly rather than via etw.ParseProvider: ParseProvider
	// resolves the GUID against the system's enumerated provider list and errors
	// if it is not found, whereas EnableProvider only needs GUID/level/keywords —
	// so a direct struct is both simpler and can never fail on a name lookup.
	prov := etw.Provider{
		GUID:        kernelFileProviderGUID,
		Name:        kernelFileProviderName,
		EnableLevel: 0xff, // all levels; we filter by event ID in the callback.
	}
	if err := session.EnableProvider(prov); err != nil {
		return err
	}

	c := etw.NewRealTimeConsumer(ctx).FromSessions(session)
	c.EventCallback = func(e *etw.Event) error {
		// Kernel-File carries the target in "FileName" (or "OpenPath"); try both.
		path, ok := e.GetPropertyString("FileName")
		if !ok || path == "" {
			path, _ = e.GetPropertyString("OpenPath")
		}
		op, ok := opForEventID(e.System.EventID)
		if !ok {
			return nil
		}
		if path == "" {
			return nil
		}
		sink(FileEvent{
			PID:  int(e.System.Execution.ProcessID),
			Path: normalizeKernelPath(path),
			Op:   op,
		})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
	logger.Info("UTMStack EDR: ransomware file-activity feed (ETW) started")

	// Watch for ctx cancellation OR an async ProcessTrace failure. golang-etw's
	// Start() only reports the synchronous OpenTrace error; a trace that dies
	// later stores its error in c.Err(). Without this poll a dead feed would
	// block here forever and the guard would believe it is still protected.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = c.Stop()
			return ctx.Err()
		case <-ticker.C:
			if err := c.Err(); err != nil {
				_ = c.Stop()
				logger.Error("UTMStack EDR: ransomware file feed (ETW) stopped: %v", err)
				return err
			}
		}
	}
}

var (
	deviceMapOnce sync.Once
	deviceMap     map[string]string // lower-cased NT device (e.g. "\device\harddiskvolume4") -> "C:"
)

// buildDeviceMap maps each drive letter's NT device name to the letter via
// QueryDosDevice, so ETW kernel paths (which use "\Device\HarddiskVolumeN\...")
// can be rewritten to the "C:\..." form the canary registry and the rest of the
// EDR use. Without this, canary tamper events never match a planted canary.
func buildDeviceMap() {
	deviceMap = map[string]string{}
	var buf [1024]uint16
	for c := byte('A'); c <= 'Z'; c++ {
		letter := string(rune(c)) + ":"
		lp, err := windows.UTF16PtrFromString(letter)
		if err != nil {
			continue
		}
		n, err := windows.QueryDosDevice(lp, &buf[0], uint32(len(buf)))
		if err != nil || n == 0 {
			continue
		}
		if dev := windows.UTF16ToString(buf[:n]); dev != "" {
			deviceMap[strings.ToLower(dev)] = letter
		}
	}
}

// normalizeKernelPath rewrites an ETW kernel file path to a drive-letter path:
// it drops the "\??\" prefix and maps a "\Device\HarddiskVolumeN\..." prefix to
// its drive letter (e.g. "C:\..."). The device map is built once, lazily.
func normalizeKernelPath(p string) string {
	p = strings.TrimPrefix(p, `\??\`)
	if strings.HasPrefix(p, `\Device\`) {
		deviceMapOnce.Do(buildDeviceMap)
		low := strings.ToLower(p)
		for dev, letter := range deviceMap {
			if strings.HasPrefix(low, dev+`\`) {
				return letter + p[len(dev):]
			}
		}
	}
	return p
}
