package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Only metadata is re-encoded while constructing this fixture. Its actual
// TBSCertificate bytes are newly signed, and all private keys stay in memory.
func certificateWireVariant(t *testing.T, cert *x509.Certificate, key *rsa.PrivateKey, hash crypto.Hash, kind string) []byte {
	t.Helper()
	marshal := func(v any) asn1.RawValue {
		b, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return primaryFields(t, b)[0]
	}
	algorithm := func(oid asn1.ObjectIdentifier, parameter []byte) asn1.RawValue {
		return primaryFields(t, derFixture(t, 0, asn1.TagSequence, oidFixture(t, oid), parameter))[0]
	}
	oid := primaryHashOID(hash, true)
	inner := algorithm(oid, []byte{5, 0})
	outer := inner
	switch kind {
	case "absent-parameter":
		inner = algorithm(oid, nil)
		outer = inner
	case "mixed-parameter":
		outer = algorithm(oid, nil)
	case "oiw-oid":
		inner = algorithm(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}, []byte{5, 0})
		outer = inner
	case "digest-oid":
		inner = algorithm(primaryHashOID(hash, false), []byte{5, 0})
		outer = inner
	case "unknown-oid":
		inner = algorithm(asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7}, []byte{5, 0})
		outer = inner
	case "invalid-oid":
		rawOID := marshal(asn1.RawValue{Tag: asn1.TagOID, Bytes: []byte{0x80}})
		inner = primaryFields(t, derFixture(t, 0, asn1.TagSequence, rawOID.FullBytes, []byte{5, 0}))[0]
		outer = inner
	case "inner-unknown-oid":
		inner = algorithm(asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6, 7}, []byte{5, 0})
	case "md2-oid":
		inner = algorithm(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 2}, []byte{5, 0})
		outer = inner
	case "rsa-oid":
		inner = algorithm(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}, []byte{5, 0})
		outer = inner
	case "bad-parameter":
		inner = algorithm(oid, []byte{4, 0})
		outer = inner
	case "nonempty-null":
		inner = algorithm(oid, []byte{5, 1, 0})
		outer = inner
	case "constructed-null":
		inner = algorithm(oid, []byte{0x25, 0})
		outer = inner
	case "outer-conflict":
		other := crypto.SHA256
		if hash == other {
			other = crypto.SHA512
		}
		outer = algorithm(primaryHashOID(other, true), []byte{5, 0})
	}
	tbs := primaryFields(t, cert.RawTBSCertificate)[0]
	fields := primaryFields(t, tbs.Bytes)
	index := 1
	if fields[0].Class == asn1.ClassContextSpecific {
		index = 2
	}
	fields[index] = inner
	tbs = primaryNode(t, 0, asn1.TagSequence, fields)
	h := hash.New()
	h.Write(tbs.FullBytes)
	digest := h.Sum(nil)
	signatureHash := hash
	if kind == "raw-digest" {
		signatureHash = 0
	}
	sig, err := signReferenceFixture(rand.Reader, key, signatureHash, digest)
	if err != nil {
		t.Fatal(err)
	}
	switch kind {
	case "leading-zero":
		sig = append([]byte{0}, sig...)
	case "signature-plus-modulus", "numeric-bitstring", "constructed-bitstring":
		value := new(big.Int).SetBytes(sig)
		value.Add(value, key.N)
		sig = value.Bytes()
	case "wrong-signature":
		sig[len(sig)-1] ^= 1
	case "oversized-signature":
		sig = append(make([]byte, 513-len(sig)), sig...)
	}
	bits := marshal(asn1.BitString{Bytes: sig, BitLength: 8 * len(sig)})
	if kind == "numeric-bitstring" {
		// The reference converts the complete BIT STRING content to an
		// unsigned integer, including its leading unused-bit-count byte.
		bits = marshal(asn1.RawValue{Tag: asn1.TagBitString, Bytes: sig})
	} else if kind == "constructed-bitstring" {
		bits = primaryFields(t, derFixture(t, 0, asn1.TagBitString, sig))[0]
	}
	return primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{tbs, outer, bits}).FullBytes
}

