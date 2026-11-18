package engine

import (
	"bytes"
	"context"
	"crypto"
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
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Supply two files extracted from the public DsHidMini v3.27.0 MSI. Neither
// third-party binary is committed to the repository.
func TestRealCurrentCatalogExemptsOnlyCoveredDriver(t *testing.T) {
	catPath, sysPath := os.Getenv("EDR_DSHIDMINI_CATALOG"), os.Getenv("EDR_DSHIDMINI_DRIVER")
	if catPath == "" && sysPath == "" {
		t.Skip("set EDR_DSHIDMINI_CATALOG and EDR_DSHIDMINI_DRIVER to pinned DsHidMini v3.27.0 files")
	}
	if catPath == "" || sysPath == "" {
		t.Fatal("both DsHidMini fixture paths are required")
	}
	cat, err := os.ReadFile(catPath)
	if err != nil {
		t.Fatal(err)
	}
	driver, err := os.ReadFile(sysPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat) != 11721 || fmt.Sprintf("%x", sha256.Sum256(cat)) != "9b6921c0662513a0e57c5025e29268f728cfda31a34a76e26013282c671adc3b" ||
		len(driver) != 56552 || fmt.Sprintf("%x", sha256.Sum256(driver)) != "49896c287dd29048e09af86f6b33adb7751b994b631b629da50535558cd987a8" {
		t.Fatal("DsHidMini fixture size or SHA-256 changed")
	}
	imageDigest, err := peCatalogDigest(driver, crypto.SHA256)
	if err != nil || hex.EncodeToString(imageDigest) != "b5ba83fac9b6eb5221b62278f65f585087bbb10a259114bb415a87c425972621" {
		t.Fatalf("driver PE image digest differs from genuine catalog: %x %v", imageDigest, err)
	}
	p7, err := pkcs7.Parse(cat)
	if err != nil || len(p7.Signers) != 1 {
		t.Fatalf("catalog PKCS#7 parse: %v", err)
	}
	if err := p7.Verify(); err != nil {
		t.Fatal(err)
	}
	signer := p7.GetOnlySigner()
	if signer == nil {
		t.Fatal("missing catalog signer")
	}
	if !time.Now().UTC().Before(signer.NotAfter) {
		t.Skip("genuine catalog signer has expired; historical signature verification remains covered by the signatures package")
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(signer.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	key, ok := signer.PublicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatal("catalog signer is not RSA")
	}
	crb := fmt.Sprintf("Trusted.DsHidMini;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
	rules := map[string][]byte{
		"trust.crb": []byte(crb),
		"ref.ndb":   []byte("Reference.MZ:0:0:4d5a90000300000004000000\n"),
	}
	check := func(withCatalog bool, object []byte, wantClean bool) {
		t.Helper()
		selected := map[string][]byte{"trust.crb": rules["trust.crb"], "ref.ndb": rules["ref.ndb"]}
		if withCatalog {
			selected["igfilter.cat"] = cat
		}
		db, err := signatures.Compile(selected)
		if err != nil {
			t.Fatal(err)
		}
		e := New(config.EDRConfig{})
		e.current.Store(&generation{db: db})
		clean, sig, err := e.Scan(context.Background(), object)
		if err != nil || clean != wantClean || wantClean && sig != "" || !wantClean && sig != "Reference.MZ" {
			t.Fatalf("catalog=%v: clean=%v sig=%q err=%v; want clean=%v", withCatalog, clean, sig, err, wantClean)
		}
	}
	check(false, driver, false)
	check(true, driver, true)
	changed := bytes.Clone(driver)
	changed[0x200] ^= 1
	changedDigest, err := peCatalogDigest(changed, crypto.SHA256)
	if err != nil || bytes.Equal(changedDigest, imageDigest) {
		t.Fatal("changed code kept the signed catalog digest")
	}
	check(true, changed, false)
}
