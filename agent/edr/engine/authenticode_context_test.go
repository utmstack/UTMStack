package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func contextMarshal(t *testing.T, value any) asn1.RawValue {
	t.Helper()
	b, err := asn1.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return primaryFields(t, b)[0]
}

func contextEdit(t *testing.T, der []byte, edit func([]asn1.RawValue)) []byte {
	t.Helper()
	outer := primaryFields(t, primaryFields(t, der)[0].Bytes)
	signed := primaryFields(t, primaryFields(t, outer[1].Bytes)[0].Bytes)
	edit(signed)
	sequence := primaryNode(t, 0, asn1.TagSequence, signed)
	outer[1] = primaryNode(t, 2, 0, []asn1.RawValue{sequence})
	return primaryNode(t, 0, asn1.TagSequence, outer).FullBytes
}

func contextScan(t *testing.T, label, kind string, body []byte, inputs map[string][]byte, want bool) {
	t.Helper()
	db, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	clean, name, err := e.Scan(context.Background(), body)
	if clean != want || err != nil {
		t.Errorf("clean=%v name=%q err=%v want=%v", clean, name, err, want)
	}
	exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
}

func contextIndirect(t *testing.T, hash crypto.Hash, digest []byte, cab bool, kind string) asn1.RawValue {
	t.Helper()
	seq := func(v ...asn1.RawValue) asn1.RawValue { return primaryNode(t, 0, asn1.TagSequence, v) }
	marshal := func(v any) asn1.RawValue { return contextMarshal(t, v) }
	object := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 15}
	if cab {
		object[len(object)-1] = 25
	}
	if kind == "wrong-object" {
		object[len(object)-1] = 99
	}
	typeInfo := seq(marshal(object))
	if kind == "description" {
		typeInfo = seq(marshal(object), seq(marshal(true)))
	}
	oid := marshal(primaryHashOID(hash, kind == "rsa-alias"))
	if kind == "oiw-digest" {
		oid = marshal(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29})
	}
	if kind == "unknown-digest" {
		oid = marshal(asn1.ObjectIdentifier{1, 2, 3, 4, 999})
	}
	if kind == "invalid-oid" {
		oid = primaryFields(t, []byte{6, 1, 0x80})[0]
	}
	if kind == "empty-oid" {
		oid = primaryFields(t, []byte{6, 0})[0]
	}
	param := primaryFields(t, []byte{5, 0})[0]
	if kind == "nonempty-null" {
		param = primaryFields(t, []byte{5, 1, 0})[0]
	}
	if kind == "constructed-null" {
		param = primaryFields(t, []byte{0x25, 0})[0]
	}
	if kind == "wrong-parameter" {
		param = marshal([]byte{})
	}
	algorithm := seq(oid, param)
	if kind == "absent-parameter" {
		algorithm = seq(oid)
	}
	if kind == "extra-parameter" {
		algorithm = seq(oid, param, param)
	}
	if kind == "short-digest" {
		digest = digest[:len(digest)-1]
	}
	if kind == "long-digest" {
		digest = append(bytes.Clone(digest), 0)
	}
	if kind == "wrong-digest" {
		digest = bytes.Clone(digest)
		digest[0] ^= 1
	}
	info := seq(algorithm, marshal(digest))
	if kind == "extra-digest" {
		info = seq(algorithm, marshal(digest), param)
	}
	if kind == "missing-digest" {
		info = seq(algorithm)
	}
	out := seq(typeInfo, info)
	if kind == "extra-content" {
		out = seq(typeInfo, info, param)
	}
	return out
}

