//go:build windows

package netblock

import (
	"context"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
)

// Microsoft-Windows-DNS-Client is the ETW provider that emits per-process DNS
// resolutions. Event 3008 ("query completed") carries the queried name
// (QueryName) and the resolved addresses (QueryResults), plus the querying
// PID in the event header. DoH/DoT bypasses this provider (documented spec
// limit). Event ID and field names are VM-confirmed (Task 10).
const (
	dnsClientProviderName = "Microsoft-Windows-DNS-Client"
	dnsClientProviderGUID = "{1C95126E-7EEA-49A9-A3FE-A378B03DDB4D}"
	dnsQueryCompletedID   = 3008 // carries QueryName + QueryResults
)

type etwDNSFeed struct{}

// NewDNSFeed returns the Windows DNS-Client ETW feed.
func NewDNSFeed() DNSFeed { return &etwDNSFeed{} }

// Run opens a real-time ETW session on Microsoft-Windows-DNS-Client, forwards
// each completed resolution (event 3008) to sink as a DNSEvent, and blocks
// until ctx is cancelled.
func (f *etwDNSFeed) Run(ctx context.Context, sink func(DNSEvent)) error {
	session := etw.NewRealTimeSession("UTMStackEDR-DNSClient")
	defer func() { _ = session.Stop() }()

	// Build the provider directly rather than via etw.ParseProvider: ParseProvider
	// resolves the GUID against the system's enumerated provider list and errors
	// if it is not found, whereas EnableProvider only needs GUID/level/keywords.
	prov := etw.Provider{
		GUID:            dnsClientProviderGUID,
		Name:            dnsClientProviderName,
		EnableLevel:     0xff,               // all levels; we filter by event ID in the callback.
		MatchAnyKeyword: 0xffffffffffffffff, // all keywords — DNS-Client gates query events on keywords.
	}
	if err := session.EnableProvider(prov); err != nil {
		return err
	}

	c := etw.NewRealTimeConsumer(ctx).FromSessions(session)
	c.EventCallback = func(e *etw.Event) error {
		if e.System.EventID != dnsQueryCompletedID {
			return nil
		}
		name, _ := e.GetPropertyString("QueryName")
		if name == "" {
			return nil
		}
		results, _ := e.GetPropertyString("QueryResults")
		ips := ParseDNSResults(results)
		if len(ips) == 0 {
			return nil
		}
		sink(DNSEvent{
			Name:    name,
			IPs:     ips,
			PID:     int(e.System.Execution.ProcessID),
			Process: "", // DNS-Client doesn't carry the image path; PID is the handle.
		})
		return nil
	}

	if err := c.Start(); err != nil {
		return err
	}
	logInfo("blocklist DNS feed (ETW) started")

	// Watch for ctx cancellation OR an async ProcessTrace failure. golang-etw's
	// Start() only reports the synchronous OpenTrace error; a trace that dies
	// later stores its error in c.Err(). Without this poll a dead feed would
	// block here forever and callers would believe DNS is still observed.
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
				logWarn("blocklist DNS feed (ETW) stopped: %v", err)
				return err
			}
		}
	}
}
