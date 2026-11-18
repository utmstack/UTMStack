package engine

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/smallstep/pkcs7"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// The official Microsoft binaries are supplied out of tree so that the
// default test suite does not depend on network access or ship executables.
// EDR_SIGNED_PE_FIXTURE_DIR points at an extracted Sigcheck.zip archive.
func TestAuthenticodeOfficialSigcheckFixtures(t *testing.T) {
	dir := os.Getenv("EDR_SIGNED_PE_FIXTURE_DIR")
	if dir == "" {
		t.Skip("set EDR_SIGNED_PE_FIXTURE_DIR to an extracted official Sigcheck.zip")
	}
	fixtures := map[string]string{
		"sigcheck.exe":    "7cf481ae647639d03782c614e268894caf94811da300deef9c6a74fccc5f7804",
		"sigcheck64.exe":  "f3d671b4bab4714746d0034104363d4f6deef591e979e8e37170803e8960487d",
		"sigcheck64a.exe": "c081f7ec670bd90c2209738801050683f1e5d908653636674f5b9f1280f7bf09",
	}
	for name, wantSHA := range fixtures {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != wantSHA {
				t.Fatalf("fixture SHA-256 %s, want %s", got, wantSHA)
			}
			img, err := parsePEImage(data)
			if err != nil || img.certificate >= img.certificateEnd {
				t.Fatalf("signed PE parse: %v", err)
			}
			for pos := img.certificate; pos < img.certificateEnd; {
				n := authenticodeSignature(data[pos:])
				if n == 0 {
					t.Fatalf("bad certificate entry at %d", pos)
				}
				der := data[pos+8 : pos+n]
				p7, err := pkcs7.Parse(der)
				if err != nil {
					t.Fatalf("PKCS#7 parse: %v", err)
				}
				if len(p7.Signers) != 1 {
					t.Fatalf("PKCS#7 signers=%d", len(p7.Signers))
				}
				if err := p7.Verify(); err != nil {
					t.Fatalf("PKCS#7 signature: %v", err)
				}
				var hasRFC3161 bool
				for _, attr := range p7.Signers[0].UnauthenticatedAttributes {
					hasRFC3161 = hasRFC3161 || attr.Type.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 3, 3, 1})
				}
				if !hasRFC3161 {
					t.Fatal("expected RFC3161 timestamp attribute in official fixture")
				}
				signer := p7.GetOnlySigner()
				var subject asn1.RawValue
				if _, err := asn1.Unmarshal(signer.RawSubject, &subject); err != nil {
					t.Fatal(err)
				}
				key, ok := signer.PublicKey.(*rsa.PublicKey)
				if !ok {
					t.Fatal("fixture signer is not RSA")
				}
				crb := fmt.Sprintf("Microsoft.Sigcheck;1;%x;;%X;ignored;1;1;1;;comment\n", sha1.Sum(subject.Bytes), key.N.Bytes())
				db, err := signatures.Compile(map[string][]byte{
					"fixture.crb": []byte(crb),
					"fixture.ndb": []byte("Raw.MZ:0:0:4d5a900003000000\n"),
				})
				if err != nil {
					t.Fatal(err)
				}
				verify := func(b []byte) (bool, error) {
					return db.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
						return peAuthenticodeDigest(b, h)
					}, time.Now().UTC())
				}
				if verified, err := verify(data); !verified || err != nil {
					t.Fatalf("valid PE was not trusted: %v", err)
				}
				engine := New(config.EDRConfig{})
				engine.current.Store(&generation{db: db})
				if clean, sig, err := engine.Scan(context.Background(), data); !clean || sig != "" || err != nil {
					t.Fatalf("valid PE scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				altered := bytes.Clone(data)
				altered[0x1000] ^= 1
				if verified, err := verify(altered); verified || err == nil {
					t.Fatalf("altered PE remained trusted: %v", err)
				}
				if clean, sig, err := engine.Scan(context.Background(), altered); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered PE scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				certificateChanged := bytes.Clone(data)
				certificateChanged[pos+32] ^= 1
				modifiedDER := certificateChanged[pos+8 : pos+n]
				verified, _ := db.VerifyTrustedAuthenticode(modifiedDER, func(h crypto.Hash) ([]byte, error) {
					return peAuthenticodeDigest(certificateChanged, h)
				}, time.Now().UTC())
				if verified {
					t.Fatal("altered certificate remained trusted")
				}
				if clean, sig, err := engine.Scan(context.Background(), certificateChanged); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered certificate scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				signerVersionChanged := bytes.Clone(data)
				versionAt, err := signerInfoVersionOffset(der)
				if err != nil {
					t.Fatal(err)
				}
				signerVersionChanged[pos+8+versionAt] = 0
				if clean, sig, err := engine.Scan(context.Background(), signerVersionChanged); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered SignerInfo version scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				// SHA-256 and SHA-384 have equally long OIDs. Changing only
				// SignedData's declared algorithm leaves the signed payload and
				// SignerInfo intact, exposing any missing agreement check.
				digestChanged := bytes.Clone(data)
				sha256OID, err := asn1.Marshal(pkcs7.OIDDigestAlgorithmSHA256)
				if err != nil {
					t.Fatal(err)
				}
				digestAt := bytes.Index(der, sha256OID)
				if digestAt < 0 || digestAt > 128 {
					t.Fatalf("SignedData digest OID offset=%d", digestAt)
				}
				digestChanged[pos+8+digestAt+len(sha256OID)-1] = 2 // SHA-384
				aliasChanged := bytes.Clone(data)
				sha256RSAOID, err := asn1.Marshal(pkcs7.OIDEncryptionAlgorithmRSASHA256)
				if err != nil || len(sha256RSAOID) != len(sha256OID) {
					t.Fatalf("unexpected SHA-256 OID aliases: %x %x %v", sha256OID, sha256RSAOID, err)
				}
				copy(aliasChanged[pos+8+digestAt:], sha256RSAOID)
				parameterChanged := bytes.Clone(data)
				parameterAt := pos + 8 + digestAt + len(sha256OID)
				if !bytes.Equal(data[parameterAt:parameterAt+2], []byte{5, 0}) {
					t.Fatalf("unexpected SignedData digest parameter: %x", data[parameterAt:parameterAt+2])
				}
				parameterChanged[parameterAt] = 4 // OCTET STRING instead of NULL
				signerParameterChanged := bytes.Clone(data)
				signerDigestAt := bytes.Index(der[versionAt+1:], sha256OID)
				if signerDigestAt < 0 || signerDigestAt > 256 {
					t.Fatalf("SignerInfo digest OID offset=%d", signerDigestAt)
				}
				signerParameterAt := pos + 8 + versionAt + 1 + signerDigestAt + len(sha256OID)
				if !bytes.Equal(data[signerParameterAt:signerParameterAt+2], []byte{5, 0}) {
					t.Fatalf("unexpected SignerInfo digest parameter: %x", data[signerParameterAt:signerParameterAt+2])
				}
				signerParameterChanged[signerParameterAt] = 4
				// Authenticode omits the optional-header checksum from its image
				// digest. It must remain trusted when that field alone changes.
				checksumChanged := bytes.Clone(data)
				peOffset := int(binary.LittleEndian.Uint32(data[0x3c:]))
				checksumChanged[peOffset+24+64] ^= 1
				if verified, err := verify(checksumChanged); !verified || err != nil {
					t.Fatalf("checksum-only mutation was not trusted: %v", err)
				}
				if clean, sig, err := engine.Scan(context.Background(), checksumChanged); !clean || sig != "" || err != nil {
					t.Fatalf("checksum-only scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				oracleDir := os.Getenv("EDR_SIGCHECK_ORACLE_DIR")
				if oracleDir == "" && os.Getenv("EDR_CLAMSCAN_BIN") != "" {
					oracleDir = t.TempDir()
				}
				if oracleDir != "" {
					out := filepath.Join(oracleDir, name)
					if err := os.MkdirAll(out, 0700); err != nil {
						t.Fatal(err)
					}
					for filename, content := range map[string][]byte{
						"fixture.crb":                 []byte(crb),
						"fixture.ndb":                 []byte("Raw.MZ:0:0:4d5a900003000000\n"),
						"original.exe":                data,
						"tampered.exe":                altered,
						"certificate.exe":             certificateChanged,
						"signer-version.exe":          signerVersionChanged,
						"digest-declaration.exe":      digestChanged,
						"digest-alias.exe":            aliasChanged,
						"digest-parameter.exe":        parameterChanged,
						"signer-digest-parameter.exe": signerParameterChanged,
						"checksum.exe":                checksumChanged,
					} {
						if err := os.WriteFile(filepath.Join(out, filename), content, 0600); err != nil {
							t.Fatal(err)
						}
					}
					if bin := os.Getenv("EDR_CLAMSCAN_BIN"); bin != "" {
						for filename, want := range map[string]bool{
							"original.exe": true, "tampered.exe": false,
							"certificate.exe": false, "checksum.exe": true,
							"signer-version.exe":          false,
							"digest-declaration.exe":      false,
							"digest-alias.exe":            true,
							"digest-parameter.exe":        false,
							"signer-digest-parameter.exe": false,
						} {
							cmd, cancel := context.WithTimeout(context.Background(), 30*time.Second)
							result, err := exec.CommandContext(cmd, bin, "--database="+out, "--no-summary", filepath.Join(out, filename)).CombinedOutput()
							cancel()
							exit := 0
							if err != nil {
								var status *exec.ExitError
								if !errors.As(err, &status) {
									t.Fatalf("clamscan %s: %v: %s", filename, err, result)
								}
								exit = status.ExitCode()
							}
							text := string(result)
							if want && (exit != 0 || !strings.Contains(text, ": OK")) ||
								!want && (exit != 1 || !strings.Contains(text, ": Raw.MZ.UNOFFICIAL FOUND")) {
								t.Fatalf("clamscan %s: exit=%d output=%q", filename, exit, text)
							}
						}
					}
				}
				if clean, sig, err := engine.Scan(context.Background(), digestChanged); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered SignedData digest declaration scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				if clean, sig, err := engine.Scan(context.Background(), parameterChanged); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered SignedData digest parameter scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				if clean, sig, err := engine.Scan(context.Background(), signerParameterChanged); clean || sig != "Raw.MZ" || err != nil {
					t.Fatalf("altered SignerInfo digest parameter scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				if clean, sig, err := engine.Scan(context.Background(), aliasChanged); !clean || sig != "" || err != nil {
					t.Fatalf("equivalent SignedData digest OID scan: clean=%v sig=%q err=%v", clean, sig, err)
				}
				pos += int((uint64(binary.LittleEndian.Uint32(data[pos:])) + 7) &^ 7)
			}
		})
	}
}

