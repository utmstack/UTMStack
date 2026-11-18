package signatures

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func htmlMatch(t *testing.T, d *Database, text, want string) {
	t.Helper()
	got, err := d.MatchHTML(context.Background(), []byte(text))
	if got != want || err != nil {
		t.Fatalf("MatchHTML=%q,%v, want %q; HTML=%s", got, err, want, text)
	}
}
func TestPhishingHostRulesAndBoundaries(t *testing.T) {
	d := compileTest(t, "pdb", "H:amazon.com\nH:amazon.co.uk")
	spoof := "Heuristics.Phishing.Email.SpoofedDomain"
	cases := []struct{ real, display, want string }{
		{"https://evil.test/", "https://www.amazon.com/", spoof},
		{"https://amazon.com.evil.test/", "www.amazon.com", spoof},
		{"https://login.amazon.com/", "www.amazon.com", ""},
		{"https://evil.co.uk/", "amazon.co.uk", spoof},
		{"https://evil.test/", "notamazon.com", ""},
		{"https://evil.test/", "amazon.com.evil.test", spoof},
		{"https://evil.test/", "www.amazon.com.evil.example", spoof},
		{"https://evil.test/", "amazon.com.evil.test/path", spoof},
		{"https://evil.test/", "amazon.co.uk.evil.test?x=1", spoof},
		{"https://evil.test/", "amazon.com.evil.net", ""},
		{"https://evil.test/", "https://amazon.com.evil.test/", ""},
		{"https://evil.test/path/amazon.com", "amazon.com", spoof},
		{"https://amazon.com@evil.test/", "amazon.com", spoof},
		{"https://AMAZON.COM./", "amazon.com", spoof},
		{"https://%61mazon.com/", "amazon.com", ""},
		{"http://amazon.com/", "https://amazon.com/", "Heuristics.Phishing.Email.SSL-Spoof"},
		{"https://amazon.com/", "http://amazon.com/", ""},
	}
	for _, c := range cases {
		t.Run(c.real+"-"+c.display, func(t *testing.T) { htmlMatch(t, d, fmt.Sprintf(`<a href="%s">%s</a>`, c.real, c.display), c.want) })
	}
	if d.DetectionCount() != 2 {
		t.Fatalf("phishing protection not in detection count: %d", d.DetectionCount())
	}
}
func TestPhishingAllowlistPrecedesProtection(t *testing.T) {
	d, e := Compile(map[string][]byte{"test.pdb": []byte("H:google.com"), "test.wdb": []byte("M:www.google.ro:www.google.com")})
	if e != nil {
		t.Fatal(e)
	}
	htmlMatch(t, d, `<a href="http://www.google.ro">www.google.com</a>`, "")
	htmlMatch(t, d, `<a href="http://www.google.ro.evil.test">www.google.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	htmlMatch(t, d, `<a href="http://evilgoogle.ro">www.google.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	htmlMatch(t, d, `<a href="http://www.google.ro">https://www.google.com</a>`, "")
	only := compileTest(t, "wdb", "M:www.google.ro:www.google.com")
	if only.DetectionCount() != 0 {
		t.Fatal("allowlist counted as malware protection")
	}
}
func TestOfficialPhishingRegexExamples(t *testing.T) {
	// Formats from https://docs.clamav.net/manual/Signatures/PhishSigs.html.
	// The fixture uses the published Amazon regex examples, not a private feed.
	d, e := Compile(map[string][]byte{
		"test.pdb": []byte(`R:.+\.amazon\.(com|co\.uk)([/?].*)?`),
		"test.wdb": []byte(`X:.+\.amazon\.(at|ca|co\.uk|co\.jp|de|fr)([/?].*)?:.+\.amazon\.com([/?].*)?:17-`),
	})
	if e != nil {
		t.Fatal(e)
	}
	htmlMatch(t, d, `<a href="https://www.amazon.de/checkout">https://www.amazon.com/login</a>`, "")
	htmlMatch(t, d, `<a href="https://evil.test/">https://www.amazon.com/login</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	htmlMatch(t, d, `<a href="https://evil.test/https://www.amazon.de/">https://www.amazon.com/login</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	htmlMatch(t, d, `<a href="https://www.amazon.de@evil.test/">https://www.amazon.com/login</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
}
func TestPhishingHTMLExtraction(t *testing.T) {
	d := compileTest(t, "pdb", "H:amazon.com")
	spoof := "Heuristics.Phishing.Email.SpoofedDomain"
	for _, html := range []string{
		`<a href="https://evil.test">ama<b>zon</b>&#46;com</a>`,
		`<a href="https://evil.test" title="amazon.com">Sign in</a>`,
		`<a href="https://evil.test"><img src="https://images.amazon.com/image.png"></a>`,
		`<form action="https://evil.test"><a href="https://amazon.com">Sign in</a></form>`,
		`<form action="https://evil.test"><img src="https://images.amazon.com/image.png"></form>`,
		`<a href="/login">amazon.com</a><base href="https://evil.test">`,
		`<a href="https://amazon.com">amazon.com<a href="https://evil.test">amazon.com</a>`,
		`<a href="https://evil.test"><iframe src="https://amazon.com"></iframe></a>`,
	} {
		htmlMatch(t, d, html, spoof)
	}
	for _, html := range []string{
		`<a href="/login">amazon.com</a><base href="https://amazon.com">`,
		`<a href="https://evil.test"><script>amazon.com</script>Click here</a>`,
		`<script>"<a href='https://evil.test'>amazon.com</a>"</script>`,
		`<a href="https://evil.test">amazon&amp;#46;com</a>`,
		`<a href="https://amazon.com">amazon.com</a>`,
		// HTML tokenization follows the browser rule: the first duplicate
		// attribute wins; the second destination is not the link target.
		`<a href="https://amazon.com" href="https://evil.test">amazon.com</a>`,
	} {
		htmlMatch(t, d, html, "")
	}
}
func TestPhishingIDNAAndPrivateDomainBoundaries(t *testing.T) {
	d := compileTest(t, "pdb", "H:xn--bcher-kva.example\nH:bank.github.io")
	htmlMatch(t, d, `<a href="https://bücher.example">xn--bcher-kva.example</a>`, "")
	htmlMatch(t, d, `<a href="https://attacker.github.io">bank.github.io</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
}
func TestPhishingIncompleteAndLimits(t *testing.T) {
	d := compileTest(t, "pdb", "H:amazon.com")
	for _, html := range []string{
		`<a href="/login">amazon.com</a>`,
		`<a href="https://2130706433/">amazon.com</a>`,
		`<a href="https://amazon.com%40evil.test/">amazon.com</a>`,
		`<a href="mailto:person@amazon.com">amazon.com</a>`,
		`<a href="https://evil.test">https://amazon.com/%zz</a>`,
		`<a href="https://evil.test">` + strings.Repeat("x", maxPhishingURL+1) + `</a>`,
	} {
		if sig, err := d.MatchHTML(context.Background(), []byte(html)); sig != "" || !errors.Is(err, ErrHTMLIncomplete) {
			t.Errorf("expected incomplete, got %s %v for %q", sig, err, html[:min(len(html), 200)])
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := d.MatchHTML(ctx, []byte(`<a href="https://evil.test">amazon.com</a>`)); e != context.Canceled {
		t.Fatal(e)
	}
}

func TestPhishingMalformedBaseOnlyAffectsRelativeLinks(t *testing.T) {
	d := compileTest(t, "pdb", "H:amazon.com")
	htmlMatch(t, d, `<base href="http://[broken"><p>No links</p>`, "")
	htmlMatch(t, d, `<base href="http://[broken"><a href="https://amazon.com/">amazon.com</a>`, "")
	htmlMatch(t, d, `<base href="http://[broken"><a href="https://evil.test/">amazon.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	if sig, err := d.MatchHTML(context.Background(), []byte(`<base href="http://[broken"><a href="/login">amazon.com</a>`)); sig != "" || !errors.Is(err, ErrHTMLIncomplete) {
		t.Fatalf("unresolved relative protected link: sig=%q err=%v", sig, err)
	}
}
func TestPhishingUnsupportedRulesFailGeneration(t *testing.T) {
	tests := map[string][]string{
		"pdb": {"H:amazon.com:0-20", "H:amazon.com:90-90", "H:amazon.com:999-", "H:evil/amazon.com", "H:*.amazon.com", `R:(?i)amazon\.com`, `R:(amazon)\1`, `R:(?=amazon).*`, `R:^amazon\.com$`, `R:https?://amazon\.com/checkout`, "Y:amazon.com"},
		"wdb": {"M:amazon.com", `X:amazon\.com`},
		"gdb": {"S3:F:" + strings.Repeat("0", 64), "S1:W:" + strings.Repeat("0", 64), "S:F:bad", "S:P:bad"},
	}
	for ext, lines := range tests {
		for _, line := range lines {
			if _, e := Compile(map[string][]byte{"test." + ext: []byte(line)}); e == nil {
				t.Errorf("accepted unsupported %s: %s", ext, line)
			}
		}
	}
	d := compileTest(t, "pdb", "H:amazon.com:20-")
	htmlMatch(t, d, `<a href="https://evil.test">amazon.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	// Allow rules the reference accepts but this engine cannot interpret
	// load as unknown when reached (unknown_test.go).
	for _, line := range []string{"M:amazon.com:evil/amazon.com", "M:amazon.com%zz:amazon.com"} {
		d, err := Compile(map[string][]byte{"test.wdb": []byte(line), "test.pdb": []byte("H:amazon.com")})
		if err != nil || d.Report().Unknown[UnknownPhishingAllowlist] != 1 {
			t.Errorf("%s: %v, unknown %v", line, err, d.Report().Unknown)
		}
	}
}

func TestPinnedPhishingFlagsAndYAlias(t *testing.T) {
	for _, line := range []string{"H102:amazon.com", "Hxyz:amazon.com"} {
		d := compileTest(t, "pdb", line)
		htmlMatch(t, d, `<a href="https://evil.test/">amazon.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	}
	for _, kind := range []string{"X", "Y", "Yabc"} {
		d, err := Compile(map[string][]byte{
			"test.pdb": []byte("H:amazon.com"),
			"test.wdb": []byte(kind + `:evil\.test.*:.*amazon\.com`),
		})
		if err != nil {
			t.Fatal(kind, err)
		}
		htmlMatch(t, d, `<a href="https://evil.test/">amazon.com</a>`, "")
		htmlMatch(t, d, `<a href="https://other.test/">amazon.com</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
	}
}

func TestOfficialWDBFullURLForms(t *testing.T) {
	const spoof = "Heuristics.Phishing.Email.SpoofedDomain"
	const sslSpoof = "Heuristics.Phishing.Email.SSL-Spoof"
	for _, tc := range []struct {
		name, protected, allow, real, display, want string
	}{
		{"scheme", "www.wellsfargo.com", `X:www\.wellsfargo\.com:https.?//www\.wellsfargo\.com`, "http://www.wellsfargo.com", "https://www.wellsfargo.com", ""},
		{"scheme-control", "www.wellsfargo.com", `X:www\.wellsfargo\.com:https.?//www\.wellsfargo\.com`, "http://other.example", "https://www.wellsfargo.com", sslSpoof},
		{"explicit-anchor", "agency.gov", `X:^.+\.gov([/?].*)?:.+\.gov([/?].*)?`, "https://abc.gov", "https://agency.gov", ""},
		{"explicit-anchor-control", "agency.gov", `X:^.+\.gov([/?].*)?:.+\.gov([/?].*)?`, "https://abc.gov.evil.test", "https://agency.gov", spoof},
		{"leading-alternative", "cic.fr", `X:(^|.+\.)ems6\.net([/?].*)?:.+\.cic\.fr([/?].*)?`, "https://mail.ems6.net", "https://www.cic.fr", ""},
		{"scheme-pair", "www.amazon.com", `X:(.+\.)?http.?//www\.amazon\.(at|be|ca|ch|co\.uk|de|es|fr|ie|in|it|nl|ph|pl|com)([/?].*)?:(.+\.)?https.?//www\.amazon\.(at|be|ca|ch|co\.uk|de|es|fr|ie|in|it|nl|ph|pl|com)([/?].*)?`, "http://www.amazon.de", "https://www.amazon.com", ""},
		{"optional-scheme", "paypal.com", `X:.+\.paypal-communication.com([/?].)?:(https.?//)?([a-zA-Z0-9-]+\.)?paypal\.(at|ca|co.uk|co.jp|com|de|fr|com.au|ph|vn|be|dk|fi|ie|it|lu|nl|no|pl|pt|ru|es|se|ch)([/?].*)?`, "https://news.paypal-communication.com", "https://www.paypal.com", ""},
		{"literal-path-after-cleanup", "rbcroyalbank.com", `X:(.+\.)?rbc.com([/?].*)?:(.+\.)?rbcroyalbank.com/([/?].*)?`, "https://rbc.com", "https://rbcroyalbank.com", spoof},
		{"display-userinfo", "example.com", `X:[^@]+:.+@.+(a|e|g|h|k|l|m|n|o|r|s|t|u|v|y|z)`, "https://evil.test", "https://user@example.com", ""},
		{"display-userinfo-control", "example.com", `X:[^@]+:.+@.+(a|e|g|h|k|l|m|n|o|r|s|t|u|v|y|z)`, "https://evil.test", "https://example.com", spoof},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, err := Compile(map[string][]byte{"test.pdb": []byte("H:" + tc.protected), "test.wdb": []byte(tc.allow)})
			if err != nil || d.Report().Unknown[UnknownPhishingAllowlist] != 0 {
				t.Fatalf("compile: %v; unknown %v", err, d.Report().Unknown)
			}
			htmlMatch(t, d, fmt.Sprintf(`<a href="%s">%s</a>`, tc.real, tc.display), tc.want)
		})
	}
}

func TestOfficialWDBStaticEncodedAndRootDotHosts(t *testing.T) {
	const spoof = "Heuristics.Phishing.Email.SpoofedDomain"
	for _, tc := range []struct {
		name, protected, allow, real, display string
	}{
		{"terminal-percent", "www.paypal.com", "M:www.paypal.com%a0:www.paypal.com", "https://www.paypal.com%a0", "https://www.paypal.com"},
		{"root-dot", "www.standardbank.co.za", "M:www.standardbank.co.za.:www.standardbank.co.za", "https://www.standardbank.co.za.", "https://www.standardbank.co.za"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			protected := []byte("H:" + tc.protected)
			data := fmt.Sprintf(`<a href="%s">%s</a>`, tc.real, tc.display)
			without := compileTest(t, "pdb", string(protected))
			htmlMatch(t, without, data, spoof)
			with, err := Compile(map[string][]byte{"test.pdb": protected, "test.wdb": []byte(tc.allow)})
			if err != nil || with.Report().Unknown[UnknownPhishingAllowlist] != 0 {
				t.Fatalf("compile: %v; unknown %v", err, with.Report().Unknown)
			}
			htmlMatch(t, with, data, "")
		})
	}
}

func TestStaticWDBBackslashPatternCannotMatchCleanedURL(t *testing.T) {
	d, err := Compile(map[string][]byte{
		"test.pdb": []byte("H:deutsche-bank.de"),
		"test.wdb": []byte(`M:\.t\.co:\.deutsche-bank\.de`),
	})
	if err != nil || d.Report().Unknown[UnknownPhishingAllowlist] != 0 {
		t.Fatalf("compile: %v; unknown %v", err, d.Report().Unknown)
	}
	htmlMatch(t, d, `<a href="https://t.co">https://deutsche-bank.de</a>`, "Heuristics.Phishing.Email.SpoofedDomain")
}
func FuzzPhishingHTML(f *testing.F) {
	d, e := Compile(map[string][]byte{"test.pdb": []byte("H:amazon.com")})
	if e != nil {
		f.Fatal(e)
	}
	for _, s := range []string{`<a href="https://evil.test">amazon.com</a>`, `<base href="https://amazon.com"><a href="/">amazon.com</a>`, "<a><script>foo", ""} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 1<<20 {
			t.Skip()
		}
		_, _ = d.MatchHTML(context.Background(), []byte(s))
	})
}
