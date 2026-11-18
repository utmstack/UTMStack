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
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func TestVerifiedCatalogCarriesMissingPESignerIntermediate(t *testing.T) {
	now := time.Now().UTC()
	makeCert := func(serial int64, name string, parent *x509.Certificate, parentKey *rsa.PrivateKey, ca bool) (*x509.Certificate, *rsa.PrivateKey, []byte) {
		t.Helper()
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		usage := x509.KeyUsageDigitalSignature
		if ca {
			usage |= x509.KeyUsageCertSign
		}
		template := &x509.Certificate{
			SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: name},
			NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
			KeyUsage: usage, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
			BasicConstraintsValid: true, IsCA: ca,
		}
		if parent == nil {
			parent, parentKey = template, key
		}
		der, err := x509.CreateCertificate(rand.Reader, template, parent, &key.PublicKey, parentKey)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		return cert, key, der
	}
	root, rootKey, rootDER := makeCert(1001, "EDR catalog chain root", nil, nil, true)
	intermediate, intermediateKey, intermediateDER := makeCert(1002, "EDR catalog intermediate", root, rootKey, true)
	leaf, leafKey, leafDER := makeCert(1003, "EDR PE signer", intermediate, intermediateKey, false)
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(root.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	crb := fmt.Sprintf("Trusted.Chain;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), rootKey.N.Bytes())
	pe, _ := signedPEWithSigner(t, leaf, leafKey, leafDER)
	cat := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20),
		intermediate, intermediateKey, append(bytes.Clone(intermediateDER), rootDER...))
	inputs := map[string][]byte{
		"trust.crb": []byte(crb),
		"raw.ndb":   []byte("Suspicious.Code:0:*:9090909090\n"),
	}
	scan := func(db *signatures.Database, data []byte) (bool, string, error) {
		t.Helper()
		e := New(config.EDRConfig{})
		e.current.Store(&generation{db: db})
		return e.Scan(context.Background(), data)
	}
	without, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := scan(without, pe); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("rootless PE without catalog: clean=%v sig=%q err=%v", clean, sig, err)
	}
	inputs["verified.cat"] = cat
	with, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := scan(with, pe); !clean || sig != "" || err != nil {
		t.Fatalf("verified catalog did not supply missing intermediate: clean=%v sig=%q err=%v", clean, sig, err)
	}
	tamperedPE := bytes.Clone(pe)
	tamperedPE[0x200] ^= 1
	if clean, sig, err := scan(with, tamperedPE); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("modified PE code kept catalog-chain trust: clean=%v sig=%q err=%v", clean, sig, err)
	}
	// A certificate outside the catalog signer's path is still usable when it
	// independently verifies under CRB trust, as in the reference store.
	otherTrusted, otherKey, otherDER := makeCert(1004, "EDR second trusted signer", root, rootKey, false)
	otherPE, _ := signedPEWithSigner(t, otherTrusted, otherKey, otherDER)
	inputs["verified.cat"] = addUnsignedCertificate(t, cat, otherDER)
	withOther, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := scan(withOther, otherPE); !clean || sig != "" || err != nil {
		t.Fatalf("CRB-verified catalog certificate was unavailable: clean=%v sig=%q err=%v", clean, sig, err)
	}
	// A certificate merely carried in the bag without a CRB-rooted chain
	// must not become a trust anchor for its own executable.
	unrelated, unrelatedKey, unrelatedDER := makeCert(1005, "EDR unrelated signer", nil, nil, false)
	unrelatedPE, _ := signedPEWithSigner(t, unrelated, unrelatedKey, unrelatedDER)
	inputs["verified.cat"] = addUnsignedCertificate(t, inputs["verified.cat"], unrelatedDER)
	withExtra, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := scan(withExtra, unrelatedPE); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("unrelated catalog certificate became trusted: clean=%v sig=%q err=%v", clean, sig, err)
	}
	tamperedCatalog := bytes.Clone(cat)
	tamperedCatalog[len(tamperedCatalog)-10] ^= 1
	tamperedCatalog = addUnsignedCertificate(t, tamperedCatalog, unrelatedDER)
	inputs["verified.cat"] = tamperedCatalog
	withTampered, err := signatures.Compile(inputs)
	if err != nil || withTampered.HasCatalogHashes() || withTampered.Report().Skipped[signatures.SkipInvalidCatalog] != 1 {
		t.Fatalf("tampered catalog was not skipped: db=%v err=%v", withTampered, err)
	}
	if clean, sig, err := scan(withTampered, pe); !clean || sig != "" || err != nil {
		t.Fatalf("independently authenticated catalog path was lost: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := scan(withTampered, unrelatedPE); clean || sig != "Suspicious.Code" || err != nil {
		t.Fatalf("unrooted certificate from invalid catalog was trusted: clean=%v sig=%q err=%v", clean, sig, err)
	}
}
