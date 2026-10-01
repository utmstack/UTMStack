//go:build windows

package ransomware

import "strings"

func t1490Scenario() (string, string, string, func(string) bool) {
	trust := func(image string) bool {
		p := strings.ToLower(strings.ReplaceAll(image, `\`, "/"))
		return strings.HasPrefix(p, "c:/program files/veeam/")
	}
	return `C:\Program Files\Veeam\veeam.exe`, `C:\Windows\System32\vssadmin.exe`,
		"vssadmin delete shadows /all /quiet", trust
}
