package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"math/big"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func counterWireVariant(t *testing.T, der []byte, key *rsa.PrivateKey, hash crypto.Hash, kind string, certBag []byte) []byte {
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
	primary := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
	unsigned := primaryFields(t, primary[6].Bytes)
	attribute := primaryFields(t, unsigned[0].Bytes)
	fields := primaryFields(t, primaryFields(t, attribute[1].Bytes)[0].Bytes)
	attrs := primaryFields(t, fields[3].Bytes)
	contentHash := hash.New()
	contentHash.Write(primary[5].Bytes)
	digest := contentHash.Sum(nil)
	if kind == "wrong-content-digest" {
		digest[0] ^= 1
	}
	for i, attr := range attrs {
		parts := primaryFields(t, attr.Bytes)
		var id asn1.ObjectIdentifier
		if _, err := asn1.Unmarshal(parts[0].FullBytes, &id); err != nil {
			t.Fatal(err)
		}
		if id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}) {
			attrs[i] = seq(parts[0], set(marshal(digest)))
		}
	}
	for _, spec := range []struct {
		kind string
		oid  asn1.ObjectIdentifier
	}{{"duplicate-content-type", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}}, {"duplicate-message-digest", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}}, {"duplicate-signing-time", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}}} {
		if kind != spec.kind {
			continue
		}
		for _, attr := range attrs {
			parts := primaryFields(t, attr.Bytes)
			if bytes.Equal(parts[0].FullBytes, oidFixture(t, spec.oid)) {
				attrs = append(attrs, attr)
				break
			}
		}
	}
	if kind == "missing-signing-time" {
		for i, attr := range attrs {
			parts := primaryFields(t, attr.Bytes)
			if bytes.Equal(parts[0].FullBytes, oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5})) {
				attrs = append(attrs[:i], attrs[i+1:]...)
				break
			}
		}
	}
	for i := 0; i < len(attrs); i++ {
		for j := i + 1; j < len(attrs); j++ {
			if bytes.Compare(attrs[i].FullBytes, attrs[j].FullBytes) > 0 {
				attrs[i], attrs[j] = attrs[j], attrs[i]
			}
		}
	}
	if kind == "signed-reverse-order" {
		attrs[0], attrs[1] = attrs[1], attrs[0]
	}
	h := hash.New()
	h.Write(set(attrs...).FullBytes)
	signatureHash := hash
	if kind == "raw-digest-signature" {
		signatureHash = 0
	}
	signature, err := signReferenceFixture(rand.Reader, key, signatureHash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	if kind == "unsigned-reverse-order" {
		attrs[0], attrs[1] = attrs[1], attrs[0]
	}
	if kind == "leading-zero-signature" {
		signature = append([]byte{0}, signature...)
	}
	if kind == "signature-plus-modulus" {
		value := new(big.Int).SetBytes(signature)
		value.Add(value, key.N)
		signature = value.Bytes()
	}
	fields[2] = algorithm(primaryHashOID(hash, false), kind == "bad-signer-parameter")
	if kind == "unknown-algorithm" {
		fields[2] = algorithm(asn1.ObjectIdentifier{1, 2, 3, 4, 999}, false)
	}
	if kind == "digest-oiw" {
		fields[2] = algorithm(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}, false)
	}
	if kind == "invalid-digest-oid" {
		fields[2] = primaryNode(t, 0, asn1.TagSequence, primaryFields(t, []byte{6, 1, 0x80, 5, 0}))
	}
	fields[3] = primaryNode(t, asn1.ClassContextSpecific, 0, attrs)
	rsaOID := asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	if kind == "rsa-digest-alias" {
		rsaOID = primaryHashOID(hash, true)
	}
	if kind == "rsa-digest-conflict" {
		other := crypto.SHA512
		if hash == crypto.SHA512 {
			other = crypto.SHA256
		}
		rsaOID = primaryHashOID(other, true)
	}
	if kind == "unknown-rsa-algorithm" {
		rsaOID = asn1.ObjectIdentifier{1, 2, 3, 4, 999}
	}
	if kind == "md2-rsa-algorithm" {
		rsaOID = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 2}
	}
	fields[4] = algorithm(rsaOID, kind == "bad-rsa-parameter")
	if kind == "rsa-oiw" {
		fields[4] = algorithm(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}, false)
	}
	if kind == "invalid-rsa-oid" {
		fields[4] = primaryNode(t, 0, asn1.TagSequence, primaryFields(t, []byte{6, 1, 0x80, 5, 0}))
	}
	fields[5] = marshal(signature)
	if kind == "signer-version-zero" {
		fields[0] = marshal(0)
	}
	if kind == "signer-version-two" {
		fields[0] = marshal(2)
	}
	if kind == "serial-mismatch" {
		identity := primaryFields(t, fields[1].Bytes)
		identity[1] = marshal(99999)
		fields[1] = seq(identity...)
	}
	attribute[1] = set(seq(fields...))
	unsigned[0] = seq(attribute...)
	primary[6] = primaryNode(t, asn1.ClassContextSpecific, 1, unsigned)
	signed[4] = set(seq(primary...))
	if certBag != nil {
		signed[3] = primaryNode(t, asn1.ClassContextSpecific, 0, primaryFields(t, certBag))
	}
	outer[1] = primaryNode(t, asn1.ClassContextSpecific, 0, []asn1.RawValue{seq(signed...)})
	return seq(outer...).FullBytes
}

