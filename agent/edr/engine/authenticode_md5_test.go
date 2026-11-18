package engine

import (
	"bytes"
	"context"
	"crypto"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestMD5ImageDigestRequiresValidSignerAndExactImage(t *testing.T) {
	// The pinned reference uses mktime for certificate UTC fields. Start
	// before the local-time offset so this fixture is valid on the test host.
	file, crb, der := signedPEFixtureWithImageHashAndValidity(t, true, crypto.MD5, 48*time.Hour, 30*24*time.Hour)
	db, err := signatures.Compile(map[string][]byte{
		"trust.crb":  []byte(crb),
		"marker.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	verify := func(data []byte) (bool, error) {
		return db.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
			return peAuthenticodeDigest(data, h)
		}, time.Now().UTC())
	}
	if trusted, err := verify(file); !trusted || err != nil {
		t.Fatalf("valid MD5-image PE not trusted: trusted=%v err=%v", trusted, err)
	}
	changed := bytes.Clone(file)
	changed[0x200] ^= 1
	if trusted, err := verify(changed); trusted || err == nil {
		t.Fatalf("changed image remained trusted: trusted=%v err=%v", trusted, err)
	}
	badSignature := bytes.Clone(der)
	badSignature[len(badSignature)-10] ^= 1
	if trusted, err := db.VerifyTrustedAuthenticode(badSignature, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(file, h)
	}, time.Now().UTC()); trusted || err != nil {
		t.Fatalf("invalid PKCS#7 signer trusted: trusted=%v err=%v", trusted, err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	if clean, sig, err := e.Scan(context.Background(), file); !clean || sig != "" || err != nil {
		t.Fatalf("valid PE verdict: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), changed); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("changed PE verdict: clean=%v sig=%q err=%v", clean, sig, err)
	}
	blocked := strings.Replace(crb, "Trusted.Signer;1;", "Revoked.Signer;0;", 1)
	blockAndTrust, err := signatures.Compile(map[string][]byte{
		"trust.crb":  []byte(crb + blocked),
		"marker.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e.current.Store(&generation{db: blockAndTrust})
	if clean, sig, err := e.Scan(context.Background(), file); clean || sig != "Revoked.Signer" || err != nil {
		t.Fatalf("CRB blocklist lost precedence: clean=%v sig=%q err=%v", clean, sig, err)
	}
	withoutTrust, err := signatures.Compile(map[string][]byte{
		"marker.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e.current.Store(&generation{db: withoutTrust})
	if clean, sig, err := e.Scan(context.Background(), file); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("PE without CRB trust was exempted: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if ref := os.Getenv("EDR_CLAMAV_PORTABLE"); ref != "" {
		t.Run("pinned-clamav", func(t *testing.T) {
			root := t.TempDir()
			for name, content := range map[string][]byte{
				"trust.crb": []byte(crb), "marker.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
				"valid.exe": file, "changed.exe": changed,
			} {
				if err := os.WriteFile(filepath.Join(root, name), content, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(filepath.Join(ref, "bin", "clamscan"), "--debug", "--database="+root,
				"--no-summary", filepath.Join(root, "valid.exe"), filepath.Join(root, "changed.exe"))
			cmd.Env = append(os.Environ(), "DYLD_LIBRARY_PATH="+filepath.Join(ref, "lib"),
				"CVD_CERTS_DIR="+filepath.Join(ref, "etc", "certs"))
			out, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok {
				t.Fatalf("reference exit: %v output=%s", err, out)
			}
			validDigest, err := peAuthenticodeDigest(file, crypto.MD5)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), filepath.Join(root, "valid.exe")+": OK") ||
				!strings.Contains(string(out), filepath.Join(root, "changed.exe")+": Suspicious.Code.UNOFFICIAL FOUND") ||
				!strings.Contains(string(out), "Authenticode: "+hex.EncodeToString(validDigest)) {
				var details []string
				for _, line := range strings.Split(string(out), "\n") {
					if strings.Contains(line, "asn1_") || strings.Contains(line, "Authenticode") || strings.Contains(line, "cert") || strings.Contains(line, "digest") || strings.Contains(line, "valid.exe:") || strings.Contains(line, "changed.exe:") || strings.Contains(line, "i: ") || strings.Contains(line, "not_after") {
						details = append(details, line)
					}
				}
				t.Fatalf("reference verdict drift: %s", strings.Join(details, "\n"))
			}
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if strings.HasSuffix(line, "valid.exe: OK") || strings.HasSuffix(line, "changed.exe: Suspicious.Code.UNOFFICIAL FOUND") || strings.Contains(line, "Authenticode:") {
					t.Log(line)
				}
			}
		})
	}
}
