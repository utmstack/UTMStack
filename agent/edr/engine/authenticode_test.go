package engine

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func certDER(t *testing.T) ([]byte, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x80ff), Subject: pkix.Name{CommonName: "Revocation fixture"},
		NotBefore: now, NotAfter: now.AddDate(1, 0, 0),
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
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(cert.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	line := fmt.Sprintf("Revoked.Signer;0;%x;%x;%X;ignored;0;0;0;;comment\n",
		sha1.Sum(subject.Bytes), sha1.Sum([]byte{0, 0x80, 0xff}), key.N.Bytes())
	return der, line
}

func derFixture(t *testing.T, class, tag int, children ...[]byte) []byte {
	t.Helper()
	var content []byte
	for _, child := range children {
		content = append(content, child...)
	}
	b, err := asn1.Marshal(asn1.RawValue{Class: class, Tag: tag, IsCompound: true, Bytes: content})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func oidFixture(t *testing.T, oid asn1.ObjectIdentifier) []byte {
	t.Helper()
	b, err := asn1.Marshal(oid)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func signedDataFixture(t *testing.T, cert, contentOID []byte) []byte {
	t.Helper()
	seq := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSequence, b...) }
	set := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSet, b...) }
	ctx := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassContextSpecific, 0, b...) }
	version, _ := asn1.Marshal(1)
	sha256 := oidFixture(t, asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1})
	signedOID := oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2})
	return seq(signedOID, ctx(seq(version, set(seq(sha256)),
		seq(contentOID, ctx(seq())), ctx(cert), set())))
}

func winCertificateFixture(der []byte) []byte {
	b := make([]byte, 8, 8+len(der)+7)
	binary.LittleEndian.PutUint16(b[4:], 0x0200)
	binary.LittleEndian.PutUint16(b[6:], 2)
	b = append(b, der...)
	b = append(b, make([]byte, (8-len(b)%8)%8)...)
	binary.LittleEndian.PutUint32(b, uint32(len(b)))
	return b
}

func TestPEAuthenticodeBlocklistVerdict(t *testing.T) {
	cert, crb := certDER(t)
	spcOID := oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 4})
	dataOID := oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1})
	blocked := winCertificateFixture(signedDataFixture(t, cert, spcOID))
	blockedZeroLength := bytes.Clone(blocked)
	binary.LittleEndian.PutUint32(blockedZeroLength, 0)
	blockedOversizedLength := bytes.Clone(blocked)
	binary.LittleEndian.PutUint32(blockedOversizedLength, ^uint32(0))
	other := winCertificateFixture(signedDataFixture(t, cert, dataOID))
	badCert := derFixture(t, asn1.ClassUniversal, asn1.TagSequence)
	badOnly := winCertificateFixture(signedDataFixture(t, badCert, spcOID))
	badThenBlocked := winCertificateFixture(signedDataFixture(t, append(badCert, cert...), spcOID))
	db, err := signatures.Compile(map[string][]byte{
		"rules.crb": []byte(crb), "rule.ndb": []byte("Payload.Test:0:*:5041594c4f4144\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if n := authenticodeSignature(blocked); n == 0 {
		t.Fatal("test WIN_CERTIFICATE was not recognized")
	} else if sig, err := db.MatchBlockedAuthenticode(blocked[8:n]); sig != "Revoked.Signer" || err != nil {
		t.Fatalf("direct Authenticode match: sig=%q err=%v", sig, err)
	}
	state := func() *scanState {
		return &scanState{ctx: context.Background(), db: db, maxFile: 16 << 20,
			remaining: 64 << 20, maxFiles: 1000, maxDepth: 12}
	}
	image := func(certTable []byte, code []byte) []byte {
		f := simplePE(code)
		f.cert = certTable
		return f.build()
	}
	cleanCode := bytes.Repeat([]byte{0x90}, 0x200)
	if img, err := parsePEImage(image(blocked, cleanCode)); err != nil {
		t.Fatal(err)
	} else if img.certificate >= img.certificateEnd {
		t.Fatalf("test PE certificate range: %d..%d", img.certificate, img.certificateEnd)
	}
	for _, tc := range []struct {
		name, want string
		cert       []byte
		code       []byte
	}{
		{"revoked embedded cert", "Revoked.Signer", blocked, cleanCode},
		{"revoked cert with zero inner length", "Revoked.Signer", blockedZeroLength, cleanCode},
		{"revoked cert with oversized inner length", "Revoked.Signer", blockedOversizedLength, cleanCode},
		{"second WIN_CERTIFICATE", "Revoked.Signer", append(append([]byte(nil), other...), blocked...), cleanCode},
		{"revoked after unreadable certificate", "Revoked.Signer", badThenBlocked, cleanCode},
		{"different content type", "", other, cleanCode},
		{"ordinary raw signature still detects", "Payload.Test", other, append([]byte("PAYLOAD"), cleanCode...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sig, err := state().visit(image(tc.cert, tc.code), 0)
			if sig != tc.want || err != nil {
				t.Fatalf("sig=%q err=%v; want %q", sig, err, tc.want)
			}
		})
	}
	if sig, err := state().visit(image(badOnly, cleanCode), 0); sig != "" || err != nil {
		t.Fatalf("known invalid empty certificate: sig=%q err=%v", sig, err)
	}
	// The synthetic PKCS#7 signer set is empty and its PE digest is not
	// verified: the blocklist verdict depends on the embedded certificate,
	// consistent with the reference's order of checks.
	if sig, err := state().visit(image(blocked, cleanCode), 0); sig != "Revoked.Signer" || err != nil {
		t.Fatalf("unverified signed digest: sig=%q err=%v", sig, err)
	}
	trustedDB, err := signatures.Compile(map[string][]byte{
		"rules.crb": []byte(strings.Replace(crb, ";0;", ";1;", 1)),
		"rule.ndb":  []byte("Payload.Test:0:*:5041594c4f4144\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	s := state()
	s.db = trustedDB
	if sig, err := s.visit(image(blocked, append([]byte("PAYLOAD"), cleanCode...)), 0); sig != "Payload.Test" || err != nil {
		t.Fatalf("unverified trust suppressed detection: sig=%q err=%v", sig, err)
	}
}
