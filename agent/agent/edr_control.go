package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
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

// edrSvcStart / edrSvcStop / edrSvcActive / edrSvcRestart wrap the service
// control so tests can substitute fakes (hermetic — no real service is started,
// stopped, restarted, or interrogated).
var (
	edrSvcStart   = func() error { return svc.Start(edrconfig.ServiceName) }
	edrSvcStop    = func() error { return svc.Stop(edrconfig.ServiceName) }
	edrSvcActive  = func() (bool, error) { return svc.IsActive(edrconfig.ServiceName) }
	edrSvcRestart = func() error { return svc.Restart(edrconfig.ServiceName) }
)

// edrCfgLoad / edrCfgSave wrap the EDR config file I/O so tests can redirect
// them to a temp dir (hermetic — never touches the real edr.json next to the
// agent binary).
var (
	edrCfgLoad = func() (edrconfig.EDRConfig, error) { return edrconfig.Load() }
	edrCfgSave = func(c edrconfig.EDRConfig) error { return edrconfig.Save(c) }
)

// edrPolicyFingerprintFile is the "known-good post-policy" snapshot stored
// alongside edr.json: the top-level config state produced by the last applied
// policy_set. The EDR service never reads it (it is not part of EDRConfig); it
// exists so the agent can detect local hand-edits (drift) against the central
// policy. A package-var seam so tests redirect it to t.TempDir().
var edrPolicyFingerprintFile = filepath.Join(edrconfig.InstallDir, "edr.policy-applied.json")

// edrPolicyDriftFile persists the drift trace of the last applied policy_set
// next to the fingerprint, so the drift is durable and inspectable even though
// it is not part of EDRConfig (the EDR service must not see it).
var edrPolicyDriftFile = filepath.Join(edrconfig.InstallDir, "edr.policy-drift.json")

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

