package signatures

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
)

// The upstream publisher signs each container header's MD5 digest with a fixed
// RSA-style key. The digits below are that published public key; they are
// data, not code. A verified header digest plus a matching payload digest
// authenticates the container's origin, unlike the MD5 check alone.
var (
	containerSigningModulus, _ = new(big.Int).SetString("118640995551645342603070001658453189751527774412027743746599405743243142607464144767361060640655844749760788890022283424922762488917565551002467771109669598189410434699034532232228621591089508178591428456220796841621637175567590476666928698770143328137383952820383197532047771780196576957695822641224262693037", 10)
	containerSigningExponent   = big.NewInt(100001027)
)

// The signature alphabet orders lowercase before uppercase, unlike base64.
const containerSignatureAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"

var errContainerSignature = errors.New("signature container publisher signature is invalid")

// verifyContainerSignature checks the header signature over the header's
// hexadecimal MD5 field. The caller must separately compare that field with
// the payload's actual MD5.
func verifyContainerSignature(md5Hex, signature string) error {
	// The reference compares lowercase hexadecimal text, so require it too.
	if len(md5Hex) != 2*md5.Size || strings.ToLower(md5Hex) != md5Hex || len(signature) == 0 || len(signature) > 512 {
		return errContainerSignature
	}
	want, err := hex.DecodeString(md5Hex)
	if err != nil {
		return errContainerSignature
	}
	// Digits are little-endian base-64: digit i contributes value << 6i.
	c := new(big.Int)
	for i := len(signature) - 1; i >= 0; i-- {
		v := strings.IndexByte(containerSignatureAlphabet, signature[i])
		if v < 0 {
			return errContainerSignature
		}
		c.Lsh(c, 6).Or(c, big.NewInt(int64(v)))
	}
	if c.Cmp(containerSigningModulus) >= 0 {
		return errContainerSignature
	}
	plain := new(big.Int).Exp(c, containerSigningExponent, containerSigningModulus).Bytes()
	if len(plain) > md5.Size {
		return errContainerSignature
	}
	// The reference writes the minimal big-endian value at the start of a
	// 16-byte buffer and zero-fills the rest; compare the same layout.
	got := make([]byte, md5.Size)
	copy(got, plain)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return errContainerSignature
	}
	return nil
}
