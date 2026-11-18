package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func replaceHistoricalCRBField(t *testing.T, line string, index int, value string) string {
	t.Helper()
	fields := strings.Split(strings.TrimSpace(line), ";")
	fields[index] = value
	return strings.Join(fields, ";") + "\n"
}

// The fixture signs the actual TBSCertificate bytes, without using a global
// GODEBUG policy to enable historical certificate algorithms in production.
func historicalCertificate(t *testing.T, cert *x509.Certificate, issuerKey *rsa.PrivateKey, hash crypto.Hash) *x509.Certificate {
	t.Helper()
	var outer asn1.RawValue
	if _, err := asn1.Unmarshal(cert.RawTBSCertificate, &outer); err != nil {
		t.Fatal(err)
	}
	var fields [][]byte
	for rest := outer.Bytes; len(rest) > 0; {
		var field asn1.RawValue
		next, err := asn1.Unmarshal(rest, &field)
		if err != nil {
			t.Fatal(err)
		}
		fields = append(fields, field.FullBytes)
		rest = next
	}
	oid := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}
	switch hash {
	case crypto.SHA256:
		oid = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}
	case crypto.MD5:
		oid = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 4}
	case crypto.SHA384:
		oid = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}
	case crypto.SHA512:
		oid = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}
	}
	algorithm := derFixture(t, asn1.ClassUniversal, asn1.TagSequence, oidFixture(t, oid), []byte{5, 0})
	index := 1
	if fields[0][0] == 0xa0 {
		index = 2
	}
	fields[index] = algorithm
	tbs := derFixture(t, asn1.ClassUniversal, asn1.TagSequence, fields...)
	h := hash.New()
	h.Write(tbs)
	sig, err := signReferenceFixture(rand.Reader, issuerKey, hash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	bits, err := asn1.Marshal(asn1.BitString{Bytes: sig, BitLength: 8 * len(sig)})
	if err != nil {
		t.Fatal(err)
	}
	der := derFixture(t, asn1.ClassUniversal, asn1.TagSequence, tbs, algorithm, bits)
	out, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

type historicalPEChain struct {
	root, parent, leaf, firstIntermediate *x509.Certificate
	rootKey, parentKey, leafKey           *rsa.PrivateKey
	bag                                   []byte
	crb                                   string
}

func historicalPEChainFixture(t *testing.T, hash crypto.Hash, depth int, embeddedRoot bool, role x509.ExtKeyUsage) historicalPEChain {
	t.Helper()
	makeCert := func(serial int64, name string, parent *x509.Certificate, parentKey *rsa.PrivateKey, ca bool) (*x509.Certificate, *rsa.PrivateKey) {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		usage := x509.KeyUsageDigitalSignature
		if ca {
			usage |= x509.KeyUsageCertSign
		}
		template := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: name}, NotBefore: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), NotAfter: time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC), KeyUsage: usage, ExtKeyUsage: []x509.ExtKeyUsage{role}, BasicConstraintsValid: true, IsCA: ca}
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
		return historicalCertificate(t, cert, parentKey, hash), key
	}
	root, rootKey := makeCert(101, "Historical root", nil, nil, true)
	parent, parentKey := root, rootKey
	var bag []byte
	var firstIntermediate *x509.Certificate
	if embeddedRoot {
		bag = append(bag, root.Raw...)
	}
	for i := 0; i < depth; i++ {
		cert, key := makeCert(int64(200+i), fmt.Sprintf("Historical intermediate %d", i), parent, parentKey, true)
		bag = append(bag, cert.Raw...)
		if firstIntermediate == nil {
			firstIntermediate = cert
		}
		parent, parentKey = cert, key
	}
	leaf, leafKey := makeCert(301, "Historical signer", parent, parentKey, false)
	bag = append(bag, leaf.Raw...)
	var subject asn1.RawValue
	if _, err := asn1.Unmarshal(root.RawSubject, &subject); err != nil {
		t.Fatal(err)
	}
	crb := fmt.Sprintf("Trusted.Historical;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), rootKey.N.Bytes())
	if role == x509.ExtKeyUsageTimeStamping {
		crb = replaceHistoricalCRBField(t, crb, 6, "0")
	}
	return historicalPEChain{root: root, parent: parent, leaf: leaf, firstIntermediate: firstIntermediate, rootKey: rootKey, parentKey: parentKey, leafKey: leafKey, bag: bag, crb: crb}
}