// edrPolicySet applies a centrally-assigned policy document:
// {"policy":{...},"version":"..."}. Only keys present in `policy`
// overwrite; absent keys are left untouched. The version is recorded in the
// config so the status reporter can ship it upstream for drift detection.
// A missing or non-object `policy` is a validation error — a version alone
// does not constitute a policy document.
//
// After the merge, the handler classifies which top-level keys changed:
//   - LIVE keys (allowlist, blocklist, and ransomware when only its
//     response_mode changed) take effect without a restart.
//   - RESTART keys (sensors, ransomware's other fields, every other
//     top-level key) require a service restart to take effect.
//   - enabled is service control: a flip starts or stops the service and is
//     reported live — never restarted.
//
// If any RESTART key changed and the service is active, the module is
// restarted. The reply reports the split plus any drift (local hand-edits to
// edr.json since the last central policy) as Opción A: the central value wins
// and the local edit is recorded, not reverted.
func edrPolicySet(payload string) (bool, string, string) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &fields); err != nil || fields == nil {
		return false, "validation_error", ""
	}
	policyRaw, hasPolicy := fields["policy"]
	if !hasPolicy {
		return false, "validation_error", ""
	}
	var policy map[string]json.RawMessage
	if err := json.Unmarshal(policyRaw, &policy); err != nil || policy == nil {
		return false, "validation_error", ""
	}
	if len(policy) == 0 {
		// An empty policy object carries nothing to apply; a version alone
		// does not constitute a policy document either.
		return false, "validation_error", ""
	}
	var version string
	if v, ok := fields["version"]; ok {
		if err := json.Unmarshal(v, &version); err != nil {
			return false, "validation_error", ""
		}
	}
	base, err := edrCfgLoad()
	if err != nil {
		return false, err.Error(), ""
	}

	// Fingerprint = the effective top-level state from the last central policy.
	// A missing or corrupt fingerprint means "no prior central policy" — no
	// drift is reported for it.
	var fpDoc map[string]json.RawMessage
	if b, ferr := os.ReadFile(edrPolicyFingerprintFile); ferr == nil {
		_ = json.Unmarshal(b, &fpDoc)
	}

	// Serialize the effective config, overlay the policy keys, then decode
	// back into a typed struct and persist. The overlay is shallow: a policy
	// key replaces that top-level field wholesale.
	b, err := json.Marshal(base)
	if err != nil {
		return false, err.Error(), ""
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(b, &doc); err != nil {
		return false, "validation_error", ""
	}
	// preDoc is the effective top-level state BEFORE this policy is applied.
	// Captured before the overlay (and before the agent stamps policy_version)
	// so it drives both the change classification and the drift comparison.
	preDoc := make(map[string]json.RawMessage, len(doc))
	for k, v := range doc {
		preDoc[k] = v
	}
	for k, v := range policy {
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
	c.PolicyVersion = version
	if err := edrCfgSave(c); err != nil {
		return false, err.Error(), ""
	}

	// Classify the top-level changes (pre vs post effective state, plus the
	// nested ransomware.response_mode special case).
	changed := changedTopLevelKeys(preDoc, c)

	// Opción A drift: any top-level key whose current value differs from the
	// last known-good central fingerprint was hand-edited locally. The central
	// policy still wins (the merge above already ran); we only leave a visible
	// trace.
	drift := detectDrift(preDoc, policy, fpDoc)

	// enabled is service control, never a restart: false→true starts the
	// service, true→false stops it, no change does nothing.
	if _, changedEnabled := changed["enabled"]; changedEnabled {
		if c.Enabled {
			if err := edrSvcStart(); err != nil {
				return false, err.Error(), ""
			}
		} else {
			if err := edrSvcStop(); err != nil {
				return false, err.Error(), ""
			}
		}
	}

	// Classify the changed top-level keys into the result arrays. The order is
	// deterministic — the well-known keys first, any other changed key after in
	// sorted order — so a repeated policy_set yields a stable payload. Every
	// array is a non-nil slice so the JSON always emits [], never null.
	ordered := []string{"allowlist", "blocklist", "sensors", "enabled"}
	extra := []string{}
	for key := range changed {
		if !slices.Contains(ordered, key) {
			extra = append(extra, key)
		}
	}
	sort.Strings(extra)
	ordered = append(ordered, extra...)

	appliedLive := []string{}
	requiresRestart := []string{}
	needsRestart := false
	var preCfg edrconfig.EDRConfig
	for _, key := range ordered {
		if !changed[key] {
			continue
		}
		// isLiveKey compares the effective pre/post state on the typed plane
		// so a policy that omits a nested field (e.g. only sets
		// kill_threshold) is not mistaken for a response_mode change.
		if key == "enabled" || isLiveKey(key, preCfg, c) {
			appliedLive = append(appliedLive, key)
		} else {
			requiresRestart = append(requiresRestart, key)
			needsRestart = true
		}
	}

	if needsRestart {
		active, _ := edrSvcActive()
		if active {
			utils.Logger.LogF(100, "EDR policy: restarting module for structural changes: %v", requiresRestart)
			if err := edrSvcRestart(); err != nil {
				return false, err.Error(), ""
			}
		} else {
			utils.Logger.LogF(100, "EDR policy: structural changes %v need a restart but the service is not active; skipping restart", requiresRestart)
		}
	}

	// Persist the new known-good fingerprint (the post-merge effective state;
	// the policy_version stamp is agent bookkeeping, not part of it) and the
	// drift trace of this policy_set.
	fpOut := make(map[string]json.RawMessage, len(doc))
	for k, v := range doc {
		if k == "policy_version" {
			continue
		}
		fpOut[k] = v
	}
	fpb, err := json.Marshal(fpOut)
	if err != nil {
		return false, err.Error(), ""
	}
	if err := os.WriteFile(edrPolicyFingerprintFile, fpb, 0o600); err != nil {
		return false, err.Error(), ""
	}
	if db, err := json.Marshal(drift); err == nil {
		if err := os.WriteFile(edrPolicyDriftFile, db, 0o600); err != nil {
			utils.Logger.LogF(100, "EDR policy: drift trace not persisted: %v", err)
		}
	}

	resp := map[string]any{
		"applied":          true,
		"version":          version,
		"applied_live":     appliedLive,
		"requires_restart": requiresRestart,
		"drift":            drift,
	}
	respB, err := json.Marshal(resp)
	if err != nil {
		return false, err.Error(), ""
	}
	return true, "", string(respB)
}

// isLiveKey reports whether a changed top-level policy key is hot-reloadable
// by the EDR service (service.go maybeReload) without a restart: the
// allowlists, the network blocklist, and the ransomware guard when its
// effective response_mode is the only thing that changed. Any other ransomware
// change (enabled, thresholds, canaries, ...) requires a restart. pre and post
// are the effective typed configs before and after the merge.
func isLiveKey(key string, pre, post edrconfig.EDRConfig) bool {
	switch key {
	case "allowlist", "blocklist":
		return true
	case "ransomware":
		return pre.Ransomware.ResponseMode != post.Ransomware.ResponseMode
	default:
		return false
	}
}

// changedTopLevelKeys returns the policy doc keys whose effective post-merge
// state differs from the pre-merge state, plus a dedicated entry for
// enabled (service control) when the effective boolean flipped.
func changedTopLevelKeys(pre map[string]json.RawMessage, post edrconfig.EDRConfig) map[string]bool {
	changed := map[string]bool{}
	var postDoc map[string]json.RawMessage
	if b, err := json.Marshal(post); err == nil {
		_ = json.Unmarshal(b, &postDoc)
	}
	for k, v := range postDoc {
		// policy_version is agent bookkeeping stamped on every apply — never
		// a "changed" policy key.
		if k == "policy_version" {
			continue
		}
		if pv, ok := pre[k]; !ok || !jsonBytesEqual(pv, v) {
			changed[k] = true
		}
	}
	if pv, ok := pre["enabled"]; ok {
		var preEnabled bool
		_ = json.Unmarshal(pv, &preEnabled)
		if preEnabled != post.Enabled {
			changed["enabled"] = true
		}
	}
	return changed
}

// edrPolicyDrift is one local hand-edit detected against the last known-good
// central policy: the key, what the local edr.json held, and the central
// value (from the new policy, or from the fingerprint when the policy does
// not specify the key).
type edrPolicyDrift struct {
	Key      string          `json:"key"`
	LocalWas json.RawMessage `json:"local_was"`
	Central  json.RawMessage `json:"central"`
}

// detectDrift compares the current (pre-merge) effective state against the
// last known-good central fingerprint. Any top-level key that differs is a
// local hand-edit. The `central` value reported is the new policy's value when
// the policy specifies the key, otherwise the fingerprint value (the last
// known central). policy_version is excluded (agent bookkeeping) and keys
// absent from the fingerprint are not drift (no prior central value).
func detectDrift(current, policy, fingerprint map[string]json.RawMessage) []edrPolicyDrift {
	if fingerprint == nil {
		return []edrPolicyDrift{}
	}
	drift := []edrPolicyDrift{}
	for k, cur := range current {
		if k == "policy_version" {
			continue
		}
		fp, ok := fingerprint[k]
		if !ok || jsonBytesEqual(fp, cur) {
			continue
		}
		central := fp
		if nv, inPolicy := policy[k]; inPolicy {
			central = nv
		}
		drift = append(drift, edrPolicyDrift{Key: k, LocalWas: cur, Central: central})
	}
	return drift
}

// jsonBytesEqual reports whether two raw JSON messages are semantically
// equal.
func jsonBytesEqual(a, b json.RawMessage) bool {
	if bytes.Equal(a, b) {
		return true
	}
	var av, bv any
	if json.Unmarshal(a, &av) != nil || json.Unmarshal(b, &bv) != nil {
		return false
	}
	return reflect.DeepEqual(av, bv)
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
