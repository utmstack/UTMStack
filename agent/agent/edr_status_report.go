package agent

import (
	"context"
	"os"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

// The intervals are package vars (not consts) so tests can shrink them to tens
// of milliseconds — the defaults are the production cadence.
var (
	// edrStatusReportInterval caps how often a report is sent when the status
	// file keeps changing: at most one per interval.
	edrStatusReportInterval = 5 * time.Minute
	// edrStatusPollInterval is how often the status file is checked for
	// changes.
	edrStatusPollInterval = 15 * time.Second
)

// edrStatusFile is the status file to ship. It defaults to the module's
// status.json; tests redirect it to a t.TempDir() file (the real path is a
// package var in agent/edr/config, not easily injectable as a const).
var edrStatusFile = edrconfig.StatusFile

// sendEdrStatusReports ships the EDR module's status.json upstream. It sends
// on the first tick where the file exists, whenever its content changes, and
// at least once per edrStatusReportInterval. A missing file means the module
// is absent: nothing is sent. A send failure is logged at level 100 and never
// kills the goroutine.
func sendEdrStatusReports(ctx context.Context, sender resultSender, cnf *config.Config) {
	ticker := time.NewTicker(edrStatusPollInterval)
	defer ticker.Stop()

	var lastSentStatus string
	var lastSentAt time.Time

	send := func(status, version string) {
		report := &EdrStatusReport{
			AgentId:       strconv.Itoa(int(cnf.AgentID)),
			StatusJson:    status,
			PolicyVersion: version,
			ReportedAt:    timestamppb.Now(),
		}
		if err := sender.Send(&BidirectionalStream{StreamMessage: &BidirectionalStream_EdrStatusReport{EdrStatusReport: report}}); err != nil {
			utils.Logger.LogF(100, "EDR status report not delivered: %v", err)
			return
		}
		lastSentStatus = status
		lastSentAt = time.Now()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, err := os.ReadFile(edrStatusFile)
			if err != nil {
				// Module absent (or the file vanished): nothing to send.
				continue
			}
			content := string(status)
			changed := content != lastSentStatus
			due := time.Since(lastSentAt) >= edrStatusReportInterval
			if !changed && !due {
				continue
			}
			// Load the policy version only when we are about to send, to
			// avoid load churn on quiet ticks.
			version := ""
			if c, cerr := edrCfgLoad(); cerr == nil {
				version = c.PolicyVersion
			}
			send(content, version)
		}
	}
}
