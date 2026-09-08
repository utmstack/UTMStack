//go:build darwin
// +build darwin

package platform

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/entities"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/tidwall/gjson"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

// esBinary already carries the com.apple.developer.endpoint-security.client
// entitlement, signed by Apple — reusing it is what makes this collector
// possible without UTMStack applying for that (restricted, Apple-approval-
// gated) entitlement itself. Ships with macOS since 12.3. Requires Full Disk
// Access granted to this agent binary (see run()'s ES_NEW_CLIENT_RESULT_ERR_
// NOT_PERMITTED handling below) — a deployment prerequisite (MDM TCC profile
// at fleet scale), not a code dependency.
const esBinary = "/usr/bin/eslogger"

// esEventTypes is what this collector subscribes to. The Unified Log (see
// Darwin, above) only carries what a daemon chose to os_log about itself —
// it has no record of process execution, and even where a CLI tool's own
// framework calls do get logged there (dscl, sudo), the arguments are almost
// always redacted as <private>. Endpoint Security is macOS's actual process-
// and file-event audit trail, with full unredacted argv/paths.
//
// This leans on the specific, purpose-built event types
// definitions/rules/macos/ needs rather than a generic "exec everything" —
// kextload, gatekeeper_user_override, tcc_modify and the od_* account events
// fire directly on the technique instead of requiring a regex over free text:
//   - exec, fork, signal        — process lineage and termination
//   - rename, create, unlink    — file tampering (log clearing, masquerading)
//   - setextattr, deleteextattr — xattr changes (Gatekeeper quarantine bypass)
//   - kextload                  — kernel extension loading
//   - gatekeeper_user_override  — explicit Gatekeeper bypass
//   - tcc_modify                — explicit TCC/privacy database changes
//   - btm_launch_item_add       — LaunchAgent/Daemon persistence registration
//   - od_create_user, od_group_add, od_group_set, od_enable_user,
//     od_modify_password        — local account/group manipulation
//   - su, sudo                  — explicit privilege-escalation events
//   - screensharing_attach      — incoming remote access
//   - trace                     — ptrace (process injection)
//   - xp_malware_detected       — XProtect's own verdict
//   - mount                     — disk image mounting (staging payloads)
var esEventTypes = []string{
	"exec", "fork", "signal",
	"rename", "create", "unlink", "setextattr", "deleteextattr",
	"kextload", "gatekeeper_user_override", "tcc_modify",
	"btm_launch_item_add",
	"od_create_user", "od_group_add", "od_group_set", "od_enable_user", "od_modify_password",
	"su", "sudo", "screensharing_attach", "trace", "xp_malware_detected", "mount",
}

// securityRelevantPathFragments gates the three high-volume filesystem event
// types (create, rename, unlink) — subscribing to those across the whole
// filesystem measured ~25-30 events/sec of routine temp-file and cache churn
// from every running app on an otherwise-idle Mac (216 create + 201 unlink in
// an 8s capture). No rule in definitions/rules/macos/ cares about a temp file
// appearing in /private/var/folders; they care about these specific
// locations.
var securityRelevantPathFragments = []string{
	"/Library/LaunchAgents/", "/Library/LaunchDaemons/",
	"/Library/Application Support/com.apple.TCC/", "/var/db/SystemPolicy",
	"/Library/Apple/System/Library/CoreServices/XProtect.bundle",
	"/usr/lib/cron/tabs/", "/etc/crontab", "/etc/hosts",
	"/.ssh/", "/Library/Extensions/", "/System/Library/Extensions/",
	"/etc/sudoers", "/etc/pam.d/",
	".bash_history", ".zsh_history", ".bash_profile", ".zshrc", ".bashrc", ".profile",
	"LoginItems", "com.apple.loginitems",
}

// esPathTarget/esExistingFile/esFsEvent pull just enough of eslogger's JSON
// to decide whether an event is worth shipping, without paying for a full
// unmarshal of the (large, mostly stat(2)-block) payload.
type esPathTarget struct {
	Path string `json:"path"`
}