func TestHistoricalAuthenticodeChains(t *testing.T) {
	// Each case has genuinely signed certificates and PKCS#7/PE image data.
	// An optional output root exports only public certificates, PE bytes and
	// rules so an independent pinned scanner can evaluate identical fixtures.
	export := os.Getenv("EDR_HISTORICAL_CHAIN_EXPORT")
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512, crypto.SHA1, crypto.MD5} {
		for _, depth := range []int{0, 1, 3} {
			for _, embeddedRoot := range []bool{false, true} {
				label := fmt.Sprintf("%s-depth%d-root%v", hash, depth, embeddedRoot)
				t.Run(label, func(t *testing.T) {
					chain := historicalPEChainFixture(t, hash, depth, embeddedRoot, x509.ExtKeyUsageCodeSigning)
					leaf, firstIntermediate, crb := chain.leaf, chain.firstIntermediate, chain.crb
					file, der := signedPEWithSigner(t, leaf, chain.leafKey, chain.bag)
					for _, kind := range []string{"verified", "tampered-image", "tampered-certificate", "tampered-intermediate", "untrusted-root", "no-cert-permission", "wrong-code-role", "expired-signer"} {
						t.Run(kind, func(t *testing.T) {
							body := bytes.Clone(file)
							signature := bytes.Clone(der)
							rules := crb
							at := time.Now().UTC()
							if kind == "tampered-image" {
								body[0x200] ^= 1
							}
							if kind == "tampered-certificate" || kind == "tampered-intermediate" {
								target := leaf
								if kind == "tampered-intermediate" && firstIntermediate != nil {
									target = firstIntermediate
								}
								offset := bytes.Index(signature, target.Raw)
								if offset < 0 {
									t.Fatal("certificate missing from fixture")
								}
								signature[offset+len(target.Raw)-1] ^= 1
								body = replacePECertificate(t, file, signature)
							}
							if kind == "wrong-code-role" {
								rules = replaceHistoricalCRBField(t, crb, 6, "0")
							}
							if kind == "untrusted-root" {
								rules = ""
							}
							if kind == "no-cert-permission" {
								rules = replaceHistoricalCRBField(t, crb, 8, "0")
							}
							if kind == "expired-signer" {
								at = time.Date(2041, 1, 1, 0, 0, 0, 0, time.UTC)
							}
							db, err := signatures.Compile(map[string][]byte{"trust.crb": []byte(rules), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")})
							if err != nil {
								t.Fatal(err)
							}
							verified, verifyErr := db.VerifyTrustedAuthenticode(signature, func(h crypto.Hash) ([]byte, error) { return peAuthenticodeDigest(body, h) }, at)
							want := kind == "verified"
							if want != verified || want && verifyErr != nil {
								t.Errorf("historical chain: trusted=%v err=%v want=%v", verified, verifyErr, want)
							}
							if kind == "expired-signer" {
								return
							} // reference CLI uses wall-clock time
							e := New(config.EDRConfig{})
							e.current.Store(&generation{db: db})
							clean, name, err := e.Scan(context.Background(), body)
							if clean != want || err != nil || !want && name != "Suspicious.Code" {
								t.Errorf("engine result: clean=%v name=%q err=%v want clean=%v", clean, name, err, want)
							}
							if export != "" {
								dir := filepath.Join(export, label, kind)
								if err := os.MkdirAll(dir, 0700); err != nil {
									t.Fatal(err)
								}
								for name, data := range map[string][]byte{"sample.exe": body, "trust.crb": []byte(rules), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")} {
									if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
										t.Fatal(err)
									}
								}
								result, _ := json.Marshal(map[string]any{"trusted": verified, "clean": clean, "name": name, "expected_clean": want, "certificate_hash": hash.String(), "intermediates": depth, "embedded_root": embeddedRoot})
								if err := os.WriteFile(filepath.Join(dir, "go.json"), result, 0600); err != nil {
									t.Fatal(err)
								}
							}
						})
					}
				})
			}
		}
	}
}

