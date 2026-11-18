package signatures

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The pinned ClamAV phishing URL scanner processes links in MIME email
// content. A standalone HTML file is not a valid oracle for GDB rules.
func TestGDBPinnedClamAVMIMEReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	clamscan := filepath.Join(root, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("clamscan is not the pinned 1.5.4 binary: %v", err)
	}
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	for _, tc := range []struct {
		name, rule, hashInput, url, want string
	}{
		{"full-path", "S:F:", "evil.example/visit", "https://evil.example/visit", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"queryless", "S2:F:", "evil.example/visit", "http://evil.example/visit?track=1", "Heuristics.Safebrowsing.Suspected-phishing_safebrowsing.clamav.net"},
		{"host-suffix", "S1:F:", "evil.example/", "https://a.b.evil.example/other", "Heuristics.Phishing.URL.Blocked"},
		{"path-prefix", "S:F:", "evil.example/one/", "https://evil.example/one/two/three", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"repeated-slash", "S:F:", "evil.example//a", "https://evil.example//a", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"single-dot", "S:F:", "evil.example/a/b", "https://evil.example/a/./b", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"parent-dot", "S:F:", "evil.example/b", "https://evil.example/a/../b", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"malformed-percent", "S:F:", "evil.example/a%25zz", "https://evil.example/a%zz", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"deep-percent", "S:F:", "evil.example/%25foo", "https://evil.example/%252525252525252525foo", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"userinfo", "S:F:", "evil.example/", "https://user@evil.example/", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"port", "S:F:", "evil.example/8443/a", "https://evil.example:8443/a", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"bare-query", "S:F:", "evil.example/bare=1", "https://evil.example?bare=1", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
		{"unrelated-host", "S:F:", "evil.example/visit", "https://safe.example/visit", ""},
		{"unrelated-path", "S:F:", "evil.example/visit", "https://evil.example/other", ""},
		{"fifth-path-prefix", "S:F:", "evil.example/one/two/three/four/five/", "https://evil.example/one/two/three/four/five/six", ""},
		{"uppercase-scheme", "S:F:", "evil.example/", "HTTPS://evil.example/", ""},
		{"query-stripped", "S:F:", "evil.example/visit", "http://evil.example/visit?track=1", "Heuristics.Safebrowsing.Suspected-malware_safebrowsing.clamav.net"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			rule := tc.rule + gdbHash(tc.hashInput) + "\n"
			if err := os.WriteFile(filepath.Join(dir, "local.gdb"), []byte(rule), 0o600); err != nil {
				t.Fatal(err)
			}
			html := fmt.Sprintf(`<html><a href="%s">Click here</a></html>`, tc.url)
			goDB, err := Compile(map[string][]byte{"local.gdb": []byte(rule)})
			if err != nil {
				t.Fatal(err)
			}
			if name, err := goDB.MatchHTML(context.Background(), []byte(html)); err != nil || name != tc.want {
				t.Fatalf("Go GDB: %q, %v; want %q", name, err, tc.want)
			}
			mail := "From: sender@example.org\nTo: receiver@example.org\nSubject: inert\nMIME-Version: 1.0\nContent-Type: text/html; charset=UTF-8\n\n" + html + "\n"
			path := filepath.Join(dir, "probe.eml")
			if err := os.WriteFile(path, []byte(mail), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(clamscan, "--database="+dir, "--no-summary", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err != nil || !strings.Contains(string(out), path+": OK") {
					t.Fatalf("ClamAV GDB clean: %v, %s", err, out)
				}
			} else {
				status, ok := err.(*exec.ExitError)
				if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), path+": "+tc.want+" FOUND") {
					t.Fatalf("ClamAV GDB: %v, %s", err, out)
				}
			}
		})
	}
}

func TestPDBPinnedClamAVMIMEReference(t *testing.T) {
	root := os.Getenv("EDR_CLAMAV_PORTABLE")
	if root == "" {
		t.Skip("set EDR_CLAMAV_PORTABLE to the pinned ClamAV 1.5.4 installation")
	}
	clamscan := filepath.Join(root, "bin", "clamscan")
	tool, err := os.ReadFile(clamscan)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(tool)) != "9b87a5d884ee22587df61bf00b61c7f276525c3734b5b3db550b3a11fa26efb7" {
		t.Fatalf("clamscan is not the pinned 1.5.4 binary: %v", err)
	}
	env := append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(root, "lib"),
		"CVD_CERTS_DIR="+filepath.Join(root, "etc", "certs"))
	const spoof = "Heuristics.Phishing.Email.SpoofedDomain"
	const ssl = "Heuristics.Phishing.Email.SSL-Spoof"
	const rules = "H:amazon.com\nH:amazon.co.uk\n"
	goDB, err := Compile(map[string][]byte{"local.pdb": []byte(rules)})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "local.pdb"), []byte(rules), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, real, display, want string }{
		{"spoofed-url", "https://evil.test/", "https://www.amazon.com/", spoof},
		{"spoofed-suffix", "https://amazon.com.evil.test/", "www.amazon.com", spoof},
		{"same-registrable", "https://login.amazon.com/", "www.amazon.com", ""},
		{"uk-spoof", "https://evil.co.uk/", "amazon.co.uk", spoof},
		{"not-protected", "https://evil.test/", "notamazon.com", ""},
		{"unknown-tail-tld", "https://evil.test/", "amazon.com.evil.test", spoof},
		{"unknown-tail-path", "https://evil.test/", "amazon.com.evil.test/path", spoof},
		{"unknown-tail-query", "https://evil.test/", "amazon.co.uk.evil.test?x=1", spoof},
		{"unknown-tail-subdomain", "https://evil.test/", "www.amazon.com.evil.test", spoof},
		{"known-tail-tld", "https://evil.test/", "amazon.com.evil.net", ""},
		{"explicit-display-url", "https://evil.test/", "https://amazon.com.evil.test/", ""},
		{"path-bait", "https://evil.test/path/amazon.com", "amazon.com", spoof},
		{"userinfo", "https://amazon.com@evil.test/", "amazon.com", spoof},
		{"root-dot", "https://AMAZON.COM./", "amazon.com", spoof},
		{"encoded-protected", "https://%61mazon.com/", "amazon.com", ""},
		{"ssl-downgrade", "http://amazon.com/", "https://amazon.com/", ssl},
		{"ssl-upgrade", "https://amazon.com/", "http://amazon.com/", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			html := fmt.Sprintf(`<html><a href="%s">%s</a></html>`, tc.real, tc.display)
			if name, err := goDB.MatchHTML(context.Background(), []byte(html)); err != nil || name != tc.want {
				t.Fatalf("Go PDB: %q, %v; want %q", name, err, tc.want)
			}
			mail := "From: sender@example.org\nTo: receiver@example.org\nSubject: inert\nMIME-Version: 1.0\nContent-Type: text/html; charset=UTF-8\n\n" + html + "\n"
			path := filepath.Join(dir, tc.name+".eml")
			if err := os.WriteFile(path, []byte(mail), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(clamscan, "--database="+dir, "--no-summary", "--alert-phishing-ssl=yes", path)
			cmd.Env = env
			out, err := cmd.CombinedOutput()
			if tc.want == "" {
				if err != nil || !strings.Contains(string(out), path+": OK") {
					t.Fatalf("ClamAV PDB clean: %v, %s", err, out)
				}
			} else {
				status, ok := err.(*exec.ExitError)
				if !ok || status.ExitCode() != 1 || !strings.Contains(string(out), path+": "+tc.want+" FOUND") {
					t.Fatalf("ClamAV PDB: %v, %s", err, out)
				}
			}
		})
	}
}
