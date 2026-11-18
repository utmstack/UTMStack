package signatures

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"math/big"
	"testing"
)

func certificateProofSeeds(tb testing.TB) ([][]byte, *rsa.PublicKey, []byte) {
	tb.Helper()
	seed, _ := primaryParserSeed(tb)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil || len(packet.certificates) != 3 {
		tb.Fatal("invalid public certificate seed", err)
	}
	parent, root := packet.certificates[1], packet.certificates[0]
	key, ok := root.PublicKey.(*rsa.PublicKey)
	if !ok {
		tb.Fatal("seed root is not RSA")
	}
	var outer asn1.RawValue
	if rest, err := asn1.Unmarshal(parent.Raw, &outer); err != nil || len(rest) != 0 {
		tb.Fatal(err)
	}
	var fields []asn1.RawValue
	for data := outer.Bytes; len(data) > 0; {
		var field asn1.RawValue
		rest, err := asn1.Unmarshal(data, &field)
		if err != nil {
			tb.Fatal(err)
		}
		fields = append(fields, field)
		data = rest
	}
	if len(fields) != 3 {
		tb.Fatal("invalid certificate framing")
	}
	encode := func(algorithm, bits []byte) []byte {
		b, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagSequence, IsCompound: true, Bytes: bytes.Join([][]byte{fields[0].FullBytes, algorithm, bits}, nil)})
		if err != nil {
			tb.Fatal(err)
		}
		return b
	}
	bitString := func(content []byte) []byte {
		b, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagBitString, Bytes: content})
		if err != nil {
			tb.Fatal(err)
		}
		return b
	}
	leading := append([]byte{0}, fields[2].Bytes...)
	numeric := new(big.Int).SetBytes(fields[2].Bytes)
	numeric.Add(numeric, key.N)
	plus := append([]byte{0}, numeric.Bytes()...)
	var algorithmOID asn1.RawValue
	if _, err := asn1.Unmarshal(fields[1].Bytes, &algorithmOID); err != nil {
		tb.Fatal(err)
	}
	absent, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagSequence, IsCompound: true, Bytes: algorithmOID.FullBytes})
	if err != nil {
		tb.Fatal(err)
	}
	return [][]byte{parent.Raw, encode(fields[1].FullBytes, bitString(leading)), encode(fields[1].FullBytes, bitString(plus)), encode(absent, fields[2].FullBytes), encode(fields[1].FullBytes, bitString(numeric.Bytes()))}, key, parent.RawTBSCertificate
}

func TestCertificateMetadataKeepsOriginalSignedBytes(t *testing.T) {
	seeds, key, originalTBS := certificateProofSeeds(t)
	for i, data := range seeds {
		cert, err := parseAuthenticodeCertificate(data)
		if err != nil || !bytes.Equal(cert.Raw, data) || !bytes.Equal(cert.RawTBSCertificate, originalTBS) {
			t.Fatalf("seed %d lost original bytes: %v", i, err)
		}
		if valid, _ := authenticodeCertificateProof(cert, key); !valid {
			t.Fatalf("seed %d lost valid cryptographic proof", i)
		}
		// Metadata normalization must never replace the actual signed input.
		changed := bytes.Clone(data)
		at := bytes.Index(changed, originalTBS)
		if at < 0 {
			t.Fatal("missing original TBSCertificate")
		}
		changed[at+len(originalTBS)-1] ^= 1
		if c, err := parseAuthenticodeCertificate(changed); err == nil {
			if valid, _ := authenticodeCertificateProof(c, key); valid {
				t.Fatalf("seed %d authenticated a changed original TBS", i)
			}
		}
		copy := *cert
		copy.Signature = bytes.Clone(cert.Signature)
		copy.Signature[len(copy.Signature)-1] ^= 1
		if valid, _ := authenticodeCertificateProof(&copy, key); valid {
			t.Fatal("parsed signature differs from its raw certificate")
		}
	}
	if _, err := x509.ParseCertificate(seeds[3]); err == nil {
		t.Fatal("standard X.509 policy changed; expected mismatched raw algorithm encodings to be refused")
	}
	for _, key := range []*rsa.PublicKey{nil, {}, {N: big.NewInt(3), E: 65537}, {N: new(big.Int).Lsh(big.NewInt(1), 4097), E: 65537}} {
		if _, ok := normalizeAuthenticodeRSA(key, make([]byte, 128)); ok {
			t.Fatal("invalid RSA key admitted")
		}
	}
}

