package signatures

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
)

func detachedFixture(t *testing.T) ([]byte, []byte, []byte) {
	t.Helper()
	now := time.Now()
	rootKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test root"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatal(err)
	}
	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	leafTemplate := &x509.Certificate{SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: "test signer"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, root, &leafKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatal(err)
	}
	data := append([]byte("ClamAV-VDB:test"), make([]byte, 512)...)
	sd, err := pkcs7.NewSignedData(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := sd.AddSignerChain(leaf, leafKey, []*x509.Certificate{root}, pkcs7.SignerInfoConfig{}); err != nil {
		t.Fatal(err)
	}
	sd.Detach()
	der, err := sd.Finish()
	if err != nil {
		t.Fatal(err)
	}
	sign := []byte("#clamsign-1.0\n90::pkcs7-pem:" + base64.StdEncoding.EncodeToString(der) + "\n")
	rootPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: rootDER})
	return data, sign, rootPEM
}

func TestDetachedCVDVerification(t *testing.T) {
	data, sign, root := detachedFixture(t)
	now := time.Now()
	if err := verifyDetachedCVD(data, sign, root, FunctionalityLevel, now, ""); err != nil {
		t.Fatal(err)
	}
	if err := verifyDetachedCVD(data, sign, root, FunctionalityLevel, now, "ClamAV_datafiles_release"); err == nil {
		t.Fatal("unexpected release signer identity accepted")
	}
	badData := append([]byte(nil), data...)
	badData[len(badData)-1] ^= 1
	if err := verifyDetachedCVD(badData, sign, root, FunctionalityLevel, now, ""); err == nil {
		t.Fatal("altered CVD verified")
	}
	wrongRootData, _, wrongRoot := detachedFixture(t)
	_ = wrongRootData
	if err := verifyDetachedCVD(data, sign, wrongRoot, FunctionalityLevel, now, ""); err == nil {
		t.Fatal("untrusted signer verified")
	}
	if err := VerifyOfficialDetachedCVD(data, sign); err == nil {
		t.Fatal("test signer verified as official")
	}
	for _, modified := range []string{
		strings.Replace(string(sign), "#clamsign-1.0", "#clamsign-2.0", 1),
		strings.Replace(string(sign), "90::", "235::", 1),
		strings.Replace(string(sign), "90::", "90:100:", 1),
		strings.Replace(string(sign), "pkcs7-pem:", "unknown:", 1),
	} {
		if err := verifyDetachedCVD(data, []byte(modified), root, FunctionalityLevel, now, ""); err == nil {
			t.Fatal("invalid signature row verified")
		}
	}
}

func TestClamAVUpstreamDetachedFixture(t *testing.T) {
	dir := filepath.Join("testdata", "cvddetached")
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	data, sign, root := read("test-1.cvd"), read("test-1.cvd.sign"), read("clamav-test-root.crt")
	// Upstream test signatures require feature level 230, now within the
	// pinned 1.5.4 runtime's 234 window.
	if err := verifyDetachedCVD(data, sign, root, FunctionalityLevel, time.Now(), ""); err != nil {
		t.Fatal(err)
	}
	if err := verifyDetachedCVD(data, sign, root, 229, time.Now(), ""); err == nil {
		t.Fatal("out-of-window signature accepted")
	}
	data[len(data)-1] ^= 1
	if err := verifyDetachedCVD(data, sign, root, FunctionalityLevel, time.Now(), ""); err == nil {
		t.Fatal("tampered upstream CVD verified")
	}
}
