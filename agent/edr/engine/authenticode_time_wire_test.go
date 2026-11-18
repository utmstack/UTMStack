package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

type timeWireCase struct {
	name     string
	accepted bool
}

func timeWireCases() []timeWireCase {
	return []timeWireCase{
		{"utc", true}, {"generalized", true}, {"utc-minute", true}, {"generalized-minute", true},
		{"utc-tail", true}, {"generalized-tail", true}, {"utc-minute-tail", true}, {"generalized-minute-tail", true},
		{"utc-nul-tail", true}, {"generalized-nul-tail", true}, {"utc-z-offset-tail", true}, {"generalized-z-offset-tail", true},
		{"utc-day-rollover", true}, {"generalized-day-rollover", true},
		{"utc-offset-zero", false}, {"generalized-offset-zero", false}, {"utc-offset-positive", false}, {"utc-offset-negative", false},
		{"generalized-fraction", false}, {"utc-fraction", false}, {"generalized-comma", false},
		{"utc-no-zone", false}, {"generalized-no-zone", false}, {"utc-lowercase-zone", false},
		{"utc-month-zero", false}, {"utc-month-thirteen", false}, {"utc-day-zero", false}, {"utc-day-thirtytwo", false},
		{"utc-hour-twentyfour", false}, {"utc-minute-sixty", false}, {"utc-second-sixty", false},
		{"utc-nondigit", false}, {"utc-constructed", false}, {"generalized-constructed", false}, {"utc-empty", false}, {"utc-wrong-tag", false},
	}
}

func timeWireValue(t *testing.T, kind string, year int) asn1.RawValue {
	t.Helper()
	generalized := strings.HasPrefix(kind, "generalized")
	tag, content := asn1.TagUTCTime, fmt.Sprintf("%02d0101000000Z", year%100)
	if generalized {
		tag, content = asn1.TagGeneralizedTime, fmt.Sprintf("%04d0101000000Z", year)
	}
	switch {
	case strings.HasSuffix(kind, "-minute-tail"):
		content = content[:len(content)-3] + "Zignored"
	case strings.HasSuffix(kind, "-minute"):
		content = content[:len(content)-3] + "Z"
	case strings.HasSuffix(kind, "-nul-tail"):
		content += "\x00ignored"
	case strings.HasSuffix(kind, "-z-offset-tail"):
		content += "+0000"
	case strings.HasSuffix(kind, "-tail"):
		content += "ignored"
	case strings.HasSuffix(kind, "-day-rollover"):
		at := 2
		if generalized {
			at = 4
		}
		content = content[:at] + "0231" + content[at+4:]
	case strings.HasSuffix(kind, "-offset-zero"):
		content = strings.TrimSuffix(content, "Z") + "+0000"
	case strings.HasSuffix(kind, "-offset-positive"):
		content = strings.TrimSuffix(content, "Z") + "+0100"
	case strings.HasSuffix(kind, "-offset-negative"):
		content = strings.TrimSuffix(content, "Z") + "-0100"
	case strings.HasSuffix(kind, "-fraction"):
		content = strings.TrimSuffix(content, "Z") + ".5Z"
	case strings.HasSuffix(kind, "-comma"):
		content = strings.TrimSuffix(content, "Z") + ",5Z"
	case strings.HasSuffix(kind, "-no-zone"):
		content = strings.TrimSuffix(content, "Z")
	case strings.HasSuffix(kind, "-lowercase-zone"):
		content = strings.TrimSuffix(content, "Z") + "z"
	case strings.HasSuffix(kind, "-month-zero"):
		content = content[:2] + "00" + content[4:]
	case strings.HasSuffix(kind, "-month-thirteen"):
		content = content[:2] + "13" + content[4:]
	case strings.HasSuffix(kind, "-day-zero"):
		content = content[:4] + "00" + content[6:]
	case strings.HasSuffix(kind, "-day-thirtytwo"):
		content = content[:4] + "32" + content[6:]
	case strings.HasSuffix(kind, "-hour-twentyfour"):
		content = content[:6] + "24" + content[8:]
	case strings.HasSuffix(kind, "-minute-sixty"):
		content = content[:8] + "60" + content[10:]
	case strings.HasSuffix(kind, "-second-sixty"):
		content = content[:10] + "60Z"
	case strings.HasSuffix(kind, "-nondigit"):
		content = "x" + content[1:]
	case strings.HasSuffix(kind, "-empty"):
		content = ""
	case strings.HasSuffix(kind, "-wrong-tag"):
		tag = asn1.TagPrintableString
	}
	encoded, err := asn1.Marshal(asn1.RawValue{Tag: tag, IsCompound: strings.HasSuffix(kind, "-constructed"), Bytes: []byte(content)})
	if err != nil {
		t.Fatal(err)
	}
	return primaryFields(t, encoded)[0]
}

