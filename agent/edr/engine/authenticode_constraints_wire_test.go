package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Sign the original extension bytes independently of the production parser.
func constraintCertificate(t *testing.T, cert *x509.Certificate, issuer *rsa.PrivateKey, hash crypto.Hash, kind string) []byte {
	t.Helper()
	extension := func(oid asn1.ObjectIdentifier, critical bool, value []byte) []byte {
		b, err := asn1.Marshal(pkix.Extension{Id: oid, Critical: critical, Value: value})
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	kuOID, ekuOID, bcOID := asn1.ObjectIdentifier{2, 5, 29, 15}, asn1.ObjectIdentifier{2, 5, 29, 37}, asn1.ObjectIdentifier{2, 5, 29, 19}
	ku := extension(kuOID, true, []byte{3, 2, 1, 0x86})
	if !cert.IsCA {
		ku = extension(kuOID, true, []byte{3, 2, 7, 0x80})
	}
	eku := extension(ekuOID, false, derFixture(t, 0, asn1.TagSequence, oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 3}), oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 8})))
	bcValue := []byte{0x30, 3, 1, 1, 0xff}
	if !cert.IsCA {
		bcValue = []byte{0x30, 0}
	}
	bc := extension(bcOID, true, bcValue)
	exts := [][]byte{ku, eku, bc}
	octet := func(value []byte) []byte {
		b, err := asn1.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	switch kind {
	case "canonical":
	case "path-zero", "path-negative", "path-large":
		path := []byte{2, 1, 0}
		if kind == "path-negative" {
			path[2] = 0xff
		}
		if kind == "path-large" {
			path = []byte{2, 5, 1, 0, 0, 0, 0}
		}
		exts[2] = extension(bcOID, true, derFixture(t, 0, asn1.TagSequence, []byte{1, 1, 0xff}, path))
	case "name-constraints", "bad-name-constraints", "policy-constraints", "policy-mapping", "inhibit-any-policy", "unknown-critical", "bad-san":
		oid, value := asn1.ObjectIdentifier{2, 5, 29, 30}, []byte{0x30, 0}
		switch kind {
		case "bad-name-constraints":
			value = []byte{0xff}
		case "policy-constraints":
			oid, value = asn1.ObjectIdentifier{2, 5, 29, 36}, []byte{0xff}
		case "policy-mapping":
			oid, value = asn1.ObjectIdentifier{2, 5, 29, 33}, []byte{0xff}
		case "inhibit-any-policy":
			oid, value = asn1.ObjectIdentifier{2, 5, 29, 54}, []byte{0xff}
		case "unknown-critical":
			oid, value = asn1.ObjectIdentifier{1, 2, 3, 4, 5}, []byte{0xff}
		case "bad-san":
			oid, value = asn1.ObjectIdentifier{2, 5, 29, 17}, []byte{0xff}
		}
		exts = append(exts, extension(oid, true, value))
	case "no-ku":
		exts = [][]byte{eku, bc}
	case "no-bc":
		exts = [][]byte{ku, eku}
	case "false-bc-no-ku":
		exts = [][]byte{eku, extension(bcOID, true, []byte{0x30, 0})}
	case "false-bc-last":
		exts[2] = extension(bcOID, true, []byte{0x30, 0})
	case "true-bc-last-no-cert-ku":
		exts = [][]byte{extension(kuOID, true, []byte{3, 2, 7, 0x80}), eku, extension(bcOID, true, []byte{0x30, 3, 1, 1, 0xff})}
	case "no-cert-ku-last":
		exts = [][]byte{bc, eku, extension(kuOID, true, []byte{3, 2, 7, 0x80})}
	case "cert-only-ku":
		exts[0] = extension(kuOID, true, []byte{3, 2, 2, 4})
	case "no-extensions":
		exts = nil
	case "empty-extensions":
		exts = [][]byte{}
	case "empty-eku":
		exts[1] = extension(ekuOID, false, []byte{0x30, 0})
	case "any-eku":
		exts[1] = extension(ekuOID, false, derFixture(t, 0, asn1.TagSequence, oidFixture(t, asn1.ObjectIdentifier{2, 5, 29, 37, 0})))
	case "no-eku":
		exts = [][]byte{ku, bc}
	case "duplicate-bc-false-last":
		exts = append(exts, extension(bcOID, true, []byte{0x30, 0}))
	case "duplicate-bc-true-last":
		exts = append(exts, extension(bcOID, true, []byte{0x30, 3, 1, 1, 0xff}))
	case "duplicate-eku-empty-last":
		exts = append(exts, extension(ekuOID, false, []byte{0x30, 0}))
	case "duplicate-ku-empty-last":
		exts = append(exts, extension(kuOID, true, []byte{3, 2, 0, 0}))
	case "critical-empty", "critical-long", "critical-noncanonical", "critical-false", "critical-wrong-tag", "extension-extra", "empty-oid", "malformed-oid", "extension-wrong-value":
		oid, flag, value := oidFixture(t, asn1.ObjectIdentifier{1, 2, 3, 4, 5}), []byte{1, 1, 0xff}, octet([]byte{0xff})
		var tail []byte
		switch kind {
		case "critical-empty":
			flag = []byte{1, 0}
		case "critical-long":
			flag = []byte{1, 2, 0, 1}
		case "critical-noncanonical":
			flag = []byte{1, 1, 1}
		case "critical-false":
			flag = []byte{1, 1, 0}
		case "critical-wrong-tag":
			flag = []byte{3, 1, 0}
		case "extension-extra":
			tail = []byte{5, 0}
		case "empty-oid":
			oid = []byte{6, 0}
		case "malformed-oid":
			oid = []byte{6, 1, 0x80}
		case "extension-wrong-value":
			value = []byte{5, 0}
		}
		exts = append(exts, derFixture(t, 0, asn1.TagSequence, oid, flag, value, tail))
	case "ku-second-octet", "ku-mask", "ku-padding", "ku-short", "ku-long", "ku-bad-length", "ku-bad-unused", "ku-wrong-tag":
		value := []byte{3, 3, 7, 4, 0}
		switch kind {
		case "ku-mask":
			value = []byte{3, 2, 7, 4}
		case "ku-padding":
			value = []byte{3, 2, 0, 4}
		case "ku-short":
			value = []byte{3, 1, 0}
		case "ku-long":
			value = []byte{3, 4, 0, 4, 0, 0}
		case "ku-bad-length":
			value = []byte{3, 3, 0, 4}
		case "ku-bad-unused":
			value = []byte{3, 2, 8, 4}
		case "ku-wrong-tag":
			value = []byte{4, 2, 0, 4}
		}
		exts = [][]byte{bc, eku, extension(kuOID, true, value)}
	case "bc-true-noncanonical", "bc-false-extra", "bc-inner-tail", "bc-outer-tail", "bc-bool-empty", "bc-bool-long", "bc-integer-empty", "bc-integer-negative", "bc-wrong-child", "bc-wrong-sequence":
		value := []byte{0x30, 3, 1, 1, 1}
		switch kind {
		case "bc-false-extra":
			value = []byte{0x30, 5, 1, 1, 0, 5, 0}
		case "bc-inner-tail":
			value = []byte{0x30, 4, 1, 1, 0xff, 0xff}
		case "bc-outer-tail":
			value = []byte{0x30, 3, 1, 1, 0xff, 0xff}
		case "bc-bool-empty":
			value = []byte{0x30, 2, 1, 0}
		case "bc-bool-long":
			value = []byte{0x30, 4, 1, 2, 0, 1}
		case "bc-integer-empty":
			value = []byte{0x30, 2, 2, 0}
		case "bc-integer-negative":
			value = []byte{0x30, 3, 2, 1, 0xff}
		case "bc-wrong-child":
			value = []byte{0x30, 2, 5, 0}
		case "bc-wrong-sequence":
			value = []byte{0x31, 0}
		}
		exts[2] = extension(bcOID, true, value)
	case "eku-code", "eku-time", "eku-code-prefix10", "eku-time-prefix10", "eku-code-prefix9", "eku-wrong-tag", "eku-extra-tail", "eku-empty-oid", "eku-duplicate-known-last", "eku-split-roles":
		purpose := []byte{0x2b, 6, 1, 5, 5, 7, 3, 3}
		if kind == "eku-time" || kind == "eku-time-prefix10" {
			purpose[7] = 8
		}
		if kind == "eku-code-prefix10" || kind == "eku-time-prefix10" {
			purpose = append(purpose, 0x80, 0)
		}
		if kind == "eku-code-prefix9" {
			purpose = append(purpose, 0)
		}
		encoded, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagOID, Bytes: purpose})
		if err != nil {
			t.Fatal(err)
		}
		value := derFixture(t, 0, asn1.TagSequence, encoded)
		switch kind {
		case "eku-wrong-tag":
			value = []byte{0x31, 0}
		case "eku-extra-tail":
			value = []byte{0x30, 0, 5, 0}
		case "eku-empty-oid":
			value = []byte{0x30, 2, 6, 0}
		case "eku-duplicate-known-last":
			exts = append(exts, eku)
			value = []byte{0x30, 0}
		case "eku-split-roles":
			exts = append(exts, extension(ekuOID, false, derFixture(t, 0, asn1.TagSequence, oidFixture(t, asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 3, 8}))))
		}
		exts[1] = extension(ekuOID, false, value)
	case "tampered-unknown":
		exts = append(exts, extension(asn1.ObjectIdentifier{1, 2, 3, 4, 5}, true, []byte("opaque authenticated extension")))
	default:
		t.Fatal(kind)
	}
	fields := primaryFields(t, primaryFields(t, cert.RawTBSCertificate)[0].Bytes)
	index := 6
	if fields[0].Class == asn1.ClassContextSpecific {
		index++
	}
	fields = fields[:index]
	if exts != nil {
		container := derFixture(t, 0, asn1.TagSequence, exts...)
		fields = append(fields, primaryFields(t, derFixture(t, asn1.ClassContextSpecific, 3, container))[0])
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
	original := primaryFields(t, primaryFields(t, cert.Raw)[0].Bytes)
	result := derFixture(t, 0, asn1.TagSequence, tbs.FullBytes, original[1].FullBytes, bits)
	if kind == "tampered-unknown" {
		offset := bytes.Index(result, []byte("opaque authenticated extension"))
		if offset < 0 {
			t.Fatal("missing tamper target")
		}
		result[offset] ^= 1
	}
	return result
}

