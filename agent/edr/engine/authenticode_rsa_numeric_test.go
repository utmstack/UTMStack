package engine

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Fixture-only integer signing has no dependency on the production verifier.
// Generated private factors/exponents stay in memory and are never exported.
type numericFixtureKey struct{ n, e, d *big.Int }

func numericStandardKey(key *rsa.PrivateKey) numericFixtureKey {
	return numericFixtureKey{key.N, big.NewInt(int64(key.E)), key.D}
}

func numericFixtureGenerate(t *testing.T, bits int, exponent *big.Int) numericFixtureKey {
	t.Helper()
	for {
		p, err := rand.Prime(rand.Reader, bits/2)
		if err != nil {
			t.Fatal(err)
		}
		q, err := rand.Prime(rand.Reader, bits-bits/2)
		if err != nil {
			t.Fatal(err)
		}
		if p.Cmp(q) == 0 {
			continue
		}
		n := new(big.Int).Mul(p, q)
		if n.BitLen() != bits {
			continue
		}
		phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))
		d := new(big.Int).ModInverse(exponent, phi)
		if d == nil {
			continue
		}
		return numericFixtureKey{n, new(big.Int).Set(exponent), d}
	}
}

func numericFixtureDigestInfo(t *testing.T, hash crypto.Hash, digest []byte) []byte {
	t.Helper()
	prefix := map[crypto.Hash]string{
		crypto.MD5:    "3020300c06082a864886f70d020505000410",
		crypto.SHA1:   "3021300906052b0e03021a05000414",
		crypto.SHA256: "3031300d060960864801650304020105000420",
		crypto.SHA384: "3041300d060960864801650304020205000430",
		crypto.SHA512: "3051300d060960864801650304020305000440",
	}[hash]
	if prefix == "" || len(digest) != hash.Size() {
		t.Fatal("invalid fixture digest")
	}
	der, err := hex.DecodeString(prefix)
	if err != nil {
		t.Fatal(err)
	}
	return append(der, digest...)
}

func (key numericFixtureKey) sign(t *testing.T, hash crypto.Hash, digest []byte, mode string) []byte {
	t.Helper()
	width := (key.n.BitLen() + 7) / 8
	info := numericFixtureDigestInfo(t, hash, digest)
	var value *big.Int
	if key.n.Sign() == 0 || key.d == nil || width < len(info)+11 {
		value = big.NewInt(1)
	} else {
		block := make([]byte, width)
		block[1] = 1
		for i := 2; i < width-len(info)-1; i++ {
			block[i] = 0xff
		}
		copy(block[width-len(info):], info)
		value = new(big.Int).Exp(new(big.Int).SetBytes(block), key.d, key.n)
		if len(value.Bytes()) < width-1 {
			value.Add(value, key.n)
		}
	}
	if mode == "plus-N" {
		value.Add(value, key.n)
	}
	encoded := value.Bytes()
	if mode == "padded" || mode == "bad-proof" || mode == "wide-field" || mode == "blocked" {
		if len(encoded) < 128 {
			encoded = value.FillBytes(make([]byte, 128))
		}
	}
	if mode == "wide-field" {
		encoded = value.FillBytes(make([]byte, 514))
	}
	if mode == "short-field" {
		encoded = make([]byte, 127)
		copy(encoded, value.Bytes())
	}
	if mode == "bad-proof" {
		encoded = bytes.Clone(encoded)
		encoded[len(encoded)-1] ^= 1
	}
	return encoded
}

func numericFixtureSPKI(key numericFixtureKey, width int) []byte {
	modulus := key.n.Bytes()
	if width < len(modulus) {
		width = len(modulus)
	}
	modulus = key.n.FillBytes(make([]byte, width))
	exponent := key.e.Bytes()
	if len(exponent) == 0 {
		exponent = []byte{0}
	}
	seq := func(data ...[]byte) []byte { return publicKeyTLV(0x30, bytes.Join(data, nil), false) }
	algorithm := seq([]byte{6, 9, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 1, 1, 1}, []byte{5, 0})
	pub := seq(publicKeyTLV(2, modulus, false), publicKeyTLV(2, exponent, false))
	return seq(algorithm, publicKeyTLV(3, append([]byte{0}, pub...), false))
}