func TestAuthenticodeCertificateWireProofs(t *testing.T) {
	kinds := []string{"canonical", "leading-zero", "signature-plus-modulus", "numeric-bitstring", "constructed-bitstring", "raw-digest", "wrong-signature", "oversized-signature", "absent-parameter", "mixed-parameter", "oiw-oid", "digest-oid", "unknown-oid", "invalid-oid", "inner-unknown-oid", "md2-oid", "rsa-oid", "bad-parameter", "nonempty-null", "constructed-null", "outer-conflict"}
	for _, surface := range []string{"pe", "catalog", "legacy-time"} {
		for _, embeddedRoot := range []bool{false, true} {
			for _, target := range []string{"leaf", "intermediate"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				chain := historicalPEChainFixture(t, crypto.SHA256, 1, embeddedRoot, role)
				var primaryBody, primaryDER []byte
				var codeCRB string
				if surface == "legacy-time" {
					code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
					template := *code.leaf
					template.NotAfter = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
					der, err := x509.CreateCertificate(rand.Reader, &template, code.parent, &code.leafKey.PublicKey, code.parentKey)
					if err != nil {
						t.Fatal(err)
					}
					expired, err := x509.ParseCertificate(der)
					if err != nil {
						t.Fatal(err)
					}
					primaryBody, primaryDER = signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, der, 1))
					codeCRB = code.crb
				}
				for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
					for _, kind := range kinds {
						label := fmt.Sprintf("certificate-%s-%s-%s-root%v", surface, target, hash, embeddedRoot)
						t.Run(label+"/"+kind, func(t *testing.T) {
							cert, key := chain.leaf, chain.parentKey
							if target == "intermediate" {
								cert, key = chain.parent, chain.rootKey
							}
							variant := certificateWireVariant(t, cert, key, hash, kind)
							bag := bytes.Replace(chain.bag, cert.Raw, variant, 1)
							body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bag)
							inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
							if surface == "catalog" {
								body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
								digest, err := peCatalogDigest(body, crypto.SHA256)
								if err != nil {
									t.Fatal(err)
								}
								inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
							} else if surface == "legacy-time" {
								timestamped := appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, bag)
								body = replacePECertificate(t, primaryBody, timestamped)
								inputs["trust.crb"] = []byte(codeCRB + chain.crb)
							}
							db, err := signatures.Compile(inputs)
							if err != nil {
								t.Fatal(err)
							}
							e := New(config.EDRConfig{})
							e.current.Store(&generation{db: db})
							clean, name, scanErr := e.Scan(context.Background(), body)
							want := kind == "canonical" || kind == "leading-zero" || kind == "signature-plus-modulus" || kind == "numeric-bitstring" || kind == "absent-parameter" || kind == "mixed-parameter" || kind == "oiw-oid" && hash == crypto.SHA1 || (kind == "digest-oid" || kind == "unknown-oid" || kind == "invalid-oid" || kind == "inner-unknown-oid") && hash == crypto.SHA512
							exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
							if scanErr != nil || clean != want || !want && name != "Suspicious.Code" {
								t.Errorf("certificate wire proof: clean=%v name=%q err=%v want clean=%v", clean, name, scanErr, want)
							}
						})
					}
				}
			}
		}
	}
}

func TestAuthenticodeCertificateBagRecovery(t *testing.T) {
	for _, embedded := range []bool{false, true} {
		chain := historicalPEChainFixture(t, crypto.SHA256, 1, embedded, x509.ExtKeyUsageCodeSigning)
		for _, kind := range []string{"bad-parameter", "nonempty-null", "constructed-null", "outer-conflict", "oversized-signature", "constructed-bitstring", "md2-oid", "rsa-oid"} {
			decoy := certificateWireVariant(t, chain.leaf, chain.parentKey, crypto.SHA256, kind)
			for _, first := range []bool{false, true} {
				bag := append(bytes.Clone(chain.bag), decoy...)
				if first {
					bag = append(bytes.Clone(decoy), chain.bag...)
				}
				for _, catalog := range []bool{false, true} {
					label := fmt.Sprintf("recovery-root%v-first%v-catalog%v", embedded, first, catalog)
					t.Run(label+"/"+kind, func(t *testing.T) {
						body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bag)
						inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
						if catalog {
							body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
							digest, err := peCatalogDigest(body, crypto.SHA256)
							if err != nil {
								t.Fatal(err)
							}
							inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
						}
						db, err := signatures.Compile(inputs)
						if err != nil {
							t.Fatal(err)
						}
						e := New(config.EDRConfig{})
						e.current.Store(&generation{db: db})
						clean, name, err := e.Scan(context.Background(), body)
						exportHistoricalScan(t, label, kind, body, inputs, clean, name, true)
						if err != nil || !clean || name != "" {
							t.Errorf("recoverable decoy invalidated authenticated signer: clean=%v name=%q err=%v", clean, name, err)
						}
					})
				}
			}
		}
	}
}
