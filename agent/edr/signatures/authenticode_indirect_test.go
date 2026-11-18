package signatures

import (
	"bytes"
	"crypto"
	"encoding/asn1"
	"runtime"
	"testing"
)

func TestAuthenticodeLargeUnknownOIDAllocationBound(t *testing.T) {
	// No signature is trusted by selecting an algorithm. This checks that
	// selection itself does not expand a hostile OID into millions of arcs.
	encoded, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagOID, Bytes: bytes.Repeat([]byte{0x7f}, 1<<20)})
	if err != nil {
		t.Fatal(err)
	}
	var oid asn1.RawValue
	if _, err = asn1.Unmarshal(encoded, &oid); err != nil {
		t.Fatal(err)
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 10; i++ {
		if referenceAuthenticodeDigest(oid) != crypto.SHA512 || !referenceAuthenticodeRSAAlgorithm(oid, crypto.SHA512) {
			t.Fatal("unknown algorithm selection changed")
		}
	}
	runtime.ReadMemStats(&after)
	if after.TotalAlloc-before.TotalAlloc > 1<<20 {
		t.Fatalf("OID selection expanded input: allocated %d bytes", after.TotalAlloc-before.TotalAlloc)
	}
}

func indirectParserSeed(t testing.TB, oid asn1.ObjectIdentifier, hash crypto.Hash, parameter []byte) []byte {
	t.Helper()
	marshal := func(v any) []byte {
		b, err := asn1.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	seq := func(v ...[]byte) []byte {
		return marshal(asn1.RawValue{Tag: asn1.TagSequence, IsCompound: true, Bytes: bytes.Join(v, nil)})
	}
	return bytes.Join([][]byte{seq(marshal(oidSPCPEImageData)), seq(seq(marshal(oid), parameter), marshal(make([]byte, hash.Size())))}, nil)
}

func TestIndirectDigestStrictParametersAndNamespace(t *testing.T) {
	sha256OID := asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	for _, tc := range []struct {
		name      string
		parameter []byte
		valid     bool
	}{
		{"absent", nil, true}, {"null", []byte{5, 0}, true}, {"nonempty-null", []byte{5, 1, 0}, false},
		{"constructed-null", []byte{0x25, 0}, false}, {"octet", []byte{4, 0}, false}, {"extra", []byte{5, 0, 5, 0}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := indirectParserSeed(t, sha256OID, crypto.SHA256, tc.parameter)
			_, _, err := indirectPEImageDigest(body)
			if (err == nil) != tc.valid {
				t.Fatalf("image err=%v", err)
			}
			_, err = parseCatalogDigest(body)
			if (err == nil) != tc.valid {
				t.Fatalf("catalog err=%v", err)
			}
		})
	}
	// SHA1/RSA OIW is recognized in the RSA context, but falls through to
	// SHA-512 in the reference's content-digest context. Catalogs reject it.
	body := indirectParserSeed(t, asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}, crypto.SHA512, []byte{5, 0})
	hash, _, err := indirectPEImageDigest(body)
	if hash != crypto.SHA512 || err != nil {
		t.Fatalf("OIW digest context: %v %v", hash, err)
	}
	if _, err := parseCatalogDigest(body); err == nil {
		t.Fatal("SHA512 catalog member admitted")
	}
}

func FuzzAuthenticodeIndirectDigest(f *testing.F) {
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		id := asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
		if hash == crypto.MD5 {
			id = asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5}
		} else if hash == crypto.SHA1 {
			id = asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}
		} else if hash == crypto.SHA384 {
			id[len(id)-1] = 2
		} else if hash == crypto.SHA512 {
			id[len(id)-1] = 3
		}
		f.Add(indirectParserSeed(f, id, hash, []byte{5, 0}))
	}
	f.Add([]byte{0x30, 0})
	f.Fuzz(func(t *testing.T, body []byte) {
		if len(body) > 16<<10 {
			return
		}
		hash, digest, kind, err := authenticodeIndirectDigest(body)
		_, _, imageErr := indirectPEImageDigest(body)
		_, catalogErr := parseCatalogDigest(body)
		if err != nil {
			if imageErr == nil || catalogErr == nil {
				t.Fatal("invalid shared digest granted an exemption")
			}
			return
		}
		if len(digest) != hash.Size() || kind != 1 && kind != 2 {
			t.Fatal("invalid parsed result")
		}
		if (imageErr == nil) != (kind == 2) {
			t.Fatal("image namespace crossed")
		}
		if (catalogErr == nil) != (hash == crypto.SHA1 || hash == crypto.SHA256) {
			t.Fatal("unsupported catalog hash admitted")
		}
	})
}
