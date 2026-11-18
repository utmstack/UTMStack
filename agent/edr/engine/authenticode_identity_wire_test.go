package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func identityCertificate(t *testing.T, cert *x509.Certificate, issuer *rsa.PrivateKey, hash crypto.Hash, changes map[int][]byte) *x509.Certificate {
	t.Helper()
	outer := primaryFields(t, primaryFields(t, cert.Raw)[0].Bytes)
	fields := primaryFields(t, outer[0].Bytes)
	for index, value := range changes {
		fields[index] = primaryFields(t, value)[0]
	}
	tbs := primaryNode(t, 0, asn1.TagSequence, fields)
	h := hash.New()
	h.Write(tbs.FullBytes)
	sig, err := signReferenceFixture(rand.Reader, issuer, hash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	bits, err := asn1.Marshal(asn1.BitString{Bytes: sig, BitLength: len(sig) * 8})
	if err != nil {
		t.Fatal(err)
	}
	copy := *cert
	copy.Raw = primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{tbs, outer[1], primaryFields(t, bits)[0]}).FullBytes
	copy.RawTBSCertificate = tbs.FullBytes
	copy.RawIssuer, copy.RawSubject = fields[3].FullBytes, fields[5].FullBytes
	return &copy
}

func identitySigner(t *testing.T, der []byte, issuer, serial []byte, counter bool) []byte {
	t.Helper()
	seq := func(v ...asn1.RawValue) asn1.RawValue { return primaryNode(t, 0, asn1.TagSequence, v) }
	set := func(v ...asn1.RawValue) asn1.RawValue { return primaryNode(t, 0, asn1.TagSet, v) }
	outer := primaryFields(t, primaryFields(t, der)[0].Bytes)
	signed := primaryFields(t, primaryFields(t, outer[1].Bytes)[0].Bytes)
	primary := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
	fields := primary
	var unsigned, attribute []asn1.RawValue
	if counter {
		unsigned = primaryFields(t, primary[6].Bytes)
		attribute = primaryFields(t, unsigned[0].Bytes)
		fields = primaryFields(t, primaryFields(t, attribute[1].Bytes)[0].Bytes)
	}
	fields[1] = seq(primaryFields(t, issuer)[0], primaryFields(t, serial)[0])
	if counter {
		attribute[1] = set(seq(fields...))
		unsigned[0] = seq(attribute...)
		primary[6] = primaryNode(t, asn1.ClassContextSpecific, 1, unsigned)
	}
	signed[4] = set(seq(primary...))
	outer[1] = primaryNode(t, asn1.ClassContextSpecific, 0, []asn1.RawValue{seq(signed...)})
	return seq(outer...).FullBytes
}

func identityName(t *testing.T, kind string) []byte {
	t.Helper()
	var body []byte
	switch kind {
	case "name-empty":
	case "name-integer":
		body = []byte{2, 1, 1}
	case "name-null":
		body = []byte{5, 0}
	case "name-opaque":
		body = []byte{0xff, 0, 0x81, 0x7f}
	case "name-deep":
		body = []byte{5, 0}
		for i := 0; i < 64; i++ {
			body = publicKeyTLV(0x30, body, false)
		}
	case "name-wrong-set":
		body = []byte{0x31, 3, 2, 1, 1}
	default:
		oid := []byte{6, 3, 0x55, 4, 3}
		value := []byte{12, 3, 'E', 'D', 'R'}
		switch kind {
		case "name-bad-utf8":
			value = []byte{12, 1, 0xff}
		case "name-empty-oid":
			oid = []byte{6, 0}
		case "name-octet":
			value = []byte{4, 3, 'E', 'D', 'R'}
		case "name-extra-attribute":
			value = append(value, 5, 0)
		default:
			t.Fatal(kind)
		}
		body = publicKeyTLV(0x31, publicKeyTLV(0x30, append(oid, value...), false), false)
	}
	return publicKeyTLV(0x30, body, false)
}

