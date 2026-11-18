package signatures

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/subtle"
	"math/big"
)

// These unsigned public integers belong only to the local AV certificate
// format. They never replace TLS, PKIX or authenticated feed verification.
type authenticodeRSAKey struct{ N, E *big.Int }

func authenticodeKey(public any) (*authenticodeRSAKey, bool) {
	var key *authenticodeRSAKey
	switch value := public.(type) {
	case *rsa.PublicKey:
		if value == nil || value.E < 0 {
			return nil, false
		}
		key = &authenticodeRSAKey{N: value.N, E: big.NewInt(int64(value.E))}
	case *authenticodeRSAKey:
		key = value
	default:
		return nil, false
	}
	if key == nil || key.N == nil || key.E == nil || key.N.Sign() < 0 || key.E.Sign() < 0 || key.N.BitLen() > 4112 || key.E.BitLen() > 4104 {
		return nil, false
	}
	return key, true
}

func (key *authenticodeRSAKey) public() any {
	if key.E.IsInt64() && uint64(key.E.Int64()) <= uint64(^uint(0)>>1) {
		return &rsa.PublicKey{N: key.N, E: int(key.E.Int64())}
	}
	return key
}

func (key *authenticodeRSAKey) size() int { return (key.N.BitLen() + 7) / 8 }

// Reference RSA integers may include leading zeroes or S+N aliases. Compare
// numeric widths before doing any modular arithmetic. The outer certificate
// and SignerInfo readers enforce their own distinct encoded-field bounds.
func normalizeAuthenticodeRSA(public any, signature []byte) ([]byte, bool) {
	key, ok := authenticodeKey(public)
	if !ok || key.N.Sign() == 0 || key.size() < 11 || len(signature) == 0 || len(signature) > 513 {
		return nil, false
	}
	value := new(big.Int).SetBytes(signature)
	if delta := (value.BitLen()+7)/8 - key.size(); delta < -1 || delta > 1 {
		return nil, false
	}
	value.Mod(value, key.N)
	return value.FillBytes(make([]byte, key.size())), true
}

func verifyAuthenticodeRSAValue(public any, hash crypto.Hash, digest, signature []byte) bool {
	// Exact short-form DigestInfo, including the NULL parameter. No raw-hash
	// or general ASN.1 re-encoding fallback is part of this wire policy.
	prefix := map[crypto.Hash]string{
		crypto.MD5:    "\x30\x20\x30\x0c\x06\x08\x2a\x86\x48\x86\xf7\x0d\x02\x05\x05\x00\x04\x10",
		crypto.SHA1:   "\x30\x21\x30\x09\x06\x05\x2b\x0e\x03\x02\x1a\x05\x00\x04\x14",
		crypto.SHA256: "\x30\x31\x30\x0d\x06\x09\x60\x86\x48\x01\x65\x03\x04\x02\x01\x05\x00\x04\x20",
		crypto.SHA384: "\x30\x41\x30\x0d\x06\x09\x60\x86\x48\x01\x65\x03\x04\x02\x02\x05\x00\x04\x30",
		crypto.SHA512: "\x30\x51\x30\x0d\x06\x09\x60\x86\x48\x01\x65\x03\x04\x02\x03\x05\x00\x04\x40",
	}[hash]
	if prefix == "" || len(digest) != hash.Size() {
		return false
	}
	key, ok := authenticodeKey(public)
	if !ok {
		return false
	}
	normalized, ok := normalizeAuthenticodeRSA(key, signature)
	if !ok {
		return false
	}
	recovered := new(big.Int).Exp(new(big.Int).SetBytes(normalized), key.E, key.N).Bytes()
	// BN recovery removes leading zeroes; type-1 padding therefore has one
	// fewer numeric byte than the modulus and begins with byte 1.
	if len(recovered) != key.size()-1 || len(recovered) < 10 || recovered[0] != 1 {
		return false
	}
	at := 1
	for at < len(recovered) && recovered[at] == 0xff {
		at++
	}
	if at < 9 || at >= len(recovered) || recovered[at] != 0 {
		return false
	}
	payload := recovered[at+1:]
	return len(payload) == len(prefix)+len(digest) && bytes.Equal(payload[:len(prefix)], []byte(prefix)) && subtle.ConstantTimeCompare(payload[len(prefix):], digest) == 1
}

// Work is charged by public modulus/exponent size as well as by proof count.
// One maximum-width proof is permitted; repeated large exponents cannot turn
// the 4,096-edge count into unbounded CPU work.
func (s *authenticodeTrustStore) chargeRSA(key *authenticodeRSAKey) error {
	width, bits := uint64(key.size()), uint64(key.E.BitLen())
	cost := (width*width*bits + 128*128*17 - 1) / (128 * 128 * 17)
	if cost == 0 {
		cost = 1
	}
	if s.rsaWork > 65536 || cost > 65536-s.rsaWork {
		return errAuthenticodeLimit
	}
	s.rsaWork += cost
	return nil
}
