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

func primaryFields(t *testing.T, data []byte) []asn1.RawValue {
	t.Helper()
	var values []asn1.RawValue
	for len(data) > 0 {
		var value asn1.RawValue
		rest, err := asn1.Unmarshal(data, &value)
		if err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
		data = rest
	}
	return values
}

func primaryNode(t *testing.T, class, tag int, values []asn1.RawValue) asn1.RawValue {
	t.Helper()
	var children [][]byte
	for _, value := range values {
		children = append(children, value.FullBytes)
	}
	data := derFixture(t, class, tag, children...)
	return primaryFields(t, data)[0]
}

func primaryHashOID(hash crypto.Hash, rsaAlias bool) asn1.ObjectIdentifier {
	if rsaAlias {
		suffix := map[crypto.Hash]int{crypto.MD5: 4, crypto.SHA1: 5, crypto.SHA256: 11, crypto.SHA384: 12, crypto.SHA512: 13}[hash]
		return asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, suffix}
	}
	switch hash {
	case crypto.MD5:
		return asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5}
	case crypto.SHA1:
		return asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
	default:
		return asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, map[crypto.Hash]int{crypto.SHA256: 1, crypto.SHA384: 2, crypto.SHA512: 3}[hash]}
	}
}

// Re-sign the serialized attributes themselves, without a PKCS#7 marshaler
// normalizing their ordering. No fixture private key is written to disk.
func primaryVariant(t *testing.T, der []byte, key *rsa.PrivateKey, hash crypto.Hash, kind string, certBag []byte) []byte {
	t.Helper()
	marshal := func(value any) asn1.RawValue {
		data, err := asn1.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return primaryFields(t, data)[0]
	}
	seq := func(values ...asn1.RawValue) asn1.RawValue { return primaryNode(t, 0, asn1.TagSequence, values) }
	set := func(values ...asn1.RawValue) asn1.RawValue { return primaryNode(t, 0, asn1.TagSet, values) }
	algorithm := func(oid asn1.ObjectIdentifier, bad bool) asn1.RawValue {
		param := marshal(asn1.RawValue{Tag: asn1.TagNull})
		if bad {
			param = marshal([]byte{})
		}
		return seq(marshal(oid), param)
	}
	outer := primaryFields(t, primaryFields(t, der)[0].Bytes)
	signed := primaryFields(t, primaryFields(t, outer[1].Bytes)[0].Bytes)
	info := primaryFields(t, signed[2].Bytes)
	content := primaryFields(t, info[1].Bytes)[0]
	contentHash := hash.New()
	contentHash.Write(content.Bytes)
	attribute := func(oid asn1.ObjectIdentifier, value asn1.RawValue) asn1.RawValue {
		return seq(marshal(oid), set(value))
	}
	contentType := attribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}, info[0])
	messageDigest := attribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}, marshal(contentHash.Sum(nil)))
	if kind == "wrong-content-digest" {
		digest := contentHash.Sum(nil)
		digest[0] ^= 1
		messageDigest = attribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}, marshal(digest))
	}
	if kind == "wrong-content-type" {
		contentType = attribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}, marshal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}))
	}
	attrs := []asn1.RawValue{contentType, messageDigest}
	if kind == "missing-message-digest" {
		attrs = []asn1.RawValue{contentType}
	}
	if kind == "missing-content-type" {
		attrs = []asn1.RawValue{messageDigest}
	}
	if kind == "duplicate-unknown-attribute" {
		unknown := attribute(asn1.ObjectIdentifier{1, 2, 3, 4, 5, 6}, marshal(true))
		attrs = append(attrs, unknown, unknown)
	}
	if kind == "duplicate-content-type" {
		attrs = append(attrs, contentType)
	}
	if kind == "duplicate-message-digest" {
		attrs = append(attrs, messageDigest)
	}
	if kind == "irrelevant-signing-time" {
		attrs = append(attrs, attribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}, marshal(time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC))))
	}
	if kind == "signed-reverse-order" {
		attrs[0], attrs[1] = attrs[1], attrs[0]
	}
	// Use DER ordering for the baseline and duplicate/time controls.
	if kind != "signed-reverse-order" {
		for i := 0; i < len(attrs); i++ {
			for j := i + 1; j < len(attrs); j++ {
				if bytes.Compare(attrs[i].FullBytes, attrs[j].FullBytes) > 0 {
					attrs[i], attrs[j] = attrs[j], attrs[i]
				}
			}
		}
	}
	signedAttrs := set(attrs...)
	h := hash.New()
	h.Write(signedAttrs.FullBytes)
	signatureHash := hash
	if kind == "raw-digest-signature" {
		signatureHash = 0
	}
	signature, err := signReferenceFixture(rand.Reader, key, signatureHash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	if kind == "leading-zero-signature" {
		signature = append([]byte{0}, signature...)
	}
	if kind == "signature-plus-modulus" {
		value := new(big.Int).SetBytes(signature)
		value.Add(value, key.N)
		signature = value.Bytes()
	}
	if kind == "unsigned-reverse-order" {
		attrs[0], attrs[1] = attrs[1], attrs[0]
	}
	fields := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
	fields[2] = algorithm(primaryHashOID(hash, kind == "signer-digest-alias"), kind == "bad-signer-parameter")
	fields[3] = primaryNode(t, asn1.ClassContextSpecific, 0, attrs)
	rsaOID := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	if kind == "rsa-digest-alias" {
		rsaOID = primaryHashOID(hash, true)
	}
	fields[4] = algorithm(rsaOID, kind == "bad-rsa-parameter")
	fields[5] = marshal(signature)
	if kind == "signer-version-zero" {
		fields[0] = marshal(0)
	}
	if kind == "signer-version-two" {
		fields[0] = marshal(2)
	}
	if kind == "serial-mismatch" {
		id := primaryFields(t, fields[1].Bytes)
		id[1] = marshal(99999)
		fields[1] = seq(id...)
	}
	signed[4] = set(seq(fields...))
	headerHash := hash
	if kind == "digest-set-mismatch" {
		headerHash = crypto.SHA256
		if hash == crypto.SHA256 {
			headerHash = crypto.SHA384
		}
	}
	signed[1] = set(algorithm(primaryHashOID(headerHash, kind == "header-digest-alias"), kind == "bad-header-parameter"))
	if kind == "unknown-algorithm" {
		unknown := algorithm(asn1.ObjectIdentifier{1, 2, 3, 4, 999}, false)
		signed[1] = set(unknown)
		fields[2] = unknown
		signed[4] = set(seq(fields...))
	}
	if kind == "signed-data-version-two" {
		signed[0] = marshal(2)
	}
	if certBag != nil {
		signed[3] = primaryNode(t, asn1.ClassContextSpecific, 0, primaryFields(t, certBag))
	}
	if kind == "empty-certificates" {
		signed[3] = primaryNode(t, asn1.ClassContextSpecific, 0, nil)
	}
	outer[1] = primaryNode(t, asn1.ClassContextSpecific, 0, []asn1.RawValue{seq(signed...)})
	return seq(outer...).FullBytes
}

