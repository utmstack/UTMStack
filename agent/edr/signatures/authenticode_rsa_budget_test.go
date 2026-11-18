package signatures

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestAuthenticodeLargeExponentWorkRemainsIncomplete(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		t.Fatal(err)
	}
	rootKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC)
	rootTemplate := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "Numeric budget root"}, RawSubject: packet.certificates[2].RawIssuer, NotBefore: at.Add(-time.Hour), NotAfter: at.Add(time.Hour), KeyUsage: x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}, IsCA: true, BasicConstraintsValid: true}
	rootDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootKey.PublicKey, rootKey)
	if err != nil {
		t.Fatal(err)
	}
	root, err := x509.ParseCertificate(rootDER)
	if err != nil {
		t.Fatal(err)
	}
	subject, _, _ := asn1Object(root.RawSubject)
	d := &Database{certRules: []certificateRule{{subject: sha1.Sum(subject.Bytes), modulus: rootKey.N.Bytes(), codeSign: true, timeSign: true, certSign: true, ignoreSerial: true}}}
	parts := func(data []byte) []asn1.RawValue {
		var out []asn1.RawValue
		for len(data) > 0 {
			v, next, err := asn1Object(data)
			if err != nil {
				t.Fatal(err)
			}
			out, data = append(out, v), next
		}
		return out
	}
	node := func(tag int, compound bool, children ...[]byte) []byte {
		out, err := asn1.Marshal(asn1.RawValue{Tag: tag, IsCompound: compound, Bytes: bytes.Join(children, nil)})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	seq := func(children ...[]byte) []byte { return node(asn1.TagSequence, true, children...) }
	wrap := func(original asn1.RawValue, children ...[]byte) []byte {
		out, err := asn1.Marshal(asn1.RawValue{Tag: original.Tag, Class: original.Class, IsCompound: original.IsCompound, Bytes: bytes.Join(children, nil)})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	for _, large := range []bool{false, true} {
		var bag []byte
		exponent := big.NewInt(65537)
		if large {
			exponent = new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 4103), big.NewInt(1))
		}
		for i := 0; i < 20; i++ {
			template := &x509.Certificate{SerialNumber: big.NewInt(301), Subject: pkix.Name{CommonName: fmt.Sprintf("Numeric budget signer %d", i)}, NotBefore: at.Add(-time.Hour), NotAfter: at.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning}, BasicConstraintsValid: true}
			der, err := x509.CreateCertificate(rand.Reader, template, root, &rootKey.PublicKey, rootKey)
			if err != nil {
				t.Fatal(err)
			}
			outer := parts(parts(der)[0].Bytes)
			fields := parts(outer[0].Bytes)
			children := make([][]byte, len(fields))
			for i, f := range fields {
				children[i] = f.FullBytes
			}
			modulus := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), 4103), big.NewInt(int64(13+2*i)))
			algorithm := seq([]byte{6, 9, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 1, 1, 1}, []byte{5, 0})
			pub := seq(node(2, false, modulus.Bytes()), node(2, false, exponent.Bytes()))
			children[6] = seq(algorithm, node(3, false, append([]byte{0}, pub...)))
			tbs := seq(children...)
			digest := sha256.Sum256(tbs)
			sig, err := rsa.SignPKCS1v15(rand.Reader, rootKey, crypto.SHA256, digest[:])
			if err != nil {
				t.Fatal(err)
			}
			value := new(big.Int).SetBytes(sig)
			if len(value.Bytes()) < 127 {
				value.Add(value, rootKey.N)
				sig = value.Bytes()
			}
			cert := seq(tbs, outer[1].FullBytes, node(3, false, append([]byte{0}, sig...)))
			if _, err := parseAuthenticodeCertificate(cert); err != nil {
				t.Fatal("invalid exponent budget control", err)
			}
			bag = append(bag, cert...)
		}
		outer := parts(seed)[0]
		top := parts(outer.Bytes)
		signed := parts(top[1].Bytes)[0]
		fields := parts(signed.Bytes)
		content := make([][]byte, len(fields))
		for i, f := range fields {
			content[i] = f.FullBytes
		}
		content[3] = wrap(fields[3], bag)
		signers := parts(fields[4].Bytes)
		signer := parts(signers[0].Bytes)
		sig := make([]byte, 513)
		sig[0], sig[len(sig)-1] = 1, 3
		signer[5] = parts(node(4, false, sig))[0]
		var body [][]byte
		for _, f := range signer {
			body = append(body, f.FullBytes)
		}
		content[4] = wrap(fields[4], seq(body...))
		der := wrap(outer, top[0].FullBytes, wrap(top[1], wrap(signed, content...)))
		called := false
		trusted, err := d.VerifyTrustedAuthenticode(der, func(crypto.Hash) ([]byte, error) { called = true; return nil, nil }, at)
		if trusted || called {
			t.Fatal("invalid weighted control authenticated an image")
		}
		if large && !errors.Is(err, ErrRuleIncomplete) {
			t.Fatal("large exponent work became known invalid", err)
		}
		if !large && err != nil {
			t.Fatal("ordinary failed proofs became incomplete", err)
		}
	}
}
