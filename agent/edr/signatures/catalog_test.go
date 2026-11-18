package signatures

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
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
)

func TestInvalidCatalogSkippedWhileOtherRulesScan(t *testing.T) {
	files := map[string][]byte{
		"bad.cat":  []byte("invalid catalog"),
		"rule.ndb": []byte("Reference.ABCD:0:*:41424344\n"),
	}
	db, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if db.HasCatalogHashes() || db.Report().Skipped[SkipInvalidCatalog] != 1 || db.Report().Loaded != 1 {
		t.Fatalf("invalid catalog was loaded or raw rule lost: %+v", db.Report())
	}
	limited, err := compileWithLimits(files, Limits{MaxSignatures: 1}, compileOptions{})
	if err != nil || limited.Report().Loaded != 1 || limited.Report().Skipped[SkipInvalidCatalog] != 1 {
		t.Fatalf("invalid catalog consumed a signature slot: db=%v err=%v", limited, err)
	}
	name, err := db.Match(context.Background(), []byte("ABCD-hello world marker payload"))
	if err != nil || name != "Reference.ABCD" {
		t.Fatalf("remaining NDB did not detect: name=%q err=%v", name, err)
	}
	audit, err := Audit(files, AuditOptions{ReferenceLevel: FunctionalityLevel})
	if err != nil || audit.Totals.ReferenceSkipped != 1 || audit.SkippedByReason[SkipInvalidCatalog] != 1 || audit.Totals.Accepted != 1 {
		t.Fatalf("invalid catalog audit mismatch: report=%+v err=%v", audit, err)
	}
}

func catalogFixture(t *testing.T, peDigest, cabDigest []byte) ([]byte, string) {
	t.Helper()
	seq := func(children ...[]byte) []byte {
		return catalogNode(t, asn1.ClassUniversal, asn1.TagSequence, children...)
	}
	set := func(children ...[]byte) []byte { return catalogNode(t, asn1.ClassUniversal, asn1.TagSet, children...) }
	ctx := func(children ...[]byte) []byte { return catalogNode(t, asn1.ClassContextSpecific, 0, children...) }
	marshal := func(v any) []byte {
		out, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(0x8190), Subject: pkix.Name{CommonName: "EDR catalog fixture"},
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
	shaOID := pkcs7.OIDDigestAlgorithmSHA256
	algorithm := seq(marshal(shaOID), []byte{0x05, 0})
	member := func(kind asn1.ObjectIdentifier, hash crypto.Hash, digest []byte) []byte {
		hashOID := pkcs7.OIDDigestAlgorithmSHA1
		if hash == crypto.SHA256 {
			hashOID = pkcs7.OIDDigestAlgorithmSHA256
		}
		indirect := seq(seq(marshal(kind), seq()), seq(seq(marshal(hashOID), []byte{0x05, 0}), marshal(digest)))
		attr := seq(marshal(oidSPCIndirectData), set(indirect))
		return seq(marshal([]byte("member")), set(attr))
	}
	content := seq(seq(marshal(oidCatalogList)), marshal([]byte("catalog-id")), marshal(now),
		seq(marshal(oidCatalogMember2), []byte{0x05, 0}), seq(member(oidSPCPEImageData, crypto.SHA256, peDigest), member(oidCatalogCAB, crypto.SHA1, cabDigest)))
	var rawContent asn1.RawValue
	if _, err := asn1.Unmarshal(content, &rawContent); err != nil {
		t.Fatal(err)
	}
	messageDigest := sha256.Sum256(rawContent.Bytes)
	attrs := [][]byte{
		seq(marshal(pkcs7.OIDAttributeContentType), set(marshal(oidCatalogCTL))),
		seq(marshal(pkcs7.OIDAttributeMessageDigest), set(marshal(messageDigest[:]))),
	}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i], attrs[j]) < 0 })
	attrsDigest := sha256.Sum256(set(attrs...))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, attrsDigest[:])
	if err != nil {
		t.Fatal(err)
	}
	signer := seq(marshal(1), seq(cert.RawIssuer, marshal(cert.SerialNumber)), algorithm, ctx(attrs...),
		seq(marshal(pkcs7.OIDEncryptionAlgorithmRSA), []byte{0x05, 0}), marshal(signature))
	signed := seq(marshal(1), set(algorithm), seq(marshal(oidCatalogCTL), ctx(content)), ctx(certDER), set(signer))
	return seq(marshal(pkcs7.OIDSignedData), ctx(signed)), crb
}

