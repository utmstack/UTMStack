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
	"testing"
	"time"
)

func TestLegacySHA1TimestampChainRequiresCRBAnchorAndValidLinks(t *testing.T) {
	at := time.Date(2020, 3, 25, 7, 28, 23, 0, time.UTC)
	makeKey := func() *rsa.PrivateKey {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	rootKey, parentKey, leafKey, wrongKey := makeKey(), makeKey(), makeKey(), makeKey()
	makeCert := func(template, issuer *x509.Certificate, key, issuerKey *rsa.PrivateKey) *x509.Certificate {
		der, err := x509.CreateCertificate(rand.Reader, template, issuer, &key.PublicKey, issuerKey)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return cert
	}
	valid := func(cn string, serial int64, isCA bool) *x509.Certificate {
		c := &x509.Certificate{
			SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: cn},
			NotBefore: at.Add(-24 * time.Hour), NotAfter: at.Add(24 * time.Hour),
			KeyUsage:              x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping},
			BasicConstraintsValid: true, IsCA: isCA,
		}
		if isCA {
			c.KeyUsage |= x509.KeyUsageCertSign
		}
		return c
	}
	rootTemplate := valid("Trusted timestamp root", 1, true)
	root := makeCert(rootTemplate, rootTemplate, rootKey, rootKey)
	parentTemplate := valid("Timestamp PCA", 2, true)
	parentTemplate.SignatureAlgorithm = x509.SHA1WithRSA
	parent := makeCert(parentTemplate, root, parentKey, rootKey)
	leafTemplate := valid("Timestamp service", 3, false)
	leafTemplate.SignatureAlgorithm = x509.SHA1WithRSA
	leaf := makeCert(leafTemplate, parent, leafKey, parentKey)
	if parent.SignatureAlgorithm != x509.SHA1WithRSA || leaf.SignatureAlgorithm != x509.SHA1WithRSA {
		t.Fatal("fixture lost SHA-1 certificate links")
	}
	var subject asn1.RawValue
	if rest, err := asn1.Unmarshal(root.RawSubject, &subject); err != nil || len(rest) != 0 {
		t.Fatalf("parse root subject: %v", err)
	}
	rootHash := sha1.Sum(subject.Bytes)
	compile := func(modulus []byte) *Database {
		line := fmt.Sprintf("Trusted.TSA;1;%x;;%X;ignored;0;1;1;;comment\n", rootHash, modulus)
		db, err := Compile(map[string][]byte{"test.crb": []byte(line)})
		if err != nil {
			t.Fatal(err)
		}
		return db
	}
	db := compile(rootKey.N.Bytes())
	if !db.verifyHistoricalTimeChain(leaf, []*x509.Certificate{leaf, parent}, at) {
		t.Fatal("valid historical SHA-1 timestamp chain lost CRB trust")
	}
	wrong := compile(wrongKey.N.Bytes())
	if wrong.verifyHistoricalTimeChain(leaf, []*x509.Certificate{leaf, parent}, at) {
		t.Fatal("unrelated CRB key trusted timestamp chain")
	}
	tamperedParent := *parent
	tamperedParent.Signature = bytes.Clone(parent.Signature)
	tamperedParent.Signature[0] ^= 1
	if db.verifyHistoricalTimeChain(leaf, []*x509.Certificate{leaf, &tamperedParent}, at) {
		t.Fatal("tampered timestamp issuer retained CRB trust")
	}
	tampered := *leaf
	tampered.Signature = bytes.Clone(leaf.Signature)
	tampered.Signature[0] ^= 1
	if db.verifyHistoricalTimeChain(&tampered, []*x509.Certificate{&tampered, parent}, at) {
		t.Fatal("tampered timestamp leaf retained trust")
	}
	if !db.verifyHistoricalTimeChain(leaf, []*x509.Certificate{leaf, parent}, at.Add(48*time.Hour)) {
		t.Fatal("reference timestamp-authority policy was replaced by TLS validity policy")
	}
}