// Independently sign the original nonstandard time bytes; no parser normalization
// is used to generate a certificate's cryptographic proof.
func certificateTimeVariant(t *testing.T, cert *x509.Certificate, key *rsa.PrivateKey, hash crypto.Hash, kind string) []byte {
	return certificateValidityParts(t, cert, key, hash, []asn1.RawValue{timeWireValue(t, kind, 2020), timeWireValue(t, kind, 2040)})
}

func certificateValidityParts(t *testing.T, cert *x509.Certificate, key *rsa.PrivateKey, hash crypto.Hash, validity []asn1.RawValue) []byte {
	t.Helper()
	fields := primaryFields(t, primaryFields(t, cert.RawTBSCertificate)[0].Bytes)
	index := 3
	if fields[0].Class == asn1.ClassContextSpecific {
		index = 4
	}
	fields[index] = primaryNode(t, 0, asn1.TagSequence, validity)
	tbs := primaryNode(t, 0, asn1.TagSequence, fields)
	h := hash.New()
	h.Write(tbs.FullBytes)
	sig, err := signReferenceFixture(rand.Reader, key, hash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	signature, err := asn1.Marshal(asn1.BitString{Bytes: sig, BitLength: len(sig) * 8})
	if err != nil {
		t.Fatal(err)
	}
	original := primaryFields(t, primaryFields(t, cert.Raw)[0].Bytes)
	return derFixture(t, 0, asn1.TagSequence, tbs.FullBytes, original[1].FullBytes, signature)
}

func scanTimeWire(t *testing.T, label, kind string, body []byte, inputs map[string][]byte, want bool) {
	t.Helper()
	db, err := signatures.Compile(inputs)
	if err != nil {
		t.Fatal(err)
	}
	e := New(config.EDRConfig{})
	e.current.Store(&generation{db: db})
	clean, name, scanErr := e.Scan(context.Background(), body)
	exportHistoricalScan(t, label, kind, body, inputs, clean, name, want)
	if os.Getenv("EDR_CERTIFICATE_TIME_EXPORT_ONLY") != "1" && (scanErr != nil || clean != want || !want && name != "Suspicious.Code") {
		t.Errorf("time wire: clean=%v name=%q err=%v want=%v", clean, name, scanErr, want)
	}
}

func TestAuthenticodeValidityTimeWire(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			for _, surface := range []string{"pe", "catalog", "legacy-time"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				chain := historicalPEChainFixture(t, hash, 1, embedded, role)
				code := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
				primaryBody, primaryDER := signedPEWithSigner(t, code.leaf, code.leafKey, code.bag)
				for _, target := range []string{"leaf", "intermediate"} {
					for _, tc := range timeWireCases() {
						label := fmt.Sprintf("validity-%s-%s-%s-root%v", surface, target, hash, embedded)
						t.Run(label+"/"+tc.name, func(t *testing.T) {
							cert, key := chain.leaf, chain.parentKey
							if target == "intermediate" {
								cert, key = chain.parent, chain.rootKey
							}
							variant := certificateTimeVariant(t, cert, key, hash, tc.name)
							bag := bytes.Replace(chain.bag, cert.Raw, variant, 1)
							body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bag)
							inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
							if surface == "catalog" {
								body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
								digest, err := peCatalogDigest(body, crypto.SHA256)
								if err != nil {
									t.Fatal(err)
								}
								inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
							} else if surface == "legacy-time" {
								packet := appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, bag)
								body = replacePECertificate(t, primaryBody, packet)
								inputs["trust.crb"] = []byte(code.crb + chain.crb)
							}
							// The pinned reader maps both 2020 and 2040 GeneralizedTime
							// values to 2020; their zero interval is rejected.
							scanTimeWire(t, label, tc.name, body, inputs, tc.accepted && !strings.HasPrefix(tc.name, "generalized"))
						})
					}
				}
			}
		}
	}
}

