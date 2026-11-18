package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func storePurposeCertificate(t *testing.T, cert, issuer *x509.Certificate, key *rsa.PrivateKey, hash crypto.Hash, purposes int) *x509.Certificate {
	t.Helper()
	template := *cert
	template.ExtKeyUsage = nil
	if purposes&1 != 0 {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageCodeSigning)
	}
	if purposes&2 != 0 {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageTimeStamping)
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

func storePurposeRule(t *testing.T, line string, purposes int) string {
	t.Helper()
	line = replaceHistoricalCRBField(t, line, 6, strconv.Itoa(purposes&1))
	return replaceHistoricalCRBField(t, line, 7, strconv.Itoa((purposes>>1)&1))
}

func storeOracle(t *testing.T) func(*testing.T, string, string, []byte, map[string][]byte) {
	t.Helper()
	exportOnly := os.Getenv("EDR_STORE_EXPORT_ONLY") == "1"
	var oracle map[string]bool
	if !exportOnly {
		data, err := os.ReadFile("testdata/authenticode_store.json")
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, &oracle); err != nil {
			t.Fatal(err)
		}
	}
	return func(t *testing.T, label, kind string, body []byte, inputs map[string][]byte) {
		db, err := signatures.Compile(inputs)
		if err != nil {
			t.Fatal(err)
		}
		e := New(config.EDRConfig{})
		e.current.Store(&generation{db: db})
		clean, name, scanErr := e.Scan(context.Background(), body)
		want, exists := oracle[label+"/"+kind]
		exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
		if !exportOnly && (!exists || scanErr != nil || clean != want || !clean && name != "Suspicious.Code") {
			t.Errorf("store policy clean=%v name=%q err=%v want=%v present=%v", clean, name, scanErr, want, exists)
		}
	}
}

func TestAuthenticodeStorePurposeSelection(t *testing.T) {
	check := storeOracle(t)
	rules := []struct {
		name  string
		roles []int
	}{
		{"code", []int{1}}, {"time", []int{2}}, {"both", []int{3}},
		{"code-time", []int{1, 2}}, {"time-code", []int{2, 1}},
		{"both-code", []int{3, 1}}, {"code-both", []int{1, 3}},
		{"both-time", []int{3, 2}}, {"time-both", []int{2, 3}}, {"none", []int{0}},
	}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		chain := historicalPEChainFixture(t, hash, 1, true, x509.ExtKeyUsageCodeSigning)
		root := storePurposeCertificate(t, chain.root, chain.root, chain.rootKey, hash, 3)
		code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
		expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
		primary, primaryDER := signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
		for parentRole := 1; parentRole <= 3; parentRole++ {
			parent := storePurposeCertificate(t, chain.parent, root, chain.rootKey, hash, parentRole)
			for leafRole := 1; leafRole <= 3; leafRole++ {
				leaf := storePurposeCertificate(t, chain.leaf, parent, chain.parentKey, hash, leafRole)
				for _, embedded := range []bool{false, true} {
					for order := 0; order < 3; order++ {
						parts := [][]byte{parent.Raw, leaf.Raw}
						if embedded {
							parts = append([][]byte{root.Raw}, parts...)
						}
						if order == 1 {
							for l, r := 0, len(parts)-1; l < r; l, r = l+1, r-1 {
								parts[l], parts[r] = parts[r], parts[l]
							}
						} else if order == 2 {
							parts = append(parts[1:], parts[0])
						}
						bag := bytes.Join(parts, nil)
						for _, rule := range rules {
							var crb string
							for _, role := range rule.roles {
								crb += storePurposeRule(t, chain.crb, role)
							}
							for _, surface := range []string{"pe", "catalog", "legacy-time"} {
								label := fmt.Sprintf("store-%s-%s-root%v", surface, hash, embedded)
								kind := fmt.Sprintf("%s-parent%d-leaf%d-order%d", rule.name, parentRole, leafRole, order)
								t.Run(label+"/"+kind, func(t *testing.T) {
									inputs := map[string][]byte{"trust.crb": []byte(crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
									body, _ := signedPEWithSigner(t, leaf, chain.leafKey, bag)
									if surface == "catalog" {
										body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
										digest, err := peCatalogDigest(body, crypto.SHA256)
										if err != nil {
											t.Fatal(err)
										}
										inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), leaf, chain.leafKey, bag)
									}
									if surface == "legacy-time" {
										inputs["trust.crb"] = []byte(code.crb + crb)
										body = replacePECertificate(t, primary, appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), leaf, chain.leafKey, bag))
									}
									check(t, label, kind, body, inputs)
								})
							}
						}
					}
				}
			}
		}
	}
}

