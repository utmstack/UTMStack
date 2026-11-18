package signatures

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"testing"
)

// Public signed fixture only; no private key is retained. Keep the same bytes
// across fuzz workers so the pinned-key authenticity assertion is meaningful.
const primaryParserSeedBase64 = "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"

func primaryParserSeed(tb testing.TB) ([]byte, *x509.Certificate) {
	tb.Helper()
	seed, err := base64.StdEncoding.DecodeString(primaryParserSeedBase64)
	if err != nil {
		tb.Fatal(err)
	}
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil || len(packet.certificates) == 0 {
		tb.Fatal("invalid pinned seed", err)
	}
	return seed, packet.certificates[0]
}

func TestOwnedPrimaryParserBoundsAndExactProof(t *testing.T) {
	seed, _ := primaryParserSeed(t)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil || packet == nil {
		t.Fatal("valid seed refused", err)
	}
	d := &Database{}
	if len(d.verifiedAuthenticodePrimarySigners(packet)) != 1 {
		t.Fatal("valid seed did not authenticate")
	}
	if _, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 0); err == nil {
		t.Fatal("certificate count bound ignored")
	}
	if _, err := parseAuthenticodeSignedData(seed, oidCatalogCTL, 256); err == nil {
		t.Fatal("wrong content type accepted")
	}
	attrs := bytes.Clone(packet.primary.attributes)
	packet.primary.attributes = bytes.Clone(attrs)
	packet.primary.attributes[len(attrs)-1] ^= 1
	if len(d.verifiedAuthenticodePrimarySigners(packet)) != 0 {
		t.Fatal("changed signed attributes authenticated")
	}
	if len(d.verifiedAuthenticodePrimarySigners(nil)) != 0 {
		t.Fatal("nil packet authenticated")
	}
}

func FuzzOwnedAuthenticodeSignedData(f *testing.F) {
	seed, cert := primaryParserSeed(f)
	packet, err := parseAuthenticodeSignedData(seed, oidSPCIndirectData, 256)
	if err != nil {
		f.Fatal(err)
	}
	content, attrs := bytes.Clone(packet.content), bytes.Clone(packet.primary.attributes)
	f.Add(seed)
	f.Add([]byte{0x30, 0})
	f.Add([]byte{0x30, 0x80, 0, 0})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 64<<10 {
			return
		}
		parsed, err := parseAuthenticodeSignedData(input, oidSPCIndirectData, 256)
		if err != nil || parsed == nil {
			return
		}
		for _, signer := range (&Database{}).verifiedAuthenticodePrimarySigners(parsed) {
			key, ok := signer.PublicKey.(*rsa.PublicKey)
			pinned := cert.PublicKey.(*rsa.PublicKey)
			if ok && key.E == pinned.E && key.N.Cmp(pinned.N) == 0 && (!bytes.Equal(parsed.content, content) || !bytes.Equal(parsed.primary.attributes, attrs)) {
				t.Fatal("mutated content or signed attributes authenticated under the pinned fuzz key")
			}
		}
	})
}
