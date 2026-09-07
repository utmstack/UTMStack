//go:build windows
// +build windows

package dependency

import (
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/agent/utils"
	sharedExec "github.com/utmstack/UTMStack/shared/exec"
	"golang.org/x/sys/windows/registry"
)

// AuditPolicyVersion is the UTMStack Windows audit policy version.
// Bump this when the settings below change, to trigger reapplying them on
// already-installed agents (see dependency.Reconcile).
const AuditPolicyVersion = "1.1.0"

// disabledByDefaultChannels are Windows Event Log channels the agent
// subscribes to (see collector/platform/windows_amd64.go, added for A10 in
// agent/GAPS_AND_IMPROVEMENTS.md) that exist but log nothing until
// explicitly enabled — this is a per-channel on/off switch, not an audit
// policy subcategory. Verified against Microsoft/DFIR documentation, not
// assumed: of everything this agent collects, these two are the ones that
// ship disabled.
var disabledByDefaultChannels = []string{
	"Microsoft-Windows-TaskScheduler/Operational",
	"Microsoft-Windows-PrintService/Operational",
}

// auditSubcategories are the Advanced Audit Policy subcategories needed to
// get useful events out of the Windows Event Log channels this agent
// already collects (see collector/platform/windows_amd64.go's Security
// channel). Windows ships most of these disabled, or in a "basic" legacy
// mode that predates Advanced Audit Policy, by default — without this, an
// event like 4688 (process creation) may never fire at all on a stock
// install, and the agent would forward an empty Security channel while
// looking like it's silently dropping events it never actually saw.
//
// Identified by GUID rather than by display name because auditpol's name
// matching is localized to the OS's display language (verified against
// Microsoft's official Subcategory/SubcategoryGUID table, [MS-GPAC]
// 2.2.3): a name like "Process Creation" would not match on a Spanish- or
// other non-English-language Windows install.
//
// Deliberately excluded: Object Access subcategories (file/registry/kernel
// object auditing) — those also need a SACL configured on each specific
// object, which is a per-host, per-path decision this agent can't safely
// make on its own the way it can for a global policy switch.
var auditSubcategories = map[string]string{
	"{0CCE922B-69AE-11D9-BED3-505054503030}": "Process Creation",
	"{0CCE922C-69AE-11D9-BED3-505054503030}": "Process Termination",
	"{0CCE9215-69AE-11D9-BED3-505054503030}": "Logon",
	"{0CCE9216-69AE-11D9-BED3-505054503030}": "Logoff",
	"{0CCE9217-69AE-11D9-BED3-505054503030}": "Account Lockout",
	"{0CCE921B-69AE-11D9-BED3-505054503030}": "Special Logon",
	"{0CCE921C-69AE-11D9-BED3-505054503030}": "Other Logon/Logoff Events",
	"{0CCE9237-69AE-11D9-BED3-505054503030}": "Security Group Management",
	"{0CCE9235-69AE-11D9-BED3-505054503030}": "User Account Management",
	"{0CCE922F-69AE-11D9-BED3-505054503030}": "Audit Policy Change",
	"{0CCE9228-69AE-11D9-BED3-505054503030}": "Sensitive Privilege Use",
}

// configureWindowsAuditPolicy enables the audit subcategories above and the
// registry settings that make their events actually useful (command line
// in process creation events, PowerShell script block/module logging). It
// only enables things — it never disables or narrows a subcategory an
// admin (or a domain GPO) may have already configured more broadly than
// this, so it's safe to layer on top of an existing policy.
//
// Note for domain-joined hosts: if a Group Policy Object also manages one
// of these subcategories, its next refresh (commonly every 90-120 minutes)
// can overwrite what we set here. That's expected — GPO is the higher
// authority — and harmless: it only means the admin's own policy is now in
// effect, which is never weaker than what we asked for here.
func configureWindowsAuditPolicy() error {
	var errs []string

	for guid, name := range auditSubcategories {
		if err := sharedExec.Run("auditpol", "", "/set", "/subcategory:"+guid, "/success:enable", "/failure:enable"); err != nil {
			errs = append(errs, fmt.Sprintf("%s (%s): %v", name, guid, err))
		}
	}

	if err := setProcessCreationCommandLineAuditing(); err != nil {
		errs = append(errs, fmt.Sprintf("command line auditing: %v", err))
	}

	if err := setPowerShellLogging(); err != nil {
		errs = append(errs, fmt.Sprintf("PowerShell logging: %v", err))
	}

	for _, channel := range disabledByDefaultChannels {
		if err := sharedExec.Run("wevtutil", "", "sl", channel, "/e:true"); err != nil {
			errs = append(errs, fmt.Sprintf("enable channel %s: %v", channel, err))
		}
	}

	if len(errs) > 0 {
		utils.Logger.ErrorF("auditpolicy: some settings could not be applied (continuing with the rest): %s", strings.Join(errs, "; "))
	}

	// Non-fatal: a subcategory a domain GPO already governs, or a
	// restricted environment that refuses one of these writes, still
	// leaves everything else applied and shouldn't block agent startup.
	return nil
}

// setProcessCreationCommandLineAuditing makes 4688 (process creation)
// events include the actual command line — without this, 4688 fires but
// carries no argument information, which is where most of its detection
// value is.
func setProcessCreationCommandLineAuditing() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Policies\System\Audit`,
		registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetDWordValue("ProcessCreationIncludeCmdLine_Enabled", 1)
}

// setPowerShellLogging enables Script Block Logging (event 4104, the
// actual PowerShell code executed, including obfuscated/decoded stages)
// and Module Logging (event 4103) — the two Windows PowerShell channels
// this agent already collects are close to empty without these.
func setPowerShellLogging() error {
	sbKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Policies\Microsoft\Windows\PowerShell\ScriptBlockLogging`,
		registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("script block logging: %w", err)
	}
	defer sbKey.Close()
	if err := sbKey.SetDWordValue("EnableScriptBlockLogging", 1); err != nil {
		return fmt.Errorf("script block logging: %w", err)
	}

	modKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Policies\Microsoft\Windows\PowerShell\ModuleLogging`,
		registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("module logging: %w", err)
	}
	defer modKey.Close()
	if err := modKey.SetDWordValue("EnableModuleLogging", 1); err != nil {
		return fmt.Errorf("module logging: %w", err)
	}

	namesKey, _, err := registry.CreateKey(modKey, "ModuleNames", registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("module logging names: %w", err)
	}
	defer namesKey.Close()
	// "*" logs every module; matches the standard "Turn on Module Logging"
	// GPO artifact layout (a ModuleNames subkey with one value per pattern).
	return namesKey.SetStringValue("*", "*")
}
