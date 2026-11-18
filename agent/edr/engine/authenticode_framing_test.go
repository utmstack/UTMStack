package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// This independent fixture encoder preserves the requested length octets,
// including nonminimal and deliberately unsupported forms.
func framingTLV(tag byte, body []byte, count int) []byte {
	if count == 0 {
		return publicKeyTLV(tag, body, false)
	}
	if count < 0 {
		return append(append([]byte{tag, 0x80}, body...), 0, 0)
	}
	header := make([]byte, 2+count)
	header[0], header[1] = tag, 0x80|byte(count)
	size := len(body)
	for i := len(header) - 1; i >= 2; i-- {
		header[i] = byte(size)
		size >>= 8
	}
	if size != 0 {
		panic("fixture length does not fit")
	}
	return append(header, body...)
}

func framingCertificate(t *testing.T, cert *x509.Certificate, issuer *rsa.PrivateKey, hash crypto.Hash, kind string) []byte {
	t.Helper()
	if kind == "empty-certificate" {
		return []byte{0x30, 0}
	}
	if kind == "not-a-certificate" {
		return []byte{5, 0}
	}
	outer := primaryFields(t, primaryFields(t, cert.Raw)[0].Bytes)
	fields := primaryFields(t, outer[0].Bytes)
	parts := make([][]byte, len(fields))
	for i, f := range fields {
		parts[i] = f.FullBytes
	}
	outerCount, tbsCount, sigCount := 0, 0, 0
	outerAlgo := outer[1].FullBytes
	length := func(index, count int) { parts[index] = framingTLV(parts[index][0], fields[index].Bytes, count) }
	innerAlgo := func(oidCount, paramCount, seqCount int) []byte {
		v := primaryFields(t, fields[2].Bytes)
		return framingTLV(0x30, append(framingTLV(6, v[0].Bytes, oidCount), framingTLV(5, nil, paramCount)...), seqCount)
	}
	uid := func(tag byte, body []byte) []byte { return framingTLV(tag, body, 0) }
	prepend := func(values ...[]byte) {
		parts = append(append(append([][]byte{}, parts[:7]...), values...), parts[7:]...)
	}
	switch kind {
	case "canonical":
	case "certificate-length3":
		outerCount = 3
	case "certificate-length4":
		outerCount = 4
	case "certificate-length5":
		outerCount = 5
	case "certificate-indefinite":
		outerCount = -1
	case "tbs-length3":
		tbsCount = 3
	case "tbs-length4":
		tbsCount = 4
	case "tbs-length5":
		tbsCount = 5
	case "tbs-indefinite":
		tbsCount = -1
	case "version-length1":
		length(0, 1)
	case "version-length4":
		length(0, 4)
	case "version-length5":
		length(0, 5)
	case "version-integer-length1":
		parts[0] = uid(0xa0, framingTLV(2, []byte{2}, 1))
	case "version-integer-length4":
		parts[0] = uid(0xa0, framingTLV(2, []byte{2}, 4))
	case "version-integer-length5":
		parts[0] = uid(0xa0, framingTLV(2, []byte{2}, 5))
	case "version-zero":
		parts[0] = uid(0xa0, []byte{2, 1, 0})
	case "version-two":
		parts[0] = uid(0xa0, []byte{2, 1, 1})
	case "version-four":
		parts[0] = uid(0xa0, []byte{2, 1, 3})
	case "version-empty":
		parts[0] = uid(0xa0, []byte{2, 0})
	case "version-wide":
		parts[0] = uid(0xa0, []byte{2, 2, 0, 2})
	case "version-extra":
		parts[0] = uid(0xa0, []byte{2, 1, 2, 5, 0})
	case "version-primitive":
		parts[0] = uid(0x80, []byte{2, 1, 2})
	case "serial-length1":
		length(1, 1)
	case "serial-length4":
		length(1, 4)
	case "serial-length5":
		length(1, 5)
	case "inner-alg-length1":
		parts[2] = innerAlgo(0, 0, 1)
	case "inner-alg-length4":
		parts[2] = innerAlgo(0, 0, 4)
	case "outer-alg-length1":
		outerAlgo = innerAlgo(0, 0, 1)
	case "outer-alg-length4":
		outerAlgo = innerAlgo(0, 0, 4)
	case "inner-oid-length1":
		parts[2] = innerAlgo(1, 0, 0)
	case "inner-oid-length4":
		parts[2] = innerAlgo(4, 0, 0)
	case "outer-oid-length1":
		outerAlgo = innerAlgo(1, 0, 0)
	case "outer-oid-length4":
		outerAlgo = innerAlgo(4, 0, 0)
	case "null-length1":
		parts[2] = innerAlgo(0, 1, 0)
		outerAlgo = parts[2]
	case "null-length4":
		parts[2] = innerAlgo(0, 4, 0)
		outerAlgo = parts[2]
	case "issuer-length1":
		length(3, 1)
	case "issuer-length4":
		length(3, 4)
	case "validity-length1":
		length(4, 1)
	case "validity-length4":
		length(4, 4)
	case "before-length1", "before-length4", "after-length1", "after-length4":
		v := primaryFields(t, fields[4].Bytes)
		at, count := 0, 1
		if kind == "after-length1" || kind == "after-length4" {
			at = 1
		}
		if kind == "before-length4" || kind == "after-length4" {
			count = 4
		}
		p := [][]byte{v[0].FullBytes, v[1].FullBytes}
		p[at] = framingTLV(v[at].FullBytes[0], v[at].Bytes, count)
		parts[4] = uid(0x30, bytes.Join(p, nil))
	case "subject-length1":
		length(5, 1)
	case "subject-length4":
		length(5, 4)
	case "spki-length3":
		length(6, 3)
	case "spki-length4":
		length(6, 4)
	case "signature-length3":
		sigCount = 3
	case "signature-length4":
		sigCount = 4
	case "signature-wrong-tag", "signature-extra", "signature-empty", "signature-invalid":
	case "tbs-extra":
		parts = append(parts, []byte{5, 0})
	case "uid1-empty":
		prepend(uid(0xa1, nil))
	case "uid1-opaque":
		prepend(uid(0xa1, []byte{0xff, 0, 0x80, 0x7f}))
	case "uid2-empty":
		prepend(uid(0xa2, nil))
	case "uid2-opaque":
		prepend(uid(0xa2, []byte{0xff, 0, 0x80, 0x7f}))
	case "uid1-uid2":
		prepend(uid(0xa1, []byte{0xff}), uid(0xa2, []byte{0xff}))
	case "uid2-uid1":
		prepend(uid(0xa2, nil), uid(0xa1, nil))
	case "uid1-after-extensions":
		parts = append(parts, uid(0xa1, nil))
	case "uid1-primitive":
		prepend(uid(0x81, []byte{0, 1}))
	case "uid2-primitive":
		prepend(uid(0x82, []byte{0, 1}))
	case "uid1-twice":
		prepend(uid(0xa1, nil), uid(0xa1, nil))
	case "uid4":
		prepend(uid(0xa4, nil))
	case "uid1-deep", "uid2-deep":
		body := []byte{5, 0}
		for i := 0; i < 64; i++ {
			body = uid(0x30, body)
		}
		tag := byte(0xa1)
		if kind == "uid2-deep" {
			tag = 0xa2
		}
		prepend(uid(tag, body))
	case "extensions-length1":
		length(7, 1)
	case "extensions-length4":
		length(7, 4)
	case "extension-sequence-length1", "extension-sequence-length4":
		v := primaryFields(t, fields[7].Bytes)[0]
		count := 1
		if kind == "extension-sequence-length4" {
			count = 4
		}
		parts[7] = uid(0xa3, framingTLV(0x30, v.Bytes, count))
	case "extension-entry-length1", "extension-entry-length4":
		v := primaryFields(t, primaryFields(t, fields[7].Bytes)[0].Bytes)
		count := 1
		if kind == "extension-entry-length4" {
			count = 4
		}
		p := [][]byte{framingTLV(0x30, v[0].Bytes, count)}
		for _, x := range v[1:] {
			p = append(p, x.FullBytes)
		}
		parts[7] = uid(0xa3, uid(0x30, bytes.Join(p, nil)))
	case "extensions-extra":
		parts[7] = uid(0xa3, append(bytes.Clone(fields[7].Bytes), 5, 0))
	default:
		t.Fatal(kind)
	}
	tbs := framingTLV(0x30, bytes.Join(parts, nil), tbsCount)
	h := hash.New()
	h.Write(tbs)
	sig, err := signReferenceFixture(rand.Reader, issuer, hash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	if kind == "signature-invalid" {
		sig[len(sig)-1] ^= 1
	}
	content := append([]byte{0}, sig...)
	if kind == "signature-empty" {
		content = nil
	}
	tag := byte(3)
	if kind == "signature-wrong-tag" {
		tag = 4
	}
	children := bytes.Join([][]byte{tbs, outerAlgo, framingTLV(tag, content, sigCount)}, nil)
	if kind == "signature-extra" {
		children = append(children, 5, 0)
	}
	return framingTLV(0x30, children, outerCount)
}

func TestAuthenticodeCertificateFraming(t *testing.T) {
	// The two bag controls distinguish a framed, known invalid certificate
	// from a member that cannot be read as a certificate at all.
	kinds := []string{"canonical", "certificate-length3", "certificate-length4", "certificate-length5", "certificate-indefinite", "tbs-length3", "tbs-length4", "tbs-length5", "tbs-indefinite", "version-length1", "version-length4", "version-length5", "version-integer-length1", "version-integer-length4", "version-integer-length5", "version-zero", "version-two", "version-four", "version-empty", "version-wide", "version-extra", "version-primitive", "serial-length1", "serial-length4", "serial-length5", "inner-alg-length1", "inner-alg-length4", "outer-alg-length1", "outer-alg-length4", "inner-oid-length1", "inner-oid-length4", "outer-oid-length1", "outer-oid-length4", "null-length1", "null-length4", "issuer-length1", "issuer-length4", "validity-length1", "validity-length4", "before-length1", "before-length4", "after-length1", "after-length4", "subject-length1", "subject-length4", "spki-length3", "spki-length4", "signature-length3", "signature-length4", "signature-wrong-tag", "signature-extra", "signature-empty", "signature-invalid", "tbs-extra", "uid1-empty", "uid1-opaque", "uid2-empty", "uid2-opaque", "uid1-uid2", "uid2-uid1", "uid1-after-extensions", "uid1-primitive", "uid2-primitive", "uid1-twice", "uid4", "uid1-deep", "uid2-deep", "extensions-length1", "extensions-length4", "extension-sequence-length1", "extension-sequence-length4", "extension-entry-length1", "extension-entry-length4", "extensions-extra"}
	var oracle map[string]bool
	kinds = append(kinds, "empty-certificate", "not-a-certificate")
	exportOnly := os.Getenv("EDR_CERTIFICATE_FRAMING_EXPORT_ONLY") == "1"
	if !exportOnly {
		data, err := os.ReadFile("testdata/authenticode_framing.json")
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(data, &oracle); err != nil {
			t.Fatal(err)
		}
	}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			for _, surface := range []string{"pe", "catalog", "legacy-time"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				chain := historicalPEChainFixture(t, hash, 1, embedded, role)
				var primaryBody, primaryDER []byte
				var codeCRB string
				if surface == "legacy-time" {
					code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
					expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
					primaryBody, primaryDER = signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
					codeCRB = code.crb
				}
				for _, target := range []string{"leaf", "intermediate"} {
					for _, kind := range kinds {
						cert, issuer := chain.leaf, chain.parentKey
						if target == "intermediate" {
							cert, issuer = chain.parent, chain.rootKey
						}
						changed := framingCertificate(t, cert, issuer, hash, kind)
						for _, mode := range []string{"only", "before-good", "after-good", "blocked"} {
							if mode == "blocked" && surface == "catalog" {
								continue
							}
							label := fmt.Sprintf("frame-%s-%s-%s-%s-root%v", surface, target, mode, hash, embedded)
							t.Run(label+"/"+kind, func(t *testing.T) {
								replacement := changed
								if mode == "before-good" {
									replacement = append(bytes.Clone(changed), cert.Raw...)
								}
								if mode == "after-good" {
									replacement = append(bytes.Clone(cert.Raw), changed...)
								}
								bag := bytes.Replace(chain.bag, cert.Raw, replacement, 1)
								body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bag)
								inputs := map[string][]byte{"trust.crb": []byte(codeCRB + chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
								if surface == "catalog" {
									body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
									digest, err := peCatalogDigest(body, crypto.SHA256)
									if err != nil {
										t.Fatal(err)
									}
									inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
								}
								if surface == "legacy-time" {
									body = replacePECertificate(t, primaryBody, appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, bag))
								}
								if mode == "blocked" {
									inputs["trust.crb"] = append([]byte(policyBlockRow(t, cert, "Blocked.Frame", "exact")), inputs["trust.crb"]...)
									inputs["raw.ndb"] = []byte("Decoy:0:*:6861726d6c6573732d6e657665722d70726573656e742d6465636f79\n")
								}
								db, err := signatures.Compile(inputs)
								if err != nil {
									t.Fatal(err)
								}
								e := New(config.EDRConfig{})
								e.current.Store(&generation{db: db})
								clean, name, scanErr := e.Scan(context.Background(), body)
								key := surface + "/" + target + "/" + mode + "/" + kind
								want, exists := oracle[key]
								if !exportOnly && !exists {
									t.Fatal("missing independent framing expectation")
								}
								exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
								expectedName := "Suspicious.Code"
								if mode == "blocked" {
									expectedName = "Blocked.Frame"
								}
								if !exportOnly && (scanErr != nil || clean != want || !clean && name != expectedName) {
									t.Errorf("framing: clean=%v name=%q err=%v want=%v/%s", clean, name, scanErr, want, expectedName)
								}
							})
						}
					}
				}
			}
		}
	}
}