func catalogNode(t *testing.T, class, tag int, children ...[]byte) []byte {
	t.Helper()
	var content []byte
	for _, child := range children {
		content = append(content, child...)
	}
	out, err := asn1.Marshal(asn1.RawValue{Class: class, Tag: tag, IsCompound: true, Bytes: content})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSignedCatalogLoadsOnlyAuthenticatedPEAndCABHashes(t *testing.T) {
	pe := sha256.Sum256([]byte("unsigned PE image hash"))
	cab := sha1.Sum([]byte("whole CAB file"))
	der, crb := catalogFixture(t, pe[:], cab[:])
	files := map[string][]byte{
		"catalog.cat": der,
		"trust.crb":   []byte(crb),
		"rule.ndb":    []byte("Raw.Marker:0:*:4d5a\n"),
	}
	db, err := Compile(files)
	if err != nil {
		t.Fatal(err)
	}
	if !db.HasCatalogHashes() || !db.CatalogAllows(2, pe[:]) || !db.CatalogAllows(1, cab[:]) ||
		db.CatalogAllows(1, pe[:]) || db.CatalogAllows(2, cab[:]) {
		t.Fatal("catalog hash namespaces were not preserved")
	}
	withoutTrust := map[string][]byte{"catalog.cat": der, "rule.ndb": files["rule.ndb"]}
	withoutTrustDB, err := Compile(withoutTrust)
	if err != nil || withoutTrustDB.HasCatalogHashes() || withoutTrustDB.Report().Skipped[SkipInvalidCatalog] != 1 || withoutTrustDB.DetectionCount() != 1 {
		t.Fatalf("untrusted catalog must be skipped while NDB remains loaded: db=%v err=%v", withoutTrustDB, err)
	}
	changed := bytes.Clone(der)
	changed[len(changed)-10] ^= 1
	files["catalog.cat"] = changed
	tamperedDB, err := Compile(files)
	if err != nil || tamperedDB.HasCatalogHashes() || tamperedDB.Report().Skipped[SkipInvalidCatalog] != 1 || tamperedDB.DetectionCount() != 1 {
		t.Fatalf("tampered catalog must be skipped while NDB remains loaded: db=%v err=%v", tamperedDB, err)
	}
}

func TestSignedCatalogAuditLimitAndReload(t *testing.T) {
	pe := sha256.Sum256([]byte("catalog reload image"))
	cab := sha1.Sum([]byte("catalog reload cabinet"))
	der, crb := catalogFixture(t, pe[:], cab[:])
	files := map[string][]byte{
		"catalog.cat": der,
		"trust.crb":   []byte(crb),
		"rule.ndb":    []byte("Raw.Marker:0:*:4d5a\n"),
	}
	limits := Limits{}.defaults()
	initial, err := compileWithLimits(files, limits, compileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	report, err := Audit(files, AuditOptions{ReferenceLevel: FunctionalityLevel})
	if err != nil {
		t.Fatal(err)
	}
	var catalogAudit *AuditFile
	for i := range report.Files {
		if report.Files[i].Format == ".cat" {
			catalogAudit = &report.Files[i]
		}
	}
	if catalogAudit == nil || catalogAudit.Rules != 1 || catalogAudit.Accepted != 1 || catalogAudit.Rejected != 0 {
		t.Fatalf("signed catalog audit was not counted as one authenticated source: %+v", catalogAudit)
	}
	if _, err := compileWithLimits(files, Limits{MaxSignatures: 3}, compileOptions{}); err != nil {
		t.Fatalf("three selected sources should fit: %v", err)
	}
	if _, err := compileWithLimits(files, Limits{MaxSignatures: 2}, compileOptions{}); err == nil {
		t.Fatal("catalog bypassed the signature count limit")
	}
	changed := map[string][]byte{"catalog.cat": bytes.Clone(der), "trust.crb": files["trust.crb"], "rule.ndb": files["rule.ndb"]}
	changed["catalog.cat"][len(der)-10] ^= 1
	changedDB, err := compileWithLimits(changed, limits, compileOptions{prior: initial})
	if err != nil || changedDB.HasCatalogHashes() || changedDB.Report().Skipped[SkipInvalidCatalog] != 1 || changedDB.DetectionCount() != 1 {
		t.Fatalf("changed catalog reused trusted hashes or lost detection rules: db=%v err=%v", changedDB, err)
	}
	badAudit, err := Audit(changed, AuditOptions{ReferenceLevel: FunctionalityLevel})
	if err != nil {
		t.Fatal(err)
	}
	if badAudit.Totals.ReferenceSkipped != 1 || badAudit.SkippedByReason[SkipInvalidCatalog] != 1 || badAudit.Totals.Rejected != 0 {
		t.Fatalf("tampered catalog was not counted as one reference-skipped source: %+v", badAudit.Totals)
	}
	changed["catalog.cat"] = der
	changed["trust.crb"] = []byte(strings.Replace(crb, ";1;", ";0;", 1))
	revokedDB, err := compileWithLimits(changed, limits, compileOptions{prior: initial})
	if err != nil || revokedDB.HasCatalogHashes() || revokedDB.Report().Skipped[SkipInvalidCatalog] != 1 || revokedDB.DetectionCount() != 1 {
		t.Fatalf("changed CRB reused catalog trust or lost detection rules: db=%v err=%v", revokedDB, err)
	}
}

func TestSignedCatalogInCVDMemberReload(t *testing.T) {
	pe := sha256.Sum256([]byte("catalog archive member"))
	cab := sha1.Sum([]byte("catalog archive cabinet"))
	der, crb := catalogFixture(t, pe[:], cab[:])
	members := []member{
		{"first.cat", der},
		{"second.ndb", []byte("Archive.Old:0:*:41424344\n")},
		{"third.crb", []byte(crb)},
	}
	files := map[string][]byte{"daily.cvd": cvdFixtureOrdered(t, members)}
	limits := Limits{}.defaults()
	initial, err := compileWithLimits(files, limits, compileOptions{})
	if err != nil || !initial.CatalogAllows(2, pe[:]) {
		t.Fatalf("catalog preceding trust member did not load: err=%v", err)
	}
	members[1].data = []byte("Archive.New:0:*:45464748\n")
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	next, err := compileWithLimits(files, limits, compileOptions{prior: initial})
	if err != nil || !next.CatalogAllows(2, pe[:]) {
		t.Fatalf("catalog did not survive archive rule reload: err=%v", err)
	}
	members[0].data = bytes.Clone(der)
	members[0].data[len(der)-10] ^= 1
	files["daily.cvd"] = cvdFixtureOrdered(t, members)
	mutated, err := compileWithLimits(files, limits, compileOptions{prior: next})
	if err != nil || mutated.HasCatalogHashes() || mutated.Report().Skipped[SkipInvalidCatalog] != 1 || mutated.DetectionCount() != 1 {
		t.Fatalf("archive catalog mutation retained trusted hashes or lost NDB: db=%v err=%v", mutated, err)
	}
}
