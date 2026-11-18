package signatures

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"
)

func historicalChainFixture(t *testing.T, kind string) (*Database, []*x509.Certificate, time.Time) {
	t.Helper()
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	var certs []*x509.Certificate
	for i := 0; i < 5; i++ {
		ca := i < 4
		usage := x509.KeyUsageDigitalSignature
		if ca {
			usage |= x509.KeyUsageCertSign
		}
		c := &x509.Certificate{SerialNumber: big.NewInt(int64(i + 1)), Subject: pkix.Name{CommonName: fmt.Sprintf("Historical chain %d", i)}, NotBefore: at.Add(-24 * time.Hour), NotAfter: at.Add(24 * time.Hour), KeyUsage: usage, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageTimeStamping}, BasicConstraintsValid: true, IsCA: ca, SignatureAlgorithm: x509.SHA1WithRSA}
		if i == 1 {
			switch kind {
			case "non-ca":
				c.IsCA = false
				c.KeyUsage = x509.KeyUsageDigitalSignature
			case "no-cert-usage":
				c.KeyUsage = x509.KeyUsageDigitalSignature
			case "path-limit":
				c.MaxPathLen = 0
				c.MaxPathLenZero = true
			case "path-valid":
				c.MaxPathLen = 2
			case "name-constraints":
				c.PermittedDNSDomainsCritical = true
				c.PermittedDNSDomains = []string{"example.test"}
			case "unknown-critical":
				c.ExtraExtensions = []pkix.Extension{{Id: asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6}, Critical: true, Value: []byte{5, 0}}}
			case "wrong-code-role":
				c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping}
			case "wrong-time-role":
				c.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}
			case "expired-parent":
				c.NotAfter = at.Add(-time.Hour)
			}
		}
		parent := c
		if i > 0 {
			parent = certs[i-1]
		}
		der, err := x509.CreateCertificate(rand.Reader, c, parent, &key.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		certs = append(certs, cert)
	}
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(certs[0].RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	crb := fmt.Sprintf("Trusted.Historical;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
	if kind == "blocked-parent" {
		if _, err := asn1.Unmarshal(certs[2].RawSubject, &subject); err != nil {
			t.Fatal(err)
		}
		crb += fmt.Sprintf("Blocked.Historical;0;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
	}
	d, err := Compile(map[string][]byte{"trust.crb": []byte(crb)})
	if err != nil {
		t.Fatal(err)
	}
	return d, certs, at
}

func TestHistoricalAuthenticodeChainConstraints(t *testing.T) {
	for _, kind := range []string{"valid", "path-valid", "non-ca", "no-cert-usage", "path-limit", "name-constraints", "unknown-critical", "wrong-code-role", "wrong-time-role", "expired-parent", "blocked-parent"} {
		t.Run(kind, func(t *testing.T) {
			d, certs, at := historicalChainFixture(t, kind)
			for _, timeSigning := range []bool{false, true} {
				want := kind == "valid" || kind == "path-valid" || kind == "expired-parent" || kind == "path-limit" || kind == "name-constraints" || kind == "unknown-critical" || kind == "no-cert-usage" || kind == "wrong-code-role" && timeSigning || kind == "wrong-time-role" && !timeSigning
				chain := d.historicalAuthenticodeChain(certs[4], certs, at, timeSigning)
				if (len(chain) > 0) != want {
					t.Fatalf("timeSigning=%v path=%d want valid=%v", timeSigning, len(chain), want)
				}
				if want {
					// Reordering and duplicate bags cannot change authentication.
					bag := []*x509.Certificate{certs[4], certs[2], certs[2], certs[0], certs[3], certs[1]}
					if len(d.historicalAuthenticodeChain(certs[4], bag, at, timeSigning)) == 0 {
						t.Fatal("order/duplicate-dependent trust")
					}
				}
			}
		})
	}
}

func TestHistoricalAuthenticodeChainBoundsUnrootedDenseGraph(t *testing.T) {
	at := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	var certs []*x509.Certificate
	for i := 0; i < 14; i++ {
		// Same subject/key/issuer with distinct serials: every signed edge
		// verifies, but no path is rooted. Traversal must be bounded even
		// after all signature checks have been cached.
		c := &x509.Certificate{SerialNumber: big.NewInt(int64(i + 1)), Subject: pkix.Name{CommonName: "Unrooted cross-signed graph"}, NotBefore: at.Add(-time.Hour), NotAfter: at.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}, BasicConstraintsValid: true, IsCA: true, SignatureAlgorithm: x509.SHA1WithRSA}
		der, err := x509.CreateCertificate(rand.Reader, c, c, &key.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		certs = append(certs, cert)
	}
	d := &Database{}
	start := time.Now()
	if len(d.historicalAuthenticodeChain(certs[0], certs, at, false)) != 0 {
		t.Fatal("unrooted graph trusted")
	}
	if _, err := d.authenticodeChainChecked(certs[0], certs, at, false); err != nil {
		t.Fatal("finite unrooted store closure became uncertain", err)
	}
	if chain, err := d.authenticodeChainChecked(certs[0], certs[:1], at, false); len(chain) != 0 || err != nil {
		t.Fatal("a fully checked untrusted cycle became a resource failure", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("dense graph exceeded bounded traversal time")
	}
	oversized := make([]*x509.Certificate, 4097)
	if len(d.historicalAuthenticodeChain(certs[0], oversized, at, false)) != 0 {
		t.Fatal("unbounded certificate bag accepted")
	}
	if _, err := d.authenticodeChainChecked(certs[0], oversized, at, false); !errors.Is(err, errAuthenticodeLimit) {
		t.Fatal("oversized graph lost its incomplete status", err)
	}
}

func TestCatalogCertificateAdmissionBoundsAndDeduplication(t *testing.T) {
	entry := func(raw []byte) catalogCertificate {
		return catalogCertificate{cert: &x509.Certificate{Raw: raw}, codeSign: true}
	}
	d := &Database{catalogCerts: []catalogCertificate{entry([]byte("existing"))}}
	added, err := d.prepareCatalogCertificates([]catalogCertificate{entry([]byte("existing")), entry([]byte("new")), entry([]byte("new"))})
	if err != nil || len(added) != 2 || len(d.catalogCerts) != 1 {
		t.Fatalf("preparation mutated source or failed deduplication: %d %v", len(added), err)
	}
	prepared, err := d.prepareCatalogCertificates([]catalogCertificate{{cert: d.catalogCerts[0].cert, timeSign: true}})
	if err != nil || len(prepared) != 2 || !prepared[0].codeSign || prepared[0].timeSign || prepared[1].codeSign || !prepared[1].timeSign || d.catalogCerts[0].timeSign {
		t.Fatal("separate authenticated purposes were combined or changed the previous store", err)
	}
	d.catalogCerts = nil
	for i := 0; i < 4096; i++ {
		raw := make([]byte, 8)
		binary.LittleEndian.PutUint64(raw, uint64(i))
		d.catalogCerts = append(d.catalogCerts, entry(raw))
	}
	if _, err := d.prepareCatalogCertificates([]catalogCertificate{entry([]byte("extra"))}); err == nil {
		t.Fatal("certificate count limit ignored")
	}
	if added, err := d.prepareCatalogCertificates([]catalogCertificate{d.catalogCerts[0]}); err != nil || len(added) != 4096 {
		t.Fatal("duplicate consumed a certificate slot", err)
	}
	d.catalogCerts = []catalogCertificate{entry(make([]byte, 8<<20))}
	if _, err := d.prepareCatalogCertificates([]catalogCertificate{entry([]byte("extra"))}); err == nil {
		t.Fatal("certificate byte limit ignored")
	}
}
