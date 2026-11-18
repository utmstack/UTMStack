package signatures

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCertificateRuleIdentityAndPolicy(t *testing.T) {
	subject := strings.Repeat("01", 20)
	serial := strings.Repeat("02", 20)
	d, err := Compile(map[string][]byte{
		"rules.crb": []byte(strings.Join([]string{
			"Revoked.Exact;0;" + subject + ";" + serial + ";00ABCD;ignored;1;0;0;;comment",
			"Revoked.Wildcard;0;" + subject + ";;ABCD;ignored;0;0;0;;comment",
			"Trusted.Anchor;1;" + subject + ";" + serial + ";ABCD;ignored;1;0;1;10;comment",
		}, "\n") + "\n"),
		"rule.ndb": []byte("Harmless:0:*:41424344\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.certRules) != 3 || d.Report().Unknown[UnknownCertificateTrust] != 1 {
		t.Fatalf("certificate rules not retained: %d, report %v", len(d.certRules), d.Report().Unknown)
	}
	subjectHash, _ := crbSHA1(subject)
	serialHash, _ := crbSHA1(serial)
	cert := certificateIdentity{subject: subjectHash, serial: serialHash, modulus: []byte{0xab, 0xcd},
		codeSign: true, certSign: true, notBefore: 12, notAfter: 20}
	if name, ok := d.blockedCertificate(cert); !ok || name != "Revoked.Wildcard" {
		t.Fatalf("blocklist precedence: %q, %v", name, ok)
	}
	if !d.trustedCertificate(cert) {
		t.Fatal("matching trust anchor rejected")
	}
	cert.serial[0]++
	if name, ok := d.blockedCertificate(cert); !ok || name != "Revoked.Wildcard" {
		t.Fatalf("wildcard serial: %q, %v", name, ok)
	}
	if !d.trustedCertificate(cert) {
		t.Fatal("legacy source-local serial wildcard was not retained")
	}
	cert.serial = serialHash
	cert.modulus[0]++
	if _, ok := d.blockedCertificate(cert); ok || d.trustedCertificate(cert) {
		t.Fatal("wrong RSA modulus matched certificate policy")
	}
	cert.modulus[0]--
	cert.notBefore = 9
	if d.trustedCertificate(cert) {
		t.Fatal("trust anchor accepted earlier validity interval")
	}
	cert.notBefore = 12
	cert.timeSign = true
	if d.trustedCertificate(cert) {
		t.Fatal("trust anchor accepted unpermitted timestamp signing")
	}
}

func TestCertificateRuleRejectsMalformedIdentity(t *testing.T) {
	subject := strings.Repeat("01", 20)
	serial := strings.Repeat("02", 20)
	for _, line := range []string{
		"Bad;2;" + subject + ";" + serial + ";ABCD;0;1;0;0;;comment",
		"Bad;0;01;" + serial + ";ABCD;0;1;0;0;;comment",
		"Bad;0;" + subject + ";02;ABCD;0;1;0;0;;comment",
		"Bad;0;" + subject + ";" + serial + ";not-hex;0;1;0;0;;comment",
		"Bad;0;" + subject + ";" + serial + ";ABCD;0;2;0;0;;comment",
	} {
		if _, err := Compile(map[string][]byte{"rules.crb": []byte(line + "\n"), "rule.ndb": []byte("Harmless:0:*:41424344\n")}); err == nil {
			t.Errorf("malformed CRB accepted: %s", line)
		}
	}
}

func TestCertificateSourceStateAndSelection(t *testing.T) {
	subject, serial := strings.Repeat("01", 20), strings.Repeat("02", 20)
	row := func(name, trust, number, before, suffix string) string {
		return name + ";" + trust + ";" + subject + ";" + number + ";ABCD;ignored;1;1;1;" + before + ";comment" + suffix + "\n"
	}
	d, err := Compile(map[string][]byte{
		"a.crb":    []byte(row("Wildcard", "1", "", "50", "") + row("Inherited.Block", "0", serial, "", "") + row("Inherited.Trust", "1", serial, "", "")),
		"b.crb":    []byte(strings.ReplaceAll(row("New.Source", "0", serial, "", ""), ";ABCD;", ";ABCF;")),
		"c.crb":    []byte(strings.ReplaceAll(row("Future.Wildcard", "1", "", "99", ";999999")+row("After.Skipped", "0", serial, "", ""), ";ABCD;", ";ABEF;")),
		"rule.ndb": []byte("Harmless:0:*:41424344\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	byName := make(map[string]certificateRule)
	for _, r := range d.certRules {
		byName[r.name] = r
	}
	if len(byName) != 4 {
		t.Fatal("selected certificate count", len(byName))
	}
	for _, name := range []string{"Wildcard", "Inherited.Block"} {
		r := byName[name]
		if !r.ignoreSerial || r.notBefore != 50 {
			t.Fatal("source state not inherited", name, r.ignoreSerial, r.notBefore)
		}
	}
	if _, ok := byName["Inherited.Trust"]; ok {
		t.Fatal("duplicate trusted record replaced existing policy")
	}
	for _, name := range []string{"New.Source", "After.Skipped"} {
		r := byName[name]
		if r.ignoreSerial || r.notBefore != 0 {
			t.Fatal("state escaped source/selection boundary", name, r.ignoreSerial, r.notBefore)
		}
	}
	// Streaming archive members enter the same source-scoped reader; reuse a
	// database across calls to prove that the reader, not global state, owns it.
	stream := &Database{limits: DefaultLimits()}
	failed := func(_ string, _ int, e error) { t.Fatal(e) }
	stream.compileTextSource(source{name: "first.crb", ext: ".crb"}, strings.NewReader(row("Stream.Wildcard", "0", "", "50", "")), failed)
	stream.compileTextSource(source{name: "second.crb", ext: ".crb"}, strings.NewReader(strings.ReplaceAll(row("Stream.Exact", "0", serial, "", ""), ";ABCD;", ";ABCF;")), failed)
	if len(stream.certRules) != 2 || !stream.certRules[0].ignoreSerial || stream.certRules[1].ignoreSerial || stream.certRules[1].notBefore != 0 {
		t.Fatal("streamed source state leaked")
	}
}

func TestPinnedOfficialCertificateRuleRetention(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("set EDR_OFFICIAL_SIGNATURE_DIR to the pinned three-CVD directory")
	}
	data, err := os.ReadFile(filepath.Join(dir, "daily.cvd"))
	if err != nil {
		t.Fatal(err)
	}
	members, _, err := unpackContainer("daily.cvd", data,
		&archiveBudget{limits: DefaultLimits(), total: int64(len(data)), files: 1})
	if err != nil {
		t.Fatal(err)
	}
	var crb []byte
	for _, m := range members {
		if m.name == "daily.crb" {
			crb = m.data
			break
		}
	}
	if len(crb) == 0 {
		t.Fatal("pinned daily CVD has no CRB member")
	}
	d, err := Compile(map[string][]byte{
		"daily.crb": crb,
		"rule.ndb":  []byte("Harmless:0:*:41424344\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	blocked, trusted := 0, 0
	for _, r := range d.certRules {
		if r.blocked {
			blocked++
		} else {
			trusted++
		}
	}
	if blocked != 22 || trusted != 8 || d.Report().Unknown[UnknownCertificateTrust] != 8 {
		t.Fatalf("pinned CRB: blocked=%d trusted=%d unknown=%v", blocked, trusted, d.Report().Unknown)
	}
}

func TestCertificateIdentityHashesRawASN1SerialAndSubject(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x80ff), Subject: pkix.Name{CommonName: "CRB raw ASN.1 test"},
		NotBefore: now, NotAfter: now.Add(24 * time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	identity, err := certificateIdentityFromX509(cert)
	if err != nil {
		t.Fatal(err)
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(cert.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	subjectHash := sha1.Sum(subject.Bytes)
	serialHash := sha1.Sum([]byte{0x00, 0x80, 0xff})
	if identity.subject != subjectHash || identity.serial != serialHash ||
		!bytes.Equal(identity.modulus, key.N.Bytes()) || !identity.codeSign || identity.timeSign || identity.certSign {
		t.Fatalf("incorrect X.509 identity: %+v", identity)
	}
	if identity.serial == sha1.Sum(cert.SerialNumber.Bytes()) {
		t.Fatal("raw serial INTEGER sign byte was lost")
	}
	line := fmt.Sprintf("Revoked.Signer;0;%x;%x;%X;ignored;0;0;0;;comment\n", subjectHash, serialHash, key.N.Bytes())
	d, err := Compile(map[string][]byte{"rules.crb": []byte(line), "rule.ndb": []byte("Harmless:0:*:41424344\n")})
	if err != nil {
		t.Fatal(err)
	}
	if name, ok := d.blockedCertificate(identity); !ok || name != "Revoked.Signer" {
		t.Fatalf("generated cert did not match CRB: %q, %v", name, ok)
	}
}

func TestCertificateDuplicateAdmissionAndDiscardedRowState(t *testing.T) {
	subject, serial := strings.Repeat("01", 20), strings.Repeat("02", 20)
	row := func(name, trust, number, flags, before string) string {
		return name + ";" + trust + ";" + subject + ";" + number + ";ABCD;ignored;" + flags + ";" + before + ";comment"
	}
	for _, tc := range []struct {
		name, first, second string
		reset               bool
		count               int
	}{
		{"blocked exact", row("First", "0", serial, "1;1;1", ""), row("Second", "0", serial, "0;0;0", "100"), false, 1},
		{"blocked wildcard", row("First", "0", "", "1;1;1", ""), row("Second", "0", serial, "0;0;0", ""), true, 1},
		{"trusted covered", row("First", "1", serial, "1;1;1", "10"), row("Second", "1", serial, "1;0;0", "20"), false, 1},
		{"trusted earlier", row("First", "1", serial, "1;1;1", "20"), row("Second", "1", serial, "1;1;1", "10"), false, 2},
		{"trusted different wildcard mode", row("First", "1", "", "1;1;1", ""), row("Second", "1", serial, "1;1;1", ""), true, 2},
		{"trusted new role", row("First", "1", serial, "1;0;1", ""), row("Second", "1", serial, "1;1;1", ""), false, 2},
		{"serial sentinel", row("First", "0", strings.Repeat("ca", 20), "1;1;1", ""), row("Second", "0", "", "1;1;1", ""), false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Database{limits: DefaultLimits()}
			state := certificateSourceState{}
			if err := d.addCertificateSourceLine(tc.first, &state); err != nil {
				t.Fatal(err)
			}
			if tc.reset {
				state = certificateSourceState{}
			}
			if err := d.addCertificateSourceLine(tc.second, &state); err != nil {
				t.Fatal(err)
			}
			if len(d.certRules) != tc.count || d.certRules[0].name != "First" {
				t.Fatal("duplicate admission changed", len(d.certRules), tc.count)
			}
			if tc.name == "serial sentinel" {
				if !state.ignoreSerial {
					t.Fatal("discarded row lost source state")
				}
				if err := d.addCertificateSourceLine(row("Third", "0", serial, "1;1;1", ""), &state); err != nil {
					t.Fatal(err)
				}
				if len(d.certRules) != 2 || !d.certRules[1].ignoreSerial {
					t.Fatal("discarded row's wildcard did not carry forward")
				}
			}
		})
	}
}

func FuzzCertificateRuleDuplicateAdmission(f *testing.F) {
	subject, serial := strings.Repeat("01", 20), strings.Repeat("02", 20)
	for _, number := range []string{serial, ""} {
		for _, trust := range []string{"0", "1"} {
			f.Add("First;" + trust + ";" + subject + ";" + number + ";ABCD;ignored;1;1;1;;comment")
		}
	}
	f.Fuzz(func(t *testing.T, line string) {
		if len(line) > 64<<10 {
			return
		}
		d := &Database{limits: DefaultLimits()}
		state := certificateSourceState{}
		if err := d.addCertificateSourceLine(line, &state); err != nil {
			return
		}
		if len(d.certRules) != 1 {
			t.Fatal("first valid row was not retained")
		}
		original := d.certRules[0]
		index := strings.IndexByte(line, ';')
		if index < 0 {
			t.Fatal("valid row had no separator")
		}
		if err := d.addCertificateSourceLine("Second"+line[index:], &state); err != nil {
			t.Fatal("identical duplicate rejected", err)
		}
		if len(d.certRules) != 1 || d.certRules[0].name != original.name {
			t.Fatal("duplicate replaced the admitted record")
		}
	})
}
