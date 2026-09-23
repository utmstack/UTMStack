package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/utmstack/UTMStack/agent/config"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/svc"
)

// edrCommandTimeout bounds every EDR subprocess; the handler must never block
// the stream indefinitely.
const edrCommandTimeout = 30 * time.Second

// edrCLI is the subprocess boundary. Production uses runEDRCLI; tests inject a
// fake so no real binary or service is ever touched.
type edrCLI interface {
	run(ctx context.Context, args ...string) (stdout string, stderr string, err error)
}

// runEDRCLI executes the EDR module binary with the given verb arguments and
// returns its captured stdout and stderr. The context deadline (30s) bounds
// the run.
type runEDRCLI struct {
	binPath string
}

func (r runEDRCLI) run(ctx context.Context, args ...string) (string, string, error) {
	cmd := exec.CommandContext(ctx, r.binPath, args...)
	var out, errBuf strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return out.String(), errBuf.String(), err
}

// edrSvcStart / edrSvcStop / edrSvcActive wrap the service control so tests
// can substitute fakes (hermetic — no real service is started, stopped, or
// interrogated).
var (
	edrSvcStart  = func() error { return svc.Start(edrconfig.ServiceName) }
	edrSvcStop   = func() error { return svc.Stop(edrconfig.ServiceName) }
	edrSvcActive = func() (bool, error) { return svc.IsActive(edrconfig.ServiceName) }
)

// edrCfgLoad / edrCfgSave wrap the EDR config file I/O so tests can redirect
// them to a temp dir (hermetic — never touches the real edr.json next to the
// agent binary).
var (
	edrCfgLoad = func() (edrconfig.EDRConfig, error) { return edrconfig.Load() }
	edrCfgSave = func(c edrconfig.EDRConfig) error { return edrconfig.Save(c) }
)

