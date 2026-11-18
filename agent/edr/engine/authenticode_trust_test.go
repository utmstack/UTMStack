package engine

import (
	"bytes"
	"context"
	"crypto"
	_ "crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/binary"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func addUnsignedCertificate(t *testing.T, der, certificate []byte) []byte {
	t.Helper()
	next := func(b []byte) (asn1.RawValue, []byte) {
		var value asn1.RawValue
		rest, err := asn1.Unmarshal(b, &value)
		if err != nil {
			t.Fatal(err)
		}
		return value, rest
	}
	outer, _ := next(der)
	oid, rest := next(outer.Bytes)
	content, _ := next(rest)
	signed, _ := next(content.Bytes)
	version, rest := next(signed.Bytes)
	digest, rest := next(rest)
	info, rest := next(rest)
	certs, signers := next(rest)
	newCerts := derFixture(t, asn1.ClassContextSpecific, 0, certs.Bytes, certificate)
	newSigned := derFixture(t, asn1.ClassUniversal, asn1.TagSequence,
		version.FullBytes, digest.FullBytes, info.FullBytes, newCerts, signers)
	return derFixture(t, asn1.ClassUniversal, asn1.TagSequence,
		oid.FullBytes, derFixture(t, asn1.ClassContextSpecific, 0, newSigned))
}

func omitLastPKCS7Certificate(t *testing.T, der []byte) []byte {
	t.Helper()
	next := func(b []byte) (asn1.RawValue, []byte) {
		var value asn1.RawValue
		rest, err := asn1.Unmarshal(b, &value)
		if err != nil {
			t.Fatal(err)
		}
		return value, rest
	}
	outer, _ := next(der)
	oid, rest := next(outer.Bytes)
	content, _ := next(rest)
	signed, _ := next(content.Bytes)
	version, rest := next(signed.Bytes)
	digest, rest := next(rest)
	info, rest := next(rest)
	certs, signers := next(rest)
	first, remainder := next(certs.Bytes)
	if len(remainder) == 0 {
		t.Fatal("fixture has no second certificate to omit")
	}
	newCerts := derFixture(t, asn1.ClassContextSpecific, 0, first.FullBytes)
	newSigned := derFixture(t, asn1.ClassUniversal, asn1.TagSequence,
		version.FullBytes, digest.FullBytes, info.FullBytes, newCerts, signers)
	return derFixture(t, asn1.ClassUniversal, asn1.TagSequence,
		oid.FullBytes, derFixture(t, asn1.ClassContextSpecific, 0, newSigned))
}

func TestTrustedCRBAnchorsSignerWithoutEmbeddedRoot(t *testing.T) {
	file, crb, der := signedPEFixture(t, true)
	rootlessDER := omitLastPKCS7Certificate(t, der)
	rootlessFile := replacePECertificate(t, file, rootlessDER)
	p7, err := pkcs7.Parse(rootlessDER)
	if err != nil || len(p7.Certificates) != 1 {
		t.Fatalf("rootless fixture has wrong certificates: %v", err)
	}
	if err := p7.Verify(); err != nil {
		t.Fatalf("rootless PKCS#7 signature is invalid: %v", err)
	}
	db, err := signatures.Compile(map[string][]byte{
		"trust.crb": []byte(crb),
		"raw.ndb":   []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	check := func(data []byte) (bool, error) {
		return db.VerifyTrustedAuthenticode(rootlessDER, func(h crypto.Hash) ([]byte, error) {
			return peAuthenticodeDigest(data, h)
		}, time.Now().UTC())
	}
	if verified, err := check(rootlessFile); !verified || err != nil {
		t.Fatalf("CRB key did not anchor rootless signer: verified=%v err=%v", verified, err)
	}
	if verified, err := db.VerifyTrustedAuthenticode(rootlessDER, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(rootlessFile, h)
	}, time.Now().Add(2*time.Hour)); verified || err != nil {
		t.Fatalf("expired rootless signer must be known untrusted: verified=%v err=%v", verified, err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	if clean, sig, err := e.Scan(context.Background(), rootlessFile); !clean || sig != "" || err != nil {
		t.Fatalf("rootless trusted PE was not exempted: clean=%v sig=%q err=%v", clean, sig, err)
	}
	changed := bytes.Clone(rootlessFile)
	changed[0x200] ^= 1
	if clean, sig, err := e.Scan(context.Background(), changed); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("changed rootless PE hid detection: clean=%v sig=%q err=%v", clean, sig, err)
	}
	noTrust, err := signatures.Compile(map[string][]byte{"raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")})
	if err != nil {
		t.Fatal(err)
	}
	e.current.Store(&generation{db: noTrust})
	if clean, sig, err := e.Scan(context.Background(), rootlessFile); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("rootless PE without CRB was exempted: clean=%v sig=%q err=%v", clean, sig, err)
	}
	parts := strings.Split(strings.TrimSpace(crb), ";")
	originalModulus := parts[4]
	parts[4] = parts[4][:len(parts[4])-2] + "01"
	if parts[4] == originalModulus {
		parts[4] = parts[4][:len(parts[4])-2] + "03"
	}
	wrongKey, err := signatures.Compile(map[string][]byte{
		"trust.crb": []byte(strings.Join(parts, ";") + "\n"),
		"raw.ndb":   []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e.current.Store(&generation{db: wrongKey})
	if clean, sig, err := e.Scan(context.Background(), rootlessFile); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("unrelated CRB key exempted rootless PE: clean=%v sig=%q err=%v", clean, sig, err)
	}
	parts[4], parts[8] = originalModulus, "0"
	noCertSigning, err := signatures.Compile(map[string][]byte{
		"trust.crb": []byte(strings.Join(parts, ";") + "\n"),
		"raw.ndb":   []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	e.current.Store(&generation{db: noCertSigning})
	if clean, sig, err := e.Scan(context.Background(), rootlessFile); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("CRB without certificate-signing permission exempted PE: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

func replacePECertificate(t *testing.T, file, der []byte) []byte {
	t.Helper()
	img, err := parsePEImage(file)
	if err != nil {
		t.Fatal(err)
	}
	table := winCertificateFixture(der)
	out := append(bytes.Clone(file[:img.certificate]), table...)
	security := int(binary.LittleEndian.Uint32(file[0x3c:])) + 24 + 96 + 32
	binary.LittleEndian.PutUint32(out[security+4:], uint32(len(table)))
	return out
}

func signedPEFixture(t *testing.T, chained bool, rootCA ...bool) ([]byte, string, []byte) {
	return signedPEFixtureWithImageHash(t, chained, crypto.SHA256, rootCA...)
}

func signedPEFixtureWithImageHash(t *testing.T, chained bool, imageAlgorithm crypto.Hash, rootCA ...bool) ([]byte, string, []byte) {
	return signedPEFixtureWithImageHashAndValidity(t, chained, imageAlgorithm, time.Hour, time.Hour, rootCA...)
}

func signedPEFixtureWithImageHashAndValidity(t *testing.T, chained bool, imageAlgorithm crypto.Hash, validBefore, validAfter time.Duration, rootCA ...bool) ([]byte, string, []byte) {
	t.Helper()
	isCA := len(rootCA) == 0 || rootCA[0]
	rootKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	usage := x509.KeyUsageDigitalSignature
	if isCA {
		usage |= x509.KeyUsageCertSign
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x8081), Subject: pkix.Name{CommonName: "EDR Authenticode trust fixture"},
		NotBefore: now.Add(-validBefore), NotAfter: now.Add(validAfter),
		KeyUsage:              usage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true, IsCA: isCA,
	}
	rootDER, err := x509.CreateCertificate(rand.Reader, template, template, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	rootCert, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatal(err)
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(rootCert.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	crb := fmt.Sprintf("Trusted.Signer;1;%x;%x;%X;ignored;1;1;1;;comment\n",
		sha1.Sum(subject.Bytes), sha1.Sum([]byte{0, 0x80, 0x81}), rootKey.N.Bytes())
	signer, signerKey, certs := rootCert, rootKey, rootDER
	if chained {
		leafKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		leaf := &x509.Certificate{
			SerialNumber: big.NewInt(0x4142), Subject: pkix.Name{CommonName: "EDR Authenticode leaf"},
			NotBefore: now.Add(-validBefore), NotAfter: now.Add(validAfter),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
			BasicConstraintsValid: true,
		}
		leafDER, err := x509.CreateCertificate(rand.Reader, leaf, rootCert, &leafKey.PublicKey, rootKey)
		if err != nil {
			t.Fatal(err)
		}
		signer, err = x509.ParseCertificate(leafDER)
		if err != nil {
			t.Fatal(err)
		}
		signerKey, certs = leafKey, append(leafDER, rootDER...)
	}
	file, der := signedPEWithSignerImageHash(t, signer, signerKey, certs, imageAlgorithm)
	return file, crb, der
}

func signedPEWithSigner(t *testing.T, signer *x509.Certificate, signerKey *rsa.PrivateKey, certs []byte) ([]byte, []byte) {
	return signedPEWithSignerImageHash(t, signer, signerKey, certs, crypto.SHA256)
}

func signedPEWithSignerImageHash(t *testing.T, signer *x509.Certificate, signerKey *rsa.PrivateKey, certs []byte, imageAlgorithm crypto.Hash) ([]byte, []byte) {
	t.Helper()
	code := bytes.Repeat([]byte{0x90}, 0x200)
	f := simplePE(code)
	f.cert = winCertificateFixture([]byte{0x30, 0})
	pre := f.build()
	imageHash, err := peAuthenticodeDigest(pre, imageAlgorithm)
	if err != nil {
		t.Fatal(err)
	}
	seq := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSequence, b...) }
	set := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSet, b...) }
	ctx := func(b ...[]byte) []byte { return derFixture(t, asn1.ClassContextSpecific, 0, b...) }
	octet := func(b []byte) []byte {
		out, err := asn1.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	version, _ := asn1.Marshal(1)
	serial, _ := asn1.Marshal(signer.SerialNumber)
	shaOID := oidFixture(t, asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1})
	imageOID := shaOID
	if imageAlgorithm == crypto.MD5 {
		imageOID = oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5})
	}
	spcOID := oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 4})
	peOID := oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 15})
	rsaOID := oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1})
	null := []byte{0x05, 0x00}
	algorithm := seq(shaOID, null)
	content := seq(seq(peOID), seq(seq(imageOID, null), octet(imageHash)))
	var rawContent asn1.RawValue
	if _, err := asn1.Unmarshal(content, &rawContent); err != nil {
		t.Fatal(err)
	}
	messageDigest := sha256.Sum256(rawContent.Bytes)
	contentTypeAttr := seq(oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}), set(spcOID))
	digestAttr := seq(oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}), set(octet(messageDigest[:])))
	attrs := [][]byte{contentTypeAttr, digestAttr}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	attrSet := set(attrs...)
	attrHash := sha256.Sum256(attrSet)
	signature, err := signReferenceFixture(rand.Reader, signerKey, crypto.SHA256, attrHash[:])
	if err != nil {
		t.Fatal(err)
	}
	signerInfo := seq(version, seq(signer.RawIssuer, serial), algorithm, ctx(attrs...), seq(rsaOID, null), octet(signature))
	signed := seq(version, set(algorithm), seq(spcOID, ctx(content)), ctx(certs), set(signerInfo))
	der := seq(oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}), ctx(signed))
	f.cert = winCertificateFixture(der)
	return f.build(), der
}

