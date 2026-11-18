package engine

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"math/big"
	"sort"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func legacyTimestampFixture(t *testing.T, der []byte, signedAt time.Time, contentOID ...asn1.ObjectIdentifier) ([]byte, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x4041), Subject: pkix.Name{CommonName: "EDR timestamp fixture"},
		NotBefore: signedAt.Add(-time.Hour), NotAfter: signedAt.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping},
		BasicConstraintsValid: true, IsCA: true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatal(err)
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(cert.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	crb := fmt.Sprintf("Trusted.TSA;1;%x;;%X;ignored;0;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
	return appendLegacyTimestampWithSigner(t, der, signedAt, cert, key, certDER, contentOID...), crb
}

func appendLegacyTimestampWithSigner(t *testing.T, der []byte, signedAt time.Time, cert *x509.Certificate, key *rsa.PrivateKey, certDER []byte, contentOID ...asn1.ObjectIdentifier) []byte {
	t.Helper()
	primary, err := pkcs7.Parse(der)
	if err != nil {
		t.Fatal(err)
	}
	seq := func(children ...[]byte) []byte {
		return derFixture(t, asn1.ClassUniversal, asn1.TagSequence, children...)
	}
	set := func(children ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSet, children...) }
	ctx := func(tag int, children ...[]byte) []byte {
		return derFixture(t, asn1.ClassContextSpecific, tag, children...)
	}
	marshal := func(v any) []byte {
		out, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	algorithm := seq(marshal(pkcs7.OIDDigestAlgorithmSHA256), []byte{0x05, 0})
	messageHash := sha256.Sum256(primary.Signers[0].EncryptedDigest)
	contentType := pkcs7.OIDData
	if len(contentOID) != 0 {
		contentType = contentOID[0]
	}
	attrs := [][]byte{
		seq(marshal(pkcs7.OIDAttributeContentType), set(marshal(contentType))),
		seq(marshal(pkcs7.OIDAttributeMessageDigest), set(marshal(messageHash[:]))),
		seq(marshal(pkcs7.OIDAttributeSigningTime), set(marshal(signedAt))),
	}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	digest := sha256.Sum256(set(attrs...))
	signature, err := signReferenceFixture(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	counter := seq(marshal(1), seq(cert.RawIssuer, marshal(cert.SerialNumber)), algorithm,
		ctx(0, attrs...), seq(marshal(pkcs7.OIDEncryptionAlgorithmRSA), []byte{0x05, 0}), marshal(signature))
	unsigned := ctx(1, seq(marshal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 6}), set(counter)))
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
	digests, rest := next(rest)
	info, rest := next(rest)
	certs, rest := next(rest)
	signers, _ := next(rest)
	signer, _ := next(signers.Bytes)
	newSigned := seq(version.FullBytes, digests.FullBytes, info.FullBytes,
		ctx(0, certs.Bytes, certDER), set(seq(signer.Bytes, unsigned)))
	return seq(oid.FullBytes, ctx(0, newSigned))
}

func TestLegacyAuthenticodeCountersignatureRequiresTrustedTimeSigner(t *testing.T) {
	file, codeCRB, der := signedPEFixture(t, false)
	now := time.Now().UTC().Truncate(time.Second)
	timestamped, timeCRB := legacyTimestampFixture(t, der, now)
	file = replacePECertificate(t, file, timestamped)
	parsed, err := pkcs7.Parse(timestamped)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed.Signers[0].UnauthenticatedAttributes) != 1 {
		t.Fatal("counter signature fixture did not retain its unsigned attribute")
	}
	if err := parsed.Verify(); err != nil {
		t.Fatalf("primary signature with timestamp: %v", err)
	}
	db, err := signatures.Compile(map[string][]byte{"test.crb": []byte(codeCRB + timeCRB)})
	if err != nil {
		t.Fatal(err)
	}
	verify := func(d *signatures.Database, b, signature []byte, at time.Time) (bool, error) {
		return d.VerifyTrustedAuthenticode(signature, func(h crypto.Hash) ([]byte, error) {
			return peAuthenticodeDigest(b, h)
		}, at)
	}
	if verified, err := verify(db, file, timestamped, now.Add(2*time.Hour)); !verified || err != nil {
		t.Fatalf("verified timestamp did not preserve expired signer trust: verified=%v err=%v", verified, err)
	}
	withoutTime, err := signatures.Compile(map[string][]byte{"test.crb": []byte(codeCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := verify(withoutTime, file, timestamped, now.Add(2*time.Hour)); verified || err != nil {
		t.Fatalf("untrusted timestamp preserved signer trust: verified=%v err=%v", verified, err)
	}
	changedSignature := bytes.Clone(timestamped)
	changedSignature[len(changedSignature)-10] ^= 1
	if verified, err := verify(db, file, changedSignature, now.Add(2*time.Hour)); verified || err != nil {
		t.Fatalf("tampered countersignature preserved trust: verified=%v err=%v", verified, err)
	}
	changedCode := bytes.Clone(file)
	changedCode[0x200] ^= 1
	if verified, err := verify(db, changedCode, timestamped, now.Add(2*time.Hour)); verified || err == nil {
		t.Fatalf("timestamp masked changed code: verified=%v err=%v", verified, err)
	}
	future, futureCRB := legacyTimestampFixture(t, der, now.Add(2*time.Hour))
	futureFile := replacePECertificate(t, file, future)
	futureDB, err := signatures.Compile(map[string][]byte{"test.crb": []byte(codeCRB + futureCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := verify(futureDB, futureFile, future, now.Add(3*time.Hour)); verified || err != nil {
		t.Fatalf("timestamp outside signer validity preserved trust: verified=%v err=%v", verified, err)
	}
	tokenOID := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 1, 4}
	token, tokenCRB := legacyTimestampFixture(t, der, now, tokenOID)
	tokenFile := replacePECertificate(t, file, token)
	tokenDB, err := signatures.Compile(map[string][]byte{"test.crb": []byte(codeCRB + tokenCRB)})
	if err != nil {
		t.Fatal(err)
	}
	if verified, err := verify(tokenDB, tokenFile, token, now.Add(2*time.Hour)); !verified || err != nil {
		t.Fatalf("legacy timestampToken content type lost trust: verified=%v err=%v", verified, err)
	}
}
