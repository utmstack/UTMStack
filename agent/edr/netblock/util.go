// edr/netblock/util.go
package netblock

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/netip"

	"github.com/utmstack/UTMStack/shared/logger"
)

func bytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func insecureTLS() *tls.Config { return &tls.Config{InsecureSkipVerify: true} }

// logWarn/logInfo funnel through the shared logger; surfaced strings never name
// the intel vendor (branding rule) — they say "blocklist".
func logWarn(format string, a ...any) { logger.Error("[UTMStack EDR] "+format, a...) }
func logInfo(format string, a ...any) { logger.Info("[UTMStack EDR] "+format, a...) }

// disabledBlocker is an always-available no-op Blocker used when WFP init fails,
// so enforcement is silently skipped rather than the module crashing.
type disabledBlocker struct{}

func newDisabledBlocker() (Blocker, error)                         { return disabledBlocker{}, nil }
func (disabledBlocker) AddIP(a netip.Addr, dir string) error       { return nil }
func (disabledBlocker) RemoveIP(a netip.Addr) error                { return nil }
func (disabledBlocker) AddPrefix(p netip.Prefix, dir string) error { return nil }
func (disabledBlocker) RemovePrefix(p netip.Prefix) error          { return nil }
func (disabledBlocker) Reset() error                               { return nil }
func (disabledBlocker) Count() int                                 { return 0 }
