package agent

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

var (
	edrSpoolFile  = filepath.Join(fs.GetExecutablePath(), "edr-spool", "events.ndjson")
	edrOffsetFile = filepath.Join(fs.GetExecutablePath(), "edr-spool", ".offset")
	edrDataType   = "utmstack_edr"
)

// drainSpool preserves the package's local test entry point. Production supplies
// the relay context so shutdown stops between records and lock retries.
func drainSpool(path, offset string, emit func(string) bool) (int64, error) {
	return event.DrainSpool(context.Background(), path, offset, emit)
}

// EDRRelay drains the EDR spool into the agent LogQueue every second until ctx is done.
func EDRRelay(ctx context.Context) {
	host, _ := os.Hostname()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, err := event.DrainSpool(ctx, edrSpoolFile, edrOffsetFile, func(raw string) bool {
				select {
				case LogQueue <- &plugins.Log{DataType: edrDataType, Raw: raw, DataSource: host}:
					return true
				default:
					utils.Logger.LogF(100, "EDR relay: LogQueue full, retrying next tick")
					return false
				}
			})
			if err != nil {
				utils.Logger.LogF(500, "EDR relay drain error: %v", err)
			}
		}
	}
}