func numericFixtureCertificate(t *testing.T, cert *x509.Certificate, public numericFixtureKey, width int, issuer numericFixtureKey, hash crypto.Hash, mode string) []byte {
	t.Helper()
	fields := primaryFields(t, primaryFields(t, cert.RawTBSCertificate)[0].Bytes)
	parts := make([][]byte, len(fields))
	for i, f := range fields {
		parts[i] = f.FullBytes
	}
	parts[6] = numericFixtureSPKI(public, width)
	tbs := publicKeyTLV(0x30, bytes.Join(parts, nil), false)
	h := hash.New()
	h.Write(tbs)
	signature := issuer.sign(t, hash, h.Sum(nil), mode)
	if len(signature) <= 512 {
		signature = append([]byte{0}, signature...)
	}
	algorithm := primaryFields(t, primaryFields(t, cert.Raw)[0].Bytes)[1].FullBytes
	return publicKeyTLV(0x30, bytes.Join([][]byte{tbs, algorithm, publicKeyTLV(3, signature, false)}, nil), false)
}

func numericFixtureAt(t *testing.T, data []byte, path []int) asn1.RawValue {
	t.Helper()
	value := primaryFields(t, data)[0]
	for _, index := range path {
		fields := primaryFields(t, value.Bytes)
		if index >= len(fields) {
			t.Fatal("invalid numeric path")
		}
		value = fields[index]
	}
	return value
}

func numericFixtureReplace(t *testing.T, data []byte, path []int, replacement []byte) []byte {
	t.Helper()
	if len(path) == 0 {
		return replacement
	}
	value := primaryFields(t, data)[0]
	fields := primaryFields(t, value.Bytes)
	parts := make([][]byte, len(fields))
	for i, f := range fields {
		parts[i] = f.FullBytes
	}
	parts[path[0]] = numericFixtureReplace(t, parts[path[0]], path[1:], replacement)
	return publicKeyTLV(data[0], bytes.Join(parts, nil), false)
}

func numericFixturePacketProof(t *testing.T, packet []byte, key numericFixtureKey, mode string, counter bool) []byte {
	t.Helper()
	path := []int{1, 0, 4, 0}
	if counter {
		path = append(path, 6, 0, 1, 0)
	}
	attrs := numericFixtureAt(t, packet, append(append([]int{}, path...), 3)).FullBytes
	set := append([]byte{0x31}, attrs[1:]...)
	h := crypto.SHA256.New()
	h.Write(set)
	sig := key.sign(t, crypto.SHA256, h.Sum(nil), mode)
	return numericFixtureReplace(t, packet, append(path, 5), publicKeyTLV(4, sig, false))
}

