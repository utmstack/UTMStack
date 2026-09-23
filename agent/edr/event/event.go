package event

import (
	"encoding/json"
	"os"
	"time"
)

const (
	DataType = "utmstack_edr"
	Product  = "UTMStack EDR"

	SourceFileWatcher    = "file_watcher"
	SourceProcessWatcher = "process_watcher"
	SourceAMSI           = "amsi"
	SourceEngine         = "edr_engine"
	SourceBehavioral     = "behavioral"
	SourceNetwork        = "network_watcher"
	SourceScheduled      = "scheduled_scan"

	ActionDetected     = "detected"
	ActionQuarantined  = "quarantined"
	ActionKilled       = "killed"
	ActionScanStart    = "scan_start"
	ActionScanComplete = "scan_complete"
	ActionHealth       = "health"
	ActionBlocked      = "blocked"
	ActionAllowed      = "allowed"

	ActionRansomwareSuspected = "ransomware_suspected"
	ActionRansomwareContained = "ransomware_contained"
)

// NewAMSIEvent builds a branded amsi event. appName is the script host that
// submitted the content (carried in ObjectPath).
func NewAMSIEvent(action, verdict, signature, appName string) Event {
	return Event{
		Source:     SourceAMSI,
		Action:     action,
		Verdict:    verdict,
		Signature:  signature,
		ObjectPath: appName,
	}
}

type ProcInfo struct {
	PID     int    `json:"pid"`
	PPID    int    `json:"ppid"`
	Image   string `json:"image"`
	Cmdline string `json:"command_line,omitempty"`
	Session int    `json:"session,omitempty"`
	User    string `json:"user,omitempty"`
}

type Event struct {
	Timestamp  string    `json:"timestamp"`
	HostID     string    `json:"host_id"`
	OS         string    `json:"os"`
	Product    string    `json:"product"`
	Source     string    `json:"source"`
	Action     string    `json:"action"`
	ObjectPath string    `json:"object_path,omitempty"`
	FileHash   string    `json:"file_hash,omitempty"`
	Verdict    string    `json:"verdict,omitempty"`
	Signature  string    `json:"signature,omitempty"`
	RemoteIP   string    `json:"remote_ip,omitempty"`
	RemotePort int       `json:"remote_port,omitempty"`
	Domain     string    `json:"domain,omitempty"`
	Direction  string    `json:"direction,omitempty"` // "outbound" | "inbound"
	Indicator  string    `json:"indicator,omitempty"` // matched blocklist value
	Engine     string    `json:"engine"`
	Severity   string    `json:"severity,omitempty"`
	Process    *ProcInfo `json:"process,omitempty"`

	// Scheduled-scan summary. Present only on scan_start/scan_complete
	// events; omitempty keeps existing event shapes unchanged.
	FilesScanned int   `json:"files_scanned,omitempty"`
	Detections   int   `json:"detections,omitempty"`
	ElapsedMs    int64 `json:"elapsed_ms,omitempty"`
}

func NewDetection(objectPath, sha256, signature, source string) Event {
	return Event{
		Source:     source,
		Action:     ActionDetected,
		ObjectPath: objectPath,
		FileHash:   sha256,
		Verdict:    "malicious",
		Signature:  signature,
	}
}

// NewProcessAction builds a branded event for a responder action on a process
// (killed / suspended / resumed).
func NewProcessAction(action, source string, p ProcInfo, verdict, signature string) Event {
	pc := p
	return Event{
		Source:     source,
		Action:     action,
		Verdict:    verdict,
		Signature:  signature,
		ObjectPath: p.Image,
		Process:    &pc,
	}
}

// NewRansomwareEvent builds a branded behavioral event for a ransomware
// escalation. signal is the top contributing detector (e.g.
// "canary:00__accounts.xlsx" or "t1490:vssadmin_delete_shadows"); severity is
// a coarse label ("high"|"critical"). Source is always behavioral so the
// platform parser routes it with the other behavioral telemetry.
func NewRansomwareEvent(action string, p ProcInfo, signal, severity string) Event {
	pc := p
	return Event{
		Source:     SourceBehavioral,
		Action:     action,
		ObjectPath: p.Image,
		Signature:  signal,
		Severity:   severity,
		Process:    &pc,
	}
}

// NewNetworkEvent builds a branded network_watcher event for a connection to/from
// a blacklisted indicator. action is blocked|detected; direction is outbound|inbound.
// domain is set only when the match came via DNS (Plan 2). p may be nil when the
// responsible process is unknown.
func NewNetworkEvent(action, direction, remoteIP string, port int, domain, indicator string, p *ProcInfo) Event {
	var pc *ProcInfo
	if p != nil {
		c := *p
		pc = &c
	}
	return Event{
		Source:     SourceNetwork,
		Action:     action,
		Verdict:    "malicious",
		Direction:  direction,
		RemoteIP:   remoteIP,
		RemotePort: port,
		Domain:     domain,
		Indicator:  indicator,
		Process:    pc,
	}
}

func (e Event) ToJSON() (string, error) {
	if e.Product == "" {
		e.Product = Product
	}
	if e.Engine == "" {
		e.Engine = Product
	}
	if e.OS == "" {
		e.OS = "windows"
	}
	if e.Timestamp == "" {
		e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if e.HostID == "" {
		if h, err := os.Hostname(); err == nil {
			e.HostID = h
		}
	}
	b, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