func counterTimeVariant(t *testing.T, der []byte, key *rsa.PrivateKey, hash crypto.Hash, value asn1.RawValue) []byte {
	t.Helper()
	der = counterWireVariant(t, der, key, hash, "valid", nil)
	outer := primaryFields(t, primaryFields(t, der)[0].Bytes)
	signed := primaryFields(t, primaryFields(t, outer[1].Bytes)[0].Bytes)
	primary := primaryFields(t, primaryFields(t, signed[4].Bytes)[0].Bytes)
	unsigned := primaryFields(t, primary[6].Bytes)
	attribute := primaryFields(t, unsigned[0].Bytes)
	fields := primaryFields(t, primaryFields(t, attribute[1].Bytes)[0].Bytes)
	attrs := primaryFields(t, fields[3].Bytes)
	for i, attr := range attrs {
		parts := primaryFields(t, attr.Bytes)
		if bytes.Equal(parts[0].FullBytes, oidFixture(t, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5})) {
			attrs[i] = primaryNode(t, 0, asn1.TagSequence, []asn1.RawValue{parts[0], primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{value})})
		}
	}
	sort.Slice(attrs, func(i, j int) bool { return bytes.Compare(attrs[i].FullBytes, attrs[j].FullBytes) < 0 })
	encoded := primaryNode(t, 0, asn1.TagSet, attrs)
	h := hash.New()
	h.Write(encoded.FullBytes)
	sig, err := signReferenceFixture(rand.Reader, key, hash, h.Sum(nil))
	if err != nil {
		t.Fatal(err)
	}
	encodedSig, err := asn1.Marshal(sig)
	if err != nil {
		t.Fatal(err)
	}
	fields[3] = primaryNode(t, asn1.ClassContextSpecific, 0, attrs)
	fields[5] = primaryFields(t, encodedSig)[0]
	attribute[1] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, fields)})
	unsigned[0] = primaryNode(t, 0, asn1.TagSequence, attribute)
	primary[6] = primaryNode(t, asn1.ClassContextSpecific, 1, unsigned)
	signed[4] = primaryNode(t, 0, asn1.TagSet, []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, primary)})
	outer[1] = primaryNode(t, asn1.ClassContextSpecific, 0, []asn1.RawValue{primaryNode(t, 0, asn1.TagSequence, signed)})
	return primaryNode(t, 0, asn1.TagSequence, outer).FullBytes
}

func TestAuthenticodeCounterTimeWire(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			code := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
			tsa := historicalPEChainFixture(t, hash, 1, embedded, x509.ExtKeyUsageTimeStamping)
			body, der := signedPEWithSigner(t, code.leaf, code.leafKey, code.bag)
			timestamped := appendLegacyTimestampWithSigner(t, der, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), tsa.leaf, tsa.leafKey, tsa.bag)
			for _, tc := range timeWireCases() {
				label := fmt.Sprintf("counter-time-%s-root%v", hash, embedded)
				t.Run(label+"/"+tc.name, func(t *testing.T) {
					packet := counterTimeVariant(t, timestamped, tsa.leafKey, hash, timeWireValue(t, tc.name, 2024))
					scanTimeWire(t, label, tc.name, replacePECertificate(t, body, packet), map[string][]byte{"trust.crb": []byte(code.crb + tsa.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}, tc.accepted)
				})
			}
		}
	}
}