func FuzzOwnedAuthenticodeCertificate(f *testing.F) {
	seeds, key, originalTBS := certificateProofSeeds(f)
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Add([]byte{0x30, 0})
	f.Add([]byte{0x30, 0x80, 0, 0})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}
		cert, err := parseAuthenticodeCertificate(input)
		if err != nil {
			return
		}
		if !bytes.Equal(cert.Raw, input) {
			t.Fatal("parser lost its original certificate bytes")
		}
		if valid, _ := authenticodeCertificateProof(cert, key); valid && !bytes.Equal(cert.RawTBSCertificate, originalTBS) {
			t.Fatal("mutated signed certificate authenticated under the pinned public key")
		}
	})
}

func TestRejectedCertificateStillConsumesBagBudget(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	fields := func(data []byte) []asn1.RawValue {
		var out []asn1.RawValue
		for len(data) > 0 {
			var v asn1.RawValue
			rest, err := asn1.Unmarshal(data, &v)
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, v)
			data = rest
		}
		return out
	}
	rebuild := func(original asn1.RawValue, children []asn1.RawValue) asn1.RawValue {
		var body []byte
		for _, child := range children {
			body = append(body, child.FullBytes...)
		}
		b, err := asn1.Marshal(asn1.RawValue{Class: original.Class, Tag: original.Tag, IsCompound: original.IsCompound, Bytes: body})
		if err != nil {
			t.Fatal(err)
		}
		return fields(b)[0]
	}
	outer := fields(seed)[0]
	top := fields(outer.Bytes)
	signed := fields(top[1].Bytes)[0]
	parts := fields(signed.Bytes)
	certs := fields(parts[3].Bytes)
	decoyFields := fields(certs[1].Bytes)
	badSignature := bytes.Clone(decoyFields[2].FullBytes)
	badSignature[0] = 0x23 // A framed certificate with a rejected signature type.
	decoyFields[2] = fields(badSignature)[0]
	decoy := rebuild(certs[1], decoyFields)
	wrap := func(extra asn1.RawValue) []byte {
		copyParts := append([]asn1.RawValue(nil), parts...)
		copyParts[3] = rebuild(parts[3], append([]asn1.RawValue{extra}, certs...))
		copyTop := append([]asn1.RawValue(nil), top...)
		copyTop[1] = rebuild(top[1], []asn1.RawValue{rebuild(signed, copyParts)})
		return rebuild(outer, copyTop).FullBytes
	}
	if _, err := parseAuthenticodeSignedData(wrap(decoy), oidSPCIndirectData, 3); err == nil {
		t.Fatal("rejected certificate bypassed the total bag budget")
	}
	packet, err := parseAuthenticodeSignedData(wrap(decoy), oidSPCIndirectData, 4)
	if err != nil || packet == nil || len(packet.certificates) != 3 || len((&Database{}).verifiedAuthenticodePrimarySigners(packet)) != 1 {
		t.Fatal("framed rejected entry prevented authenticating the valid signer", err)
	}
	// The independent empty-certificate framing controls confirm that a
	// framed empty certificate is known invalid and discarded. It still
	// consumes the same bag budget as every other entry.
	empty := fields([]byte{0x30, 0})[0]
	if packet, err = parseAuthenticodeSignedData(wrap(empty), oidSPCIndirectData, 4); err != nil || len(packet.certificates) != 3 {
		t.Fatal("known invalid certificate prevented reading the remaining bag", err)
	}
	if _, err = parseAuthenticodeSignedData(wrap(empty), oidSPCIndirectData, 3); err == nil {
		t.Fatal("empty certificate bypassed the total bag budget")
	}
	unframed := fields([]byte{5, 0})[0]
	if _, err = parseAuthenticodeSignedData(wrap(unframed), oidSPCIndirectData, 4); err == nil {
		t.Fatal("unframed bag member was silently ignored")
	}
}
