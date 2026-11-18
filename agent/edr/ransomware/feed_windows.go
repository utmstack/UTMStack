//go:build windows

package ransomware

import (
	"context"
	"errors"
	"fmt"
	"github.com/utmstack/UTMStack/agent/edr/responder"
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

// The event IDs this feed decodes and what each means are documented with the
// portable decoder (kernelfile.go), which is where the decisions live.

type etwFeed struct {
	mu            sync.Mutex
	running       bool
	failure       error
	rawRecords    uint64
	consumer      *nativeFileConsumer
	nativeRecord  func(uint16) // bounded acceptance diagnostics; set before Run only
	nativeDecoded func(kernelFileRecord)
	isDecoy       func(path string) bool // set by the guard before Run
}

// NewFeed returns the Windows ETW file-activity feed.
func NewFeed() FileActivityFeed { return &etwFeed{} }

func (f *etwFeed) Health() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failure != nil {
		return f.failure
	}
	if f.consumer != nil {
		if err := f.consumer.Err(); err != nil {
			return err
		}
	}
	if !f.running {
		return errors.New("ransomware ETW feed not initialized")
	}
	if f.rawRecords == 0 {
		return errors.New("ransomware ETW provider has not delivered a native record")
	}
	return nil
}

func (f *etwFeed) setInterest(isDecoy func(string) bool) { f.isDecoy = isDecoy }

// Run opens a real-time ETW session on Microsoft-Windows-Kernel-File, forwards
// each modification of a decoy file to sink as a FileEvent, and blocks until
// ctx is cancelled.
func (f *etwFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	f.mu.Lock()
	if f.consumer != nil {
		select {
		case <-f.consumer.done:
		default:
			f.mu.Unlock()
			return errors.New("ransomware ETW previous native consumer has not joined")
		}
	}
	f.rawRecords = 0
	f.mu.Unlock()
	session, err := startFileTrace(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := session.Stop(); err != nil {
			f.mu.Lock()
			if f.failure == nil {
				f.failure = err
			}
			f.mu.Unlock()
		}
	}()

	dec := newKernelFileDecoder(f.isDecoy)
	provider := etw.MustParseGUIDFromString(kernelFileProviderGUID)
	lostProvider := etw.MustParseGUIDFromString("{6A399AE0-4BC6-4DE9-870B-3657F8947E7E}")
	// Drop reads, queries and the other unused events before their properties
	// are parsed; they are the bulk of the provider's volume.
	consume := func(consumerCtx context.Context, er *etw.EventRecord) error {
		if consumerCtx.Err() != nil {
			return consumerCtx.Err()
		}
		if er.EventHeader.ProviderId.Equals(lostProvider) {
			return errors.New("ransomware ETW native lost-event notification")
		}
		if !er.EventHeader.ProviderId.Equals(provider) {
			return nil
		}
		f.mu.Lock()
		f.rawRecords++
		f.mu.Unlock()
		if f.nativeRecord != nil {
			f.nativeRecord(er.EventHeader.EventDescriptor.Id)
		}
		if !kernelFileEventWanted(er.EventHeader.EventDescriptor.Id) {
			return nil
		}
		record, err := nativeKernelFileRecord(er)
		if err != nil {
			return fmt.Errorf("ransomware ETW event %d could not be decoded: %w", er.EventHeader.EventDescriptor.Id, err)
		}
		if f.nativeDecoded != nil {
			f.nativeDecoded(record)
		}
		fe, decoded := dec.decode(record)
		coverageErr := dec.Err()
		if decoded && consumerCtx.Err() == nil {
			fe.IdentityRequired = true
			before, err := (responder.OSKiller{}).ProcessIdentity(fe.PID)
			timestamp := er.EventHeader.UTCTimeStamp()
			if err == nil && creationPredatesEvent(before, timestamp) && !timestamp.After(time.Now().Add(time.Second)) {
				image, ppid, readErr := responder.LiveProcess(fe.PID)
				after, checkErr := (responder.OSKiller{}).ProcessIdentity(fe.PID)
				if readErr == nil && checkErr == nil && before == after {
					fe.OSIdentity = before
					fe.Image = image
					fe.PPID = ppid
				}
			}
			if fe.OSIdentity <= 0 {
				f.mu.Lock()
				if f.failure == nil {
					f.failure = errors.New("ransomware ETW event original process identity unavailable")
				}
				f.mu.Unlock()
			}
			if consumerCtx.Err() == nil {
				sink(fe)
			}
		}
		return coverageErr
	}

	c, err := openFileConsumer(ctx, session.TraceName(), consume)
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Stop(); err != nil {
			f.mu.Lock()
			if f.failure == nil {
				f.failure = err
			}
			f.mu.Unlock()
		}
	}()
	f.mu.Lock()
	f.running = true
	f.consumer = c
	f.mu.Unlock()
	defer func() { f.mu.Lock(); f.running = false; f.mu.Unlock() }()
	logger.Info("UTMStack EDR: ransomware file-activity feed (ETW) started")

	// Own ProcessTrace completion and synchronized parsing/native error health.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return c.Err()
		case <-ticker.C:
			if err := session.Loss(); err != nil {
				f.mu.Lock()
				if f.failure == nil {
					f.failure = err
				}
				f.mu.Unlock()
			}
			if err := c.Err(); err != nil {
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