func TestAuthenticodeStoreCatalogLearning(t *testing.T) {
	check := storeOracle(t)
	arrangements := []struct {
		name string
		bags []string
	}{
		{"code", []string{"a"}}, {"time", []string{"b"}},
		{"code-time", []string{"a", "b"}}, {"time-code", []string{"b", "a"}},
		{"together-code-time", []string{"ab"}}, {"together-time-code", []string{"ba"}},
		{"code-duplicate-time", []string{"a", "a", "b"}}, {"time-duplicate-code", []string{"b", "b", "a"}},
	}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		chain := historicalPEChainFixture(t, hash, 1, true, x509.ExtKeyUsageCodeSigning)
		root := storePurposeCertificate(t, chain.root, chain.root, chain.rootKey, hash, 3)
		other := historicalPEChainFixture(t, hash, 0, true, x509.ExtKeyUsageCodeSigning)
		template := *other.root
		template.Subject = pkix.Name{CommonName: "Store alternate root"}
		template.RawSubject = nil
		der, err := x509.CreateCertificate(rand.Reader, &template, &template, other.root.PublicKey, other.rootKey)
		if err != nil {
			t.Fatal(err)
		}
		otherRoot, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		otherRoot = storePurposeCertificate(t, otherRoot, otherRoot, other.rootKey, hash, 3)
		parentA := storePurposeCertificate(t, chain.parent, root, chain.rootKey, hash, 1)
		parentB := storePurposeCertificate(t, chain.parent, otherRoot, other.rootKey, hash, 2)
		otherCRB := replaceHistoricalCRBField(t, policyBlockRow(t, otherRoot, "Trusted.Other", "wildcard"), 1, "1")
		code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
		expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
		primary, primaryDER := signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
		for leafRole := 1; leafRole <= 3; leafRole++ {
			leaf := storePurposeCertificate(t, chain.leaf, parentA, chain.parentKey, hash, leafRole)
			for _, arrange := range arrangements {
				for _, badCatalog := range []bool{false, true} {
					for bagMode := 0; bagMode < 3; bagMode++ {
						bag := leaf.Raw
						if bagMode == 1 {
							bag = append(bytes.Clone(parentA.Raw), leaf.Raw...)
						}
						if bagMode == 2 {
							bag = bytes.Join([][]byte{parentB.Raw, leaf.Raw, parentA.Raw}, nil)
						}
						for _, surface := range []string{"pe", "catalog", "legacy-time"} {
							label := fmt.Sprintf("learn-%s-%s", surface, hash)
							kind := fmt.Sprintf("%s-leaf%d-bad%v-bag%d", arrange.name, leafRole, badCatalog, bagMode)
							t.Run(label+"/"+kind, func(t *testing.T) {
								inputs := map[string][]byte{"trust.crb": []byte(chain.crb + otherCRB + code.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
								for i, sequence := range arrange.bags {
									parts := [][]byte{root.Raw, otherRoot.Raw}
									for _, which := range sequence {
										parent := parentA
										if which == 'b' {
											parent = parentB
										}
										parts = append(parts, parent.Raw)
									}
									parts = append(parts, code.bag)
									packet := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0x5a}, 32), bytes.Repeat([]byte{0xa5}, 20), code.leaf, code.leafKey, bytes.Join(parts, nil))
									if badCatalog {
										packet = bytes.Clone(packet)
										packet[len(packet)-10] ^= 1
									}
									inputs[fmt.Sprintf("%d-learn.cat", i)] = packet
								}
								body, _ := signedPEWithSigner(t, leaf, chain.leafKey, bag)
								if surface == "catalog" {
									body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
									digest, err := peCatalogDigest(body, crypto.SHA256)
									if err != nil {
										t.Fatal(err)
									}
									inputs["z-target.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), leaf, chain.leafKey, bag)
								}
								if surface == "legacy-time" {
									body = replacePECertificate(t, primary, appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), leaf, chain.leafKey, bag))
								}
								check(t, label, kind, body, inputs)
							})
						}
					}
				}
			}
		}
	}
}