func (t *esPathTarget) path() string {
	if t == nil {
		return ""
	}
	return t.Path
}

type esExistingFile struct {
	ExistingFile *esPathTarget `json:"existing_file"`
	// NewPath covers the other half of ES's create/rename destination union
	// (ES_DESTINATION_TYPE_NEW_PATH: a not-yet-existing file named under a
	// directory) — best-effort, since no sample of this variant was captured
	// during testing to confirm the exact field layout.
	NewPath *struct {
		Dir      esPathTarget `json:"dir"`
		Filename string       `json:"filename"`
	} `json:"new_path"`
}

func (d *esExistingFile) path() string {
	if d == nil {
		return ""
	}
	if d.ExistingFile != nil {
		return d.ExistingFile.Path
	}
	if d.NewPath != nil {
		return d.NewPath.Dir.Path + "/" + d.NewPath.Filename
	}
	return ""
}

type esFsEvent struct {
	Attr        string          `json:"extattr"`     // setextattr / deleteextattr
	Target      *esPathTarget   `json:"target"`      // setextattr / deleteextattr / unlink
	Source      *esPathTarget   `json:"source"`      // rename
	Destination *esExistingFile `json:"destination"` // create / rename
}

type esEnvelope struct {
	Event map[string]json.RawMessage `json:"event"`
}

// shouldShip decides whether a raw eslogger line is worth sending. exec,
// fork, signal and the purpose-built security event types (kextload,
// tcc_modify, the od_* account events, su, sudo, screensharing_attach,
// trace, xp_malware_detected, mount, gatekeeper_user_override,
// btm_launch_item_add) are already narrow — they always ship.
// create/rename/unlink/setextattr/deleteextattr are filtered to the one
// thing any rule actually checks them for: a quarantine-flag change, or a
// touch on a security-relevant path.
func shouldShip(raw string) bool {
	var env esEnvelope
	if err := json.Unmarshal([]byte(raw), &env); err != nil {
		// Shouldn't happen from eslogger's own output — ship rather than
		// silently drop something a human might need to see.
		return true
	}

	for eventType, body := range env.Event {
		switch eventType {
		case "setextattr", "deleteextattr":
			var fs esFsEvent
			if err := json.Unmarshal(body, &fs); err != nil {
				return true
			}
			return fs.Attr == "com.apple.quarantine"

		case "create", "rename", "unlink":
			var fs esFsEvent
			if err := json.Unmarshal(body, &fs); err != nil {
				return true
			}
			for _, p := range []string{fs.Target.path(), fs.Source.path(), fs.Destination.path()} {
				if p != "" && isSecurityRelevantPath(p) {
					return true
				}
			}
			return false
		}
	}

	return true
}

func isSecurityRelevantPath(path string) bool {
	for _, frag := range securityRelevantPathFragments {
		if strings.Contains(path, frag) {
			return true
		}
	}
	return false
}

// esLog is the shape the macOS OSLog collector (main.swift, in the
// UTMStackEnterprise repo) already emits, and the one
// definitions/filters/macos/macos.yaml and every rule under
// definitions/rules/macos/ is written against: one flat object with
// process/message/subsystem/level/category. reshape below targets exactly
// this — not a schema of its own — so Endpoint Security events flow through
// the same filter and the same rules as OSLog events, both tagged dataType
// "macos", with no second pipeline to keep in sync.
//
// It's also what makes CEL's string-only contains/regexMatch usable against
// argv at all: event.exec.args is a JSON array, which every one of those
// helpers silently treats as non-matching (they only match string-typed
// values), so it gets joined into plain "message" text here rather than
// shipped as JSON for a filter step to — there is no such step — flatten
// later.
type esLog struct {
	Timestamp        string `json:"timestamp"`
	ClassName        string `json:"class_name"`
	Process          string `json:"process"`
	Message          string `json:"message"`
	Subsystem        string `json:"subsystem"`
	Category         string `json:"category"`
	Level            string `json:"level"`
	StoreCat         string `json:"store_category"`
	Host             string `json:"host"`
	SigningId        string `json:"signing_id,omitempty"`
	TeamId           string `json:"team_id,omitempty"`
	IsPlatformBinary bool   `json:"is_platform_binary,omitempty"`
}

