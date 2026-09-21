package ransomware

import "testing"

func TestMatchT1490_Positives(t *testing.T) {
	cases := []struct {
		img, cmd, want string
	}{
		{`C:\Windows\System32\vssadmin.exe`, `vssadmin  delete   shadows /all /quiet`, "vssadmin_delete_shadows"},
		{`C:\Windows\System32\wbem\WMIC.exe`, `wmic shadowcopy delete`, "wmic_shadowcopy_delete"},
		{`C:\Windows\System32\wbadmin.exe`, `wbadmin delete catalog -quiet`, "wbadmin_delete_catalog"},
		{`C:\Windows\System32\bcdedit.exe`, `bcdedit /set {default} recoveryenabled no`, "bcdedit_recovery_disable"},
		{`C:\Windows\System32\reagentc.exe`, `reagentc /disable`, "reagentc_disable"},
	}
	for _, c := range cases {
		name, ok := MatchT1490(c.img, c.cmd, nil)
		if !ok || name != c.want {
			t.Errorf("MatchT1490(%q) = %q,%v ; want %q,true", c.cmd, name, ok, c.want)
		}
	}
}

func TestMatchT1490_NegativesAndAllowlist(t *testing.T) {
	if _, ok := MatchT1490(`C:\Windows\System32\vssadmin.exe`, `vssadmin list shadows`, nil); ok {
		t.Error("listing shadows is benign")
	}
	if _, ok := MatchT1490(`C:\app\backup.exe`, `backup.exe --run`, nil); ok {
		t.Error("unrelated command matched")
	}
	// Allowlisted exact command is exempt (e.g. a sanctioned backup job).
	if _, ok := MatchT1490(`C:\Windows\System32\wbadmin.exe`, `wbadmin delete catalog -quiet`,
		[]string{"wbadmin delete catalog -quiet"}); ok {
		t.Error("allowlisted command should not match")
	}
}

func TestMatchT1490_SubstringAllowlist(t *testing.T) {
	// A distinctive token from sanctioned tooling suppresses the rule without
	// reproducing the exact command line.
	if _, ok := MatchT1490(`C:\Program Files\Veeam\backup.exe`,
		`vssadmin delete shadows /for=c: /oldest`, []string{"/oldest"}); ok {
		t.Error("command containing an allowlisted token should be exempt")
	}
	// A non-matching allowlist entry must NOT suppress a real T1490 command.
	if name, ok := MatchT1490(`C:\x\enc.exe`, `vssadmin delete shadows /all /quiet`,
		[]string{"acronis", "veeam"}); !ok || name != "vssadmin_delete_shadows" {
		t.Errorf("unrelated allowlist must not exempt; got %q,%v", name, ok)
	}
	// Empty allowlist entries are ignored (don't match everything).
	if _, ok := MatchT1490(`C:\x\enc.exe`, `vssadmin delete shadows`, []string{""}); !ok {
		t.Error("empty allowlist entry must not exempt every command")
	}
}