func TestAuthenticodeStoreRepeatedCertificatePurposes(t *testing.T) {
	check := storeOracle(t)
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		base := historicalPEChainFixture(t, hash, 1, true, x509.ExtKeyUsageCodeSigning)
		other := historicalPEChainFixture(t, hash, 1, true, x509.ExtKeyUsageCodeSigning)
		rootA := storePurposeCertificate(t, base.root, base.root, base.rootKey, hash, 3)
		template := *other.root
		template.Subject, template.RawSubject = pkix.Name{CommonName: "Repeated alternate root"}, nil
		der, err := x509.CreateCertificate(rand.Reader, &template, &template, other.root.PublicKey, other.rootKey)
		if err != nil {
			t.Fatal(err)
		}
		rootB, err := x509.ParseCertificate(der)
		if err != nil {
			t.Fatal(err)
		}
		rootB = storePurposeCertificate(t, rootB, rootB, other.rootKey, hash, 3)
		parentA := storePurposeCertificate(t, base.parent, rootA, base.rootKey, hash, 1)
		parentB := storePurposeCertificate(t, base.parent, rootB, other.rootKey, hash, 2)
		// Q is byte-for-byte identical in both learning steps. Its issuer P
		// has first one authenticated purpose and later another one.
		shared := storePurposeCertificate(t, other.parent, parentA, base.parentKey, hash, 3)
		otherCRB := replaceHistoricalCRBField(t, policyBlockRow(t, rootB, "Trusted.Repeated", "wildcard"), 1, "1")
		code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
		expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
		primary, primaryDER := signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
		for leafRole := 1; leafRole <= 3; leafRole++ {
			leaf := storePurposeCertificate(t, base.leaf, shared, other.parentKey, hash, leafRole)
			for _, reverse := range []bool{false, true} {
				first, second := parentA, parentB
				if reverse {
					first, second = second, first
				}
				for _, repeat := range []bool{false, true} {
					phases := [][]byte{first.Raw, shared.Raw, second.Raw, shared.Raw}
					if repeat {
						phases = append(phases, shared.Raw)
					}
					for _, badCatalog := range []bool{false, true} {
						for bagMode := 0; bagMode < 3; bagMode++ {
							bag := leaf.Raw
							if bagMode == 1 {
								bag = append(bytes.Clone(shared.Raw), leaf.Raw...)
							}
							if bagMode == 2 {
								bag = bytes.Join([][]byte{second.Raw, shared.Raw, leaf.Raw, first.Raw}, nil)
							}
							for _, surface := range []string{"pe", "catalog", "legacy-time"} {
								label := fmt.Sprintf("repeated-%s-%s", surface, hash)
								kind := fmt.Sprintf("reverse%v-repeat%v-leaf%d-bad%v-bag%d", reverse, repeat, leafRole, badCatalog, bagMode)
								t.Run(label+"/"+kind, func(t *testing.T) {
									inputs := map[string][]byte{"trust.crb": []byte(base.crb + otherCRB + code.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
									for i, learned := range phases {
										carried := bytes.Join([][]byte{rootA.Raw, rootB.Raw, learned, code.bag}, nil)
										packet := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0x5a}, 32), bytes.Repeat([]byte{0xa5}, 20), code.leaf, code.leafKey, carried)
										if badCatalog {
											packet[len(packet)-10] ^= 1
										}
										inputs[fmt.Sprintf("%d-learn.cat", i)] = packet
									}
									body, _ := signedPEWithSigner(t, leaf, base.leafKey, bag)
									if surface == "catalog" {
										body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
										digest, err := peCatalogDigest(body, crypto.SHA256)
										if err != nil {
											t.Fatal(err)
										}
										inputs["z-target.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), leaf, base.leafKey, bag)
									}
									if surface == "legacy-time" {
										body = replacePECertificate(t, primary, appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), leaf, base.leafKey, bag))
									}
									check(t, label, kind, body, inputs)
								})
							}
						}
					}
				}
			}
		}
	}
}