// edrDepFileName mirrors dependency.EDRFile("") — the dependency package
// cannot be imported here (it imports agent/agent).
func edrDepFileName() string {
	name := fmt.Sprintf("utmstack_edr_%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func edrBinPath() string {
	return filepath.Join(fs.GetExecutablePath(), edrDepFileName())
}

// edrCLIEnvelope is the --json envelope emitted by the EDR CLI (Y1.2):
// {"ok":bool,"error":"...","data":{...}} with error omitted on success.
type edrCLIEnvelope struct {
	Ok    bool            `json:"ok"`
	Error string          `json:"error"`
	Data  json.RawMessage `json:"data"`
}

func parseEDREnvelope(stdout string) (*edrCLIEnvelope, error) {
	trimmed := strings.TrimSpace(stdout)
	if trimmed == "" {
		return nil, errors.New("empty output")
	}
	line := strings.SplitN(trimmed, "\n", 2)[0]
	var env edrCLIEnvelope
	if err := json.Unmarshal([]byte(line), &env); err != nil {
		return nil, err
	}
	return &env, nil
}

// edrCommandProcessor handles a typed EDR control command: it audits the
// request, dispatches the action, and always replies with a typed EdrResult.
// It never returns an error — a failed send is logged and the stream loop
// continues.
func edrCommandProcessor(sender resultSender, cnf *config.Config, cmd *EdrCommand) {
	utils.Logger.LogF(100, "EDR command received: action=%s cmd_id=%s executed_by=%s reason=%s",
		cmd.GetAction(), cmd.GetCmdId(), cmd.GetExecutedBy(), cmd.GetReason())

	binPath := edrBinPath()
	ok, errStr, dataPayload := edrDispatch(cmd.GetAction(), cmd.GetPayload(), binPath, runEDRCLI{binPath: binPath})
	sendEdrResult(sender, cnf, cmd, ok, errStr, dataPayload)
}

// edrDispatch maps one of the 14 EDR actions to a mechanism — service control
// (shared/svc), a config write (agent/edr/config), a CLI subprocess, or an
// immediate unsupported reply — and returns the result triple. The binary
// exists check runs for every action that needs the module.
func edrDispatch(action, payload, binPath string, cli edrCLI) (ok bool, errStr string, dataPayload string) {
	switch action {
	case "enable":
		return edrEnable(binPath)
	case "disable":
		return edrDisable(binPath)
	case "status":
		return edrCLIAction(cli, binPath, "status")
	case "policy_set":
		return edrPolicySet(payload)
	case "quarantine_list":
		return edrCLIAction(cli, binPath, "quarantine", "list")
	case "quarantine_restore":
		id, err := payloadString(payload, "id")
		if err != nil {
			return false, "validation_error", ""
		}
		return edrCLIAction(cli, binPath, "quarantine", "restore", id)
	case "quarantine_purge":
		args, err := purgeArgs(payload)
		if err != nil {
			return false, "validation_error", ""
		}
		return edrCLIAction(cli, binPath, args...)
	case "scan_path":
		p, err := payloadString(payload, "path")
		if err != nil {
			return false, "validation_error", ""
		}
		return edrCLIAction(cli, binPath, "scan", p)
	case "allow_add":
		cat, val, err := allowPayload(payload)
		if err != nil {
			return false, "validation_error", ""
		}
		return edrCLIAction(cli, binPath, "allow", cat, "add", val)
	case "allow_remove":
		cat, val, err := allowPayload(payload)
		if err != nil {
			return false, "validation_error", ""
		}
		return edrCLIAction(cli, binPath, "allow", cat, "remove", val)
	case "kill_process", "full_scan", "isolate", "release_isolation", "blocklist_refresh":
		// No endpoint backend yet.
		return false, "unsupported_action", ""
	default:
		return false, "unsupported_action", ""
	}
}

// edrEnable flips the module on in edr.json and starts the service.
func edrEnable(binPath string) (bool, string, string) {
	if !fs.Exists(binPath) {
		return false, "module_not_installed", ""
	}
	c, err := edrCfgLoad()
	if err != nil {
		return false, err.Error(), ""
	}
	c.Enabled = true
	if err := edrCfgSave(c); err != nil {
		return false, err.Error(), ""
	}
	if err := edrSvcStart(); err != nil {
		return false, err.Error(), ""
	}
	return true, "", `{"enabled":true,"service":"started"}`
}

// edrDisable stops the service and flips the module off in edr.json. A service
// that is not running is not an error. The failure is reported as
// service_not_running when the service was already down AND the config save
// also failed (nothing to disable and the state could not be persisted);
// otherwise a save failure surfaces its own error.
func edrDisable(binPath string) (bool, string, string) {
	if !fs.Exists(binPath) {
		return false, "module_not_installed", ""
	}
	active, _ := edrSvcActive()
	stopErr := edrSvcStop()
	c, err := edrCfgLoad()
	if err != nil {
		return false, err.Error(), ""
	}
	c.Enabled = false
	if saveErr := edrCfgSave(c); saveErr != nil {
		if !active {
			return false, "service_not_running", ""
		}
		return false, saveErr.Error(), ""
	}
	if stopErr != nil {
		utils.Logger.LogF(100, "EDR disable: service stop failed (already stopped?): %v", stopErr)
	}
	return true, "", `{"enabled":false,"service":"stopped"}`
}

// edrPolicySet shallow-merges the policy document into edr.json. Only keys
// present in the payload overwrite; absent keys are left untouched.
func edrPolicySet(payload string) (bool, string, string) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &fields); err != nil || fields == nil {
		return false, "validation_error", ""
	}
	base, err := edrCfgLoad()
	if err != nil {
		return false, err.Error(), ""
	}
	// Serialize the effective config, overlay the payload keys, then decode
	// back into a typed struct and persist. The overlay is shallow: a payload
	// key replaces that top-level field wholesale.
	b, err := json.Marshal(base)
	if err != nil {
		return false, err.Error(), ""
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(b, &doc); err != nil {
		return false, "validation_error", ""
	}
	for k, v := range fields {
		doc[k] = v
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return false, err.Error(), ""
	}
	var c edrconfig.EDRConfig
	if err := json.Unmarshal(out, &c); err != nil {
		return false, "validation_error", ""
	}
	if err := edrCfgSave(c); err != nil {
		return false, err.Error(), ""
	}
	return true, "", `{"applied":true}`
}

