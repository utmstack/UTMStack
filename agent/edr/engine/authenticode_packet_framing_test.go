package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Descend only through the original canonical fixture. Rebuilding each parent
// preserves the modified child's exact nonminimal bytes without parsing them.
func packetFramingRewrite(t *testing.T, data []byte, path []int, count int) []byte {
	t.Helper()
	value := primaryFields(t, data)[0]
	if len(path) == 0 {
		if count == -1 {
			changed := bytes.Clone(value.Bytes)
			changed[len(changed)-1] ^= 1
			return framingTLV(data[0], changed, 0)
		}
		return framingTLV(data[0], value.Bytes, count)
	}
	children := primaryFields(t, value.Bytes)
	parts := make([][]byte, len(children))
	for i, child := range children {
		parts[i] = child.FullBytes
	}
	if path[0] >= len(parts) {
		t.Fatal("invalid framing fixture path")
	}
	parts[path[0]] = packetFramingRewrite(t, parts[path[0]], path[1:], count)
	return framingTLV(data[0], bytes.Join(parts, nil), 0)
}

func TestAuthenticodePacketDefiniteFraming(t *testing.T) {
	points := []struct {
		name string
		path []int
		long int
	}{
		{"outer", nil, 3}, {"type", []int{0}, 1}, {"wrapper", []int{1}, 3}, {"signed", []int{1, 0}, 3},
		{"version", []int{1, 0, 0}, 1}, {"digest-set", []int{1, 0, 1}, 1}, {"digest-algorithm", []int{1, 0, 1, 0}, 1}, {"digest-oid", []int{1, 0, 1, 0, 0}, 1}, {"digest-null", []int{1, 0, 1, 0, 1}, 1},
		{"content-info", []int{1, 0, 2}, 1}, {"content-type", []int{1, 0, 2, 0}, 1}, {"content-wrapper", []int{1, 0, 2, 1}, 1}, {"content-sequence", []int{1, 0, 2, 1, 0}, 2},
		{"certificate-set", []int{1, 0, 3}, 3}, {"signer-set", []int{1, 0, 4}, 3}, {"signer", []int{1, 0, 4, 0}, 3}, {"signer-version", []int{1, 0, 4, 0, 0}, 1},
		{"signer-identity", []int{1, 0, 4, 0, 1}, 1}, {"signer-issuer", []int{1, 0, 4, 0, 1, 0}, 1}, {"signer-serial", []int{1, 0, 4, 0, 1, 1}, 1},
		{"signer-algorithm", []int{1, 0, 4, 0, 2}, 1}, {"signer-oid", []int{1, 0, 4, 0, 2, 0}, 1}, {"signer-null", []int{1, 0, 4, 0, 2, 1}, 1},
		{"signature-algorithm", []int{1, 0, 4, 0, 4}, 1}, {"signature-oid", []int{1, 0, 4, 0, 4, 0}, 1}, {"signature-null", []int{1, 0, 4, 0, 4, 1}, 1}, {"signature", []int{1, 0, 4, 0, 5}, 2},
	}
	for _, embedded := range []bool{false, true} {
		for _, surface := range []string{"pe", "catalog", "legacy-time"} {
			role := x509.ExtKeyUsageCodeSigning
			if surface == "legacy-time" {
				role = x509.ExtKeyUsageTimeStamping
			}
			chain := historicalPEChainFixture(t, crypto.SHA256, 1, embedded, role)
			body, packet := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
			crb := chain.crb
			var codeCRB string
			if surface == "catalog" {
				body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
				digest, err := peCatalogDigest(body, crypto.SHA256)
				if err != nil {
					t.Fatal(err)
				}
				packet = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, chain.bag)
			}
			if surface == "legacy-time" {
				code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
				expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
				var primary []byte
				body, primary = signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
				packet = appendLegacyTimestampWithSigner(t, primary, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, chain.bag)
				crb = code.crb + chain.crb
				codeCRB = code.crb
			}
			for _, point := range points {
				for _, count := range []int{point.long, 4} {
					label := fmt.Sprintf("packet-%s-root%v", surface, embedded)
					kind := fmt.Sprintf("%s-length%d", point.name, count)
					t.Run(label+"/"+kind, func(t *testing.T) {
						changed := packetFramingRewrite(t, packet, point.path, count)
						inputs := map[string][]byte{"trust.crb": []byte(crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
						object := body
						if surface == "catalog" {
							inputs["verified.cat"] = changed
						} else {
							object = replacePECertificate(t, body, changed)
						}
						db, err := signatures.Compile(inputs)
						if err != nil {
							t.Fatal(err)
						}
						e := New(config.EDRConfig{})
						e.current.Store(&generation{db: db})
						clean, name, scanErr := e.Scan(context.Background(), object)
						exportHistoricalScan(t, label, kind, object, inputs, clean, name, true)
						if os.Getenv("EDR_CERTIFICATE_FRAMING_EXPORT_ONLY") != "1" && (scanErr != nil || !clean || name != "") {
							t.Errorf("packet framing clean=%v name=%q err=%v", clean, name, scanErr)
						}
					})
				}
			}
			negative := []string{"signature-invalid"}
			if surface == "legacy-time" {
				negative = append(negative, "counter-signature-invalid", "time-root-absent")
			}
			for _, kind := range negative {
				label := fmt.Sprintf("packet-%s-root%v", surface, embedded)
				t.Run(label+"/"+kind, func(t *testing.T) {
					changed := packetFramingRewrite(t, packet, []int{1, 0, 4, 0, 5}, -1)
					inputs := map[string][]byte{"trust.crb": []byte(crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
					if kind == "counter-signature-invalid" {
						changed = bytes.Clone(packet)
						changed[len(changed)-10] ^= 1
					}
					if kind == "time-root-absent" {
						changed = packet
						inputs["trust.crb"] = []byte(codeCRB)
					}
					object := body
					if surface == "catalog" {
						inputs["verified.cat"] = changed
					} else {
						object = replacePECertificate(t, body, changed)
					}
					db, err := signatures.Compile(inputs)
					if err != nil {
						t.Fatal(err)
					}
					e := New(config.EDRConfig{})
					e.current.Store(&generation{db: db})
					clean, name, scanErr := e.Scan(context.Background(), object)
					exportHistoricalScan(t, label, kind, object, inputs, clean, name, false)
					if os.Getenv("EDR_CERTIFICATE_FRAMING_EXPORT_ONLY") != "1" && (scanErr != nil || clean || name != "Suspicious.Code") {
						t.Errorf("invalid proof clean=%v name=%q err=%v", clean, name, scanErr)
					}
				})
			}
		}
	}
}