func TestAuthenticodeTimeBlocklist(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			chain := historicalPEChainFixture(t, hash, 1, embedded, x509.ExtKeyUsageCodeSigning)
			for _, target := range []string{"leaf", "intermediate"} {
				for _, tc := range timeWireCases() {
					label := fmt.Sprintf("time-block-%s-%s-root%v", target, hash, embedded)
					t.Run(label+"/"+tc.name, func(t *testing.T) {
						cert, key := chain.leaf, chain.parentKey
						if target == "intermediate" {
							cert, key = chain.parent, chain.rootKey
						}
						variant := certificateTimeVariant(t, cert, key, hash, tc.name)
						body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bytes.Replace(chain.bag, cert.Raw, variant, 1))
						inputs := map[string][]byte{"block.crb": []byte(policyBlockRow(t, cert, "Blocked.Time", "exact")), "raw.ndb": []byte("Unmatched.Decoy:0:*:4544522d414253454e542d424f4459\n")}
						db, err := signatures.Compile(inputs)
						if err != nil {
							t.Fatal(err)
						}
						e := New(config.EDRConfig{})
						e.current.Store(&generation{db: db})
						clean, name, scanErr := e.Scan(context.Background(), body)
						want := !(tc.accepted && !strings.HasPrefix(tc.name, "generalized"))
						exportHistoricalScan(t, label, tc.name, body, inputs, clean, name, want)
						wantName := ""
						if !want {
							wantName = "Blocked.Time"
						}
						if os.Getenv("EDR_CERTIFICATE_TIME_EXPORT_ONLY") != "1" && (scanErr != nil || clean != want || name != wantName) {
							t.Errorf("block time: clean=%v name=%q err=%v want clean=%v name=%q", clean, name, scanErr, want, wantName)
						}
					})
				}
			}
		}
	}
}

func TestAuthenticodeTimeCenturies(t *testing.T) {
	raw := func(tag int, s string) asn1.RawValue {
		data, err := asn1.Marshal(asn1.RawValue{Tag: tag, Bytes: []byte(s)})
		if err != nil {
			t.Fatal(err)
		}
		return primaryFields(t, data)[0]
	}
	u := asn1.TagUTCTime
	g := asn1.TagGeneralizedTime
	cases := []struct {
		name   string
		values []asn1.RawValue
		want   bool
	}{
		{"generalized-century-range", []asn1.RawValue{raw(g, "20200101000000Z"), raw(g, "21400101000000Z")}, true},
		{"generalized-century-nondigits", []asn1.RawValue{raw(g, "20x?0101000000Z"), raw(g, "21x?0101000000Z")}, true},
		{"generalized-zero-to-max-year", []asn1.RawValue{raw(g, "00000101000000Z"), raw(g, "99990101000000Z")}, true},
		{"utc-pivot-range", []asn1.RawValue{raw(u, "500101000000Z"), raw(u, "490101000000Z")}, true},
		{"utc-pivot-reversed", []asn1.RawValue{raw(u, "490101000000Z"), raw(u, "500101000000Z")}, false},
		{"zero-normalized-interval", []asn1.RawValue{raw(g, "20200101000000Z"), raw(g, "20400101000000Z")}, false},
		{"reversed-interval", []asn1.RawValue{raw(u, "400101000000Z"), raw(u, "200101000000Z")}, false},
		{"zero-interval", []asn1.RawValue{raw(u, "200101000000Z"), raw(u, "200101000000Z")}, false},
		{"extra-validity-entry", []asn1.RawValue{raw(u, "200101000000Z"), raw(u, "400101000000Z"), raw(u, "300101000000Z")}, false},
	}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			for _, surface := range []string{"pe", "catalog", "legacy-time"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				chain := historicalPEChainFixture(t, hash, 1, embedded, role)
				code := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
				codeBody, codeDER := signedPEWithSigner(t, code.leaf, code.leafKey, code.bag)
				for _, tc := range cases {
					label := fmt.Sprintf("century-%s-%s-root%v", surface, hash, embedded)
					t.Run(label+"/"+tc.name, func(t *testing.T) {
						variant := certificateValidityParts(t, chain.leaf, chain.parentKey, hash, tc.values)
						bag := bytes.Replace(chain.bag, chain.leaf.Raw, variant, 1)
						body, _ := signedPEWithSigner(t, chain.leaf, chain.leafKey, bag)
						inputs := map[string][]byte{"trust.crb": []byte(chain.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
						if surface == "catalog" {
							body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
							digest, err := peCatalogDigest(body, crypto.SHA256)
							if err != nil {
								t.Fatal(err)
							}
							inputs["verified.cat"] = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), chain.leaf, chain.leafKey, bag)
						} else if surface == "legacy-time" {
							packet := appendLegacyTimestampWithSigner(t, codeDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), chain.leaf, chain.leafKey, bag)
							body = replacePECertificate(t, codeBody, packet)
							inputs["trust.crb"] = []byte(code.crb + chain.crb)
						}
						scanTimeWire(t, label, tc.name, body, inputs, tc.want)
					})
				}
			}
		}
	}
}

