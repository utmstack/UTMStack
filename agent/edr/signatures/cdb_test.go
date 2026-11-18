package signatures

import (
	"context"
	"errors"
	"testing"
)

func memberMatch(t *testing.T, d *Database, archive FileType, size int64, m MemberMeta) string {
	t.Helper()
	name, err := d.MatchMember(context.Background(), archive, size, m)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

func TestContainerPOSIXRegexEscapesAndConditional(t *testing.T) {
	// The bundled ClamAV ERE parser interprets escaped ordinary bytes as
	// literals, including digits, n and b. Go interprets these differently.
	d := compileTest(t, "cdb", "Digit:CL_TYPE_ZIP:*:^(ab)\\1$:*:*:*:*:*:\n")
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "ab1", HasName: true, Index: 1}); got != "Digit" {
		t.Fatalf("escaped digit: %q", got)
	}
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "abab", HasName: true, Index: 1}); got != "" {
		t.Fatalf("escaped digit treated as a backreference: %q", got)
	}
	d = compileTest(t, "cdb", "LiteralN:CL_TYPE_ZIP:*:^a\\nb$:*:*:*:*:*:\n")
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "anb", HasName: true, Index: 1}); got != "LiteralN" {
		t.Fatalf("escaped n: %q", got)
	}
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "a\nb", HasName: true, Index: 1}); got != "" {
		t.Fatalf("escaped n treated as a newline: %q", got)
	}
	d = compileTest(t, "cdb", "LiteralB:CL_TYPE_ZIP:*:^a\\bb$:*:*:*:*:*:\n")
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "abb", HasName: true, Index: 1}); got != "LiteralB" {
		t.Fatalf("escaped b: %q", got)
	}
	// This bracket shape occurs in the signed daily.cdb feed. Both slash
	// bytes are literal class members in the bundled ClamAV ERE parser.
	d = compileTest(t, "cdb", `Bracket:CL_TYPE_ZIP:*:^[^/\\]+\.js$:*:*:*:*:*:`+"\n")
	for _, tc := range []struct {
		name, want string
	}{
		{"report.js", "Bracket"},
		{"dir/report.js", ""},
		{"dir\\report.js", ""},
	} {
		if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: tc.name, HasName: true, Index: 1}); got != tc.want {
			t.Fatalf("bracket filename %q: %q, want %q", tc.name, got, tc.want)
		}
	}

	// Directional word boundaries differ from Go's symmetric boundary.
	d = compileTest(t, "cdb", "Word:CL_TYPE_ZIP:*:\\<report\\>:*:*:*:*:*:\n")
	if got := d.Report().Unknown[UnknownContainerRegex]; got != 1 {
		t.Fatalf("conditional container rules = %d, want 1", got)
	}
	for _, tc := range []struct {
		archive FileType
		member  MemberMeta
		unknown bool
	}{
		{TypeZIP, MemberMeta{Name: "report", HasName: true, Index: 1}, true},
		{TypeRAR, MemberMeta{Name: "report", HasName: true, Index: 1}, false},
		{TypeZIP, MemberMeta{Index: 1}, false},
	} {
		name, err := d.MatchMember(context.Background(), tc.archive, 100, tc.member)
		if name != "" || errors.Is(err, ErrRuleIncomplete) != tc.unknown {
			t.Fatalf("archive %d member %+v: %q, %v", tc.archive, tc.member, name, err)
		}
	}
	// A definite detection still wins when a higher-priority conditional
	// expression was encountered first.
	d = compileTest(t, "cdb", "Known:CL_TYPE_ZIP:*:^report$:*:*:*:*:*:\nWord:CL_TYPE_ZIP:*:\\<report\\>:*:*:*:*:*:\n")
	if got := memberMatch(t, d, TypeZIP, 100, MemberMeta{Name: "report", HasName: true, Index: 1}); got != "Known" {
		t.Fatalf("definite detection after unknown: %q", got)
	}
	// A second syntax error must still reject the feed.
	if _, err := Compile(map[string][]byte{"bad.cdb": []byte("Bad:CL_TYPE_ZIP:*:((ab)\\1:*:*:*:*:*:\n")}); err == nil {
		t.Fatal("malformed POSIX regex compiled")
	}
}

