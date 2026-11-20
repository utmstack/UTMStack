//go:build windows

package ransomware

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
	"github.com/utmstack/UTMStack/shared/logger"
)

// Microsoft-Windows-Kernel-Registry is the ETW provider that emits per-process
// registry mutations. Each event carries the PID in the header and the key
// name in the payload.
const (
	registryProviderName = "Microsoft-Windows-Kernel-Registry"
	registryProviderGUID = "{70eb4f03-c1de-4f73-a051-33d13d5413bd}"
	// registryMutateKeywords is the OR of the keyword masks for mutating ops
	// (create/delete key, set/delete value, set information). Enabling only
	// these at the provider level keeps the event stream small.
	registryMutateKeywords = uint64(0x5340)
)

// extractKeyName tries to extract the key name from a Kernel-Registry event.
// Different event templates use different property names:
//   - Most mutating events (DeleteKey, SetValueKey, etc.) use "KeyName".
//   - CreateKey uses "BaseName" and "RelativeName" separately.
func extractKeyName(e *etw.Event) string {
	if k, ok := e.GetPropertyString("KeyName"); ok && k != "" {
		return k
	}
	base, _ := e.GetPropertyString("BaseName")
	rel, _ := e.GetPropertyString("RelativeName")
	if base != "" || rel != "" {
		// Reconstruct the full path from the two parts.
		if base != "" && rel != "" {
			return base + `\\` + rel
		}
		return base + rel
	}
	return ""
}

type registryFeed struct {
	mu        sync.Mutex
	running   bool
	failure   error
	rawEvents uint64
}

// NewRegistryFeed returns the Windows ETW registry-mutation feed.
func NewRegistryFeed() RegistryFeed { return &registryFeed{} }

// Health reports whether the feed is running and has delivered at least one
// native record, mirroring the file feed's health pattern so the service can
// fold it into the ransomware sensor-health aggregate.
func (f *registryFeed) Health() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failure != nil {
		return f.failure
	}
	if !f.running {
		return errors.New("ransomware registry ETW feed not initialized")
	}
	if f.rawEvents == 0 {
		return errors.New("ransomware registry ETW provider has not delivered a native record")
	}
	return nil
}

// Run opens a real-time ETW session on Microsoft-Windows-Kernel-Registry,
// forwards each mutating event to sink as a RegistryEvent (the guard filters
// sensitive keys), and blocks until ctx is cancelled.
func (f *registryFeed) Run(ctx context.Context, sink func(RegistryEvent)) error {
	session := etw.NewRealTimeSession("UTMStackEDR-RansomReg")
	defer func() { _ = session.Stop() }()
	prov := etw.Provider{
		GUID:            registryProviderGUID,
		Name:            registryProviderName,
		EnableLevel:     0xff,
		MatchAnyKeyword: registryMutateKeywords,
	}
	if err := session.EnableProvider(prov); err != nil {
		return err
	}

	c := etw.NewRealTimeConsumer(ctx).FromSessions(session)
	c.EventCallback = func(e *etw.Event) error {
		key := extractKeyName(e)
		if key == "" {
			return nil
		}
		f.mu.Lock()
		f.rawEvents++
		f.mu.Unlock()
		sink(RegistryEvent{
			PID: int(e.System.Execution.ProcessID),
			Key: key,
		})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
	f.mu.Lock()
	f.running = true
	f.mu.Unlock()
	defer func() { f.mu.Lock(); f.running = false; f.mu.Unlock() }()
	logger.Info("UTMStack EDR: ransomware registry feed (ETW) started")

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
				f.mu.Lock()
				if f.failure == nil {
					f.failure = err
				}
				f.mu.Unlock()
				logger.Error("UTMStack EDR: ransomware registry feed (ETW) stopped: %v", err)
				return err
			}
		}
	}
}