func TestAuthenticodeCounterTimeBoundaries(t *testing.T) {
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, embedded := range []bool{false, true} {
			for _, futurePrimary := range []bool{false, true} {
				code := historicalPEChainFixture(t, crypto.SHA256, 1, true, x509.ExtKeyUsageCodeSigning)
				tsa := historicalPEChainFixture(t, hash, 1, embedded, x509.ExtKeyUsageTimeStamping)
				cert := code.leaf
				if futurePrimary {
					template := *cert
					template.NotBefore = time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
					template.SignatureAlgorithm = 0
					der, err := x509.CreateCertificate(rand.Reader, &template, code.parent, &code.leafKey.PublicKey, code.parentKey)
					if err != nil {
						t.Fatal(err)
					}
					cert, err = x509.ParseCertificate(der)
					if err != nil {
						t.Fatal(err)
					}
				}
				bag := bytes.Replace(code.bag, code.leaf.Raw, cert.Raw, 1)
				body, der := signedPEWithSigner(t, cert, code.leafKey, bag)
				timestamped := appendLegacyTimestampWithSigner(t, der, time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC), tsa.leaf, tsa.leafKey, tsa.bag)
				cases := []struct {
					name, wire string
					tag        int
					want       bool
				}{
					{"future-in-primary-window", "310101000000Z", asn1.TagUTCTime, true},
					{"earliest-primary-boundary", "200101000000Z", asn1.TagUTCTime, !futurePrimary},
					{"latest-primary-boundary", "400101000000Z", asn1.TagUTCTime, true},
					{"outside-primary-window", "410101000000Z", asn1.TagUTCTime, false},
					{"generalized-future-folds-to-2020", "20310101000000Z", asn1.TagGeneralizedTime, !futurePrimary},
					{"generalized-nondigits-fold-to-2020", "20x?0101000000Z", asn1.TagGeneralizedTime, !futurePrimary},
					{"utc-pivot-1950", "500101000000Z", asn1.TagUTCTime, false},
					{"utc-pivot-2049", "490101000000Z", asn1.TagUTCTime, false},
				}
				for _, tc := range cases {
					label := fmt.Sprintf("counter-boundary-%s-root%v-primaryfuture%v", hash, embedded, futurePrimary)
					t.Run(label+"/"+tc.name, func(t *testing.T) {
						encoded, err := asn1.Marshal(asn1.RawValue{Tag: tc.tag, Bytes: []byte(tc.wire)})
						if err != nil {
							t.Fatal(err)
						}
						packet := counterTimeVariant(t, timestamped, tsa.leafKey, hash, primaryFields(t, encoded)[0])
						scanTimeWire(t, label, tc.name, replacePECertificate(t, body, packet), map[string][]byte{"trust.crb": []byte(code.crb + tsa.crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}, tc.want)
					})
				}
			}
		}
	}
}
