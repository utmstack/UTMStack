package signatures

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"testing"
)

func identityWireSeeds(tb testing.TB) ([][]byte, *rsa.PublicKey, []byte) {
	tb.Helper()
	seed, key, original := certificateProofSeeds(tb)
	certificate := seed[0]
	wire, err := readAuthenticodeCertificate(certificate)
	if err != nil {
		tb.Fatal(err)
	}
	outer, _, _ := asn1Object(certificate)
	tbs, rest, _ := asn1Object(outer.Bytes)
	encode := func(tag int, compound bool, b []byte) []byte {
		d, err := asn1.Marshal(asn1.RawValue{Tag: tag, IsCompound: compound, Bytes: b})
		if err != nil {
			tb.Fatal(err)
		}
		return d
	}
	change := func(index int, replacement []byte) []byte {
		var body []byte
		for i, field := range wire.fields {
			if i == index {
				body = append(body, replacement...)
			} else {
				body = append(body, field.FullBytes...)
			}
		}
		return encode(asn1.TagSequence, true, append(encode(tbs.Tag, true, body), rest...))
	}
	out := [][]byte{certificate}
	for _, s := range [][]byte{{0x80}, {0, 0, 1}, bytes.Repeat([]byte{0xff}, 256)} {
		out = append(out, change(wire.algorithm-1, encode(asn1.TagInteger, false, s)))
	}
	for _, name := range [][]byte{{0xff, 0, 0x81, 0x7f}, {5, 0}, {2, 1, 0xff}} {
		out = append(out, change(wire.algorithm+1, encode(asn1.TagSequence, true, name)), change(wire.algorithm+3, encode(asn1.TagSequence, true, name)))
	}
	deep := []byte{5, 0}
	for i := 0; i < 64; i++ {
		deep = encode(asn1.TagSequence, true, deep)
	}
	out = append(out, change(wire.algorithm+3, encode(asn1.TagSequence, true, deep)))
	return out, key, original
}

func TestOwnedIdentityMetadataPreservesRawProofAndHashes(t *testing.T) {
	seeds, key, original := identityWireSeeds(t)
	for i, data := range seeds {
		before := bytes.Clone(data)
		cert, err := parseAuthenticodeCertificate(data)
		if err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
		wire, err := readAuthenticodeCertificate(data)
		if err != nil {
			t.Fatal(err)
		}
		serial, issuer, subject := wire.fields[wire.algorithm-1], wire.fields[wire.algorithm+1], wire.fields[wire.algorithm+3]
		id, err := certificateIdentityFromX509(cert)
		if err != nil || id.serial != sha1.Sum(serial.Bytes) || id.subject != sha1.Sum(subject.Bytes) {
			t.Fatal("raw identity changed", i, err)
		}
		if !bytes.Equal(before, data) || !bytes.Equal(cert.Raw, data) || !bytes.Equal(cert.RawTBSCertificate, wire.tbs.FullBytes) || !bytes.Equal(cert.RawIssuer, issuer.FullBytes) || !bytes.Equal(cert.RawSubject, subject.FullBytes) {
			t.Fatal("original bytes changed", i)
		}
		valid, _ := authenticodeCertificateProof(cert, key)
		if valid != bytes.Equal(wire.tbs.FullBytes, original) {
			t.Fatal("proof did not bind original TBS", i)
		}
		nodes := 0
		if !boundedCertificateIdentityDER(data, 4, &nodes) || nodes > 131072 {
			t.Fatal("opaque identity was treated as recursive metadata", i, nodes)
		}
		if i > 0 {
			if _, err := x509.ParseCertificate(data); err == nil {
				t.Fatal("seed did not exercise metadata difference", i)
			}
		}
	}
}

func TestIdentityWorkBoundsAndKnownRejections(t *testing.T) {
	seeds, _, _ := identityWireSeeds(t)
	wire, err := readAuthenticodeCertificate(seeds[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{wire.algorithm - 1, wire.algorithm + 1, wire.algorithm + 3} {
		fields := make([][]byte, len(wire.fields))
		for i, f := range wire.fields {
			fields[i] = f.FullBytes
		}
		wireCopy := *wire
		wireCopy.fields = append([]asn1.RawValue(nil), wire.fields...)
		wireCopy.fields[index].Bytes = nil
		if err := normalizeCertificateIdentity(&wireCopy, fields); !errors.Is(err, errRejectedAuthenticodeCertificate) {
			t.Fatal("empty identity was not rejected", index, err)
		}
	}
	nodes := 131071
	if boundedCertificateIdentityDER(seeds[0], 4, &nodes) {
		t.Fatal("certificate reset shared work budget")
	}
	seq := func(v ...[]byte) []byte {
		b, e := certificateSequence(v...)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	signer := seq([]byte{2, 1, 1}, seq(seq([]byte{0xff, 0}), []byte{2, 1, 1}), []byte{5, 0})
	nodes = 0
	if !boundedSignerIdentityDER(signer, 4, &nodes) {
		t.Fatal("opaque signer issuer was not bounded by position")
	}
	nodes = 131067
	if boundedSignerIdentityDER(signer, 4, &nodes) {
		t.Fatal("signer reset shared work budget")
	}
	deep := []byte{5, 0}
	for i := 0; i < 50; i++ {
		deep = seq(deep)
	}
	bad := seq([]byte{2, 1, 1}, seq(seq([]byte{5, 0}), []byte{2, 1, 1}), deep)
	nodes = 0
	if boundedSignerIdentityDER(bad, 4, &nodes) {
		t.Fatal("deep nonidentity data escaped work bound")
	}
}

func FuzzOwnedCertificateIdentities(f *testing.F) {
	seeds, key, original := identityWireSeeds(f)
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}
		before := bytes.Clone(input)
		cert, err := parseAuthenticodeCertificate(input)
		if !bytes.Equal(input, before) {
			t.Fatal("input mutated")
		}
		if err != nil {
			return
		}
		wire, err := readAuthenticodeCertificate(input)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(cert.Raw, input) || !bytes.Equal(cert.RawTBSCertificate, wire.tbs.FullBytes) || !bytes.Equal(cert.RawIssuer, wire.fields[wire.algorithm+1].FullBytes) || !bytes.Equal(cert.RawSubject, wire.fields[wire.algorithm+3].FullBytes) {
			t.Fatal("raw identity lost")
		}
		if valid, _ := authenticodeCertificateProof(cert, key); valid && !bytes.Equal(cert.RawTBSCertificate, original) {
			t.Fatal("mutated certificate authenticated")
		}
	})
}

func FuzzOwnedSignerIdentityBounds(f *testing.F) {
	seed := []byte{0x30, 0x0e, 2, 1, 1, 0x30, 7, 0x30, 2, 5, 0, 2, 1, 1, 5, 0}
	f.Add(seed, uint8(4), uint32(0))
	f.Add(seed, uint8(4), uint32(131067))
	f.Add([]byte{}, uint8(49), uint32(0))
	f.Fuzz(func(t *testing.T, input []byte, depth uint8, start uint32) {
		if len(input) > 64<<10 {
			return
		}
		nodes := int(start % 131080)
		initial := nodes
		accepted := boundedSignerIdentityDER(input, int(depth), &nodes)
		if accepted && (nodes > 131072 || nodes < initial || depth > 48) {
			t.Fatal("work/depth budget escaped", nodes, depth)
		}
	})
}
