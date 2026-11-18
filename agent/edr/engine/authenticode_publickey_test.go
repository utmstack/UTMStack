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

func publicKeyTLV(tag byte, content []byte, long bool) []byte {
	if long {
		n := len(content)
		return append([]byte{tag, 0x84, byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)}, content...)
	}
	if len(content) < 128 {
		return append([]byte{tag, byte(len(content))}, content...)
	}
	if len(content) < 256 {
		return append([]byte{tag, 0x81, byte(len(content))}, content...)
	}
	return append([]byte{tag, 0x82, byte(len(content) >> 8), byte(len(content))}, content...)
}

func publicKeyVariant(t *testing.T, cert *x509.Certificate, issuer *rsa.PrivateKey, kind string) []byte {
	t.Helper()
	spki := primaryFields(t, cert.RawSubjectPublicKeyInfo)[0]
	parts := primaryFields(t, spki.Bytes)
	algorithm := primaryFields(t, parts[0].Bytes)
	key := primaryFields(t, parts[1].Bytes[1:])[0]
	integers := primaryFields(t, key.Bytes)
	n, e := bytes.Clone(integers[0].Bytes), bytes.Clone(integers[1].Bytes)
	switch kind {
	case "modulus-zero", "both-zero":
		n = append([]byte{0}, n...)
	case "modulus-many-zero":
		n = append(make([]byte, 8), n...)
	case "modulus-limit", "both-limit":
		n = append(make([]byte, 513-len(n)), n...)
	case "unsigned-modulus":
		if n[0] == 0 {
			n = n[1:]
		}
	case "modulus-empty":
		n = nil
	case "modulus-short":
		n = n[len(n)-127:]
	case "modulus-long":
		n = append(make([]byte, 514-len(n)), n...)
	}
	switch kind {
	case "exponent-zero", "both-zero":
		e = append([]byte{0}, e...)
	case "exponent-many-zero":
		e = append(make([]byte, 8), e...)
	case "exponent-limit", "both-limit":
		e = append(make([]byte, len(n)-len(e)), e...)
	case "unsigned-exponent":
		if e[0] == 0 {
			e = e[1:]
		}
	case "exponent-empty":
		e = nil
	case "exponent-long":
		e = append(make([]byte, len(n)+1-len(e)), e...)
	}
	keyBytes := append(publicKeyTLV(2, n, kind == "modulus-ber-length"), publicKeyTLV(2, e, kind == "exponent-ber-length")...)
	if kind == "key-extra-field" {
		keyBytes = append(keyBytes, 5, 0)
	}
	keyDER := publicKeyTLV(0x30, keyBytes, kind == "sequence-ber-length")
	if kind == "key-extra-bytes" {
		keyDER = append(keyDER, 5, 0)
	}
	if kind == "wrong-key-sequence" {
		keyDER[0] = 0x31
	}
	bitContent := append([]byte{0}, keyDER...)
	if kind == "unused-bits" {
		bitContent[0] = 1
	}
	bitTag := byte(3)
	if kind == "constructed-bitstring" {
		bitTag = 0x23
	}
	parameter := []byte{5, 0}
	if kind == "absent-parameter" {
		parameter = nil
	}
	if kind == "bad-parameter" {
		parameter = []byte{4, 0}
	}
	if kind == "nonempty-null" {
		parameter = []byte{5, 1, 0}
	}
	if kind == "constructed-null" {
		parameter = []byte{0x25, 0}
	}
	oid := algorithm[0].FullBytes
	if kind == "wrong-algorithm" {
		oid = oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11})
	}
	alg := publicKeyTLV(0x30, append(bytes.Clone(oid), parameter...), false)
	body := append(alg, publicKeyTLV(bitTag, bitContent, false)...)
	if kind == "spki-extra-field" {
		body = append(body, 5, 0)
	}
	spkiDER := publicKeyTLV(0x30, body, false)
	outer := primaryFields(t, cert.Raw)[0]
	certParts := primaryFields(t, outer.Bytes)
	tbsParts := primaryFields(t, certParts[0].Bytes)
	index := 1
	if tbsParts[0].Class == asn1.ClassContextSpecific {
		index = 2
	}
	tbsParts[index+4] = primaryFields(t, spkiDER)[0]
	tbs := primaryNode(t, 0, asn1.TagSequence, tbsParts)
	h := crypto.SHA256.New()
	h.Write(tbs.FullBytes)
	sig, err := signReferenceFixture(rand.Reader, issuer, crypto.SHA256, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	bits, err := asn1.Marshal(asn1.BitString{Bytes: sig, BitLength: len(sig) * 8})
	if err != nil {
		t.Fatal(err)
	}
	return primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{tbs, certParts[1], primaryFields(t, bits)[0]}).FullBytes
}

