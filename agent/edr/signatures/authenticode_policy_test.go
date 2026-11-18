package signatures

import (
	"bytes"
	"crypto/x509/pkix"
	"encoding/asn1"
	"testing"
)

func FuzzAuthenticodeOpaqueExtension(f *testing.F) {
	seeds, issuer, _ := certificateProofSeeds(f)
	wire, err := readAuthenticodeCertificate(seeds[0])
	if err != nil {
		f.Fatal(err)
	}
	originalPolicy, err := certificateExtensionPolicy(wire.fields, wire.algorithm)
	if err != nil {
		f.Fatal(err)
	}
	outer, _, _ := asn1Object(seeds[0])
	_, remainder, _ := asn1Object(outer.Bytes)
	for _, data := range [][]byte{nil, {0xff}, []byte("opaque extension payload"), bytes.Repeat([]byte{0x30, 0x80}, 128)} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 64<<10 {
			return
		}
		before := bytes.Clone(payload)
		fields := append([]asn1.RawValue(nil), wire.fields...)
		last := len(fields) - 1
		if fields[last].Class != asn1.ClassContextSpecific || fields[last].Tag != 3 {
			t.Fatal("seed extensions missing")
		}
		exts, tail, err := authObject(fields[last].Bytes, 0, asn1.TagSequence)
		if err != nil || len(tail) != 0 {
			t.Fatal("seed extension framing", err)
		}
		extra, err := asn1.Marshal(pkix.Extension{Id: asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 55555, 269}, Critical: true, Value: payload})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := certificateSequence(exts.Bytes, extra)
		if err != nil {
			t.Fatal(err)
		}
		fields[last].FullBytes, err = asn1.Marshal(asn1.RawValue{Class: asn1.ClassContextSpecific, Tag: 3, IsCompound: true, Bytes: encoded})
		if err != nil {
			t.Fatal(err)
		}
		var parts [][]byte
		for _, field := range fields {
			parts = append(parts, field.FullBytes)
		}
		tbs, err := certificateSequence(parts...)
		if err != nil {
			t.Fatal(err)
		}
		data, err := certificateSequence(tbs, remainder)
		if err != nil {
			t.Fatal(err)
		}
		cert, err := parseAuthenticodeCertificate(data)
		if err != nil {
			t.Fatal("opaque extension was interpreted as PKIX metadata", err)
		}
		policy, err := certificatePolicyFromTBS(cert.RawTBSCertificate)
		if err != nil || policy != originalPolicy {
			t.Fatal("opaque payload changed signing purposes", err)
		}
		if !bytes.Equal(cert.Raw, data) || !bytes.Equal(cert.RawTBSCertificate, tbs) || !bytes.Equal(payload, before) {
			t.Fatal("original signed input changed")
		}
		// Only the metadata is normalized. Adding even an ignored extension
		// changes the signed TBS, so its retained original signature must fail.
		if valid, _ := authenticodeCertificateProof(cert, issuer); valid {
			t.Fatal("unsigned extension authenticated through normalized metadata")
		}
	})
}