func TestExportAuthenticodeNumericKeys(t *testing.T) {
	root := os.Getenv("EDR_RSA_NUMERIC_EXPORT")
	if root == "" {
		t.Skip("explicit owned-fixture export only")
	}
	boundRoot := os.Getenv("EDR_RSA_BOUND_ROOT_EXPORT") == "1"
	type profile struct {
		name        string
		bits, width int
		exponent    *big.Int
	}
	profiles := []profile{}
	for _, bits := range []int{512, 768, 1016, 1023, 1024, 1025, 2048, 4096, 4104} {
		profiles = append(profiles, profile{fmt.Sprintf("n%d-e65537", bits), bits, max(128, (bits+7)/8), big.NewInt(65537)})
	}
	for _, e := range []int64{1, 3, 17, 2147483647, 2147483649, 4294967311, 9223372036854775807} {
		profiles = append(profiles, profile{fmt.Sprintf("n1024-e%d", e), 1024, 128, big.NewInt(e)})
	}
	for _, bits := range []int{64, 128, 1024, 4096, 4104} {
		e := new(big.Int).Add(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)), big.NewInt(51))
		profiles = append(profiles, profile{fmt.Sprintf("n1024-e%dbits", bits), 1024, max(128, (bits+7)/8), e})
	}
	keys := map[string]numericFixtureKey{}
	for _, p := range profiles {
		keys[p.name] = numericFixtureGenerate(t, p.bits, p.exponent)
	}
	// Invalid mathematical keys still have readable unsigned wire identities.
	for _, e := range []int64{0, 2} {
		name := fmt.Sprintf("n1024-e%d", e)
		p := profile{name, 1024, 128, big.NewInt(e)}
		profiles = append(profiles, p)
		k := keys["n1024-e65537"]
		k.e = p.exponent
		k.d = nil
		keys[name] = k
	}
	profiles = append(profiles, profile{"n0-e0", 0, 128, big.NewInt(0)})
	keys["n0-e0"] = numericFixtureKey{big.NewInt(0), big.NewInt(0), nil}
	for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
		for _, surface := range []string{"pe", "catalog", "legacy-time"} {
			role := x509.ExtKeyUsageCodeSigning
			if surface == "legacy-time" {
				role = x509.ExtKeyUsageTimeStamping
			}
			base := historicalPEChainFixture(t, hash, 1, true, role)
			code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
			expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
			primary, primaryDER := signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
			for _, p := range profiles {
				for _, target := range []string{"leaf", "intermediate", "root"} {
					if boundRoot && target != "root" {
						continue
					}
					for _, embedded := range []bool{false, true} {
						for _, mode := range []string{"padded", "plus-N", "bad-proof", "short-field", "wide-field", "blocked"} {
							label := fmt.Sprintf("numeric-%s-%s-%s-root%v", surface, target, hash, embedded)
							kind := p.name + "-" + mode
							t.Run(label+"/"+kind, func(t *testing.T) {
								rootKey, parentKey, leafKey := numericStandardKey(base.rootKey), numericStandardKey(base.parentKey), numericStandardKey(base.leafKey)
								rootWidth, parentWidth, leafWidth := 128, 128, 128
								key := keys[p.name]
								if target == "root" {
									rootKey = key
									rootWidth = p.width
								}
								if target == "intermediate" {
									parentKey = key
									parentWidth = p.width
								}
								if target == "leaf" {
									leafKey = key
									leafWidth = p.width
								}
								rootDER := numericFixtureCertificate(t, base.root, rootKey, rootWidth, rootKey, hash, "padded")
								if boundRoot {
									rootDER = numericFixtureCertificate(t, base.root, rootKey, rootWidth, numericStandardKey(base.rootKey), hash, "padded")
								}
								parentMode, leafMode := "padded", "padded"
								if target == "root" {
									parentMode = mode
								}
								if target == "intermediate" {
									leafMode = mode
								}
								parentDER := numericFixtureCertificate(t, base.parent, parentKey, parentWidth, rootKey, hash, parentMode)
								leafDER := numericFixtureCertificate(t, base.leaf, leafKey, leafWidth, parentKey, hash, leafMode)
								bag := bytes.Join([][]byte{parentDER, leafDER}, nil)
								if embedded {
									bag = append(bytes.Clone(rootDER), bag...)
								}
								crb := base.crb
								if target == "root" {
									crb = replaceHistoricalCRBField(t, crb, 4, fmt.Sprintf("%X", key.n))
								}
								if key.n.Sign() == 0 && target == "root" {
									crb = replaceHistoricalCRBField(t, crb, 4, "00")
								}
								if boundRoot {
									crb = base.crb + crb
								}
								inputs := map[string][]byte{"trust.crb": []byte(crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
								body, packet := signedPEWithSigner(t, base.leaf, base.leafKey, bag)
								if surface == "catalog" {
									body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
									digest, err := peCatalogDigest(body, crypto.SHA256)
									if err != nil {
										t.Fatal(err)
									}
									packet = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), base.leaf, base.leafKey, bag)
								}
								if surface == "legacy-time" {
									packet = appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), base.leaf, base.leafKey, bag)
									body = primary
									inputs["trust.crb"] = []byte(code.crb + crb)
								}
								proofMode := "padded"
								if target == "leaf" {
									proofMode = mode
								}
								packet = numericFixturePacketProof(t, packet, leafKey, proofMode, surface == "legacy-time")
								if surface == "catalog" {
									inputs["verified.cat"] = packet
								} else {
									body = replacePECertificate(t, body, packet)
								}
								if mode == "blocked" {
									cert := base.leaf
									pub := leafKey
									if target == "intermediate" {
										cert = base.parent
										pub = parentKey
									}
									if target == "root" {
										cert = base.root
										pub = rootKey
									}
									line := policyBlockRow(t, cert, "Blocked.RSA", "exact")
									mod := fmt.Sprintf("%X", pub.n)
									if pub.n.Sign() == 0 {
										mod = "00"
									}
									line = replaceHistoricalCRBField(t, line, 4, mod)
									inputs["trust.crb"] = append([]byte(line), inputs["trust.crb"]...)
									inputs["raw.ndb"] = []byte("Decoy:0:*:6861726d6c6573732d6e657665722d70726573656e742d6465636f79\n")
								}
								db, err := signatures.Compile(inputs)
								if err != nil {
									t.Fatal(err)
								}
								e := New(config.EDRConfig{})
								e.current.Store(&generation{db: db})
								clean, name, _ := e.Scan(context.Background(), body)
								t.Setenv("EDR_HISTORICAL_CHAIN_EXPORT", root)
								exportHistoricalScan(t, label, kind, body, inputs, clean, name, false)
							})
						}
					}
				}
			}
		}
	}
}