func TestAuthenticodeIndirectWireContexts(t *testing.T) {
	chain := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
	pe, embedded := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
	unsigned := simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
	cab := cabFixture(t, 0, [][]byte{bytes.Repeat([]byte{0x90}, 32)}, []cabFixtureFile{{name: "marker.bin", size: 32}}, false)
	baseCatalog := signedCatalogWithSignerForEngine(t, bytes.Repeat([]byte{1}, 32), bytes.Repeat([]byte{2}, 20), chain.leaf, chain.leafKey, chain.bag)
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		for _, format := range []string{"image", "catalog-pe", "catalog-cab"} {
			for _, kind := range []string{"valid", "rsa-alias", "oiw-digest", "unknown-digest", "invalid-oid", "empty-oid", "absent-parameter", "nonempty-null", "constructed-null", "wrong-parameter", "extra-parameter", "short-digest", "long-digest", "wrong-digest", "extra-digest", "missing-digest", "wrong-object", "description", "extra-content"} {
				label := "indirect-" + format + "-" + hash.String()
				t.Run(label+"/"+kind, func(t *testing.T) {
					body, base := pe, embedded
					var digest []byte
					var err error
					if format == "image" {
						digest, err = peAuthenticodeDigest(pe, hash)
					} else if format == "catalog-pe" {
						body = unsigned
						digest, err = peCatalogDigest(body, hash)
					} else {
						body = cab
						h := hash.New()
						h.Write(body)
						digest = h.Sum(nil)
					}
					if err != nil {
						t.Fatal(err)
					}
					indirect := contextIndirect(t, hash, digest, format == "catalog-cab", kind)
					if format != "image" {
						base = baseCatalog
					}
					variant := contextEdit(t, base, func(signed []asn1.RawValue) {
						info := primaryFields(t, signed[2].Bytes)
						content := indirect
						if format != "image" {
							fields := primaryFields(t, primaryFields(t, info[1].Bytes)[0].Bytes)
							member := primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{contextMarshal(t, asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 4}), primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{indirect})})
							row := primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{contextMarshal(t, []byte("member")), primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{member})})
							fields[4] = primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{row})
							content = primaryNode(t, 0, asn1.TagSequence, fields)
						}
						info[1] = primaryNode(t, 2, 0, []asn1.RawValue{content})
						signed[2] = primaryNode(t, 0, asn1.TagSequence, info)
					})
					variant = primaryVariant(t, variant, chain.leafKey, crypto.SHA256, "valid", nil)
					inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
					if format == "image" {
						body = replacePECertificate(t, pe, variant)
					} else {
						inputs["target.cat"] = variant
					}
					want := kind == "valid" || kind == "rsa-alias" || kind == "absent-parameter" || kind == "description" || (kind == "oiw-digest" || kind == "unknown-digest" || kind == "invalid-oid") && hash == crypto.SHA512
					if format != "image" && hash != crypto.SHA1 && hash != crypto.SHA256 {
						want = false
					}
					contextScan(t, label, kind, body, inputs, want)
				})
			}
		}
	}
}

