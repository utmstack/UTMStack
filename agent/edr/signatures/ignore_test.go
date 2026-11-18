package signatures

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestIgnoreListAppliesToEveryFileAndContainer(t *testing.T) {
	daily := cvdFixtureOrdered(t, []member{
		{"daily.ign2", []byte("# reviewed\nMain.Ignored\nLate.Rule\n")},
		{"daily.ndb", []byte("Daily.Rule:0:*:41424344\n")},
	})
	main := cvdFixtureOrdered(t, []member{{"main.ndb", []byte("Main.Ignored:0:*:45464748\nMain.Kept:0:*:47484950\n")}})
	// A loose file sorting before every container is still covered.
	d, err := Compile(map[string][]byte{"daily.cvd": daily, "main.cvd": main, "a.ndb": []byte("Late.Rule:4:*:zz\n")})
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Report().Skipped[SkipIgnoreList]; got != 2 || d.Count() != 2 {
		t.Fatalf("ignored %d, loaded %d", got, d.Count())
	}
	assertMatch(t, d, []byte("EFGH"), "")
	assertMatch(t, d, []byte("GHIP"), "Main.Kept")
	// An ignore list inside a later container applies to earlier ones too.
	late := cvdFixtureOrdered(t, []member{{"zz.ign", []byte("Main.Kept\n")}})
	d, err = Compile(map[string][]byte{"main.cvd": main, "zz.cvd": late})
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, d, []byte("GHIP"), "")
}

func TestIgnoreListFormatsAndNames(t *testing.T) {
	const md5Hex = "44d88612fea8a8f36de82e1278abb02f"
	line := "Hashed:0:*:41424344"
	f := callbackWithLabel([]byte("UTMStack.Bytecode"), true)
	f.logical = "Bytecode.Trigger.{A,B};Engine:51-255,Target:0;0;4142"
	files := map[string][]byte{
		// MD5 entries come last: a later name-only line would reuse their MD5.
		"list.ign2": []byte("Hash.Rule\nAllow.Rule\nLogic.Rule\nLegacy.Rule\nBytecode.Trigger\nUnofficial.Rule\nab\n" +
			"db:12:Old.Form\nWrong.case\n" + fmt.Sprintf("Hashed:%x\n", md5.Sum([]byte(line))) + "Other:" + md5Hex + "\n"),
		"a.hdb":   []byte(md5Hex + ":68:Hash.Rule\n"),
		"a.fp":    []byte(md5Hex + ":68:Allow.Rule\n"),
		"a.ldb":   []byte("Logic.Rule;Engine:51-255,Target:0;0;zz\n"),
		"a.db":    []byte("Legacy.Rule=zz\nOld.Form=zz\n"),
		"a.cbc":   f.encode(),
		"b.ndb":   []byte("Unofficial.Rule.UNOFFICIAL:0:*:4142\nab:0:*:4142\n" + line + "\nOther:0:*:41424344\nWrong.Case:0:*:5a5a5a5a\n"),
		"pdb.pdb": []byte("H:hash.rule.example\n"),
	}
	d, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Report().Skipped[SkipIgnoreList]; got != 9 {
		t.Fatalf("ignored %d rules, want 9", got)
	}
	// An MD5 entry only matches its exact line; names are case-sensitive.
	assertMatch(t, d, []byte("ABCD"), "Other")
	assertMatch(t, d, []byte("ZZZZ"), "Wrong.Case")
	if d.phishing.count() != 1 {
		t.Fatal("ignore lists must not affect phishing rules")
	}
}

func TestSignatureNameNormalization(t *testing.T) {
	for name, want := range map[string]string{
		"A.{B}": "A  ", "Win.X-1.{A,B}": "Win.X-1", "Win.X-1.UNOFFICIAL": "Win.X-1", "ab": "ab ", "": "   ",
		"X.UNOFFICIAL": "X  ", "a}": "a} ", "Name}": "Name}", ".UNOFFICIAL": ".UNOFFICIAL", "Win.X-1": "Win.X-1",
	} {
		if got := normalizeSignatureName(name); got != want {
			t.Errorf("normalizeSignatureName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestIgnoreListMalformed(t *testing.T) {
	for _, list := range []string{"a:b:c:d\n", "Name:zz\n", "Name:" + "0123456789abcdef0123456789abcde\n", "One\n\nTwo\n", ":\n"} {
		if _, err := Compile(map[string][]byte{"a.ndb": []byte("Ok:0:*:41424344"), "x.ign2": []byte(list)}); err == nil {
			t.Errorf("accepted ignore list %q", list)
		}
	}
}

// The reference reuses its line buffer and MD5 pointer, so a name-only line
// after a "name:md5" line inherits whatever the buffer holds at that position.
func TestIgnoreListReusedMD5Pointer(t *testing.T) {
	const kept = "Bcde:0:*:41424344"
	list := fmt.Sprintf("Abcd:%x\nBcde\n", md5.Sum([]byte("unrelated")))
	// Container members are copied without their newline: Bcde inherits the
	// MD5 of the line before and, since it does not match, stays loaded.
	d, err := Compile(map[string][]byte{"x.cvd": cvdFixtureOrdered(t, []member{{"x.ign2", []byte(list)}, {"x.ndb", []byte(kept)}})})
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, d, []byte("ABCD"), "Bcde")
	// A loose file keeps the newline, which leaves an empty MD5 there.
	if _, err := Compile(map[string][]byte{"x.ign2": []byte(list), "x.ndb": []byte(kept)}); err == nil {
		t.Fatal("the loose list must fail like the reference")
	}
	// Inheriting a matching MD5 does ignore the rule.
	list = fmt.Sprintf("Abcd:%x\nBcde\n", md5.Sum([]byte(kept)))
	d, err = Compile(map[string][]byte{"x.cvd": cvdFixtureOrdered(t, []member{{"x.ign2", []byte(list)}, {"x.ndb", []byte(kept + "\nC:0:*:45464748")}})})
	if err != nil {
		t.Fatal(err)
	}
	assertMatch(t, d, []byte("ABCD"), "")
	// Padding a short name overwrites the start of its own MD5.
	if _, err := Compile(map[string][]byte{"x.ign2": []byte(fmt.Sprintf("A:%x\n", md5.Sum([]byte(kept)))), "x.ndb": []byte(kept)}); err == nil {
		t.Fatal("a padded name with an MD5 must fail like the reference")
	}
}