func publicExponentKey(t *testing.T, exponent int) *rsa.PrivateKey {
	t.Helper()
	for {
		key, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		phi := new(big.Int).Mul(new(big.Int).Sub(key.Primes[0], big.NewInt(1)), new(big.Int).Sub(key.Primes[1], big.NewInt(1)))
		d := new(big.Int).ModInverse(big.NewInt(int64(exponent)), phi)
		if d == nil {
			continue
		}
		key.E, key.D = exponent, d
		key.Precomputed = rsa.PrecomputedValues{}
		key.Precompute()
		if err := key.Validate(); err != nil {
			t.Fatal(err)
		}
		return key
	}
}

func publicExponentChain(t *testing.T, embedded bool, exponent int, role x509.ExtKeyUsage) historicalPEChain {
	t.Helper()
	chain := historicalPEChainFixture(t, crypto.SHA256, 1, embedded, role)
	if exponent == 65537 {
		return chain
	}
	parentKey, leafKey := publicExponentKey(t, exponent), publicExponentKey(t, exponent)
	parentDER, err := x509.CreateCertificate(rand.Reader, chain.parent, chain.root, &parentKey.PublicKey, chain.rootKey)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := x509.ParseCertificate(parentDER)
	if err != nil {
		t.Fatal(err)
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, chain.leaf, parent, &leafKey.PublicKey, parentKey)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(leafDER)
	if err != nil {
		t.Fatal(err)
	}
	chain.bag = bytes.Replace(chain.bag, chain.parent.Raw, parentDER, 1)
	chain.bag = bytes.Replace(chain.bag, chain.leaf.Raw, leafDER, 1)
	chain.parent, chain.parentKey, chain.leaf, chain.leafKey = parent, parentKey, leaf, leafKey
	return chain
}

func TestAuthenticodePublicKeyEncodings(t *testing.T) {
	kinds := []string{"canonical", "absent-parameter", "modulus-zero", "modulus-many-zero", "modulus-limit", "unsigned-modulus", "exponent-zero", "exponent-many-zero", "exponent-limit", "unsigned-exponent", "both-zero", "both-limit", "sequence-ber-length", "modulus-ber-length", "exponent-ber-length", "key-extra-field", "key-extra-bytes", "spki-extra-field", "modulus-empty", "modulus-short", "modulus-long", "exponent-empty", "exponent-long", "unused-bits", "constructed-bitstring", "wrong-algorithm", "bad-parameter", "nonempty-null", "constructed-null", "wrong-key-sequence"}
	for _, surface := range []string{"pe", "catalog", "legacy-time"} {
		for _, embedded := range []bool{false, true} {
			for _, exponent := range []int{65537, 131} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				chain := publicExponentChain(t, embedded, exponent, role)
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
				for _, target := range []string{"leaf", "intermediate", "root"} {
					if target == "root" && !embedded {
						continue
					}
					for _, kind := range kinds {
						label := fmt.Sprintf("key-%s-%s-e%d-root%v", surface, target, exponent, embedded)
						t.Run(label+"/"+kind, func(t *testing.T) {
							cert, issuer := chain.leaf, chain.parentKey
							if target == "intermediate" {
								cert, issuer = chain.parent, chain.rootKey
							}
							if target == "root" {
								cert, issuer = chain.root, chain.rootKey
							}
							variant := publicKeyVariant(t, cert, issuer, kind)
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
							}
							if surface == "legacy-time" {
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
							want := target == "root" || kind == "canonical" || kind == "absent-parameter" || kind == "modulus-zero" || kind == "modulus-many-zero" || kind == "modulus-limit" || kind == "unsigned-modulus" || kind == "exponent-zero" || kind == "exponent-many-zero" || kind == "exponent-limit" || kind == "unsigned-exponent" || kind == "both-zero" || kind == "both-limit" || kind == "sequence-ber-length" || kind == "modulus-ber-length" || kind == "exponent-ber-length"
							exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
							if scanErr != nil || clean != want || !want && name != "Suspicious.Code" {
								t.Errorf("key encoding: clean=%v name=%q err=%v want=%v", clean, name, scanErr, want)
							}
						})
					}
				}
			}
		}
	}
}