func exportHistoricalScan(t *testing.T, label, kind string, body []byte, inputs map[string][]byte, clean bool, name string, want bool) {
	t.Helper()
	root := os.Getenv("EDR_HISTORICAL_CHAIN_EXPORT")
	if root == "" {
		return
	}
	dir := filepath.Join(root, label, kind)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for name, data := range inputs {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "sample.exe"), body, 0600); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(map[string]any{"clean": clean, "name": name, "expected_clean": want, "case": kind})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestHistoricalAuthenticodeCatalogsAndTimestamps(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA384, crypto.SHA512, crypto.SHA1, crypto.MD5} {
		for _, depth := range []int{0, 1, 3} {
			for _, embeddedRoot := range []bool{false, true} {
				label := fmt.Sprintf("%s-depth%d-root%v", hash, depth, embeddedRoot)
				t.Run(label, func(t *testing.T) {
					chain := historicalPEChainFixture(t, hash, depth, embeddedRoot, x509.ExtKeyUsageCodeSigning)
					check := func(t *testing.T, kind string, body []byte, inputs map[string][]byte, want bool) {
						t.Helper()
						db, err := signatures.Compile(inputs)
						if err != nil {
							t.Fatal(err)
						}
						e := New(config.EDRConfig{})
						e.current.Store(&generation{db: db})
						clean, name, err := e.Scan(context.Background(), body)
						if clean != want || err != nil || !want && name != "Suspicious.Code" {
							t.Errorf("%s: clean=%v name=%q err=%v want=%v", kind, clean, name, err, want)
						}
						exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
					}
					unsigned := simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
					digest, err := peCatalogDigest(unsigned, crypto.SHA256)
					if err != nil {
						t.Fatal(err)
					}
					cat := signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, chain.bag)
					for _, kind := range []string{"catalog-verified", "catalog-tampered-image", "catalog-tampered-signature"} {
						t.Run(kind, func(t *testing.T) {
							body, catalog := bytes.Clone(unsigned), bytes.Clone(cat)
							if kind == "catalog-tampered-image" {
								body[0x200] ^= 1
							}
							if kind == "catalog-tampered-signature" {
								catalog[len(catalog)-1] ^= 1
							}
							check(t, kind, body, map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n"), "historical.cat": catalog}, kind == "catalog-verified")
						})
					}
					if depth > 0 {
						// The member digest deliberately does not cover this PE.
						// Only a verified catalog's authenticated intermediate can
						// supply the missing issuer for this embedded leaf signature.
						catalog := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20), chain.parent, chain.parentKey, chain.bag)
						pe, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.leaf.Raw)
						for _, kind := range []string{"catalog-certificate-store", "catalog-invalid-store"} {
							t.Run(kind, func(t *testing.T) {
								cat := bytes.Clone(catalog)
								if kind == "catalog-invalid-store" {
									cat[len(cat)-1] ^= 1
								}
								check(t, kind, pe, map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n"), "historical.cat": cat}, true)
							})
						}
					}
					// The primary signer expired in 2025. Only a valid legacy
					// countersignature at 2024 may keep it trusted at scan time.
					expiredTemplate := *chain.leaf
					expiredTemplate.NotAfter = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
					expiredTemplate.SignatureAlgorithm = x509.SHA256WithRSA
					raw, err := x509.CreateCertificate(rand.Reader, &expiredTemplate, chain.parent, &chain.leafKey.PublicKey, chain.parentKey)
					if err != nil {
						t.Fatal(err)
					}
					expired, err := x509.ParseCertificate(raw)
					if err != nil {
						t.Fatal(err)
					}
					expired = historicalCertificate(t, expired, chain.parentKey, hash)
					bag := bytes.Replace(chain.bag, chain.leaf.Raw, expired.Raw, 1)
					pe, der := signedPEWithSigner(t, expired, chain.leafKey, bag)
					tsa := historicalPEChainFixture(t, hash, depth, embeddedRoot, x509.ExtKeyUsageTimeStamping)
					timestamped := appendLegacyTimestampWithSigner(t, der, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tsa.leaf, tsa.leafKey, tsa.bag)
					for _, kind := range []string{"legacy-time-verified", "legacy-time-tampered", "legacy-time-wrong-role", "legacy-time-missing"} {
						t.Run(kind, func(t *testing.T) {
							signature := bytes.Clone(timestamped)
							timeCRB := tsa.crb
							if kind == "legacy-time-tampered" {
								signature[len(signature)-1] ^= 1
							}
							if kind == "legacy-time-wrong-role" {
								timeCRB = replaceHistoricalCRBField(t, timeCRB, 7, "0")
							}
							if kind == "legacy-time-missing" {
								signature = der
							}
							body := replacePECertificate(t, pe, signature)
							check(t, kind, body, map[string][]byte{"trust.crb": []byte(chain.crb + timeCRB), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}, kind == "legacy-time-verified")
						})
					}
					if depth > 0 {
						missing := appendLegacyTimestampWithSigner(t, der, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tsa.leaf, tsa.leafKey, tsa.leaf.Raw)
						body := replacePECertificate(t, pe, missing)
						cat := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20), chain.parent, chain.parentKey, append(bytes.Clone(chain.bag), tsa.bag...))
						for _, kind := range []string{"legacy-time-catalog", "legacy-time-invalid-catalog", "legacy-time-catalog-untrusted"} {
							t.Run(kind, func(t *testing.T) {
								catalog := bytes.Clone(cat)
								rules := chain.crb + tsa.crb
								if kind == "legacy-time-invalid-catalog" {
									catalog[len(catalog)-1] ^= 1
								}
								if kind == "legacy-time-catalog-untrusted" {
									rules = chain.crb
								}
								check(t, kind, body, map[string][]byte{"trust.crb": []byte(rules), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n"), "historical.cat": catalog}, kind != "legacy-time-catalog-untrusted")
							})
						}
					}
				})
			}
		}
	}
}