func identitySerial(t *testing.T, kind string) []byte {
	t.Helper()
	var b []byte
	tag := byte(2)
	switch kind {
	case "serial-canonical":
		b = []byte{1, 0x2d}
	case "serial-zero":
		b = []byte{0}
	case "serial-negative":
		b = []byte{0xff}
	case "serial-unsigned":
		b = []byte{0x80}
	case "serial-padding":
		b = []byte{0, 0, 1, 0x2d}
	case "serial-empty":
	case "serial-large":
		b = bytes.Repeat([]byte{0x80}, 256)
	case "serial-zeroes":
		b = make([]byte, 128)
	case "serial-wrong-tag":
		tag, b = 4, []byte{1}
	case "serial-constructed":
		tag, b = 0x22, []byte{2, 1, 1}
	default:
		t.Fatal(kind)
	}
	return publicKeyTLV(tag, b, false)
}

func scanIdentityControl(t *testing.T, label, kind string, body []byte, inputs map[string][]byte, want bool) {
	t.Helper()
	db, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	clean, name, scanErr := e.Scan(context.Background(), body)
	exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
	if scanErr != nil || clean != want || !want && name != "Suspicious.Code" {
		t.Errorf("raw identity: clean=%v name=%q err=%v want=%v", clean, name, scanErr, want)
	}
}

