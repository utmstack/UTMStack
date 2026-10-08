package ransomware

import "testing"

func TestIsSensitiveKey(t *testing.T) {
	cases := []struct {
		key  string
		want bool
	}{
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\SystemRestore`, true},
		{`HKLM\SYSTEM\CurrentControlSet\Services\swprv`, false},
		{`HKLM\SYSTEM\CurrentControlSet\Control\Shadow Copies`, true},
		{`HKCU\Software\SomeApp`, false},
		{`HKLM\SYSTEM\Recovery`, true},
		{`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options`, false},
		// The Kernel-Registry provider emits the leaf name for the acted-upon
		// key, not always the full path: a bare leaf must still be matched.
		{`SystemRestore`, true},
		{`Recovery`, true},
		{`shadowstorage61`, true},
		{`winlogon`, false},
	}
	for _, c := range cases {
		if got := isSensitiveKey(c.key); got != c.want {
			t.Errorf("isSensitiveKey(%q) = %v, want %v", c.key, got, c.want)
		}
	}
}

func TestRegistryFeedNonNil(t *testing.T) {
	f := NewRegistryFeed()
	if f == nil {
		t.Fatal("NewRegistryFeed returned nil")
	}
}
