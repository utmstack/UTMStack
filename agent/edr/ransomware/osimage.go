package ransomware

import (
	"os"
	"path"
	"strings"
)

// windowsDir is the normalized OS directory (e.g. "c:/windows"). A package var
// so tests on a non-Windows host see the usual default.
var windowsDir = func() string {
	d := os.Getenv("SystemRoot")
	if d == "" {
		d = os.Getenv("windir")
	}
	if d == "" {
		d = `C:\Windows`
	}
	return normWinPath(d)
}()

// normWinPath lower-cases a Windows path, uses forward slashes, drops a
// `\\?\` / `\??\` prefix and resolves "." / ".." so prefix checks are exact.
func normWinPath(p string) string {
	p = strings.ReplaceAll(strings.ToLower(p), `\`, "/")
	for _, pre := range []string{"//?/", "/??/"} {
		p = strings.TrimPrefix(p, pre)
	}
	if p == "" {
		return ""
	}
	return strings.TrimRight(path.Clean(p), "/")
}

// isWindowsDirImage reports whether an image lives under the Windows directory.
//
// Ransomware containment never quarantines such an image. Those are shared OS
// binaries (powershell.exe, cmd.exe, rundll32.exe, …): moving one breaks Windows
// for every user, Windows Resource Protection denies it anyway ("Access is
// denied"), and when one really tampers the harmful part is the script or
// command line it was given, not the file. So a signed system binary that
// tampers with a decoy is still suspended/killed per response_mode (that is what
// stops the encryption), the incident is recorded as contained with no
// quarantine ID, and the event carries its command line for the analyst.
// Limitation: a payload an attacker stages in a writable folder under the
// Windows directory (e.g. C:\Windows\Temp) is killed but left on disk too.
func isWindowsDirImage(image string) bool {
	p := normWinPath(image)
	return p != "" && (p == windowsDir || strings.HasPrefix(p, windowsDir+"/"))
}
