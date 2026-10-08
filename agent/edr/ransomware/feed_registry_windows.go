//go:build windows

package ransomware

import (
	"context"
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

type registryFeed struct{}

// NewRegistryFeed returns the Windows ETW registry-mutation feed.
func NewRegistryFeed(_ ...string) RegistryFeed { return &registryFeed{} }

// Run opens a real-time ETW session on Microsoft-Windows-Kernel-Registry,
// forwards each sensitive-key mutation to sink as a RegistryEvent, and blocks
// until ctx is cancelled.
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
		if key == "" || !isSensitiveKey(key) {
			return nil
		}
		sink(RegistryEvent{
			PID: int(e.System.Execution.ProcessID),
			Key: key,
		})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
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
				logger.Error("UTMStack EDR: ransomware registry feed (ETW) stopped: %v", err)
				return err
			}
		}
	}
}
