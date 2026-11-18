package signatures

import (
	"crypto"
	"encoding/asn1"
	"testing"
)

func TestAuthenticodeImageHashOIDsMatchReference(t *testing.T) {
	cases := []struct {
		name string
		oid  asn1.ObjectIdentifier
		hash crypto.Hash
	}{
		{"MD5", asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5}, crypto.MD5},
		{"MD5-RSA", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 4}, crypto.MD5},
		{"SHA1", asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}, crypto.SHA1},
		{"OIW-RSA-in-digest-position", asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}, crypto.SHA512},
		{"SHA1-RSA", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}, crypto.SHA1},
		{"SHA256", asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}, crypto.SHA256},
		{"SHA256-RSA", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}, crypto.SHA256},
		{"SHA384", asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}, crypto.SHA384},
		{"SHA384-RSA", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}, crypto.SHA384},
		{"SHA512", asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}, crypto.SHA512},
		{"SHA512-RSA", asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}, crypto.SHA512},
	}
	marshal := func(v any) []byte {
		b, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	seq := func(children ...[]byte) []byte {
		return catalogNode(t, asn1.ClassUniversal, asn1.TagSequence, children...)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			digest := make([]byte, tc.hash.Size())
			content := seq(seq(marshal(oidSPCPEImageData)), seq(seq(marshal(tc.oid), []byte{0x05, 0}), marshal(digest)))
			var outer asn1.RawValue
			if _, err := asn1.Unmarshal(content, &outer); err != nil {
				t.Fatal(err)
			}
			gotHash, gotDigest, err := indirectPEImageDigest(outer.Bytes)
			if err != nil || gotHash != tc.hash || len(gotDigest) != len(digest) {
				t.Fatalf("image digest: hash=%v bytes=%d err=%v", gotHash, len(gotDigest), err)
			}
			var oid asn1.RawValue
			if _, err := asn1.Unmarshal(marshal(tc.oid), &oid); err != nil || referenceAuthenticodeDigest(oid) != tc.hash {
				t.Fatalf("hash OID alias: got=%v err=%v", referenceAuthenticodeDigest(oid), err)
			}
		})
	}
}
