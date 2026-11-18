package signatures

import (
	"crypto/md5"
	"fmt"
	"strings"
	"testing"
)

func TestDynamicConfigLevelCheck(t *testing.T) {
	for line, want := range map[string]bool{
		"PE:0x0":         true, // no level: every level
		"PE:0x0:217:255": false,
		"PE:0x0:216:216": true,
		"PE:0x0:10":      true,
		"PE:0x0:10:20":   false,
		"PE:0x0::217":    false, // runs of ':' are one separator, so 217 is the minimum
		"PE:0x0:x":       false,
		"PE:0x0:10:x":    false,
		"PE:0x0:10:":     true, // an empty maximum is missing
		"PE:0x0: 10":     false,
	} {
		if got := dconfApplies(line, 216); got != want {
			t.Errorf("dconfApplies(%q) = %v", line, got)
		}
	}
	for value, want := range map[string]uint32{"0x1f": 31, "0x 1f:10:20": 31, "0x0x1f": 31, "0x-1": 0xffffffff, "0xffffffff": 0xffffffff, "0x0": 0} {
		if got, ok := scanDconfValue(value); !ok || got != want {
			t.Errorf("scanDconfValue(%q) = %x, %v", value, got, ok)
		}
	}
	for _, value := range []string{"0X1", "0x", "0xg", "1f", "0x123456789", " 0x1"} {
		if _, ok := scanDconfValue(value); ok {
			t.Errorf("scanDconfValue(%q) accepted", value)
		}
	}
}

func compileConfig(t *testing.T, cfg string, files map[string][]byte) (*Database, error) {
	t.Helper()
	all := map[string][]byte{"daily.cfg": []byte(cfg), "a.ndb": []byte("Ok:0:*:41424344")}
	for n, b := range files {
		all[n] = b
	}
	return Compile(all)
}

func TestDynamicConfigAppliesAndGatesFormats(t *testing.T) {
	pdb := map[string][]byte{"a.pdb": []byte("H:amazon.com\nH:bank.example\n")}
	// Comments, unknown modules and directives outside the level are ignored.
	d, err := compileConfig(t, "# comment\nPHISHING:0x0:235:255\nFOO:0x0\nDOCUMENT:0x5:11:13\n", pdb)
	if err != nil || d.DynamicConfig() != DefaultDynamicConfig() || d.phishing.count() != 2 {
		t.Fatalf("non-applying directives changed the configuration: %v %+v", err, d)
	}
	d, err = compileConfig(t, "PHISHING:0x2\n", pdb)
	if err != nil {
		t.Fatal(err)
	}
	if d.phishing.count() != 0 || d.Report().Skipped[SkipDynamicConfig] != 2 || d.DynamicConfig().Phishing != 2 {
		t.Fatalf("phishing engine off: %d rules, skipped %v", d.phishing.count(), d.Report().Skipped)
	}
	// The last applying line replaces the whole mask.
	d, err = compileConfig(t, "PHISHING:0x0\nPHISHING:0x3:1:255\n", pdb)
	if err != nil || d.phishing.count() != 2 {
		t.Fatalf("last directive must win: %v", err)
	}
	f := callbackWithLabel([]byte("UTMStack.Bytecode"), true)
	d, err = compileConfig(t, "BYTECODE:0x0\n", map[string][]byte{"a.cbc": f.encode()})
	if err != nil || d.Report().Skipped[SkipDynamicConfig] != 1 || d.Count() != 1 {
		t.Fatalf("bytecode off: %v %v", err, d.Report().Skipped)
	}
	if _, err = compileConfig(t, "BYTECODE:0x1\n", map[string][]byte{"a.cbc": f.encode()}); err != nil {
		t.Fatalf("interpreter-only bytecode: %v", err)
	}
	section := map[string][]byte{"a.mdb": []byte(fmt.Sprintf("512:%x:Section", md5.Sum([]byte("x"))))}
	d, err = compileConfig(t, fmt.Sprintf("PE:0x%x\n", DefaultDynamicConfig().PE&^dconfPEMD5Sect), section)
	if err != nil || d.sections.len() != 0 || d.Report().Skipped[SkipDynamicConfig] != 1 {
		t.Fatalf("section hashes off: %v %v", err, d.Report().Skipped)
	}
	d, err = compileConfig(t, fmt.Sprintf("PE:0x%x\n", DefaultDynamicConfig().PE&^dconfPECerts), map[string][]byte{"a.crb": []byte("x;1;2;3;4;5;6;7;8;9;10\n")})
	if err != nil || d.Report().Skipped[SkipDynamicConfig] != 1 {
		t.Fatalf("certificate rules off: %v", err)
	}
	if _, err = compileConfig(t, "STATS:0x3\nPE:0xfff7f\n", nil); err != nil {
		t.Fatalf("statistics switches and defaults are honored: %v", err)
	}
}

func TestDynamicConfigThisEngineCannotHonorFails(t *testing.T) {
	for cfg, want := range map[string]string{
		"DOCUMENT:0x7fb\n":   "DOCUMENT features 0x4 ",
		"PE:0x0\n":           "PE features 0xcff6f ",
		"BYTECODE:0x2\n":     "without the interpreter",
		"PE:zz\n":            "malformed dynamic configuration value for PE",
		"ARCHIVE:0x0\nx\n":   "ARCHIVE features",
		"PCRE:0x0:1:255\n":   "PCRE features",
		"MAIL:0x1\nMAIL:0x3": "",
	} {
		_, err := compileConfig(t, cfg, nil)
		switch {
		case want == "" && err != nil:
			t.Errorf("%q: %v", cfg, err)
		case want != "" && (err == nil || !strings.Contains(err.Error(), want)):
			t.Errorf("%q: got %v, want %q", cfg, err, want)
		case err != nil && strings.Contains(strings.ToLower(err.Error()), "clam"):
			t.Errorf("diagnostic leaks engine branding: %v", err)
		}
	}
}

func TestLoadOrderFollowsReference(t *testing.T) {
	names := []string{"a.ndb", "b.ign", "c.IGN2", "daily.cfg", "daily.cld", "daily.cvd", "local.gdb", "main.cvd", "x.crb"}
	want := "b.ign c.IGN2 daily.cld daily.cvd local.gdb daily.cfg x.crb a.ndb main.cvd"
	if got := strings.Join(loadOrder(names), " "); got != want {
		t.Fatalf("load order %q, want %q", got, want)
	}
	// daily loads before main, so its rules come first; members keep archive order.
	daily := cvdFixtureOrdered(t, []member{{"z.ndb", []byte("Daily.Z:0:*:41424344")}, {"a.ndb", []byte("Daily.A:0:*:41424344")}})
	main := cvdFixtureOrdered(t, []member{{"main.ndb", []byte("Main.Rule:0:*:41424344")}})
	d, err := Compile(map[string][]byte{"main.cvd": main, "daily.cvd": daily})
	if err != nil {
		t.Fatal(err)
	}
	// Three equal four-byte BM entries exhaust their two triplet choices.
	// The main entry then retains prefix0 and precedes the first daily entry.
	assertMatch(t, d, []byte("ABCD"), "Main.Rule")
	if files := d.Report().Files; files[0] != "daily.cvd/z.ndb" || files[2] != "main.cvd/main.ndb" {
		t.Fatalf("files in load order: %v", files)
	}
}