func TestAuthenticodeCounterWireVerification(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		code := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
		template := *code.leaf
		template.NotAfter = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		raw, err := x509.CreateCertificate(rand.Reader, &template, code.parent, &code.leafKey.PublicKey, code.parentKey)
		if err != nil {
			t.Fatal(err)
		}
		expired, err := x509.ParseCertificate(raw)
		if err != nil {
			t.Fatal(err)
		}
		codeBag := bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1)
		pe, der := signedPEWithSigner(t, expired, code.leafKey, codeBag)
		tsa := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageTimeStamping)
		timestamped := appendLegacyTimestampWithSigner(t, der, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tsa.leaf, tsa.leafKey, tsa.bag)
		carrier := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{0xa5}, 32), bytes.Repeat([]byte{0x5a}, 20), code.root, code.rootKey, append(bytes.Clone(code.bag), tsa.bag...))
		for _, kind := range []string{"valid", "signed-reverse-order", "unsigned-reverse-order", "duplicate-content-type", "duplicate-message-digest", "duplicate-signing-time", "bad-signer-parameter", "bad-rsa-parameter", "rsa-digest-alias", "serial-mismatch", "leading-zero-signature", "signature-plus-modulus", "raw-digest-signature", "wrong-content-digest", "missing-signing-time", "signer-version-two", "signer-version-zero", "catalog-certificate", "rsa-digest-conflict", "unknown-algorithm", "unknown-rsa-algorithm", "md2-rsa-algorithm", "digest-oiw", "rsa-oiw", "invalid-digest-oid", "invalid-rsa-oid"} {
			t.Run(hash.String()+"/"+kind, func(t *testing.T) {
				var bag []byte
				if kind == "catalog-certificate" {
					bag = codeBag
				}
				variant := counterWireVariant(t, timestamped, tsa.leafKey, hash, kind, bag)
				body := replacePECertificate(t, pe, variant)
				inputs := map[string][]byte{"trust.crb": []byte(code.crb + tsa.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
				if kind == "catalog-certificate" {
					inputs["00-carrier.cat"] = carrier
				}
				want := kind == "valid" || kind == "signed-reverse-order" || kind == "leading-zero-signature" || kind == "signature-plus-modulus" || kind == "signer-version-zero" || kind == "catalog-certificate" || kind == "rsa-digest-alias" || kind == "rsa-oiw" && hash == crypto.SHA1 || (kind == "unknown-algorithm" || kind == "unknown-rsa-algorithm" || kind == "digest-oiw" || kind == "invalid-digest-oid" || kind == "invalid-rsa-oid") && hash == crypto.SHA512
				db, err := signatures.Compile(inputs)
				if err != nil {
					t.Fatal(err)
				}
				e := New(config.EDRConfig{})
				e.current.Store(&generation{db: db})
				clean, name, err := e.Scan(context.Background(), body)
				if clean != want || err != nil {
					t.Errorf("counter verification: clean=%v name=%q err=%v want=%v", clean, name, err, want)
				}
				exportHistoricalScan(t, "counter-"+hash.String(), kind, body, inputs, clean, name, want)
			})
		}
	}
}
