package ransomware

import "strings"

// t1490Rule matches a recovery/backup-destruction command (MITRE T1490). All
// tokens must be present (in order-independent substring form) in the
// whitespace-normalized, lower-cased command line.
type t1490Rule struct {
	name   string
	tokens []string
}

// t1490Rules is defined per platform: rules_windows.go (vssadmin/wbadmin/
// bcdedit/...), rules_linux.go (backup-tree/snapshot/journal tampering), and
// rules_other.go (empty — no recovery-tampering sensor off those platforms).

func normCmd(s string) string { return strings.Join(strings.Fields(strings.ToLower(s)), " ") }

// MatchT1490 returns the first matching recovery-tampering rule name for the
// given process image + command line, or ("",false). A command line whose
// normalized form CONTAINS any (normalized) allowlist entry is exempt — so a
// distinctive token from sanctioned admin/backup tooling (e.g. a product name
// or a specific switch) suppresses the rule without having to reproduce the
// exact command line. Allowlist entries should be distinctive to stay safe.
func MatchT1490(image, cmdline string, allowlist []string) (string, bool) {
	nc := normCmd(cmdline)
	for _, a := range allowlist {
		if na := normCmd(a); na != "" && strings.Contains(nc, na) {
			return "", false
		}
	}
	for _, r := range t1490Rules {
		all := true
		for _, tok := range r.tokens {
			if !strings.Contains(nc, tok) {
				all = false
				break
			}
		}
		if all {
			return r.name, true
		}
	}
	return "", false
}
