package signatures

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func gdbHash(s string) string {
	digest := sha256.Sum256([]byte(s))
	return hex.EncodeToString(digest[:])
}

func TestGDBFullHashesAndURLVariants(t *testing.T) {
	const malware = "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"
	const phish = "Heuristics.Safebrowsing.Suspected-phishing_safebrowsing.clamav.net"
	const blocked = "Heuristics.Phishing.URL.Blocked"
	for _, tc := range []struct{ name, rule, hashInput, url, want string }{
		{"full", "S:F:", "evil.example/visit", "https://evil.example/visit", malware},
		{"queryless", "S2:F:", "evil.example/visit", "http://evil.example/visit?track=1", phish},
		{"host-suffix", "S1:F:", "evil.example/", "https://a.b.evil.example/other", blocked},
		{"long-full-host", "S1:F:", "a.b.c.d.e.f.example/", "https://a.b.c.d.e.f.example/", blocked},
		{"path-prefix", "S:F:", "evil.example/one/", "https://evil.example/one/two/three", malware},
		{"repeated-slash", "S:F:", "evil.example//a", "https://evil.example//a", malware},
		{"single-dot", "S:F:", "evil.example/a/b", "https://evil.example/a/./b", malware},
		{"parent-dot", "S:F:", "evil.example/b", "https://evil.example/a/../b", malware},
		{"chained-dot-preserved", "S:F:", "evil.example/b/./c", "https://evil.example/a/../b/./c", malware},
		{"chained-parent-preserved", "S:F:", "evil.example/a/../c", "https://evil.example/a/b/../../c", malware},
		{"malformed-percent", "S:F:", "evil.example/a%25zz", "https://evil.example/a%zz", malware},
		{"deep-percent", "S:F:", "evil.example/%25foo", "https://evil.example/%252525252525252525foo", malware},
		{"encoded-null", "S:F:", "evil.example/a%01b", "https://evil.example/a%00b", malware},
		{"literal-path-space", "S:F:", "evil.example/a%20b", "https://evil.example/a b", malware},
		{"unicode-host-bytes", "S:F:", "%c3%bc.example/", "https://ü.example/", malware},
		{"fourth-path-prefix", "S:F:", "evil.example/one/two/three/four/", "https://evil.example/one/two/three/four/five", malware},
		{"userinfo", "S:F:", "evil.example/", "https://user@evil.example/", malware},
		{"userinfo-with-port-colon", "S:F:", "user/pass@evil.example/", "https://user:pass@evil.example/", malware},
		{"mailto-protocol", "S:F:", "evil.example/", "mailto://evil.example/", malware},
		{"port", "S:F:", "evil.example/8443/a", "https://evil.example:8443/a", malware},
		{"bare-query", "S:F:", "evil.example/bare=1", "https://evil.example?bare=1", malware},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Compile(map[string][]byte{"local.gdb": []byte(tc.rule + gdbHash(tc.hashInput))})
			if err != nil {
				t.Fatal(err)
			}
			if d.DetectionCount() != 1 {
				t.Fatalf("detection count %d", d.DetectionCount())
			}
			htmlMatch(t, d, fmt.Sprintf(`<a href="%s">Click here</a>`, tc.url), tc.want)
			htmlMatch(t, d, fmt.Sprintf(`<a href="%s"></a>`, tc.url), tc.want)
			htmlMatch(t, d, `<a href="https://safe.example/">Click here</a>`, "")
		})
	}
}

