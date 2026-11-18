package engine

import (
	"bytes"
	"context"
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
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func signedCatalogForEngine(t *testing.T, peDigest, cabDigest []byte) ([]byte, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x9031), Subject: pkix.Name{CommonName: "EDR catalog scan fixture"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
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
	crb := fmt.Sprintf("Trusted.Catalog;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
	return signedCatalogWithSignerForEngine(t, peDigest, cabDigest, cert, key, certDER), crb
}

func signedCatalogWithSignerForEngine(t *testing.T, peDigest, cabDigest []byte, cert *x509.Certificate, key *rsa.PrivateKey, certDER []byte) []byte {
	t.Helper()
	seq := func(children ...[]byte) []byte {
		return derFixture(t, asn1.ClassUniversal, asn1.TagSequence, children...)
	}
	set := func(children ...[]byte) []byte { return derFixture(t, asn1.ClassUniversal, asn1.TagSet, children...) }
	ctx := func(children ...[]byte) []byte { return derFixture(t, asn1.ClassContextSpecific, 0, children...) }
	marshal := func(value any) []byte {
		out, err := asn1.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	now := time.Now().UTC()
	ctlOID := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 10, 1}
	indirectOID := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 4}
	member := func(kind asn1.ObjectIdentifier, hashOID asn1.ObjectIdentifier, digest []byte) []byte {
		indirect := seq(seq(marshal(kind), seq()), seq(seq(marshal(hashOID), []byte{0x05, 0}), marshal(digest)))
		return seq(marshal([]byte("member")), set(seq(marshal(indirectOID), set(indirect))))
	}
	content := seq(
		seq(marshal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 12, 1, 1})),
		marshal([]byte("list")), marshal(now), seq(marshal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 12, 1, 3}), []byte{0x05, 0}),
		seq(member(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 15}, pkcs7.OIDDigestAlgorithmSHA256, peDigest),
			member(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 25}, pkcs7.OIDDigestAlgorithmSHA1, cabDigest)))
	var rawContent asn1.RawValue
	if _, err := asn1.Unmarshal(content, &rawContent); err != nil {
		t.Fatal(err)
	}
	messageHash := sha256.Sum256(rawContent.Bytes)
	attrs := [][]byte{
		seq(marshal(pkcs7.OIDAttributeContentType), set(marshal(ctlOID))),
		seq(marshal(pkcs7.OIDAttributeMessageDigest), set(marshal(messageHash[:]))),
	}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	attrHash := sha256.Sum256(set(attrs...))
	signature, err := signReferenceFixture(rand.Reader, key, crypto.SHA256, attrHash[:])
	if err != nil {
		t.Fatal(err)
	}
	algorithm := seq(marshal(pkcs7.OIDDigestAlgorithmSHA256), []byte{0x05, 0})
	signer := seq(marshal(1), seq(cert.RawIssuer, marshal(cert.SerialNumber)), algorithm, ctx(attrs...),
		seq(marshal(pkcs7.OIDEncryptionAlgorithmRSA), []byte{0x05, 0}), marshal(signature))
	signed := seq(marshal(1), set(algorithm), seq(marshal(ctlOID), ctx(content)), ctx(certDER), set(signer))
	return seq(marshal(pkcs7.OIDSignedData), ctx(signed))
}

func TestSignedCatalogExemptsOnlyMatchingPEAndCAB(t *testing.T) {
	code := bytes.Repeat([]byte("PAYLOAD"), 80)
	pe := simplePE(code).build()
	cab := cabFixture(t, 0, [][]byte{[]byte("PAYLOAD")}, []cabFixtureFile{{name: "payload.bin", size: 7}}, false)
	peDigest, err := peCatalogDigest(pe, crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	cabDigest := sha1.Sum(cab)
	cat, crb := signedCatalogForEngine(t, peDigest, cabDigest[:])
	db, err := signatures.Compile(map[string][]byte{
		"catalog.cat": cat,
		"trust.crb":   []byte(crb),
		"raw.ndb":     []byte("Raw.Payload:0:*:5041594c4f4144\n"),
	})
	if err != nil {
		t.Fatal(err)
	}
	engine := New(config.EDRConfig{})
	engine.current.Store(&generation{db: db})
	for _, file := range [][]byte{pe, cab} {
		if clean, sig, err := engine.Scan(context.Background(), file); !clean || sig != "" || err != nil {
			t.Fatalf("catalog member was not exempted: clean=%v sig=%q err=%v", clean, sig, err)
		}
		changed := bytes.Clone(file)
		if bytes.HasPrefix(file, []byte("MZ")) {
			changed[0x200] ^= 1
		} else {
			changed[60] ^= 1 // cabinet member name, outside its data-block checksum
		}
		if clean, sig, err := engine.Scan(context.Background(), changed); clean || sig != "Raw.Payload" {
			t.Fatalf("changed member kept catalog trust: clean=%v sig=%q err=%v", clean, sig, err)
		}
	}
}