func TestAuthenticodeRawCertificateIdentities(t *testing.T) {
	kinds := []string{"serial-canonical", "serial-zero", "serial-negative", "serial-unsigned", "serial-padding", "serial-empty", "serial-large", "serial-zeroes", "serial-wrong-tag", "serial-constructed", "name-empty", "name-integer", "name-null", "name-opaque", "name-deep", "name-wrong-set", "name-bad-utf8", "name-empty-oid", "name-octet", "name-extra-attribute"}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			for _, surface := range []string{"pe", "catalog", "legacy-time"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				base := historicalPEChainFixture(t, hash, 1, embedded, role)
				var primaryBody, primaryDER []byte
				var codeCRB string
				if surface == "legacy-time" {
					code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
					template := *code.leaf
					template.NotAfter = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
					der, err := x509.CreateCertificate(rand.Reader, &template, code.parent, &code.leafKey.PublicKey, code.parentKey)
					if err != nil {
						t.Fatal(err)
					}
					cert, err := x509.ParseCertificate(der)
					if err != nil {
						t.Fatal(err)
					}
					primaryBody, primaryDER = signedPEWithSigner(t, cert, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, der, 1))
					codeCRB = code.crb
				}
				for _, target := range []string{"leaf", "issuer", "root"} {
					for _, kind := range kinds {
						label := fmt.Sprintf("identity-%s-%s-%s-root%v", surface, target, hash, embedded)
						t.Run(label+"/"+kind, func(t *testing.T) {
							chain := base
							if strings.HasPrefix(kind, "serial-") {
								cert, key := chain.leaf, chain.parentKey
								if target == "issuer" {
									cert, key = chain.parent, chain.rootKey
								}
								if target == "root" {
									cert, key = chain.root, chain.rootKey
								}
								changed := identityCertificate(t, cert, key, hash, map[int][]byte{1: identitySerial(t, kind)})
								chain.bag = bytes.Replace(chain.bag, cert.Raw, changed.Raw, 1)
								if target == "leaf" {
									chain.leaf = changed
								}
							} else {
								name := identityName(t, kind)
								if target == "leaf" {
									chain.leaf = identityCertificate(t, base.leaf, chain.parentKey, hash, map[int][]byte{5: name})
								} else if target == "issuer" {
									chain.parent = identityCertificate(t, base.parent, chain.rootKey, hash, map[int][]byte{5: name})
									chain.leaf = identityCertificate(t, base.leaf, chain.parentKey, hash, map[int][]byte{3: name})
								} else {
									chain.root = identityCertificate(t, base.root, chain.rootKey, hash, map[int][]byte{3: name, 5: name})
									chain.parent = identityCertificate(t, base.parent, chain.rootKey, hash, map[int][]byte{3: name})
									fields := strings.Split(strings.TrimSpace(chain.crb), ";")
									fields[2] = fmt.Sprintf("%x", sha1.Sum(primaryFields(t, name)[0].Bytes))
									chain.crb = strings.Join(fields, ";") + "\n"
								}
								for _, pair := range [][2]*x509.Certificate{{base.root, chain.root}, {base.parent, chain.parent}, {base.leaf, chain.leaf}} {
									chain.bag = bytes.Replace(chain.bag, pair[0].Raw, pair[1].Raw, 1)
								}
							}
							serial := primaryFields(t, chain.leaf.RawTBSCertificate)[0]
							serial = primaryFields(t, serial.Bytes)[1]
							body, packet := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
							packet = identitySigner(t, packet, chain.leaf.RawIssuer, serial.FullBytes, false)
							body = replacePECertificate(t, body, packet)
							inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
							if surface == "catalog" {
								body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
								digest, err := peCatalogDigest(body, crypto.SHA256)
								if err != nil {
									t.Fatal(err)
								}
								cat := signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, chain.bag)
								inputs["verified.cat"] = identitySigner(t, cat, chain.leaf.RawIssuer, serial.FullBytes, false)
							}
							if surface == "legacy-time" {
								packet = appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, chain.bag)
								packet = identitySigner(t, packet, chain.leaf.RawIssuer, serial.FullBytes, true)
								body = replacePECertificate(t, primaryBody, packet)
								inputs["trust.crb"] = []byte(codeCRB + chain.crb)
							}
							want := kind != "name-empty" && (target == "root" || kind != "serial-wrong-tag" && kind != "serial-constructed" && kind != "serial-empty")
							scanIdentityControl(t, label, kind, body, inputs, want)
							if target == "root" || kind != "serial-padding" && kind != "name-opaque" && kind != "name-bad-utf8" {
								return
							}
							modes := []string{"changed-image", "changed-tbs", "wrong-signer-id", "untrusted-root", "revoked-exact"}
							if kind == "serial-padding" {
								modes = append(modes, "inherited-block-wildcard", "different-raw-block-serial-first")
							}
							for _, mode := range modes {
								t.Run(mode, func(t *testing.T) {
									data := bytes.Clone(body)
									rules := make(map[string][]byte, len(inputs))
									for n, b := range inputs {
										rules[n] = bytes.Clone(b)
									}
									want := mode == "different-raw-block-serial-first"
									if mode == "changed-image" {
										data[0x200] ^= 1
									}
									if mode == "changed-tbs" {
										changed := bytes.Clone(chain.leaf.Raw)
										offset := bytes.Index(changed, chain.leaf.RawSubject)
										if offset < 0 {
											t.Fatal("missing raw subject")
										}
										changed[offset+len(chain.leaf.RawSubject)-1] ^= 1
										if surface == "catalog" {
											rules["verified.cat"] = bytes.Replace(rules["verified.cat"], chain.leaf.Raw, changed, 1)
										} else {
											data = bytes.Replace(data, chain.leaf.Raw, changed, 1)
										}
									}
									if mode == "wrong-signer-id" {
										issuer, serialDER := identityName(t, "name-null"), serial.FullBytes
										if kind == "serial-padding" && target == "leaf" {
											issuer, serialDER = chain.leaf.RawIssuer, identitySerial(t, "serial-canonical")
										}
										if surface == "catalog" {
											rules["verified.cat"] = identitySigner(t, rules["verified.cat"], issuer, serialDER, false)
										} else {
											data = replacePECertificate(t, data, identitySigner(t, packet, issuer, serialDER, surface == "legacy-time"))
										}
									}
									if mode == "untrusted-root" {
										rules["trust.crb"] = []byte(codeCRB + replaceHistoricalCRBField(t, chain.crb, 2, strings.Repeat("0", 40)))
									}
									if mode == "revoked-exact" || mode == "inherited-block-wildcard" || mode == "different-raw-block-serial-first" {
										serialBytes := serial.Bytes
										if mode != "revoked-exact" {
											serialBytes = append([]byte{0}, serialBytes...)
										}
										subject := primaryFields(t, chain.leaf.RawSubject)[0]
										blocked := fmt.Sprintf("Suspicious.Code;0;%x;%x;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), sha1.Sum(serialBytes), chain.leafKey.N.Bytes())
										if mode == "different-raw-block-serial-first" {
											rules["trust.crb"] = append([]byte(blocked), rules["trust.crb"]...)
										} else {
											rules["trust.crb"] = append(rules["trust.crb"], []byte(blocked)...)
										}
										if surface != "catalog" {
											rules["raw.ndb"] = []byte("Decoy:0:*:6861726d6c6573732d6e657665722d70726573656e742d6465636f79\n")
										}
									}
									scanIdentityControl(t, label, kind+"-"+mode, data, rules, want)
								})
							}
						})
					}
				}
			}
		}
	}
}
