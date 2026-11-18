package signatures

import (
	"bufio"
	"compress/gzip"
	"crypto"
	"crypto/rsa"
	"encoding/hex"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestOwnedAVRSAReference(t *testing.T) {
	file, err := os.Open("testdata/authenticode_rsa_vectors.txt.gz")
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
	reader.Buffer(make([]byte, 4096), 8192)
	hashes := map[int]crypto.Hash{1: crypto.MD5, 2: crypto.SHA1, 3: crypto.SHA256, 4: crypto.SHA384, 5: crypto.SHA512}
	count, accepted := 0, 0
	for reader.Scan() {
		p := strings.Fields(reader.Text())
		if len(p) != 6 {
			t.Fatal("invalid numeric oracle row")
		}
		values := make([]*big.Int, 3)
		for i := range values {
			var ok bool
			values[i], ok = new(big.Int).SetString(p[i], 16)
			if !ok {
				t.Fatal("invalid oracle integer")
			}
		}
		h, err := strconv.Atoi(p[3])
		if err != nil || hashes[h] == 0 {
			t.Fatal("invalid oracle hash")
		}
		digest, err := hex.DecodeString(p[4])
		if err != nil {
			t.Fatal(err)
		}
		key := &authenticodeRSAKey{N: values[0], E: values[1]}
		valid := verifyAuthenticodeRSAValue(key, hashes[h], digest, values[2].Bytes())
		if valid != (p[5] == "1") {
			t.Fatalf("numeric reference row %d: valid=%v want=%s Nbits=%d Ebits=%d", count, valid, p[5], key.N.BitLen(), key.E.BitLen())
		}
		if valid {
			accepted++
		}
		count++
	}
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 423974 || accepted != 6872 {
		t.Fatal("incomplete numeric oracle", count, accepted)
	}
}

func FuzzOwnedAVRSARecovery(f *testing.F) {
	seeds, key, _ := certificateProofSeeds(f)
	wire, err := readAuthenticodeCertificate(seeds[0])
	if err != nil {
		f.Fatal(err)
	}
	h := wire.hash.New()
	h.Write(wire.tbs.FullBytes)
	original := h.Sum(nil)
	f.Add(wire.signature, original, uint8(2))
	f.Add([]byte{0}, []byte{}, uint8(0))
	f.Add(make([]byte, 513), make([]byte, 64), uint8(4))
	hashes := []crypto.Hash{crypto.MD5, crypto.SHA1, crypto.SHA256, crypto.SHA384, crypto.SHA512}
	f.Fuzz(func(t *testing.T, signature, digest []byte, which uint8) {
		if len(signature) > 513 || len(digest) > 64 {
			return
		}
		hash := hashes[int(which)%len(hashes)]
		got := verifyAuthenticodeRSAValue(key, hash, digest, signature)
		normalized, ok := normalizeAuthenticodeRSA(key, signature)
		want := ok && rsa.VerifyPKCS1v15(key, hash, digest, normalized) == nil
		if got != want {
			t.Fatal("local AV recovery disagrees with independent standard-key RSA proof", got, want)
		}
	})
}