// reshape converts one already shouldShip()-approved eslogger line into the
// esLog shape above. ok is false only for an event type this collector
// doesn't know how to describe (shouldn't happen given esEventTypes is a
// fixed, matching list — defensive rather than expected). host is stamped
// onto every event the same way darwin.go stamps it onto OSLog entries —
// every Endpoint Security event is local to this Mac by definition, so
// macos.yaml can promote it to origin.host unconditionally.
//
// Field paths for the less common event types (tcc_modify, the od_*
// account events, btm_launch_item_add, ...) are best-effort against Apple's
// documented es_events_t layout: only exec/fork/signal/rename/create/unlink
// actually fired during local testing (an 8s capture on an idle dev Mac), so
// those are confirmed against a real sample and the rest are not. Expect to
// adjust these once real triggering during rule verification surfaces the
// actual field names. signing_id/team_id/is_platform_binary on the exec
// target are the same category of best-effort: they're sibling fields of
// es_process_t.executable per Apple's public header, not yet confirmed
// against a captured sample the way target.executable.path itself was.
func reshape(raw, host string) (esLog, bool) {
	root := gjson.Parse(raw)
	out := esLog{
		Timestamp: root.Get("time").String(),
		Level:     "notice",
		StoreCat:  "undefined",
		Host:      host,
	}

	caller := basename(root.Get("process.executable.path").String())

	switch {
	case root.Get("event.exec").Exists():
		e := root.Get("event.exec")
		out.ClassName = "es.exec"
		out.Process = basename(e.Get("target.executable.path").String())
		out.Message = joinArgs(e.Get("args"))
		out.SigningId = e.Get("target.signing_id").String()
		out.TeamId = e.Get("target.team_id").String()
		out.IsPlatformBinary = e.Get("target.is_platform_binary").Bool()

	case root.Get("event.fork").Exists():
		out.ClassName = "es.fork"
		out.Process = basename(root.Get("event.fork.child.executable.path").String())
		out.Message = "fork by " + caller

	case root.Get("event.signal").Exists():
		out.ClassName = "es.signal"
		out.Process = basename(root.Get("event.signal.target.executable.path").String())
		out.Message = "signal " + root.Get("event.signal.sig").String() + " sent by " + caller

	case root.Get("event.rename").Exists():
		out.ClassName = "es.rename"
		out.Process = caller
		out.Message = "rename " + root.Get("event.rename.source.path").String() + " to " +
			root.Get("event.rename.destination.existing_file.path").String()

	case root.Get("event.create").Exists():
		out.ClassName = "es.create"
		out.Process = caller
		out.Message = "create " + root.Get("event.create.destination.existing_file.path").String()

	case root.Get("event.unlink").Exists():
		out.ClassName = "es.unlink"
		out.Process = caller
		out.Message = "unlink " + root.Get("event.unlink.target.path").String()

	case root.Get("event.setextattr").Exists():
		out.ClassName = "es.setextattr"
		out.Process = caller
		// Worded as the equivalent CLI invocation (xattr -w) so it reads the
		// same way a rule already expects xattr activity to look — even
		// though this is quarantine being *set* (routine — Safari/curl
		// tagging a fresh download), not removed. See deleteextattr below
		// for the actual bypass case.
		out.Message = "xattr -w " + root.Get("event.setextattr.extattr").String() + " " +
			root.Get("event.setextattr.target.path").String()

	case root.Get("event.deleteextattr").Exists():
		out.ClassName = "es.deleteextattr"
		out.Process = caller
		out.Message = "xattr -d " + root.Get("event.deleteextattr.extattr").String() + " " +
			root.Get("event.deleteextattr.target.path").String()

	case root.Get("event.kextload").Exists():
		out.ClassName = "es.kextload"
		out.Process = "kextload"
		out.Subsystem = "com.apple.kernel"
		out.Message = "kext load: " + root.Get("event.kextload.identifier").String()

	case root.Get("event.gatekeeper_user_override").Exists():
		out.ClassName = "es.gatekeeper_user_override"
		out.Process = "syspolicyd"
		out.Subsystem = "com.apple.security.assessment"
		out.Message = "assessment bypass override for " + root.Get("event.gatekeeper_user_override.file.path").String()

	case root.Get("event.tcc_modify").Exists():
		out.ClassName = "es.tcc_modify"
		out.Process = "tccd"
		out.Subsystem = "com.apple.TCC"
		out.Message = "TCC grant modified: service=" + root.Get("event.tcc_modify.service").String() +
			" identity=" + root.Get("event.tcc_modify.identity").String()

	case root.Get("event.btm_launch_item_add").Exists():
		out.ClassName = "es.btm_launch_item_add"
		out.Process = "launchd"
		out.Message = "launch item added: " + root.Get("event.btm_launch_item_add.item.executable_path").String()

	case root.Get("event.od_create_user").Exists():
		out.ClassName = "es.od_create_user"
		out.Process = "opendirectoryd"
		out.Message = "dscl create user " + root.Get("event.od_create_user.user_name").String()

	case root.Get("event.od_group_add").Exists():
		out.ClassName = "es.od_group_add"
		out.Process = "dseditgroup"
		out.Message = "dseditgroup -o edit -a " + caller + " -t user " + root.Get("event.od_group_add.group_name").String()

	case root.Get("event.od_group_set").Exists():
		out.ClassName = "es.od_group_set"
		out.Process = "dseditgroup"
		out.Message = "dseditgroup -o edit " + root.Get("event.od_group_set.group_name").String()

	case root.Get("event.od_enable_user").Exists():
		out.ClassName = "es.od_enable_user"
		out.Process = "sysadminctl"
		out.Message = "-guestAccount on " + root.Get("event.od_enable_user.user_name").String()

	case root.Get("event.od_modify_password").Exists():
		out.ClassName = "es.od_modify_password"
		out.Process = "opendirectoryd"
		out.Message = "password modified for " + root.Get("event.od_modify_password.user_name").String()

	case root.Get("event.su").Exists():
		out.ClassName = "es.su"
		out.Process = "su"
		out.Message = "su to " + root.Get("event.su.to_username").String() +
			" success=" + root.Get("event.su.success").String()

	case root.Get("event.sudo").Exists():
		out.ClassName = "es.sudo"
		out.Process = "sudo"
		out.Message = "sudo success=" + root.Get("event.sudo.success").String()

	case root.Get("event.screensharing_attach").Exists():
		out.ClassName = "es.screensharing_attach"
		out.Process = "screensharingd"
		out.Message = "incoming connection from " + root.Get("event.screensharing_attach.source_address").String()

	case root.Get("event.trace").Exists():
		out.ClassName = "es.trace"
		out.Process = caller
		out.Message = "ptrace target " + basename(root.Get("event.trace.target.executable.path").String())

	case root.Get("event.xp_malware_detected").Exists():
		out.ClassName = "es.xp_malware_detected"
		out.Process = "XProtectService"
		out.Message = "malware detected: " + root.Get("event.xp_malware_detected.signature_version.malware_name").String()

	case root.Get("event.mount").Exists():
		out.ClassName = "es.mount"
		out.Process = caller
		out.Message = "mount " + root.Get("event.mount.statfs.f_mntonname").String()

	default:
		return esLog{}, false
	}

	return out, true
}

