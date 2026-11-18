package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func policyBlockRow(t *testing.T, cert *x509.Certificate, name, kind string) string {
	t.Helper()
	subject := primaryFields(t, cert.RawSubject)[0]
	serial := primaryFields(t, primaryFields(t, cert.RawTBSCertificate)[0].Bytes)[1].Bytes
	subjectHash, serialHash := fmt.Sprintf("%x", sha1.Sum(subject.Bytes)), fmt.Sprintf("%x", sha1.Sum(serial))
	modulus := cert.PublicKey.(*rsa.PublicKey).N.Bytes()
	flags, before := "1;1;1", ""
	switch kind {
	case "wildcard":
		serialHash = ""
	case "other-serial":
		serialHash = fmt.Sprintf("%x", sha1.Sum(append([]byte{0}, serial...)))
	case "other-subject":
		subjectHash = strings.Repeat("00", 20)
	case "other-key":
		modulus = bytes.Clone(modulus)
		modulus[len(modulus)-1] ^= 2
	case "no-roles":
		flags = "0;0;0"
	case "later-before":
		before = "100"
	case "exact":
	default:
		t.Fatal(kind)
	}
	return fmt.Sprintf("%s;0;%s;%s;%X;ignored;%s;%s;comment\n", name, subjectHash, serialHash, modulus, flags, before)
}

func policyDatedCertificate(t *testing.T, cert, issuer *x509.Certificate, key *rsa.PrivateKey, hash crypto.Hash, kind string) *x509.Certificate {
	t.Helper()
	template := *cert
	template.SignatureAlgorithm = 0
	switch kind {
	case "primary-expired":
		template.NotAfter = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	case "parent-expired", "signer-expired":
		template.NotAfter = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	case "parent-future", "signer-future":
		template.NotBefore = time.Date(2035, 1, 1, 0, 0, 0, 0, time.UTC)
	case "parent-pre-epoch":
		template.NotBefore = time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
		template.NotAfter = time.Date(1961, 1, 1, 0, 0, 0, 0, time.UTC)
	case "parent-start-pre-epoch":
		template.NotBefore = time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)
	case "parent-zero-interval":
		template.NotAfter = template.NotBefore
	default:
		t.Fatal(kind)
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, issuer, cert.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return historicalCertificate(t, parsed, key, hash)
}

