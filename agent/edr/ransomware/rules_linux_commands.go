package ransomware

import (
	"path"
	"strings"
)

// These are exact direct-tool verbs, not shell substring matches. They flag
// destructive recovery operations, not proof of malicious intent; sanctioned
// administration needs the existing command/process allowlists or alert mode.
func matchLinuxRecovery(image, command string, allowlist []string) (string, bool) {
	if len(command) > 64<<10 {
		return "", false
	}
	normalized := strings.Join(strings.Fields(command), " ")
	for _, allow := range allowlist {
		a := strings.Join(strings.Fields(allow), " ")
		if a != "" && strings.Contains(normalized, a) {
			return "", false
		}
	}
	args := strings.Fields(command)
	if len(args) < 2 || len(args) > 256 {
		return "", false
	}
	tool := path.Base(strings.TrimSuffix(image, " (deleted)"))
	if path.Base(args[0]) != tool {
		return "", false
	}
	for _, arg := range args {
		if arg == "--help" || arg == "--dry-run" || arg == "--test" {
			return "", false
		}
	}
	switch tool {
	case "zfs":
		if args[1] != "destroy" || len(args) < 3 {
			return "", false
		}
		for _, arg := range args[2:] {
			if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && strings.Contains(arg, "n") {
				return "", false
			}
		}
		return "linux_zfs_destroy", true
	case "btrfs":
		if len(args) >= 4 && args[1] == "subvolume" && args[2] == "delete" {
			return "linux_btrfs_subvolume_delete", true
		}
	case "systemctl":
		action := ""
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "-") {
				continue
			}
			if action == "" {
				action = arg
				continue
			}
			if action != "stop" && action != "mask" && action != "disable" {
				return "", false
			}
			unit := strings.TrimSuffix(arg, ".service")
			switch unit {
			case "borgmatic", "restic", "veeamservice", "urbackupclientbackend", "bacula-fd", "bareos-fd":
				return "linux_backup_service_" + action, true
			}
		}
	}
	return "", false
}