func TestGDBAllowOverridesAndLevel(t *testing.T) {
	hash := gdbHash("evil.example/")
	d, err := Compile(map[string][]byte{
		"local.gdb": []byte("S:W:" + hash + "\n"),
		"other.gdb": []byte("S:F:" + hash + "\nS2:F:" + gdbHash("phish.example/") + ":999-\n"),
	})
	if err != nil || d.DetectionCount() != 0 || d.Report().Skipped[SkipLevel] != 1 {
		t.Fatalf("expected allow-only database after level skip: database=%v error=%v", d, err)
	}
	d, err = Compile(map[string][]byte{
		"local.gdb": []byte("S:W:" + hash + "\n"),
		"other.gdb": []byte("S:F:" + hash + "\nS2:F:" + gdbHash("phish.example/") + "\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if d.DetectionCount() != 1 {
		t.Fatalf("allow override count %d", d.DetectionCount())
	}
	htmlMatch(t, d, `<a href="https://evil.example/">Click</a>`, "")
	htmlMatch(t, d, `<a href="https://phish.example/">Click</a>`, "Heuristics.Safebrowsing.Suspected-phishing_safebrowsing.clamav.net")
	if d.Report().Skipped[SkipLevel] != 0 {
		t.Fatalf("unexpected skip %v", d.Report().Skipped)
	}
}

func TestGDBAmbiguousURLIsIncomplete(t *testing.T) {
	d, err := Compile(map[string][]byte{"local.gdb": []byte("S:F:" + gdbHash("evil.example/"))})
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{"https://evil.example/" + strings.Repeat("a", safeBrowsingURLLimit)} {
		html := fmt.Sprintf(`<a href="%s">Click</a>`, u)
		if sig, err := d.MatchHTML(context.Background(), []byte(html)); sig != "" || !errors.Is(err, ErrHTMLIncomplete) {
			t.Errorf("%s: %q, %v", u, sig, err)
		}
	}
}

func TestGDBPrefixDoesNotGateFullHash(t *testing.T) {
	hash := gdbHash("evil.example/")
	d, err := Compile(map[string][]byte{"local.gdb": []byte(strings.Join([]string{
		"S:P:00000000", "S:F:" + hash,
	}, "\n"))})
	if err != nil {
		t.Fatal(err)
	}
	htmlMatch(t, d, `<a href="https://evil.example/">Click</a>`, "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net")
}

func TestGDBPathPrefixLimitMatchesPinnedMatcher(t *testing.T) {
	d, err := Compile(map[string][]byte{"local.gdb": []byte("S:F:" + gdbHash("evil.example/one/two/three/four/five/"))})
	if err != nil {
		t.Fatal(err)
	}
	htmlMatch(t, d, `<a href="https://evil.example/one/two/three/four/five/six">Click</a>`, "")
	consecutive, err := Compile(map[string][]byte{"local.gdb": []byte("S:F:" + gdbHash("evil.example//"))})
	if err != nil {
		t.Fatal(err)
	}
	htmlMatch(t, consecutive, `<a href="https://evil.example//a">Click</a>`, "")
	query, err := Compile(map[string][]byte{"local.gdb": []byte("S:F:" + gdbHash("evil.example/one/two/three/four/"))})
	if err != nil {
		t.Fatal(err)
	}
	htmlMatch(t, query, `<a href="https://evil.example/one/two/three/four/five?x=1">Click</a>`, "")
}

func TestGDBURLCanonicalBytes(t *testing.T) {
	for _, tc := range []struct{ input, host, path string }{
		{"https://EVIL.Example/a%252Fb", "evil.example", "a/b"},
		{"https://evil.example/a#frag", "evil.example", "a%23frag"},
		{"https://evil.example/a%20b", "evil.example", "a%20b"},
		{"https://evil.example/a%zz", "evil.example", "a%25zz"},
		{"https://evil.example/%252525252525252525foo", "evil.example", "%25foo"},
		{"https://%65vil.example/", "evil.example", ""},
		{"https://%252565vil.example/", "%252565vil.example", ""},
		{"https://evil.example/a%00b", "evil.example", "a%01b"},
		{"https://ü.example/a", "%c3%bc.example", "a"},
		{"https://evil.example/a/./b", "evil.example", "a/b"},
		{"https://evil.example/a/../b", "evil.example", "b"},
		{"https://evil.example/a/..", "evil.example", ""},
		{"https://evil.example/a/%2E%2E/b", "evil.example", "b"},
		{"https://evil.example/a/../b/./c", "evil.example", "b/./c"},
		{"https://evil.example/a/b/../../c", "evil.example", "a/../c"},
		{"https://evil.example/a/./b/../c", "evil.example", "a/c"},
		{"https://evil.example/a/../b/../c", "evil.example", "b/../c"},
		{"https://evil.example/a/./b/./c", "evil.example", "a/b/c"},
		{"https://evil.example/a/b/../../../c", "evil.example", "a/c"},
		{"https://evil.example/a/../", "evil.example", ""},
		{"https://evil.example/a/b/../../", "evil.example", "a/../"},
		{"https://evil.example/./a", "evil.example", "a"},
		{"https://evil.example/a/.", "evil.example", "a/."},
		{"https://user@evil.example/a?b=1", "evil.example", "a?b=1"},
		{"https://user:pass@evil.example/", "user", "pass@evil.example/"},
		{"https://evil.example:8443/a", "evil.example", "8443/a"},
		{"https://evil.example?bare=1", "evil.example", "bare=1"},
	} {
		host, path, eligible, err := canonicalSafeBrowsingURL(tc.input)
		if err != nil || !eligible || host != tc.host || path != tc.path {
			t.Errorf("%s => %q %q %t %v", tc.input, host, path, eligible, err)
		}
	}
}

func TestGDBInvalidLeadingEscapeIsIncomplete(t *testing.T) {
	_, _, eligible, err := canonicalSafeBrowsingURL("https://%zz.example/")
	if !eligible || !errors.Is(err, ErrHTMLIncomplete) {
		t.Fatalf("invalid leading escape: eligible=%t err=%v", eligible, err)
	}
}

func TestGDBHonorsPhishingDynamicConfig(t *testing.T) {
	d, err := compileConfig(t, "PHISHING:0x2\n", map[string][]byte{
		"local.gdb": []byte("S:F:" + gdbHash("evil.example/")),
	})
	if err != nil || d.DetectionCount() != 1 || d.Report().Skipped[SkipDynamicConfig] != 1 {
		t.Fatalf("GDB dconf gate: database=%v error=%v", d, err)
	}
	htmlMatch(t, d, `<a href="https://evil.example/">Click</a>`, "")
}

func TestGDBDoesNotFoldProtocolCase(t *testing.T) {
	d, err := Compile(map[string][]byte{"local.gdb": []byte("S:F:" + gdbHash("evil.example/"))})
	if err != nil {
		t.Fatal(err)
	}
	htmlMatch(t, d, `<a href="HTTPS://evil.example/">Click</a>`, "")
}