func TestAuthenticodeRuleSelectionAndIssuerDates(t *testing.T) {
	kinds := []string{"normal", "exact-twice", "wildcard-twice", "wildcard-then-exact", "exact-then-wildcard", "other-serial-then-exact", "exact-then-other-serial", "other-key-then-exact", "other-subject-then-exact", "duplicate-no-roles", "duplicate-later-before", "split-exact-twice", "split-wildcard-then-exact", "split-source-reset", "parent-expired", "parent-future", "parent-pre-epoch", "parent-start-pre-epoch", "parent-zero-interval", "signer-expired", "signer-future", "tampered-parent", "untrusted-root", "unnamed-block", "unnamed-first-duplicate", "unnamed-second-duplicate"}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, depth := range []int{1, 3} {
			for _, embedded := range []bool{false, true} {
				for _, surface := range []string{"pe", "catalog", "legacy-time"} {
					role := x509.ExtKeyUsageCodeSigning
					if surface == "legacy-time" {
						role = x509.ExtKeyUsageTimeStamping
					}
					base := historicalPEChainFixture(t, hash, depth, embedded, role)
					var primaryBody, primaryDER []byte
					var codeCRB string
					if surface == "legacy-time" {
						code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
						cert := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
						primaryBody, primaryDER = signedPEWithSigner(t, cert, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, cert.Raw, 1))
						codeCRB = code.crb
					}
					for _, kind := range kinds {
						label := fmt.Sprintf("policy-%s-%s-depth%d-root%v", surface, hash, depth, embedded)
						t.Run(label+"/"+kind, func(t *testing.T) {
							chain := base
							want := true
							wantName := ""
							inputs := map[string][]byte{"trust.crb": []byte(codeCRB + chain.crb), "raw.ndb": []byte("Raw.Marker:0:*:9090909090\n")}
							if strings.HasPrefix(kind, "parent-") {
								changed := policyDatedCertificate(t, base.firstIntermediate, chain.root, chain.rootKey, hash, kind)
								chain.bag = bytes.Replace(chain.bag, base.firstIntermediate.Raw, changed.Raw, 1)
								if kind == "parent-zero-interval" {
									want = false
									wantName = "Raw.Marker"
								}
							} else if strings.HasPrefix(kind, "signer-") {
								chain.leaf = policyDatedCertificate(t, base.leaf, chain.parent, chain.parentKey, hash, kind)
								chain.bag = bytes.Replace(chain.bag, base.leaf.Raw, chain.leaf.Raw, 1)
								want = false
								wantName = "Raw.Marker"
								if surface == "legacy-time" {
									want, wantName = true, ""
								}
							} else if kind == "tampered-parent" {
								changed := bytes.Clone(base.firstIntermediate.Raw)
								changed[len(changed)-1] ^= 1
								chain.bag = bytes.Replace(chain.bag, base.firstIntermediate.Raw, changed, 1)
								want = false
								wantName = "Raw.Marker"
							} else if kind == "untrusted-root" {
								inputs["trust.crb"] = []byte(codeCRB + replaceHistoricalCRBField(t, chain.crb, 2, strings.Repeat("0", 40)))
								want = false
								wantName = "Raw.Marker"
							} else if kind != "normal" {
								first, second := "exact", "exact"
								want = false
								wantName = "Block.First"
								switch kind {
								case "wildcard-twice":
									first, second = "wildcard", "wildcard"
								case "wildcard-then-exact", "split-wildcard-then-exact":
									first = "wildcard"
								case "exact-then-wildcard":
									second = "wildcard"
									wantName = "Block.Second"
								case "other-serial-then-exact":
									first = "other-serial"
									wantName = "Block.Second"
								case "exact-then-other-serial":
									second = "other-serial"
								case "other-key-then-exact":
									first = "other-key"
									wantName = "Block.Second"
								case "other-subject-then-exact":
									first = "other-subject"
									wantName = "Block.Second"
								case "duplicate-no-roles":
									second = "no-roles"
								case "duplicate-later-before":
									second = "later-before"
								case "split-source-reset":
									first = "other-key"
									second = "other-serial"
									want = true
									wantName = ""
								case "unnamed-block":
									second = "other-key"
									wantName = "(unnamed CRB rule)"
								case "unnamed-first-duplicate":
									wantName = "(unnamed CRB rule)"
								}
								firstName, secondName := "Block.First", "Block.Second"
								if kind == "unnamed-block" || kind == "unnamed-first-duplicate" {
									firstName = ""
								}
								if kind == "unnamed-second-duplicate" {
									secondName = ""
								}
								a, b := policyBlockRow(t, chain.leaf, firstName, first), policyBlockRow(t, chain.leaf, secondName, second)
								if kind == "split-source-reset" {
									a = strings.Replace(a, strings.Split(a, ";")[3]+";", ";", 1)
								}
								if strings.HasPrefix(kind, "split-") {
									inputs["a-first.crb"], inputs["b-second.crb"] = []byte(a), []byte(b)
								} else {
									inputs["trust.crb"] = append([]byte(a+b), inputs["trust.crb"]...)
								}
								if surface == "catalog" && !want {
									wantName = "Raw.Marker"
								}
								if strings.HasPrefix(kind, "unnamed-") {
									inputs["raw.ndb"] = []byte("Decoy:0:*:6861726d6c6573732d6e657665722d70726573656e742d6465636f79\n")
									if surface == "catalog" {
										want, wantName = true, ""
									}
								}
							}
							body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
							if surface == "catalog" {
								body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
								digest, err := peCatalogDigest(body, crypto.SHA256)
								if err != nil {
									t.Fatal(err)
								}
								inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, chain.bag)
							}
							if surface == "legacy-time" {
								packet := appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, chain.bag)
								body = replacePECertificate(t, primaryBody, packet)
							}
							db, err := signatures.Compile(inputs)
							if err != nil {
								t.Fatal(err)
							}
							e := New(config.EDRConfig{})
							e.current.Store(&generation{db: db})
							clean, name, scanErr := e.Scan(context.Background(), body)
							exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
							if root := os.Getenv("EDR_HISTORICAL_CHAIN_EXPORT"); root != "" {
								p := filepath.Join(root, label, kind, "go.json")
								data, err := os.ReadFile(p)
								if err != nil {
									t.Fatal(err)
								}
								var row map[string]any
								if err = json.Unmarshal(data, &row); err != nil {
									t.Fatal(err)
								}
								row["expected_name"] = wantName
								data, err = json.MarshalIndent(row, "", "  ")
								if err != nil {
									t.Fatal(err)
								}
								if err = os.WriteFile(p, data, 0600); err != nil {
									t.Fatal(err)
								}
							}
							if scanErr != nil || clean != want || name != wantName {
								t.Errorf("policy: clean=%v name=%q err=%v want=%v/%q", clean, name, scanErr, want, wantName)
							}
						})
					}
				}
			}
		}
	}
}
