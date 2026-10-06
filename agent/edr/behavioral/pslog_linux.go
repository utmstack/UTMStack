//go:build linux

package behavioral

// interpreterNames is the basename set of shells and interpreters whose
// execve we surface as shell_activity telemetry (the Linux counterpart of
// the PowerShell 4104 script-block records). A process whose image basename
// matches one of these emits shell_activity in addition to process_create.
//
// Deliberately conservative: only interactive/invoked interpreters. The
// shell builtins (echo, ls, ...) are child processes of these, so catching
// the interpreter's cmdline at execve covers what was "asked to run".
var interpreterNames = map[string]bool{
	// shells
	"bash": true, "sh": true, "dash": true, "zsh": true, "ksh": true,
	"ash": true, "busybox": true,
	// interpreted languages
	"python": true, "python2": true, "python3": true,
	"perl": true, "ruby": true, "php": true,
	"node": true, "deno": true, "bun": true,
	"lua": true, "tclsh": true, "wish": true,
	"expect": true, "pwsh": true, "powershell": true,
}

// IsInterpreter reports whether the image path refers to a known shell or
// interpreter, returning its basename alongside. A non-interpreter image
// returns ("", false) — the name is only meaningful on a match.
func IsInterpreter(image string) (string, bool) {
	name := basename(image)
	if !interpreterNames[name] {
		return "", false
	}
	return name, true
}

// basename is filepath.Base without importing path/filepath (keeps this file
// import-light and lets the test target the exact string used). It handles
// both / and \ separators for robustness against config-supplied images.
func basename(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}