func TestAuthenticodeNumericKeyReference(t *testing.T) {
	file, err := os.Open("testdata/authenticode_rsa_signed.jsonl.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	z, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	reader := bufio.NewScanner(z)
	reader.Buffer(make([]byte, 4096), 2<<20)
	count := 0
	for reader.Scan() {
		var row struct {
			Label  string
			Body   []byte
			Inputs map[string][]byte
			Clean  bool
			Name   string
		}
		if err := json.Unmarshal(reader.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		t.Run(strings.ReplaceAll(row.Label, " ", "_"), func(t *testing.T) {
			db, err := signatures.Compile(row.Inputs)
			if err != nil {
				t.Fatal(err)
			}
			e := New(config.EDRConfig{})
			e.current.Store(&generation{db: db})
			clean, name, err := e.Scan(context.Background(), row.Body)
			if err != nil || clean != row.Clean || name != row.Name {
				t.Errorf("numeric clean=%v name=%q err=%v want=%v/%q", clean, name, err, row.Clean, row.Name)
			}
		})
		count++
	}
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 6984 {
		t.Fatal("missing numeric reference cases")
	}
}

func TestExportAuthenticodeCRBKeyWidths(t *testing.T) {
	root := os.Getenv("EDR_RSA_CRYPTOWIDTH_EXPORT")
	if root == "" {
		t.Skip("explicit owned CRB-width export only")
	}
	for _, bits := range []int{4105, 4112} {
		key := numericFixtureGenerate(t, bits, big.NewInt(65537))
		for _, hash := range []crypto.Hash{crypto.SHA256, crypto.SHA1} {
			for _, surface := range []string{"pe", "catalog", "legacy-time"} {
				role := x509.ExtKeyUsageCodeSigning
				if surface == "legacy-time" {
					role = x509.ExtKeyUsageTimeStamping
				}
				base := historicalPEChainFixture(t, hash, 1, false, role)
				var parentDER []byte
				for serial := 1; serial <= 4096; serial++ {
					clone := *base.parent
					fields := primaryFields(t, primaryFields(t, clone.RawTBSCertificate)[0].Bytes)
					parts := make([][]byte, len(fields))
					for i, f := range fields {
						parts[i] = f.FullBytes
					}
					value, err := asn1.Marshal(big.NewInt(int64(10000 + serial)))
					if err != nil {
						t.Fatal(err)
					}
					parts[1] = value
					clone.RawTBSCertificate = publicKeyTLV(0x30, bytes.Join(parts, nil), false)
					candidate := numericFixtureCertificate(t, &clone, numericStandardKey(base.parentKey), 128, key, hash, "padded")
					sig := primaryFields(t, primaryFields(t, candidate)[0].Bytes)[2].Bytes
					if len(sig) == 513 && sig[0] != 0 {
						parentDER = candidate
						break
					}
				}
				if parentDER == nil {
					t.Fatal("could not obtain a positive 513-byte numeric issuer proof")
				}
				code := historicalPEChainFixture(t, crypto.SHA256, 0, true, x509.ExtKeyUsageCodeSigning)
				expired := policyDatedCertificate(t, code.leaf, code.root, code.rootKey, crypto.SHA256, "primary-expired")
				primary, primaryDER := signedPEWithSigner(t, expired, code.leafKey, bytes.Replace(code.bag, code.leaf.Raw, expired.Raw, 1))
				for _, mode := range []string{"valid", "zero-prefix", "bad-proof", "short-numeric", "too-wide", "plus-N"} {
					label := fmt.Sprintf("issuer-width-%s-%s-n%d", surface, hash, bits)
					t.Run(label+"/"+mode, func(t *testing.T) {
						outer := primaryFields(t, primaryFields(t, parentDER)[0].Bytes)
						sig := bytes.Clone(outer[2].Bytes)
						switch mode {
						case "zero-prefix":
							sig = append([]byte{0}, sig...)
						case "bad-proof":
							sig[len(sig)-1] ^= 1
						case "short-numeric":
							sig = sig[1:]
						case "too-wide":
							sig = append(sig, 0)
						case "plus-N":
							sig = new(big.Int).Add(new(big.Int).SetBytes(sig), key.n).Bytes()
						}
						changed := publicKeyTLV(0x30, bytes.Join([][]byte{outer[0].FullBytes, outer[1].FullBytes, publicKeyTLV(3, sig, false)}, nil), false)
						bag := append(changed, base.leaf.Raw...)
						crb := replaceHistoricalCRBField(t, base.crb, 4, fmt.Sprintf("%X", key.n))
						inputs := map[string][]byte{"trust.crb": []byte(crb), "raw.ndb": []byte("Suspicious.Code:0:*:9090909090\n")}
						body, packet := signedPEWithSigner(t, base.leaf, base.leafKey, bag)
						if surface == "catalog" {
							body = simplePE(bytes.Repeat([]byte{0x90}, 0x200)).build()
							digest, err := peCatalogDigest(body, crypto.SHA256)
							if err != nil {
								t.Fatal(err)
							}
							packet = signedCatalogWithSignerForEngine(t, digest, bytes.Repeat([]byte{0xa5}, 20), base.leaf, base.leafKey, bag)
							inputs["verified.cat"] = packet
						}
						if surface == "legacy-time" {
							packet = appendLegacyTimestampWithSigner(t, primaryDER, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), base.leaf, base.leafKey, bag)
							inputs["trust.crb"] = []byte(code.crb + crb)
							body = replacePECertificate(t, primary, packet)
						}
						db, err := signatures.Compile(inputs)
						if err != nil {
							t.Fatal(err)
						}
						e := New(config.EDRConfig{})
						e.current.Store(&generation{db: db})
						clean, name, _ := e.Scan(context.Background(), body)
						t.Setenv("EDR_HISTORICAL_CHAIN_EXPORT", root)
						exportHistoricalScan(t, label, mode, body, inputs, clean, name, false)
					})
				}
			}
		}
	}
}
