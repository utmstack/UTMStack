package signatures

import (
	"archive/zip"
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
)

// The signed driver archive is supplied out of tree. Its catalog was signed
// with a code-signing certificate that expired in January 2025, and carries
// an RFC3161 timestamp the pinned ClamAV 1.5.4 verifier does not validate.
func TestRealOpenVPNSignedCatalog(t *testing.T) {
	path := os.Getenv("EDR_OPENVPN_CATALOG_ZIP")
	if path == "" {
		t.Skip("set EDR_OPENVPN_CATALOG_ZIP to OpenVPN tap-windows6 9.27.0 dist.win10.zip")
	}
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() != 189720 {
		t.Fatalf("unexpected real-catalog archive: %v", err)
	}
	archive, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(archive)); got != "36e2609b7ceefedcb978ce5c48caf9e0e5af83423717c4e2e3c1d7ebca8f62a5" {
		t.Fatalf("archive SHA-256 changed: %s", got)
	}
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatal(err)
	}
	var catalog []byte
	for _, f := range zr.File {
		if f.Name != "dist.win10/amd64/tap0901.cat" {
			continue
		}
		if f.UncompressedSize64 != 11882 {
			t.Fatal("catalog size changed")
		}
		r, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		catalog, err = io.ReadAll(io.LimitReader(r, 11883))
		closeErr := r.Close()
		if err != nil || closeErr != nil || len(catalog) != 11882 {
			t.Fatalf("catalog read: %v, close: %v", err, closeErr)
		}
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(catalog)); got != "ee062e5ef2743ceab10c64830e4cefe52e35cc1ece85947ac4e61ddd1c0b05f7" {
		t.Fatalf("catalog SHA-256 changed: %s", got)
	}
	p7, err := pkcs7.Parse(catalog)
	if err != nil || len(p7.Signers) != 1 {
		t.Fatalf("catalog PKCS#7 parse: %v", err)
	}
	if err := p7.Verify(); err != nil {
		t.Fatalf("catalog PKCS#7 signature: %v", err)
	}
	signer := p7.GetOnlySigner()
	if signer == nil || signer.Subject.CommonName != "Microsoft Windows Hardware Compatibility Publisher" {
		t.Fatal("unexpected catalog signer")
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(signer.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	key, ok := signer.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("catalog signer is not RSA")
	}
	crb := fmt.Sprintf("Trusted.OpenVPN;1;%x;;%X;ignored;1;1;1;;comment", sha1.Sum(subject.Bytes), key.N.Bytes())
	db := newDatabase(DefaultLimits(), FunctionalityLevel)
	if err := db.addCertificateLine(crb); err != nil {
		t.Fatal(err)
	}
	withinValidity := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	entries, err := db.verifiedCatalogEntries(catalog, withinValidity)
	if err != nil || len(entries) != 2 {
		t.Fatalf("real signed catalog at signer-valid time: entries=%d err=%v", len(entries), err)
	}
	want := map[int64]string{
		1: "1327ab3a8c50691f04bea8e2ca356c5b604092a719e219464f8cc4b42e192de9",
		2: "9e176400e823172804afb95b5eb0c03c39b8e96a6b15fab745fa9190da45ebbb",
	}
	for _, entry := range entries {
		if got := hex.EncodeToString(entry.digest); got != want[entry.kind] {
			t.Fatalf("unexpected signed member kind=%d digest=%s", entry.kind, got)
		}
	}
	if _, err := db.verifiedCatalogEntries(catalog, time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expired signer with RFC3161-only timestamp was trusted")
	}
	otherRules := map[string][]byte{
		"trust.crb": []byte(crb + "\n"), "rule.ndb": []byte("Raw.MZ:0:0:4d5a90000300000004000000\n"),
	}
	if _, err := Compile(otherRules); err != nil {
		t.Fatalf("valid reference rules without catalog: %v", err)
	}
	withCatalog := map[string][]byte{
		"tap.cat": catalog, "trust.crb": otherRules["trust.crb"], "rule.ndb": otherRules["rule.ndb"],
	}
	expiredDB, err := Compile(withCatalog)
	if err != nil || expiredDB.HasCatalogHashes() || expiredDB.Report().Skipped[SkipInvalidCatalog] != 1 || expiredDB.DetectionCount() != 1 {
		t.Fatalf("production compiler did not skip expired catalog while retaining NDB: db=%v err=%v", expiredDB, err)
	}
	changed := bytes.Clone(catalog)
	memberOffset := bytes.Index(changed, entries[1].digest)
	if memberOffset < 0 {
		t.Fatal("signed member digest not found in catalog bytes")
	}
	changed[memberOffset] ^= 1
	if _, err := db.verifiedCatalogEntries(changed, withinValidity); err == nil {
		t.Fatal("tampered real catalog was trusted")
	}
	withoutTrust := newDatabase(DefaultLimits(), FunctionalityLevel)
	if _, err := withoutTrust.verifiedCatalogEntries(catalog, withinValidity); err == nil || !strings.Contains(err.Error(), "untrusted") {
		t.Fatalf("real catalog without CRB trust: %v", err)
	}
}
