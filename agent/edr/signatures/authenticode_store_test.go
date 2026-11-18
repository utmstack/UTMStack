package signatures

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/asn1"
	"errors"
	"math/big"
	"testing"
	"time"
)

func TestBlockedCatalogNeverLearnsBlockedDirectApproval(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		t.Fatal(err)
	}
	cert := packet.certificates[0]
	identity, err := certificateIdentityFromX509(cert)
	if err != nil {
		t.Fatal(err)
	}
	approved := certificateRule{subject: identity.subject, modulus: identity.modulus, ignoreSerial: true, codeSign: true, timeSign: true, certSign: true}
	blocked := approved
	blocked.blocked = true
	d := &Database{certRules: []certificateRule{approved, blocked}}
	store, err := d.resolveAuthenticodeStore(packet.certificates[:1], true)
	if !errors.Is(err, errRejectedAuthenticodeCertificate) || store == nil || len(store.learned) != 0 {
		t.Fatal("blocked identity became authority through catalog side effects", err)
	}
	verified, err := d.prepareCatalogCertificates(store.learned)
	if err != nil || len(verified) != 0 {
		t.Fatal("blocked identity entered persistent catalog store", err)
	}
}

func TestAuthenticodeStoreProofExhaustionRemainsIncomplete(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		t.Fatal(err)
	}
	leaf, parent := packet.certificates[2], packet.certificates[1]
	wire, err := readAuthenticodeCertificate(leaf.Raw)
	if err != nil {
		t.Fatal(err)
	}
	serial := wire.fields[wire.algorithm-1]
	if bytes.Count(leaf.Raw, serial.FullBytes) != 1 {
		t.Fatal("nonunique seed serial encoding")
	}
	var certificates []byte
	for i := 1; i <= 50; i++ {
		data := bytes.Clone(leaf.Raw)
		data[bytes.Index(data, serial.FullBytes)+len(serial.FullBytes)-1] += byte(i)
		if _, err := parseAuthenticodeCertificate(data); err != nil {
			t.Fatal("framing-invalid budget control", err)
		}
		certificates = append(certificates, data...)
	}
	d := &Database{}
	issuer, _, _ := asn1Object(leaf.RawIssuer)
	for i := 1; i <= 94; i++ {
		// Distinct non-signing public keys with the correct issuer name force
		// real RSA work. No private keys or successful path are manufactured.
		modulus := new(big.Int).Add(parent.PublicKey.(*rsa.PublicKey).N, big.NewInt(int64(i*2)))
		d.certRules = append(d.certRules, certificateRule{subject: sha1.Sum(issuer.Bytes), modulus: modulus.Bytes(), codeSign: true, timeSign: true, certSign: true, ignoreSerial: true})
	}
	parts := func(data []byte) []asn1.RawValue {
		var values []asn1.RawValue
		for len(data) > 0 {
			value, tail, err := asn1Object(data)
			if err != nil {
				t.Fatal(err)
			}
			values, data = append(values, value), tail
		}
		return values
	}
	wrap := func(value asn1.RawValue, children ...[]byte) []byte {
		data, err := asn1.Marshal(asn1.RawValue{Class: value.Class, Tag: value.Tag, IsCompound: value.IsCompound, Bytes: bytes.Join(children, nil)})
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	outer := parts(seed)[0]
	top := parts(outer.Bytes)
	signed := parts(top[1].Bytes)[0]
	fields := parts(signed.Bytes)
	var content [][]byte
	for i, field := range fields {
		data := field.FullBytes
		if i == 3 {
			data = wrap(field, certificates)
		}
		content = append(content, data)
	}
	der := wrap(outer, top[0].FullBytes, wrap(top[1], wrap(signed, content...)))
	packet, err = parseAuthenticodeSignedData(der, oidSPCIndirectData, 256)
	if err != nil || len(packet.certificates) != 50 {
		t.Fatal("invalid fixed resource control", err)
	}
	if _, err := d.resolveAuthenticodeStore(packet.certificates, false); !errors.Is(err, errAuthenticodeLimit) {
		t.Fatal("RSA work exhaustion became known untrusted", err)
	}
	digestCalled := false
	trusted, err := d.VerifyTrustedAuthenticode(der, func(crypto.Hash) ([]byte, error) {
		digestCalled = true
		return nil, nil
	}, time.Date(2026, 11, 11, 0, 0, 0, 0, time.UTC))
	if trusted || !errors.Is(err, ErrRuleIncomplete) || digestCalled {
		t.Fatal("public verification lost budget uncertainty", trusted, err, digestCalled)
	}
	// Catalogs resolve the certificate prefix before checking their signer.
	// An exhausted prefix must reject the generation, rather than be counted
	// as an ordinary invalid catalog and silently omitted from an update.
	info := parts(fields[2].Bytes)
	ctlOID, err := asn1.Marshal(oidCatalogCTL)
	if err != nil {
		t.Fatal(err)
	}
	content[2] = wrap(fields[2], ctlOID, info[1].FullBytes)
	cat := wrap(outer, top[0].FullBytes, wrap(top[1], wrap(signed, content...)))
	if _, err := d.verifyCatalog(cat, time.Now()); !errors.Is(err, errAuthenticodeLimit) {
		t.Fatal("catalog verification lost budget uncertainty", err)
	}
	var recorded error
	d.compileCatalog(source{name: "budget.cat", data: cat}, func(_ string, _ int, err error) { recorded = err })
	if !errors.Is(recorded, errAuthenticodeLimit) || len(d.catalogCerts) != 0 || d.report.Skipped[SkipInvalidCatalog] != 0 {
		t.Fatal("catalog exhaustion silently published a partial policy", recorded)
	}
}