func TestContainerMetadataRules(t *testing.T) {
	d := compileTest(t, "cdb", `Zip.Invoice:CL_TYPE_ZIP:*:invoice\.exe$:40:100:1:1:deadbeef:
Any.Big:*:*:*:*:1000-0:*:*:*:
Zip.Named:CL_TYPE_ZIP:*:(?i)^report:*:*:*:2-3:*:
Rar.Exact:CL_TYPE_RAR:*:*:0:*:0:*:*:
`)
	member := MemberMeta{Name: "docs/invoice.exe", HasName: true, CompressedSize: 40, Size: 100, Encrypted: true, Index: 1}
	if got := memberMatch(t, d, TypeZIP, 500, member); got != "Zip.Invoice" {
		t.Fatalf("encrypted member: %q (the CRC field is never read)", got)
	}
	if got := memberMatch(t, d, TypeOOXMLWord, 500, member); got != "" {
		t.Fatalf("the container type is compared exactly: %q", got)
	}
	unencrypted := member
	unencrypted.Encrypted = false
	if got := memberMatch(t, d, TypeZIP, 500, unencrypted); got != "" {
		t.Fatalf("encryption flag: %q", got)
	}
	if got := memberMatch(t, d, TypeRAR, 1, MemberMeta{CompressedSize: 10, Size: 5000, Index: 1}); got != "Any.Big" {
		t.Fatalf("open-ended range: %q", got)
	}
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Name: "REPORT.txt", HasName: true, Index: 2}); got != "Zip.Named" {
		t.Fatalf("case-insensitive anchored name: %q", got)
	}
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Name: "old REPORT.txt", HasName: true, Index: 2}); got != "" {
		t.Fatalf("anchor: %q", got)
	}
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Index: 2}); got != "" {
		t.Fatalf("a regex never matches a member without a name: %q", got)
	}
	if got := memberMatch(t, d, TypeRAR, 1, MemberMeta{CompressedSize: 0, Encrypted: false, Index: 7}); got != "Rar.Exact" {
		t.Fatalf("exact zero: %q", got)
	}
	if got := memberMatch(t, d, TypeRAR, 1, MemberMeta{CompressedSize: 3, Index: 7}); got != "" {
		t.Fatalf("exact zero must not match 3: %q", got)
	}
	if d.DetectionCount() < 4 {
		t.Fatalf("container rules not counted: %d", d.DetectionCount())
	}
}

func TestContainerMetadataRulesLaterLineFirst(t *testing.T) {
	d := compileTest(t, "cdb", "First:*:*:*:*:*:*:*:*:\nSecond:*:*:*:*:*:*:*:*:\n")
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Index: 1}); got != "Second" {
		t.Fatalf("got %q, want the later line", got)
	}
}

func TestContainerMetadataRuleErrorsAndWindows(t *testing.T) {
	for _, line := range []string{
		"Bad:CL_TYPE_ZIP:*:*:*:*:*:*:*",              // nine fields
		"Bad:CL_TYPE_BOGUS:*:*:*:*:*:*:*:",           // unknown container type
		"Bad:CL_TYPE_ZIP:*:(unclosed:*:*:*:*:*:",     // bad regex
		"Bad:CL_TYPE_ZIP:*:*:*:*:2:*:*:",             // encryption flag
		"Bad:CL_TYPE_ZIP:*:*:x-1:*:*:*:*:",           // range
		"Bad:CL_TYPE_ZIP:*:*:*:*:*:*:*::abc",         // level
		"Bad:CL_TYPE_ZIP:*:*:*:*:*:*:*::1:2:3:extra", // fourteen fields
	} {
		if _, err := Compile(map[string][]byte{"a.cdb": []byte(line + "\n")}); err == nil {
			t.Errorf("%q compiled", line)
		}
	}
	d, err := Compile(map[string][]byte{"a.cdb": []byte("Future:*:*:*:*:*:*:*:*::235\nNow:*:*:*:*:*:*:*:*::51:\n"), "a.ndb": []byte("M:0:*:41424344\n")})
	if err != nil {
		t.Fatal(err)
	}
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Index: 1}); got != "" {
		t.Fatalf("an empty maximum level skips the rule: %q", got)
	}
	d = compileTest(t, "cdb", "Future:*:*:*:*:*:*:*:*::235\nNow:*:*:*:*:*:*:*:*::51:255\n")
	if got := memberMatch(t, d, TypeZIP, 1, MemberMeta{Index: 1}); got != "Now" {
		t.Fatalf("window: %q", got)
	}
}
