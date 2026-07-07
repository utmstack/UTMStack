package behavioral

import (
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// ProcTelemetry builds a behavioral telemetry event for a process start.
func ProcTelemetry(p proctable.Proc) event.Event {
	pc := event.ProcInfo{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline}
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "process_create",
		ObjectPath: p.Image,
		Process:    &pc,
	}
}

// ScriptBlockTelemetry builds a behavioral event carrying deobfuscated script text.
func ScriptBlockTelemetry(scriptText, appName string) event.Event {
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "script_block",
		ObjectPath: appName,
		Signature:  truncate(scriptText, 8192),
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
