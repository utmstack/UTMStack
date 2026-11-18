package service

import (
	"encoding/json"
	"os"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
)

func writeStoppedStatus(queue *orchestrator.Orchestrator, spool *event.Spool) {
	raw, err := os.ReadFile(config.StatusFile)
	if err != nil || len(raw) > 1<<20 {
		return
	}
	var doc statusDoc
	if json.Unmarshal(raw, &doc) != nil {
		return
	}
	doc.Running = false
	doc.EngineHealthy = false
	doc.EngineResident = false
	doc.SensorFileWatcher = false
	doc.SensorProcessGuard = false
	doc.SensorAMSI = false
	doc.AMSILocalReady = false
	doc.SensorBehavioral = false
	doc.RansomwareEnabled = false
	doc.RansomwareFeedHealthy = false
	doc.BlocklistEnabled = false
	doc.BlocklistEnforce = false
	doc.BlocklistActiveFilters = 0
	doc.ScanQueue = nil
	if queue != nil {
		stats := queue.Snapshot()
		doc.ScanQueue = &stats
	}
	doc.EventSpool = nil
	if spool != nil {
		stats := spool.Snapshot()
		doc.EventSpool = &stats
	}
	for name, st := range doc.SensorStates {
		st.State = "stopped"
		doc.SensorStates[name] = st
	}
	doc.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	raw, err = json.MarshalIndent(doc, "", "  ")
	if err == nil {
		_ = writeStatusFile(raw)
	}
}