// Locate the first SignerInfo's version in the fixture PKCS#7 without
// depending on the lengths of the certificates or timestamp token.
func signerInfoVersionOffset(der []byte) (int, error) {
	read := func(b []byte) (asn1.RawValue, []byte, error) {
		var value asn1.RawValue
		rest, err := asn1.Unmarshal(b, &value)
		return value, rest, err
	}
	outer, _, err := read(der)
	if err != nil {
		return 0, err
	}
	_, rest, err := read(outer.Bytes) // signedData OID
	if err != nil {
		return 0, err
	}
	content, _, err := read(rest)
	if err != nil {
		return 0, err
	}
	signed, _, err := read(content.Bytes)
	if err != nil {
		return 0, err
	}
	rest = signed.Bytes
	for range 4 { // version, digest algorithms, content info, certificates
		_, rest, err = read(rest)
		if err != nil {
			return 0, err
		}
	}
	signerSet, _, err := read(rest)
	if err != nil {
		return 0, err
	}
	signer, _, err := read(signerSet.Bytes)
	if err != nil {
		return 0, err
	}
	version, _, err := read(signer.Bytes)
	if err != nil || !bytes.Equal(version.FullBytes, []byte{2, 1, 1}) {
		return 0, fmt.Errorf("unexpected SignerInfo version: %x: %v", version.FullBytes, err)
	}
	start := bytes.Index(der, signer.FullBytes)
	if start < 0 {
		return 0, fmt.Errorf("SignerInfo missing from DER")
	}
	return start + len(signer.FullBytes) - len(signer.Bytes) + 2, nil
}