func basename(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func joinArgs(args gjson.Result) string {
	if !args.IsArray() {
		return ""
	}
	items := args.Array()
	parts := make([]string, 0, len(items))
	for _, a := range items {
		parts = append(parts, a.String())
	}
	return strings.Join(parts, " ")
}

type EndpointSecurity struct{}

func (e EndpointSecurity) Name() string { return "darwin-endpointsecurity" }

func (e EndpointSecurity) Start(ctx context.Context, enqueue func(*plugins.Log) error) {
	if _, err := os.Stat(esBinary); os.IsNotExist(err) {
		utils.Logger.ErrorF("eslogger not found at %s (requires macOS 12.3+)", esBinary)
		return
	}

	host, err := os.Hostname()
	if err != nil {
		utils.Logger.ErrorF("error getting hostname: %v", err)
		host = "unknown"
	}

	restartDelay := baseRestartDelay

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("Endpoint Security collector stopping due to context cancellation")
			return
		default:
		}

		exitCode := e.run(ctx, host, enqueue)

		if exitCode == 0 {
			utils.Logger.Info("Endpoint Security collector exited normally")
		} else {
			// ES_NEW_CLIENT_RESULT_ERR_NOT_PERMITTED (no Full Disk Access
			// granted to this binary) surfaces here as a nonzero exit with the
			// reason already logged from stderr below — restarting won't fix
			// that on its own, but backing off instead of exiting outright
			// means it recovers on its own once FDA is granted, with no
			// service restart required.
			utils.Logger.ErrorF("Endpoint Security collector exited with code %d, restarting in %v", exitCode, restartDelay)
		}

		if ctx.Err() != nil {
			return
		}
		time.Sleep(restartDelay)

		restartDelay *= 2
		if restartDelay > maxRestartDelay {
			restartDelay = maxRestartDelay
		}
	}
}

