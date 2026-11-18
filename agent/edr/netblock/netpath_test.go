package netblock

import "testing"

func TestNormalizeDevicePathPassthrough(t *testing.T) {
	// Non-device paths and already-drive-letter paths are returned unchanged.
	for _, p := range []string{`C:\Windows\x.exe`, ``, `System`, `\\?\C:\y`} {
		if got := NormalizeDevicePath(p); got != stripQ(p) {
			t.Errorf("NormalizeDevicePath(%q) = %q", p, got)
		}
	}
}