func TestAuthenticodeUnsignedWireContexts(t *testing.T) {
	chain := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
	pe, embedded := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
	unsigned := simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
	digest, err := peCatalogDigest(unsigned, crypto.SHA256)
	if err != nil {
		t.Fatal(err)
	}
	catalog := signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{1}, 20), chain.leaf, chain.leafKey, chain.bag)
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		for _, format := range []string{"image", "catalog"} {
			for _, kind := range []string{"empty", "nested", "timestamp", "both", "duplicate-nested", "duplicate-timestamp", "duplicate-unknown", "unknown-no-value", "nested-no-set", "timestamp-no-set", "nested-extra-value", "timestamp-extra-value", "unknown-invalid-oid", "unknown-empty-oid", "unknown-constructed-oid", "nested-and-timestamp-values"} {
				label := "unsigned-" + format + "-" + hash.String()
				t.Run(label+"/"+kind, func(t *testing.T) {
					base := embedded
					if format == "catalog" {
						base = catalog
					}
					base = primaryVariant(t, base, chain.leafKey, hash, "valid", nil)
					variant := contextEdit(t, base, func(signed []asn1.RawValue) {
						fields := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
						attr := func(oid asn1.ObjectIdentifier, mode string) asn1.RawValue {
							v := []asn1.RawValue{contextMarshal(t, oid)}
							if mode != "absent" {
								v = append(v, primaryNode(t, 0, asn1.TagSet, nil))
							}
							if mode == "extra" {
								v = append(v, contextMarshal(t, true))
							}
							return primaryNode(t, 0, asn1.TagSequence, v)
						}
						nested := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 4, 1}
						timestamp := asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 3, 3, 1}
						unknown := asn1.ObjectIdentifier{1, 2, 3, 4, 999}
						var attrs []asn1.RawValue
						switch kind {
						case "unknown-invalid-oid", "unknown-empty-oid", "unknown-constructed-oid":
							raw := []byte{6, 1, 0x80}
							if kind == "unknown-empty-oid" {
								raw = []byte{6, 0}
							}
							if kind == "unknown-constructed-oid" {
								raw = []byte{0x26, 0}
							}
							attrs = []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, primaryFields(t, raw))}
						case "nested-and-timestamp-values":
							n := attr(nested, "")
							v := primaryFields(t, n.Bytes)
							v[1] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{contextMarshal(t, true), contextMarshal(t, false)})
							attrs = []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, v), attr(timestamp, "")}
						case "nested":
							attrs = []asn1.RawValue{attr(nested, "")}
						case "timestamp":
							attrs = []asn1.RawValue{attr(timestamp, "")}
						case "both":
							attrs = []asn1.RawValue{attr(nested, ""), attr(timestamp, "")}
						case "duplicate-nested":
							attrs = []asn1.RawValue{attr(nested, ""), attr(nested, "")}
						case "duplicate-timestamp":
							attrs = []asn1.RawValue{attr(timestamp, ""), attr(timestamp, "")}
						case "duplicate-unknown":
							attrs = []asn1.RawValue{attr(unknown, ""), attr(unknown, "")}
						case "unknown-no-value":
							attrs = []asn1.RawValue{attr(unknown, "absent")}
						case "nested-no-set":
							attrs = []asn1.RawValue{attr(nested, "absent")}
						case "timestamp-no-set":
							attrs = []asn1.RawValue{attr(timestamp, "absent")}
						case "nested-extra-value":
							attrs = []asn1.RawValue{attr(nested, "extra")}
						case "timestamp-extra-value":
							attrs = []asn1.RawValue{attr(timestamp, "extra")}
						}
						fields = append(fields, primaryNode(t, 2, 1, attrs))
						signed[4] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, fields)})
					})
					inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
					body := replacePECertificate(t, pe, variant)
					if format == "catalog" {
						body = unsigned
						inputs["target.cat"] = variant
					}
					want := kind == "empty" || kind == "nested" || kind == "timestamp" || kind == "both" || kind == "duplicate-unknown" || kind == "unknown-no-value" || kind == "unknown-invalid-oid" || kind == "unknown-empty-oid" || kind == "nested-and-timestamp-values"
					contextScan(t, label, kind, body, inputs, want)
				})
			}
		}
	}
}

func TestAuthenticodeDigestOIDContexts(t *testing.T) {
	chain := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
	pe, embedded := signedPEWithSigner(t, chain.leaf, chain.leafKey, chain.bag)
	for _, hash := range []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512} {
		for _, position := range []string{"header", "signer", "both"} {
			for _, kind := range []string{"oiw", "invalid", "empty"} {
				label := fmt.Sprintf("algorithm-%s-%s", position, hash)
				t.Run(label+"/"+kind, func(t *testing.T) {
					base := primaryVariant(t, embedded, chain.leafKey, hash, "valid", nil)
					variant := contextEdit(t, base, func(signed []asn1.RawValue) {
						id := contextMarshal(t, asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29})
						if kind == "invalid" {
							id = primaryFields(t, []byte{6, 1, 0x80})[0]
						}
						if kind == "empty" {
							id = primaryFields(t, []byte{6, 0})[0]
						}
						algorithm := primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{id, primaryFields(t, []byte{5, 0})[0]})
						if position != "signer" {
							signed[1] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{algorithm})
						}
						if position != "header" {
							fields := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
							fields[2] = algorithm
							signed[4] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, fields)})
						}
					})
					contextScan(t, label, kind, replacePECertificate(t, pe, variant), map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}, hash == crypto.SHA512 && kind != "empty")
				})
			}
		}
	}
}
