package ransomware

import (
	"strings"
)

// sensitiveKeyPatterns are substrings that indicate a registry key associated
// with ransomware behaviour (recovery, VSS/shadow copies, system restore).
// The Kernel-Registry provider emits the key name for the acted-upon object,
// which is typically a leaf (e.g. "SystemRestore"), so the patterns are
// matched case-insensitively WITHOUT a leading separator: they still hit in
// full paths ("...\\SystemRestore") and in leaf names alike.
var sensitiveKeyPatterns = []string{
	"systemrestore",
	"shadow",
	"recovery",
	"restoresettings",
}

// isSensitiveKey reports whether the registry key path or leaf contains a
// pattern associated with ransomware behavior.
func isSensitiveKey(key string) bool {
	low := strings.ToLower(key)
	for _, p := range sensitiveKeyPatterns {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}
