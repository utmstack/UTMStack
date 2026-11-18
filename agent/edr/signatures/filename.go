package signatures

import (
	"errors"
	"strings"
)

// ValidateFileName rejects path traversal, Windows device names and names whose
// spelling changes on Windows. Bundles are identical across endpoint platforms.
func ValidateFileName(name string) error {
	if name == "" || len(name) > 128 || name[0] == '.' || strings.HasSuffix(name, ".") {
		return errors.New("invalid signature artifact filename")
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
			return errors.New("invalid signature artifact filename")
		}
	}
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	switch stem {
	case "CON", "PRN", "AUX", "NUL":
		return errors.New("reserved signature artifact filename")
	}
	if len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) && stem[3] >= '1' && stem[3] <= '9' {
		return errors.New("reserved signature artifact filename")
	}
	return nil
}
