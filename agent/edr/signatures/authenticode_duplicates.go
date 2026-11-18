package signatures

import (
	"crypto"
	"crypto/sha1"
	"crypto/x509"
)

type authenticodeDuplicateKey struct {
	subject, issuer, serial [sha1.Size]byte
	hash                    crypto.Hash
	exponent                string
	modulus                 string
}

// Certificate insertion is ordered. The reference discards a later entry
// covered by an earlier identity's validity and purpose envelope, before
// validating its issuer signature. Trying every signed-byte variant instead
// can incorrectly rescue a poisoned first entry and grant a trust exemption.
type authenticodeDuplicateIndex map[authenticodeDuplicateKey][]certificateIdentity

func (index authenticodeDuplicateIndex) discard(cert *x509.Certificate) bool {
	identity, err := certificateIdentityFromX509(cert)
	if err != nil {
		return false
	}
	return index.discardIdentity(cert, identity)
}

func (index authenticodeDuplicateIndex) discardIdentity(cert *x509.Certificate, identity certificateIdentity) bool {
	issuer, tail, err := asn1Object(cert.RawIssuer)
	if err != nil || len(tail) != 0 {
		return false
	}
	key, ok := authenticodeKey(cert.PublicKey)
	if !ok {
		return false
	}
	bucket := authenticodeDuplicateKey{subject: identity.subject, issuer: sha1.Sum(issuer.Bytes), serial: identity.serial, hash: authenticodeCertificateHash(cert.SignatureAlgorithm), exponent: string(key.E.Bytes()), modulus: string(identity.modulus)}
	for _, prior := range index[bucket] {
		if identity.notBefore >= prior.notBefore && identity.notAfter <= prior.notAfter && (!identity.certSign || prior.certSign) && (!identity.codeSign || prior.codeSign) && (!identity.timeSign || prior.timeSign) {
			return true
		}
	}
	index[bucket] = append(index[bucket], identity)
	return false
}
