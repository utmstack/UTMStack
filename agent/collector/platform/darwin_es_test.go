//go:build darwin
// +build darwin

package platform

import (
	"strings"
	"testing"
)

// Samples are reconstructed from a real 8-second eslogger capture on an
// otherwise-idle dev Mac (see darwin_es.go's package comment for the
// numbers): the noisy cases are what dominated that capture, the relevant
// ones are what a rule in definitions/rules/macos/ actually needs to see.
func TestShouldShip(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "setextattr: routine logd metadata attr -- noise",
			line: `{"event":{"setextattr":{"extattr":"com.apple.logd.metadata","target":{"path":"/private/var/folders/zz/T/lsfw-tmp.tracev3"}}}}`,
			want: false,
		},
		{
			name: "setextattr: quarantine flag removed -- Gatekeeper bypass",
			line: `{"event":{"setextattr":{"extattr":"com.apple.quarantine","target":{"path":"/Users/test/Downloads/app.dmg"}}}}`,
			want: true,
		},
		{
			name: "create: app writing its own temp file -- noise",
			line: `{"event":{"create":{"destination":{"existing_file":{"path":"/opt/datadog-agent/run/registry.json.tmp1488990576"}},"destination_type":0}}}`,
			want: false,
		},
		{
			name: "create: new LaunchAgent plist -- persistence",
			line: `{"event":{"create":{"destination":{"existing_file":{"path":"/Users/test/Library/LaunchAgents/com.evil.plist"}},"destination_type":0}}}`,
			want: true,
		},
		{
			name: "unlink: app's own temp file cleanup -- noise",
			line: `{"event":{"unlink":{"parent_dir":{"path":"/private/var/folders/zz/T"},"target":{"path":"/private/var/folders/zz/T/lsfw-tmp.tracev3"}}}}`,
			want: false,
		},
		{
			name: "unlink: shell history cleared -- anti-forensics",
			line: `{"event":{"unlink":{"parent_dir":{"path":"/Users/test"},"target":{"path":"/Users/test/.zsh_history"}}}}`,
			want: true,
		},
		{
			name: "rename: app's own atomic-write temp swap -- noise",
			line: `{"event":{"rename":{"source":{"path":"/opt/datadog-agent/run/registry.json.tmp1488990576"},"destination":{"existing_file":{"path":"/opt/datadog-agent/run/registry.json"}},"destination_type":0}}}`,
			want: false,
		},
		{
			name: "exec always ships",
			line: `{"event":{"exec":{"target":{"executable":{"path":"/bin/bash"}}}}}`,
			want: true,
		},
		{
			name: "tcc_modify always ships",
			line: `{"event":{"tcc_modify":{}}}`,
			want: true,
		},
		{
			name: "malformed JSON ships rather than silently drops",
			line: `not json`,
			want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := shouldShip(c.line)
			if got != c.want {
				t.Errorf("shouldShip(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestReshape confirms the ES -> OSLog-shaped conversion produces what the
// existing macos.yaml filter and definitions/rules/macos/ rules expect:
// plain process/message strings, not the nested ES JSON a rule can't search.
func TestReshape(t *testing.T) {
	// Real exec sample captured locally (see the eslogger validation in the
	// session this collector came from): args carries the full command line,
	// unredacted, which the Unified Log-based collector could never give us.
	exec := `{"time":"2026-09-07T11:59:28Z","process":{"executable":{"path":"/sbin/launchd"}},"event":{"exec":{"target":{"executable":{"path":"/usr/libexec/xpcproxy"}},"args":["xpcproxy","com.apple.endpointsecurity.endpointsecurityd"]}}}`
	got, ok := reshape(exec, "Test-Mac")
	if !ok {
		t.Fatal("reshape returned ok=false for a real exec sample")
	}
	if got.Process != "xpcproxy" {
		t.Errorf("exec Process = %q, want %q", got.Process, "xpcproxy")
	}
	if got.Message != "xpcproxy com.apple.endpointsecurity.endpointsecurityd" {
		t.Errorf("exec Message = %q, want the joined argv", got.Message)
	}
	if got.Host != "Test-Mac" {
		t.Errorf("exec Host = %q, want %q", got.Host, "Test-Mac")
	}

	// signing_id/team_id/is_platform_binary are best-effort field paths (not
	// yet confirmed against a real capture — see reshape's doc comment), so
	// this is a synthetic sample checking the extraction logic itself, not
	// that eslogger actually names the fields this way.
	signed := `{"time":"2026-09-07T12:01:00Z","process":{"executable":{"path":"/bin/bash"}},"event":{"exec":{"target":{"executable":{"path":"/usr/bin/curl"},"signing_id":"com.apple.curl","team_id":"","is_platform_binary":true},"args":["curl","https://example.com"]}}}`
	got, ok = reshape(signed, "Test-Mac")
	if !ok {
		t.Fatal("reshape returned ok=false for a synthetic signed exec sample")
	}
	if got.SigningId != "com.apple.curl" {
		t.Errorf("exec SigningId = %q, want %q", got.SigningId, "com.apple.curl")
	}
	if !got.IsPlatformBinary {
		t.Error("exec IsPlatformBinary = false, want true")
	}

	// gatekeeper_xattr_bypass.yaml (unchanged, pre-existing rule) checks
	// contains(eventMessage,"xattr") && contains(eventMessage,"-d") &&
	// contains(eventMessage,"com.apple.quarantine") — deleteextattr's message
	// must satisfy all three unmodified for that rule to still work against
	// Endpoint Security data.
	del := `{"time":"2026-09-07T12:00:00Z","process":{"executable":{"path":"/usr/bin/xattr"}},"event":{"deleteextattr":{"extattr":"com.apple.quarantine","target":{"path":"/Users/test/Downloads/app.dmg"}}}}`
	got, ok = reshape(del, "Test-Mac")
	if !ok {
		t.Fatal("reshape returned ok=false for deleteextattr")
	}
	want := "xattr -d com.apple.quarantine /Users/test/Downloads/app.dmg"
	if got.Message != want {
		t.Errorf("deleteextattr Message = %q, want %q", got.Message, want)
	}
	if !strings.Contains(got.Message, "xattr") || !strings.Contains(got.Message, "-d") || !strings.Contains(got.Message, "com.apple.quarantine") {
		t.Errorf("deleteextattr Message %q does not satisfy gatekeeper_xattr_bypass.yaml's existing where clause", got.Message)
	}

	// An event type this collector doesn't know how to describe should never
	// reach here in practice (esEventTypes only subscribes to known types),
	// but reshape must not fabricate a log for it.
	if _, ok := reshape(`{"event":{"something_unknown":{}}}`, "Test-Mac"); ok {
		t.Error("reshape returned ok=true for an unrecognized event type")
	}
}