func TestVerifyTrustedAuthenticodeRequiresSignerChainAndImageDigest(t *testing.T) {
	file, crb, der := signedPEFixture(t, false)
	db, err := signatures.Compile(map[string][]byte{
		"test.crb": []byte(crb),
		"test.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	check := func(file, der []byte) (bool, error) {
		return db.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
			return peAuthenticodeDigest(file, h)
		}, time.Now().UTC())
	}
	if verified, err := check(file, der); !verified || err != nil {
		t.Fatalf("valid signed PE: verified=%v err=%v", verified, err)
	}
	state := func() *scanState {
		return &scanState{ctx: context.Background(), db: db, maxFile: 16 << 20,
			remaining: 64 << 20, maxFiles: 1000, maxDepth: 12}
	}
	if sig, err := state().visit(file, 0); sig != "" || err != nil {
		t.Fatalf("trusted PE was not exempted: sig=%q err=%v", sig, err)
	}
	engine := New(config.EDRConfig{})
	engine.current.Store(&generation{db: db})
	if clean, sig, err := engine.Scan(context.Background(), file); !clean || sig != "" || err != nil {
		t.Fatalf("engine API did not exempt trusted PE: clean=%v sig=%q err=%v", clean, sig, err)
	}
	blockedRule := strings.Replace(crb, "Trusted.Signer;1;", "Revoked.Signer;0;", 1)
	blockAndTrust, err := signatures.Compile(map[string][]byte{
		"test.crb": []byte(crb + blockedRule),
		"test.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	blockedState := state()
	blockedState.db = blockAndTrust
	if sig, err := blockedState.visit(file, 0); sig != "Revoked.Signer" || err != nil {
		t.Fatalf("trust overrode certificate blocklist: sig=%q err=%v", sig, err)
	}
	changedCode := bytes.Clone(file)
	changedCode[0x200] ^= 0xff
	if verified, err := check(changedCode, der); verified || err == nil {
		t.Fatalf("modified code was trusted: verified=%v err=%v", verified, err)
	}
	if sig, err := state().visit(changedCode, 0); sig != "Suspicious.Code" || err != nil {
		t.Fatalf("tampered PE hid raw detection: sig=%q err=%v", sig, err)
	}
	if clean, sig, err := engine.Scan(context.Background(), changedCode); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("engine API hid tampered PE: clean=%v sig=%q err=%v", clean, sig, err)
	}
	changedChecksum := bytes.Clone(file)
	checksum := int(binary.LittleEndian.Uint32(file[0x3c:])) + 24 + 64
	changedChecksum[checksum] ^= 0xff
	if verified, err := check(changedChecksum, der); !verified || err != nil {
		t.Fatalf("excluded checksum changed trust: verified=%v err=%v", verified, err)
	}
	appended := append(bytes.Clone(file), 0x90)
	if verified, err := check(appended, der); verified || err == nil {
		t.Fatalf("appended data was trusted: verified=%v err=%v", verified, err)
	}
	img, err := parsePEImage(file)
	if err != nil {
		t.Fatal(err)
	}
	// ClamAV 1.5.4 deliberately accepts a WIN_CERTIFICATE length that
	// disagrees with the security directory (MS13-098 compatibility). Our
	// trust exemption stays strict: this field must not hide a raw alert.
	certificateLength := binary.LittleEndian.Uint32(file[img.certificate:])
	for _, value := range []uint32{0, 1, 7, 8, certificateLength - 1, certificateLength + 1, ^uint32(0)} {
		changed := bytes.Clone(file)
		binary.LittleEndian.PutUint32(changed[img.certificate:], value)
		if verified, err := check(changed, der); verified || err == nil {
			t.Fatalf("malformed certificate length %d trusted: verified=%v err=%v", value, verified, err)
		}
		if sig, err := state().visit(changed, 0); sig != "Suspicious.Code" || err != nil {
			t.Fatalf("malformed certificate length %d hid raw detection: sig=%q err=%v", value, sig, err)
		}
	}
	paddingAttack := append(bytes.Clone(file), []byte("UNSIGNED")...)
	length := int(binary.LittleEndian.Uint32(file[img.certificate:]))
	binary.LittleEndian.PutUint32(paddingAttack[img.certificate:], uint32(length+8))
	security := int(binary.LittleEndian.Uint32(file[0x3c:])) + 24 + 96 + 32
	binary.LittleEndian.PutUint32(paddingAttack[security+4:], uint32(length+8))
	if verified, err := check(paddingAttack, der); verified || err == nil {
		t.Fatalf("certificate padding payload was trusted: verified=%v err=%v", verified, err)
	}
	if sig, err := state().visit(paddingAttack, 0); sig != "Suspicious.Code" || err != nil {
		t.Fatalf("certificate padding hid raw detection: sig=%q err=%v", sig, err)
	}
	changedSignature := bytes.Clone(der)
	changedSignature[len(changedSignature)-10] ^= 0xff
	if verified, err := check(file, changedSignature); verified || err != nil {
		t.Fatalf("modified signed data was trusted: verified=%v err=%v", verified, err)
	}
	if verified, err := db.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(file, h)
	}, time.Now().Add(2*time.Hour)); verified || err != nil {
		t.Fatalf("expired signing chain must be known untrusted: verified=%v err=%v", verified, err)
	}
	timeOnlyCRB := strings.Replace(crb, ";1;1;1;;comment", ";0;1;1;;comment", 1)
	timeOnlyDB, err := signatures.Compile(map[string][]byte{"test.crb": []byte(timeOnlyCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := timeOnlyDB.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(file, h)
	}, time.Now().UTC()); verified || err != nil {
		t.Fatalf("time-signing-only CRB trusted code: verified=%v err=%v", verified, err)
	}
	chainedFile, chainedCRB, chainedDER := signedPEFixture(t, true)
	chainedDB, err := signatures.Compile(map[string][]byte{"test.crb": []byte(chainedCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := chainedDB.VerifyTrustedAuthenticode(chainedDER, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(chainedFile, h)
	}, time.Now().UTC()); !verified || err != nil {
		t.Fatalf("trusted root did not authenticate leaf signer: verified=%v err=%v", verified, err)
	}
	leafFile, leafCRB, leafDER := signedPEFixture(t, false, false)
	leafDB, err := signatures.Compile(map[string][]byte{"test.crb": []byte(leafCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := leafDB.VerifyTrustedAuthenticode(leafDER, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(leafFile, h)
	}, time.Now().UTC()); !verified || err != nil {
		t.Fatalf("directly trusted non-CA signer was rejected: verified=%v err=%v", verified, err)
	}
	_, unrelatedCRB, unrelatedDER := signedPEFixture(t, false)
	unrelatedPKCS7, err := pkcs7.Parse(unrelatedDER)
	if err != nil {
		t.Fatal(err)
	}
	withUnrelatedCert := addUnsignedCertificate(t, der, unrelatedPKCS7.GetOnlySigner().Raw)
	withUnrelatedFile := replacePECertificate(t, file, withUnrelatedCert)
	unrelatedDB, err := signatures.Compile(map[string][]byte{
		"test.crb": []byte(unrelatedCRB), "test.ndb": []byte("Suspicious.Code:0:*:9090909090\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := unrelatedDB.VerifyTrustedAuthenticode(withUnrelatedCert, func(h crypto.Hash) ([]byte, error) {
		return peAuthenticodeDigest(withUnrelatedFile, h)
	}, time.Now().UTC()); verified || err != nil {
		t.Fatalf("unrelated embedded trusted cert authenticated signer: verified=%v err=%v", verified, err)
	}
	s := state()
	s.db = unrelatedDB
	if sig, err := s.visit(withUnrelatedFile, 0); sig != "Suspicious.Code" || err != nil {
		t.Fatalf("unrelated trust cert hid detection: sig=%q err=%v", sig, err)
	}
}