func TestAuthenticodeConstraintWirePolicy(t *testing.T) {
	kinds := []string{"canonical", "path-zero", "path-negative", "path-large", "name-constraints", "bad-name-constraints", "policy-constraints", "policy-mapping", "inhibit-any-policy", "unknown-critical", "bad-san", "no-ku", "no-bc", "false-bc-no-ku", "false-bc-last", "true-bc-last-no-cert-ku", "no-cert-ku-last", "cert-only-ku", "no-extensions", "empty-extensions", "empty-eku", "any-eku", "no-eku", "duplicate-bc-false-last", "duplicate-bc-true-last", "duplicate-eku-empty-last", "duplicate-ku-empty-last"}
	kinds = append(kinds, "critical-empty", "critical-long", "critical-noncanonical", "critical-false", "critical-wrong-tag", "extension-extra", "empty-oid", "malformed-oid", "extension-wrong-value", "ku-second-octet", "ku-mask", "ku-padding", "ku-short", "ku-long", "ku-bad-length", "ku-bad-unused", "ku-wrong-tag", "bc-true-noncanonical", "bc-false-extra", "bc-inner-tail", "bc-outer-tail", "bc-bool-empty", "bc-bool-long", "bc-integer-empty", "bc-integer-negative", "bc-wrong-child", "bc-wrong-sequence", "eku-code", "eku-time", "eku-code-prefix10", "eku-time-prefix10", "eku-code-prefix9", "eku-wrong-tag", "eku-extra-tail", "eku-empty-oid", "eku-duplicate-known-last", "eku-split-roles", "tampered-unknown")
	var oracle map[string]bool
	if os.Getenv("EDR_CERTIFICATE_CONSTRAINT_EXPORT_ONLY") != "1" {
		data, err := os.ReadFile("testdata/authenticode_constraints.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &oracle); err != nil {
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
				chain := historicalPEChainFixture(t, hash, 3, embedded, role)
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
						label := fmt.Sprintf("constraint-%s-%s-%s-root%v", surface, target, hash, embedded)
						t.Run(label+"/"+kind, func(t *testing.T) {
							cert, issuer := chain.leaf, chain.parentKey
							if target == "intermediate" {
								cert, issuer = chain.firstIntermediate, chain.rootKey
							}
							changed := constraintCertificate(t, cert, issuer, hash, kind)
							bag := bytes.Replace(chain.bag, cert.Raw, changed, 1)
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
							db, err := signatures.Compile(inputs)
							if err != nil {
								t.Fatal(err)
							}
							e := New(config.EDRConfig{})
							e.current.Store(&generation{db: db})
							clean, name, scanErr := e.Scan(context.Background(), body)
							want, exists := oracle[surface+"/"+target+"/"+kind]
							if !exists && os.Getenv("EDR_CERTIFICATE_CONSTRAINT_EXPORT_ONLY") != "1" {
								t.Fatal("missing independent reference expectation")
							}
							exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
							if os.Getenv("EDR_CERTIFICATE_CONSTRAINT_EXPORT_ONLY") != "1" && (scanErr != nil || clean != want || !want && name != "Suspicious.Code") {
								t.Errorf("constraint wire: clean=%v name=%q err=%v want=%v", clean, name, scanErr, want)
							}
						})
					}
				}
			}
		}
	}
}