func TestAuthenticodePrimaryWireVerification(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		chain := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
		// A different key with the same issuer/serial is intentionally placed
		// before the real signer. Its certificate is valid under the same root.
		decoyKey, err := rsa.GenerateKey(rand.Reader, 1024)
		if err != nil {
			t.Fatal(err)
		}
		decoyTemplate := *chain.leaf
		raw, err := x509.CreateCertificate(rand.Reader, &decoyTemplate, chain.parent, &decoyKey.PublicKey, chain.parentKey)
		if err != nil {
			t.Fatal(err)
		}
		decoy, err := x509.ParseCertificate(raw)
		if err != nil {
			t.Fatal(err)
		}
		pe, embedded := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
		unsigned := simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
		digest, err := peCatalogDigest(unsigned, crypto.SHA256)
		if err != nil {
			t.Fatal(err)
		}
		catalog := signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, chain.bag)
		carrier := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20), chain.root, chain.rootKey, chain.bag)
		for _, format := range []string{"embedded", "catalog"} {
			for _, kind := range []string{"valid", "signed-reverse-order", "unsigned-reverse-order", "duplicate-content-type", "duplicate-message-digest", "irrelevant-signing-time", "header-digest-alias", "signer-digest-alias", "rsa-digest-alias", "bad-header-parameter", "bad-signer-parameter", "bad-rsa-parameter", "signer-version-zero", "signer-version-two", "signed-data-version-two", "digest-set-mismatch", "serial-mismatch", "decoy-first", "duplicate-certificate", "empty-certificates", "leading-zero-signature", "signature-plus-modulus", "raw-digest-signature", "wrong-content-digest", "duplicate-unknown-attribute", "missing-message-digest", "missing-content-type", "unknown-algorithm", "wrong-content-type"} {
				label := fmt.Sprintf("%s-%s", hash, format)
				t.Run(label+"/"+kind, func(t *testing.T) {
					base := embedded
					if format == "catalog" {
						base = catalog
					}
					var bag []byte
					if kind == "decoy-first" {
						bag = append(bytes.Clone(decoy.Raw), chain.bag...)
					}
					if kind == "duplicate-certificate" {
						bag = append(bytes.Clone(chain.leaf.Raw), chain.bag...)
					}
					variant := primaryVariant(t, base, chain.leafKey, hash, kind, bag)
					inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
					body := replacePECertificate(t, pe, variant)
					if format == "catalog" {
						inputs["10-target.cat"] = variant
						body = unsigned
					}
					if kind == "empty-certificates" {
						inputs["00-carrier.cat"] = carrier
					}
					want := kind == "valid" || kind == "signed-reverse-order" || kind == "irrelevant-signing-time" || kind == "header-digest-alias" || kind == "signer-digest-alias" || kind == "decoy-first" || kind == "duplicate-certificate" || kind == "empty-certificates" || kind == "leading-zero-signature" || kind == "signature-plus-modulus" || kind == "duplicate-unknown-attribute" || kind == "unknown-algorithm" && hash == crypto.SHA512
					db, err := signatures.Compile(inputs)
					if err != nil {
						t.Fatal(err)
					}
					e := New(config.EDRConfig{})
					e.current.Store(&generation{db: db})
					clean, name, err := e.Scan(context.Background(), body)
					if clean != want || err != nil {
						t.Errorf("primary verification: clean=%v name=%q err=%v want=%v", clean, name, err, want)
					}
					exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
				})
			}
		}
	}
}
