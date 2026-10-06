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

// ShellActivityTelemetry builds a behavioral event carrying what a shell or
// interpreter was asked to run: its command line, attributed to the
// interpreter name. This is the Linux counterpart of ScriptBlockTelemetry
// (the PowerShell 4104 record) and feeds the same correlation rules. It is
// platform-shared (a pure event builder); the IsInterpreter gate that decides
// WHEN it fires is linux-only.
func ShellActivityTelemetry(cmdline, interpreter string) event.Event {
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "shell_activity",
		ObjectPath: interpreter,
		Signature:  truncate(cmdline, 8192),
	}
}
