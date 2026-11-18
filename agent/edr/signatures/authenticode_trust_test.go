package signatures

import (
	"crypto"
	"encoding/asn1"
	"testing"

	"github.com/smallstep/pkcs7"
)

func TestAuthenticodeSignedDigestParametersAndAliases(t *testing.T) {
	node := func(class, tag int, children ...[]byte) []byte {
		var body []byte
		for _, child := range children {
			body = append(body, child...)
		}
		out, err := asn1.Marshal(asn1.RawValue{Class: class, Tag: tag, IsCompound: true, Bytes: body})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	seq := func(children ...[]byte) []byte { return node(asn1.ClassUniversal, asn1.TagSequence, children...) }
	set := func(children ...[]byte) []byte { return node(asn1.ClassUniversal, asn1.TagSet, children...) }
	ctx := func(children ...[]byte) []byte { return node(asn1.ClassContextSpecific, 0, children...) }
	marshal := func(v any) []byte {
		out, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	for _, tc := range []struct {
		name  string
		oid   asn1.ObjectIdentifier
		param []byte
		valid bool
	}{
		{"SHA-256 without parameters", pkcs7.OIDDigestAlgorithmSHA256, nil, true},
		{"SHA-256 with NULL", pkcs7.OIDDigestAlgorithmSHA256, []byte{5, 0}, true},
		{"SHA-256 RSA alias", pkcs7.OIDEncryptionAlgorithmRSASHA256, []byte{5, 0}, true},
		{"non-NULL parameter", pkcs7.OIDDigestAlgorithmSHA256, []byte{4, 0}, false},
		{"extra parameter", pkcs7.OIDDigestAlgorithmSHA256, []byte{5, 0, 5, 0}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			algorithm := seq(marshal(tc.oid), tc.param)
			signed := seq(marshal(1), set(algorithm), seq(marshal(oidSPCIndirectData), ctx(seq())), ctx(), set())
			der := seq(marshal(oidPKCS7SignedData), ctx(signed))
			var digestOID asn1.ObjectIdentifier
			certs, recognized, err := authenticodeCertificates(der, &digestOID)
			if !recognized || len(certs) != 0 || (err == nil) != tc.valid {
				t.Fatalf("certs=%d recognized=%v err=%v", len(certs), recognized, err)
			}
			if tc.valid && (!digestOID.Equal(tc.oid) || authenticodeHashID(digestOID) != crypto.SHA256) {
				t.Fatalf("digest OID=%v hash=%v", digestOID, authenticodeHashID(digestOID))
			}
		})
	}
}

func TestAuthenticodeDERDepthBound(t *testing.T) {
	der := []byte{0x05, 0x00}
	for i := 0; i < 50; i++ {
		var err error
		der, err = asn1.Marshal(asn1.RawValue{
			Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: der,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	nodes := 0
	if boundedAuthenticodeDER(der, 0, &nodes) {
		t.Fatal("deeply nested unsigned data reached PKCS#7 parser")
	}
}