// edrCLIAction verifies the module binary exists, runs the CLI verb with
// --json, and maps the envelope onto the result triple.
func edrCLIAction(cli edrCLI, binPath string, args ...string) (bool, string, string) {
	if !fs.Exists(binPath) {
		return false, "module_not_installed", ""
	}
	args = append(args, "--json")
	ctx, cancel := context.WithTimeout(context.Background(), edrCommandTimeout)
	defer cancel()

	stdout, _, err := cli.run(ctx, args...)
	if err != nil {
		// A timed-out process never produces a usable envelope. In production
		// ctx.Err() is DeadlineExceeded once the 30s bound passes; a fake CLI
		// (tests) may surface the same error directly.
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return false, "timeout", ""
		}
		// Non-zero exit: the envelope is authoritative when parseable.
		if env, perr := parseEDREnvelope(stdout); perr == nil {
			return mapEnvelope(env)
		}
		return false, "invalid_module_output", ""
	}
	env, perr := parseEDREnvelope(stdout)
	if perr != nil {
		return false, "invalid_module_output", ""
	}
	return mapEnvelope(env)
}

// mapEnvelope converts a parsed CLI envelope into the result triple.
func mapEnvelope(env *edrCLIEnvelope) (bool, string, string) {
	if env.Ok {
		data := ""
		if len(env.Data) > 0 {
			data = string(env.Data)
		}
		return true, "", data
	}
	errStr := env.Error
	if errStr == "" {
		errStr = "invalid_module_output"
	}
	return false, errStr, ""
}

// sendEdrResult delivers the typed reply. A send failure is logged; the
// stream loop must keep running.
func sendEdrResult(sender resultSender, cnf *config.Config, cmd *EdrCommand, ok bool, errStr string, dataPayload string) {
	res := &EdrResult{
		AgentId:    strconv.Itoa(int(cnf.AgentID)),
		CmdId:      cmd.GetCmdId(),
		Ok:         ok,
		Error:      errStr,
		ExecutedAt: timestamppb.Now(),
	}
	if ok && dataPayload != "" {
		res.Payload = dataPayload
	}
	if err := sender.Send(&BidirectionalStream{StreamMessage: &BidirectionalStream_EdrResult{EdrResult: res}}); err != nil {
		utils.Logger.LogF(100, "EDR result not delivered: %v", err)
	}
}

// payloadString extracts a required non-empty string field from the payload.
func payloadString(payload string, key string) (string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return "", err
	}
	v, ok := m[key]
	if !ok {
		return "", fmt.Errorf("%s missing", key)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("%s invalid", key)
	}
	return s, nil
}

// purgeArgs builds the CLI args for quarantine purge: {"id":"..."} or
// {"expired":true}.
func purgeArgs(payload string) ([]string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(payload), &m); err != nil {
		return nil, err
	}
	if id, ok := m["id"]; ok {
		s, ok := id.(string)
		if !ok || s == "" {
			return nil, fmt.Errorf("id invalid")
		}
		return []string{"quarantine", "purge", s}, nil
	}
	if exp, ok := m["expired"]; ok {
		b, ok := exp.(bool)
		if !ok || !b {
			return nil, fmt.Errorf("expired must be true")
		}
		return []string{"quarantine", "purge", "--expired"}, nil
	}
	return nil, fmt.Errorf("quarantine purge requires id or expired")
}

var allowCategories = map[string]string{
	"path":    "path",
	"process": "process",
	"command": "command",
	"network": "network",
}

// allowPayload extracts and validates {category, value} for allow verbs.
func allowPayload(payload string) (category string, value string, err error) {
	var m map[string]any
	if err = json.Unmarshal([]byte(payload), &m); err != nil {
		return "", "", err
	}
	cat, ok := m["category"].(string)
	if !ok || cat == "" {
		return "", "", fmt.Errorf("category missing")
	}
	mapped, ok := allowCategories[cat]
	if !ok {
		return "", "", fmt.Errorf("category %q not one of path/process/command/network", cat)
	}
	val, ok := m["value"].(string)
	if !ok || val == "" {
		return "", "", fmt.Errorf("value missing")
	}
	return mapped, val, nil
}
