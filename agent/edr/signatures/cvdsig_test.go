package signatures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Public headers of three published containers (Sep 2026 snapshot).
var publishedContainerHeaders = []string{
	"ClamAV-VDB:16 Dec 2025 23-18 +0000:63:3287027:90:9c353a6b32555186a45a61c6441f38c0:gI9w3cezaElJbedEoiykyQeX1f6j9Xbs3a+cf9U/kH5QScGMe2mRz1K2JlYOrqfBuOUMbiKmvkVvAquOsMeDp/ejLVDx5WulC6klZVhLOfBHda29OjaLZwEGR9GPIHDr0mXriv7pivuMH60/loxpASIJHLOszyj95pxVK+EaTCh:tomjudge:1765927102",
	"ClamAV-VDB:23 Sep 2026 06-24 +0000:28132:355666:90:c84c59fc0ade2ae707ef6a2d924fa297:qnNy3Ia8IhP6vboqdLunHOTKHPtOYHRmscDvmMAMcMFOr0lMV9syeIqe92QtfuHg29Ryarr9AR72vALNTxr7WSCY0KM8dJ1PSYNbqooEYuS6xwKZ986x5q+Sv9BH/CAdRRcFrsok90Xg6N6BVHntzRrZfhxQju6QN5aTMRX+TSc:svc.clamav-publisher:1790144682",
	"ClamAV-VDB:11 Sep 2025 08-29 -0400:339:80:90:8bdb03f60b90cfc4c8d500543233de96:ow+EI2l1J6dXNMseMugc4ZK0cdbnUyDZY7x/Or+pIuzjIaYmWBt6MflLRusmT+y2dQamP0RjLrSAHjeU8DwyuWmTIkoTo1wrY1B3ZEUDC/Ta5vpv0cxYy4MPgMkeNovlXmEKCWX4m3orZS+3VJ/9J9grZ2rxcJubE9RAU5CXmEg:nrandolp:1757593759",
}

func TestPublishedContainerSignaturesVerify(t *testing.T) {
	for _, h := range publishedContainerHeaders {
		f := strings.Split(h, ":")
		if err := verifyContainerSignature(f[5], f[6]); err != nil {
			t.Fatalf("published header %s rejected: %v", f[2], err)
		}
	}
}

func TestContainerSignatureRejectsTampering(t *testing.T) {
	f := strings.Split(publishedContainerHeaders[1], ":")
	md5Hex, sig := f[5], f[6]
	flip := func(s string, i int, alphabet string) string {
		c := alphabet[(strings.IndexByte(alphabet, s[i])+1)%len(alphabet)]
		return s[:i] + string(c) + s[i+1:]
	}
	cases := map[string][2]string{
		"digest changed":       {flip(md5Hex, 0, "0123456789abcdef"), sig},
		"signature changed":    {md5Hex, flip(sig, 10, containerSignatureAlphabet)},
		"other container sig":  {md5Hex, strings.Split(publishedContainerHeaders[0], ":")[6]},
		"invalid character":    {md5Hex, sig[:5] + "*" + sig[6:]},
		"empty signature":      {md5Hex, ""},
		"short digest":         {md5Hex[:30], sig},
		"non-hex digest":       {"zz" + md5Hex[2:], sig},
		"oversized signature":  {md5Hex, strings.Repeat("a", 600)},
		"value above modulus":  {md5Hex, strings.Repeat("/", 200)},
		"uppercase hex digest": {strings.ToUpper(md5Hex), sig},
	}
	for name, c := range cases {
		if verifyContainerSignature(c[0], c[1]) == nil {
			t.Errorf("%s: tampered input verified", name)
		}
	}
}

func TestUnsignedContainerIsNotAuthenticated(t *testing.T) {
	d, err := Compile(map[string][]byte{"daily.cvd": cvdFixture(t, map[string][]byte{"daily.ndb": []byte("Test:0:*:41424344")})})
	if err != nil {
		t.Fatal(err)
	}
	if c := d.Report().Containers[0]; !c.IntegrityVerified || c.Authenticated {
		t.Fatalf("container info = %+v", c)
	}
}

// Set EDR_OFFICIAL_SIGNATURE_DIR to a directory holding a published
// bytecode.cvd to check a real publisher signature end to end.
func TestPublishedContainerIsAuthenticated(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	b, err := os.ReadFile(filepath.Join(dir, "bytecode.cvd"))
	if err != nil {
		t.Fatal(err)
	}
	budget := archiveBudget{limits: Limits{}.defaults(), total: int64(len(b)), files: 1}
	_, info, err := unpackContainer("bytecode.cvd", b, &budget)
	if err != nil || !info.IntegrityVerified || !info.Authenticated {
		t.Fatalf("published container: %+v, %v", info, err)
	}
	tampered := append([]byte(nil), b...)
	tampered[len(tampered)/2] ^= 1
	budget = archiveBudget{limits: Limits{}.defaults(), total: int64(len(b)), files: 1}
	if _, info, err = unpackContainer("bytecode.cvd", tampered, &budget); err == nil || info.Authenticated {
		t.Fatalf("tampered container accepted: %+v", info)
	}
}

// Set EDR_OFFICIAL_SIGNATURE_DIR to a directory holding the published
// databases to check that the whole feed compiles: one rejected rule fails
// the generation, so a runtime change can break every official update.
func TestPublishedFeedCompiles(t *testing.T) {
	dir := os.Getenv("EDR_OFFICIAL_SIGNATURE_DIR")
	if dir == "" {
		t.Skip("EDR_OFFICIAL_SIGNATURE_DIR not set")
	}
	if _, err := Load(dir); err != nil {
		t.Fatal(err)
	}
}