func TestCatalogCertificateRolesCannotExpandAuthority(t *testing.T) {
	chain := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
	parentTemplate := *chain.parent
	parentTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageTimeStamping}
	raw, err := x509.CreateCertificate(rand.Reader, &parentTemplate, chain.root, &chain.parentKey.PublicKey, chain.rootKey)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := x509.ParseCertificate(raw)
	if err != nil {
		t.Fatal(err)
	}
	timeTemplate := *chain.leaf
	timeTemplate.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping}
	raw, err = x509.CreateCertificate(rand.Reader, &timeTemplate, parent, &chain.leafKey.PublicKey, chain.parentKey)
	if err != nil {
		t.Fatal(err)
	}
	tsa, err := x509.ParseCertificate(raw)
	if err != nil {
		t.Fatal(err)
	}
	file, codeCRB, der := signedPEFixture(t, false)
	timestamped := appendLegacyTimestampWithSigner(t, der, time.Now().UTC().Truncate(time.Second), tsa, chain.leafKey, tsa.Raw)
	body := replacePECertificate(t, file, timestamped)
	bag := append(bytes.Clone(parent.Raw), chain.root.Raw...)
	catalog := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20), parent, chain.parentKey, bag)
	for _, kind := range []string{"catalog-code-only", "catalog-code-and-time"} {
		t.Run(kind, func(t *testing.T) {
			timeFlag := "0"
			want := false
			if kind == "catalog-code-and-time" {
				timeFlag = "1"
				want = true
			}
			inputs := map[string][]byte{"trust.crb": []byte(codeCRB + replaceHistoricalCRBField(t, chain.crb, 7, timeFlag)), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n"), "roles.cat": catalog}
			db, err := signatures.Compile(inputs)
			if err != nil {
				t.Fatal(err)
			}
			e := New(config.EDRConfig{})
			e.current.Store(&generation{db: db})
			clean, name, err := e.Scan(context.Background(), body)
			if clean != want || err != nil {
				t.Errorf("catalog expanded authority: clean=%v name=%q err=%v want clean=%v", clean, name, err, want)
			}
			exportHistoricalScan(t, "roles", kind, body, inputs, clean, name, want)
		})
	}
}