func FuzzAuthenticodeStoreProofPaths(f *testing.F) {
	seed, root := primaryParserSeed(f)
	identity, err := certificateIdentityFromX509(root)
	if err != nil {
		f.Fatal(err)
	}
	d := &Database{certRules: []certificateRule{{subject: identity.subject, modulus: identity.modulus, ignoreSerial: true, codeSign: true, timeSign: true, certSign: true}}}
	for _, data := range [][]byte{seed, nil, {0x30, 0}, {0x30, 0x80, 0, 0}} {
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		packet, err := parseAuthenticodeSignedData(data, oidSPCIndirectData, 256)
		if err != nil || packet == nil {
			return
		}
		for _, catalog := range []bool{false, true} {
			unrooted, err := (&Database{}).resolveAuthenticodeStore(packet.certificates, catalog)
			if err == nil && (len(unrooted.entries) != 0 || len(unrooted.learned) != 0) {
				t.Fatal("unrooted certificate graph created its own authority")
			}
			store, err := d.resolveAuthenticodeStore(packet.certificates, catalog)
			if err != nil {
				continue
			}
			for _, entry := range store.entries {
				if entry.cert == nil {
					continue
				}
				if len(entry.path) == 0 || !bytes.Equal(entry.path[0].Raw, entry.cert.Raw) {
					t.Fatal("missing original certificate proof path")
				}
				for i, cert := range entry.path {
					policy, err := certificateIdentityFromX509(cert)
					if err != nil || entry.identity.codeSign && !policy.codeSign || entry.identity.timeSign && !policy.timeSign {
						t.Fatal("path manufactured a requested purpose")
					}
					if i+1 < len(entry.path) {
						parent := entry.path[i+1]
						if valid, _ := authenticodeCertificateProof(cert, parent.PublicKey); !valid {
							t.Fatal("path authenticated changed signed certificate bytes")
						}
					} else {
						key, ok := authenticodeKey(cert.PublicKey)
						if !ok {
							t.Fatal("path lost its original public integers")
						}
						if !d.trustedCertificate(policy) || key.E.Cmp(big.NewInt(65537)) != 0 {
							if valid, _ := authenticodeCertificateProof(cert, root.PublicKey); !valid {
								t.Fatal("path terminated without the authorized key or its proof")
							}
						}
					}
				}
			}
		}
	})
}
