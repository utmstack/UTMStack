package signatures

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
)

// This independent signed catalog is supplied out of tree from the public
// DsHidMini v3.27.0 release. Its signer is valid from May 2026 to May 2027.
func TestRealCurrentSignedCatalog(t *testing.T) {
	path := os.Getenv("EDR_DSHIDMINI_CATALOG")
	if path == "" {
		t.Skip("set EDR_DSHIDMINI_CATALOG to the pinned DsHidMini v3.27.0 igfilter.cat")
	}
	st, err := os.Stat(path)
	if err != nil || !st.Mode().IsRegular() || st.Size() != 11721 {
		t.Fatalf("unexpected catalog file: %v", err)
	}
	cat, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(cat)); got != "9b6921c0662513a0e57c5025e29268f728cfda31a34a76e26013282c671adc3b" {
		t.Fatalf("catalog SHA-256 changed: %s", got)
	}
	p7, err := pkcs7.Parse(cat)
	if err != nil || len(p7.Signers) != 1 {
		t.Fatalf("catalog PKCS#7 parse: %v", err)
	}
	if err := p7.Verify(); err != nil {
		t.Fatalf("catalog signature: %v", err)
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
	crb := fmt.Sprintf("Trusted.DsHidMini;1;%x;;%X;ignored;1;1;1;;comment", sha1.Sum(subject.Bytes), key.N.Bytes())
	db := newDatabase(DefaultLimits(), FunctionalityLevel)
	if err := db.addCertificateLine(crb); err != nil {
		t.Fatal(err)
	}
	validAt := time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)
	entries, err := db.verifiedCatalogEntries(cat, validAt)
	if err != nil || len(entries) != 2 {
		t.Fatalf("genuine signed catalog at signer-valid time: entries=%d err=%v", len(entries), err)
	}
	want := map[int64]string{
		1: "feb832e327bf7c117593e450c471f5bdd2c7ca8db4ddf6b12247800bde857239",
		2: "b5ba83fac9b6eb5221b62278f65f585087bbb10a259114bb415a87c425972621",
	}
	for _, entry := range entries {
		if got := hex.EncodeToString(entry.digest); got != want[entry.kind] {
			t.Fatalf("unexpected signed member kind=%d digest=%s", entry.kind, got)
		}
	}
	if _, err := db.verifiedCatalogEntries(cat, time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("expired signer was trusted")
	}
	changed := bytes.Clone(cat)
	memberOffset := bytes.Index(changed, entries[0].digest)
	if memberOffset < 0 {
		t.Fatal("signed member digest not found in catalog bytes")
	}
	changed[memberOffset] ^= 1
	if _, err := db.verifiedCatalogEntries(changed, validAt); err == nil {
		t.Fatal("tampered catalog was trusted")
	}
	withoutTrust := newDatabase(DefaultLimits(), FunctionalityLevel)
	if _, err := withoutTrust.verifiedCatalogEntries(cat, validAt); err == nil {
		t.Fatal("catalog without CRB trust was accepted")
	}
	if time.Now().UTC().Before(signer.NotAfter) {
		if _, err := Compile(map[string][]byte{
			"igfilter.cat": cat,
			"trust.crb":    []byte(crb + "\n"),
			"ref.ndb":      []byte("Reference.MZ:0:0:4d5a90000300000004000000\n"),
		}); err != nil {
			t.Fatalf("current production compiler rejected valid catalog: %v", err)
		}
	}
}
