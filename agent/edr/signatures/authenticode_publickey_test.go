package signatures

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"testing"
)

func publicKeyEncodingSeeds(tb testing.TB) ([][]byte, *rsa.PublicKey) {
	tb.Helper()
	seed, _ := primaryParserSeed(tb)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		tb.Fatal(err)
	}
	cert := packet.certificates[1]
	key := cert.PublicKey.(*rsa.PublicKey)
	var spki asn1.RawValue
	if _, err := asn1.Unmarshal(cert.RawSubjectPublicKeyInfo, &spki); err != nil {
		tb.Fatal(err)
	}
	var algorithm, bits, seq, n, e asn1.RawValue
	rest, err := asn1.Unmarshal(spki.Bytes, &algorithm)
	if err != nil {
		tb.Fatal(err)
	}
	if _, err = asn1.Unmarshal(rest, &bits); err != nil {
		tb.Fatal(err)
	}
	if _, err = asn1.Unmarshal(bits.Bytes[1:], &seq); err != nil {
		tb.Fatal(err)
	}
	rest, err = asn1.Unmarshal(seq.Bytes, &n)
	if err != nil {
		tb.Fatal(err)
	}
	if _, err = asn1.Unmarshal(rest, &e); err != nil {
		tb.Fatal(err)
	}
	var oid asn1.RawValue
	if _, err = asn1.Unmarshal(algorithm.Bytes, &oid); err != nil {
		tb.Fatal(err)
	}
	encode := func(tag int, compound bool, content []byte) []byte {
		b, err := asn1.Marshal(asn1.RawValue{Tag: tag, IsCompound: compound, Bytes: content})
		if err != nil {
			tb.Fatal(err)
		}
		return b
	}
	makeKey := func(modulus, exponent, keyDER, alg []byte) []byte {
		if keyDER == nil {
			keyDER = encode(asn1.TagSequence, true, append(encode(asn1.TagInteger, false, modulus), encode(asn1.TagInteger, false, exponent)...))
		}
		return encode(asn1.TagSequence, true, append(bytes.Clone(alg), encode(asn1.TagBitString, false, append([]byte{0}, keyDER...))...))
	}
	seeds := [][]byte{cert.RawSubjectPublicKeyInfo}
	seeds = append(seeds, makeKey(n.Bytes, e.Bytes, nil, encode(asn1.TagSequence, true, oid.FullBytes)))
	unsigned := n.Bytes
	if len(unsigned) > 0 && unsigned[0] == 0 {
		unsigned = unsigned[1:]
	}
	seeds = append(seeds, makeKey(unsigned, e.Bytes, nil, algorithm.FullBytes))
	seeds = append(seeds, makeKey(append([]byte{0}, n.Bytes...), append([]byte{0}, e.Bytes...), nil, algorithm.FullBytes))
	// The reference permits four-octet definite lengths with leading zeroes
	// inside the opaque RSA payload, even though the Go scalar DER parser does not.
	content := append(bytes.Clone(n.FullBytes), e.FullBytes...)
	long := append([]byte{0x30, 0x84, 0, 0, byte(len(content) >> 8), byte(len(content))}, content...)
	seeds = append(seeds, makeKey(nil, nil, long, algorithm.FullBytes))
	return seeds, key
}

func TestOwnedPublicKeyNormalizationPreservesValues(t *testing.T) {
	seeds, key := publicKeyEncodingSeeds(t)
	for i, seed := range seeds {
		before := bytes.Clone(seed)
		encoded, err := normalizeCertificatePublicKey(seed)
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		parsed, err := x509.ParsePKIXPublicKey(encoded)
		if err != nil {
			t.Fatal(err)
		}
		actual, ok := parsed.(*rsa.PublicKey)
		if !ok || actual.E != key.E || actual.N.Cmp(key.N) != 0 || !bytes.Equal(before, seed) {
			t.Fatal("key value or original encoding changed")
		}
		if i == 1 || i == 2 || i == 3 || i == 4 {
			if _, err = x509.ParsePKIXPublicKey(seed); err == nil {
				t.Fatalf("seed %d did not exercise strict metadata parsing", i)
			}
		}
	}
	for _, bad := range [][]byte{{0x30, 0x80, 0, 0}, {0x30, 0x85, 0, 0, 0, 0, 0}, {0x30, 0x84, 255, 255, 255, 255}, {0x30}} {
		if _, err := normalizeCertificatePublicKey(bad); !errors.Is(err, errRejectedAuthenticodeCertificate) {
			t.Fatal("invalid framing was not a known rejection", err)
		}
	}
}

func FuzzOwnedAuthenticodePublicKey(f *testing.F) {
	seeds, _ := publicKeyEncodingSeeds(f)
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Add([]byte{0x30, 0x80, 0, 0})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}
		before := bytes.Clone(input)
		encoded, err := normalizeCertificatePublicKey(input)
		if !bytes.Equal(before, input) {
			t.Fatal("input mutated")
		}
		if err != nil {
			return
		}
		if len(encoded) > 1200 {
			t.Fatal("key canonicalization exceeded its bounded integer representation")
		}
		key, err := x509.ParsePKIXPublicKey(encoded)
		if err != nil {
			return
		} // Numeric values can exceed the metadata parser's range.
		canonical, err := x509.MarshalPKIXPublicKey(key)
		if err != nil || !bytes.Equal(encoded, canonical) {
			t.Fatal("noncanonical metadata output", err)
		}
	})
}