func (e EndpointSecurity) run(ctx context.Context, host string, enqueue func(*plugins.Log) error) int {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in Endpoint Security collector: %v", r)
		}
	}()

	cmd := exec.CommandContext(ctx, esBinary, esEventTypes...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		utils.Logger.ErrorF("error creating stdout pipe: %v", err)
		return -1
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		utils.Logger.ErrorF("error creating stderr pipe: %v", err)
		return -1
	}

	if err := cmd.Start(); err != nil {
		utils.Logger.ErrorF("error starting eslogger: %v", err)
		return -1
	}

	utils.Logger.Info("Endpoint Security collector started successfully")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in eslogger stdout reader: %v", r)
			}
		}()

		scanner := bufio.NewScanner(stdout)
		// exec events carry full argv/env; bufio.Scanner's 64KB default token
		// size truncates the whole stream read on the first line that exceeds
		// it, not just that one event.
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			logLine := scanner.Text()

			if !shouldShip(logLine) {
				continue
			}

			reshaped, ok := reshape(logLine, host)
			if !ok {
				continue
			}

			shipped, err := json.Marshal(reshaped)
			if err != nil {
				utils.Logger.ErrorF("error marshaling reshaped Endpoint Security log: %v", err)
				continue
			}

			validatedLog, _, err := entities.ValidateString(string(shipped), false)
			if err != nil {
				utils.Logger.ErrorF("error validating Endpoint Security log: %v", err)
				continue
			}

			if err := enqueue(&plugins.Log{
				DataType:   string(config.DataTypeMacOs),
				DataSource: host,
				Raw:        validatedLog,
			}); err != nil {
				utils.Logger.ErrorF("failed to persist Endpoint Security log: %v", err)
			}
		}

		if err := scanner.Err(); err != nil {
			utils.Logger.ErrorF("error reading eslogger stdout: %v", err)
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in eslogger stderr reader: %v", r)
			}
		}()

		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errLine := scanner.Text()
			utils.Logger.ErrorF("eslogger error: %s", errLine)
		}

		if err := scanner.Err(); err != nil {
			utils.Logger.ErrorF("error reading eslogger stderr: %v", err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		utils.Logger.ErrorF("Endpoint Security collector process ended with error: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return -1
	}

	return 0
}

func (e EndpointSecurity) Stop() {}
