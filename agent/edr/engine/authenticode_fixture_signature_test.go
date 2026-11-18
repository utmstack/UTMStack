package engine

import (
	"crypto"
	"crypto/rsa"
	"io"
	"math/big"
)

// Positive reference fixtures must stay inside its numeric signature window.
// A random ordinary PKCS#1 signature can have two leading zero bytes; the
// reference rejects that integer even though its RSA proof is valid. Encode
// the equivalent S+N only for that rare case. The fixed numeric-boundary cohort
// independently verifies both representations against the actual scanner.
// Deliberate short-signature controls retain their original bytes.
func referenceSizedFixtureSignature(key *rsa.PublicKey, signature []byte) []byte {
	value := new(big.Int).SetBytes(signature)
	if len(value.Bytes()) < key.Size()-1 {
		value.Add(value, key.N)
		return value.Bytes()
	}
	return signature
}

func signReferenceFixture(random io.Reader, key *rsa.PrivateKey, hash crypto.Hash, digest []byte) ([]byte, error) {
	signature, err := rsa.SignPKCS1v15(random, key, hash, digest)
	if err != nil {
		return nil, err
	}
	return referenceSizedFixtureSignature(&key.PublicKey, signature), nil
}
