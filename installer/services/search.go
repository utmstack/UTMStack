package services

import (
	"fmt"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

// logIndexMappings pins the canonical fields for v11-log-* documents.
// "event" is OpenSearch's flat_object type: it stores every sub-key as a keyword
// and never infers types, so heterogeneous vendor values can no longer produce
// mapper_parsing_exception conflicts. "controls" holds compliance control tags.
// Top-level Event fields keep real types so the UI/SQL/sort work on them.
//
// This mappings template is scoped to v11-log-* ONLY (NOT v11-alert-*): alert
// documents carry e.g. severity as an integer, so they must not inherit these
// log-side text mappings. It carries no settings — settings come from the
// settings-only template (utmstack_log_indexes) that v11 applies to both.
const logIndexMappings = `
{
  "@timestamp": {"type":"date"},
  "dataType": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "dataSource": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "action": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "protocol": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "actionResult": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "severity": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
  "connectionStatus": {"type":"keyword"},
  "statusCode": {"type":"long"},
  "origin": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip"},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "target": {"type":"object","dynamic":true,"properties":{
    "ip": {"type":"ip"},
    "port": {"type":"long"},
    "user": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "host": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "path": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "file": {"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},
    "bytesSent": {"type":"double"},
    "bytesReceived": {"type":"double"}
  }},
  "event": {"type":"flat_object"},
  "controls": {"type":"keyword"}
}`

// newBagMappings are the mappings added to EXISTING v11-log-* indices on
// upgrade (event/controls are new keys — safe to add in-place; typed
// top-levels are not, so they stay out of the retro PUT).
const newBagMappings = `{"event":{"type":"flat_object"},"controls":{"type":"keyword"}}`

// alertIndexMappings pins the nested event bags inside alert documents so they
// use flat_object (same conflict-proof behaviour as the log index). Without
// this, lastEvent.event and events[N].event are dynamically mapped and two
// vendor events with the same field but different types will collide.
// The alert's own top-level fields (severity, status, tags, etc.) keep their
// natural dynamic types — only the embedded event bags need pinning.
const alertIndexMappings = `
{
  "lastEvent": {"type":"object","properties":{
    "event": {"type":"flat_object"},
    "controls": {"type":"keyword"}
  }},
  "events": {"type":"object","properties":{
    "event": {"type":"flat_object"},
    "controls": {"type":"keyword"}
  }}
}`

func getOpenSearchContainerID() (string, error) {
	containerIDs, err := utils.RunCmdWithOutput("docker", "ps", "-q", "-f", "name=utmstack_node1")
	if err != nil {
		return "", fmt.Errorf("error getting opensearch container: %v", err)
	}
	if len(containerIDs) == 0 {
		return "", fmt.Errorf("opensearch container not found")
	}
	return containerIDs[0], nil
}

func execCurl(containerID string, method, url, data string) error {
	cnf := config.GetConfig()
	args := []string{"exec", containerID, "curl", "-s", "-k", "-u", "admin:" + cnf.OpenSearchPassword, "-X", method}
	if data != "" {
		args = append(args, "-H", "Content-Type: application/json", "-d", data)
	}
	args = append(args, url)

	_, err := utils.RunCmdWithOutput("docker", args...)
	return err
}

func InitOpenSearch() error {
	containerID, err := getOpenSearchContainerID()
	if err != nil {
		return err
	}

	// Wait for OpenSearch to be ready
	for intent := 0; intent <= 10; intent++ {
		time.Sleep(1 * time.Minute)

		err := execCurl(containerID, "GET", "https://localhost:9200/_cluster/health?wait_for_status=green&timeout=50s", "")
		if err != nil {
			if intent >= 10 {
				return err
			}
		} else {
			break
		}
	}

	// Create snapshot repository
	snapshotData := `{"type":"fs","settings":{"location":"/usr/share/opensearch/.utm_geoip/","compress":true}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_snapshot/.utm_geoip", snapshotData); err != nil {
		return err
	}

	// Create index template
	templateData := `{"index_patterns":["v11-alert-*","v11-log-*",".utm-*",".utmstack-*"],"template":{"settings":{"index.number_of_shards":1,"index.number_of_replicas":0,"index.mapping.total_fields.limit":50000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_indexes", templateData); err != nil {
		return err
	}

	// v11 settings-only template (covers log + alert) — unchanged from v11.
	logTemplateData := `{"index_patterns":["v11-log-*","v11-alert-*"],"template":{"settings":{"index.max_shards":30000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}

	// Event-bag mappings template: scoped to v11-log-* ONLY (alert docs carry
	// severity as an integer — they must not inherit these log-side types).
	// No settings here; settings come from utmstack_log_indexes above.
	// Priority 10 so it composes with (not collides with) the settings-only
	// utmstack_indexes template, which also matches v11-log-* at priority 0.
	eventMappingsData := `{"index_patterns":["v11-log-*"],"priority":10,"template":{"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_event_mappings", eventMappingsData); err != nil {
		return err
	}

	// Alert-document mappings template: alerts embed full Event copies at
	// lastEvent and events[*] — each carrying an `event` bag. Without this,
	// those nested bags are dynamically mapped, and two alerts whose events
	// disagree on a bag field's type (int vs string) will collide, failing
	// alert ingestion. Pin them as flat_object (conflict-proof). The alert's
	// own top-level fields (severity int, status, tags…) stay dynamic.
	alertMappingsData := `{"index_patterns":["v11-alert-*"],"priority":10,"template":{"mappings":{"properties":` + alertIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_alert_event_mappings", alertMappingsData); err != nil {
		return err
	}


	// Restore geoip snapshot
	restoreData := `{"indices":".utm-geoip","include_global_state":false}`
	if err := execCurl(containerID, "POST", "https://localhost:9200/_snapshot/.utm_geoip/.utm_geoip/_restore", restoreData); err != nil {
		return err
	}

	return nil
}

func UpdateOpenSearch() error{

	containerID, err := getOpenSearchContainerID()
	if err != nil {
		return err
	}

	// (Re)create the v11 settings-only log template (covers log + alert).
	logTemplateData := `{"index_patterns":["v11-log-*","v11-alert-*"],"template":{"settings":{"index.max_shards":30000}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_indexes", logTemplateData); err != nil {
		return err
	}

	// (Re)create the event-bag mappings template (v11-log-* only).
	// Priority 10 so it composes with (not collides with) the settings-only
	// utmstack_indexes template, which also matches v11-log-* at priority 0.
	eventMappingsData := `{"index_patterns":["v11-log-*"],"priority":10,"template":{"mappings":{"properties":` + logIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_log_event_mappings", eventMappingsData); err != nil {
		return err
	}

	// (Re)create the alert-document event-bag mappings template (v11-alert-*).
	alertMappingsData := `{"index_patterns":["v11-alert-*"],"priority":10,"template":{"mappings":{"properties":` + alertIndexMappings + `}}}`
	if err := execCurl(containerID, "PUT", "https://localhost:9200/_index_template/utmstack_alert_event_mappings", alertMappingsData); err != nil {
		return err
	}

	// Add ONLY the new bag mappings to already-existing v11-log-* indices
	// (non-destructive: event/controls are new keys - no type clash with the
	// legacy dynamic log mapping). Typed top-levels are NOT retro-applied: they
	// can conflict with existing dynamic mappings and would 400 the upgrade.
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*/_mapping?allow_no_indices=true",
		`{"properties":` + newBagMappings + `}`); err != nil {
		return err
	}

	// Add the nested event-bag mappings to already-existing v11-alert-* indices.
	// lastEvent.event / events[*].event are new keys post-rename (previously
	// lastEvent.log), so this is additive - no conflict with existing mappings.
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-alert-*/_mapping?allow_no_indices=true",
		`{"properties":` + alertIndexMappings + `}`); err != nil {
		return err
	}

	// updated already existing index (update case)
	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-log-*,v11-alert-*/_settings?allow_no_indices=true", `{"index.max_shards":30000}`); err != nil {
		return err
	}

	if err := execCurl(containerID, "PUT", "https://localhost:9200/v11-alert-*,v11-log-*,.utm-*,.utmstack-*/_settings?allow_no_indices=true", `{"index.mapping.total_fields.limit":50000}`); err != nil {
		return err
	}
	return nil

}

