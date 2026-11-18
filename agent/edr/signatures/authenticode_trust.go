package signatures

import (
	"bytes"
	"crypto"
	_ "crypto/md5"
	"crypto/sha1"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"time"
)

var oidSPCPEImageData = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 15}
var oidLegacyCountersignature = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 6}
var oidTimestampToken = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 1, 4}

// VerifyTrustedAuthenticode authenticates a PE signature against CRB trust
// identities and independently authenticated catalog certificate bags.
// A trust candidate must be in the verified signer's chain; merely carrying
// that certificate is insufficient.
// The callback computes the PE image digest only after the PKCS#7 signature
// and certificate chain have passed. Callers must still scan for blocklisted
// certificates first, and should only exempt an object after this returns
// true. Legacy PKCS#9 countersignatures are validated. ClamAV 1.5.4 recognizes
// but does not verify RFC3161 timestamp tokens, so they do not extend signer
// validity beyond the current time.
func (d *Database) VerifyTrustedAuthenticode(der []byte, imageDigest func(crypto.Hash) ([]byte, error), at time.Time) (bool, error) {
	hasTrust := false
	for _, rule := range d.certRules {
		if !rule.blocked {
			hasTrust = true
			break
		}
	}
	if !hasTrust {
		return false, nil
	}
	if len(der) > 8<<20 {
		return false, incompleteCertificate()
	}
	_, recognized, err := authenticodeCertificates(der, nil)
	if !recognized {
		return false, nil
	}
	if err != nil {
		if errors.Is(err, errRejectedAuthenticodeCertificate) {
			return false, nil
		}
		return false, incompleteCertificate()
	}
	packet, parseErr := parseAuthenticodeSignedData(der, oidSPCIndirectData, 256)
	if packet == nil {
		return false, incompleteCertificate()
	}
	store, storeErr := d.resolveAuthenticodeStore(packet.certificates, false)
	if storeErr != nil {
		if errors.Is(storeErr, errRejectedAuthenticodeCertificate) {
			return false, nil
		}
		return false, incompleteCertificate()
	}
	if parseErr != nil || imageDigest == nil {
		return false, incompleteCertificate()
	}
	algorithm, storedDigest, err := indirectPEImageDigest(packet.content)
	if err != nil {
		return false, incompleteCertificate()
	}
	signer, err := store.primarySigner(packet)
	if err != nil {
		return false, incompleteCertificate()
	}
	if signer == nil {
		return false, nil
	}
	verifyAt, err := d.legacyAuthenticodeSigningTime(packet, signer, store, at)
	if err != nil {
		if errors.Is(err, errRejectedAuthenticodeCertificate) {
			return false, nil
		}
		return false, incompleteCertificate()
	}
	if verifyAt.Before(signer.NotBefore) || verifyAt.After(signer.NotAfter) {
		return false, nil
	}
	computed, err := imageDigest(algorithm)
	if err != nil || len(computed) != len(storedDigest) || subtle.ConstantTimeCompare(computed, storedDigest) != 1 {
		return false, incompleteCertificate()
	}
	return true, nil
}

// Legacy PKCS#9 countersignatures authenticate a signing time over the
// primary SignerInfo's encrypted digest. RFC3161 timestamp tokens are a
// separate format; absent a validated legacy countersignature, validity is
// checked at scan time. An invalid present countersignature cannot grant
// trust, even when the primary certificate is otherwise current.
func (d *Database) legacyAuthenticodeSigningTime(packet *authenticodeSignedData, signer *x509.Certificate, store *authenticodeTrustStore, now time.Time) (time.Time, error) {
	var found bool
	var signedAt time.Time
	for _, attr := range packet.primary.unsigned {
		if !attr.typeOID.Equal(oidLegacyCountersignature) {
			continue
		}
		if found {
			return time.Time{}, errRejectedAuthenticodeCertificate
		}
		found = true
		value, rest, err := asn1Object(attr.value.Bytes)
		if err != nil || len(rest) != 0 || !derTag(value, asn1.ClassUniversal, asn1.TagSequence) {
			return time.Time{}, errInvalidAuthenticode
		}
		signedAt, err = d.verifyLegacyCountersigner(value.Bytes, packet.primary.signature, store)
		// The reference checks an authenticated legacy timestamp against the
		// primary signer's interval, including future times inside that window.
		// Scan time applies only when no legacy countersignature is present.
		if err != nil {
			return time.Time{}, err
		}
		if signedAt.Before(signer.NotBefore) || signedAt.After(signer.NotAfter) {
			return time.Time{}, errRejectedAuthenticodeCertificate
		}
	}
	if found {
		return signedAt, nil
	}
	return now, nil
}

func (d *Database) verifyLegacyCountersigner(info, signedMessage []byte, store *authenticodeTrustStore) (time.Time, error) {
	invalid := func() (time.Time, error) { return time.Time{}, errInvalidAuthenticode }
	version, rest, err := asn1Object(info)
	if err != nil || !derTag(version, asn1.ClassUniversal, asn1.TagInteger) || len(version.Bytes) != 1 || version.Bytes[0] > 1 {
		return invalid()
	}
	identity, rest, err := asn1Object(rest)
	if err != nil || !derTag(identity, asn1.ClassUniversal, asn1.TagSequence) {
		return invalid()
	}
	issuer, serialDER, err := asn1Object(identity.Bytes)
	if err != nil || !derTag(issuer, asn1.ClassUniversal, asn1.TagSequence) {
		return invalid()
	}
	serial, remaining, err := asn1Object(serialDER)
	if err != nil || len(remaining) != 0 || !derTag(serial, asn1.ClassUniversal, asn1.TagInteger) {
		return invalid()
	}
	serialHash := sha1.Sum(serial.Bytes)
	var candidates []*x509.Certificate
	for _, entry := range store.entries {
		if entry.cert != nil && entry.identity.timeSign && entry.issuer == sha1.Sum(issuer.Bytes) && entry.identity.serial == serialHash {
			candidates = append(candidates, entry.cert)
			if len(candidates) > 64 {
				return time.Time{}, errAuthenticodeLimit
			}
		}
	}
	if len(candidates) == 0 {
		if store.uncertain {
			return invalid()
		}
		return time.Time{}, errRejectedAuthenticodeCertificate
	}
	digestOID, rest, err := authAlgorithm(rest)
	if err != nil {
		return invalid()
	}
	hash := referenceAuthenticodeDigest(digestOID)
	if hash == 0 || !hash.Available() {
		return invalid()
	}
	attributeMap := rest
	attrs, rest, err := asn1Object(rest)
	if err != nil || !derTag(attrs, asn1.ClassContextSpecific, 0) {
		return invalid()
	}
	signedAt, err := legacyCounterAttributes(attrs.Bytes, attributeMap[len(attrs.FullBytes)-len(attrs.Bytes):], signedMessage, hash)
	if err != nil {
		return invalid()
	}
	signatureOID, rest, err := authAlgorithm(rest)
	if err != nil || !referenceAuthenticodeRSAAlgorithm(signatureOID, hash) {
		return invalid()
	}
	signature, rest, err := asn1Object(rest)
	if err != nil || len(rest) != 0 || signature.Class != asn1.ClassUniversal || signature.Tag != asn1.TagOctetString {
		return invalid()
	}
	digest := authenticodeSignatureDigest(hash, attrs.FullBytes)
	for _, signer := range candidates {
		identity, err := certificateIdentityFromX509(signer)
		if err != nil || !identity.timeSign {
			continue
		}
		if _, blocked := d.blockedCertificate(identity); blocked {
			continue
		}
		key, ok := authenticodeKey(signer.PublicKey)
		if ok {
			if err := store.chargeRSA(key); err != nil {
				return time.Time{}, err
			}
		}
		if !ok || !verifyAuthenticodeRSA(key, hash, digest, signature.Bytes) {
			continue
		}
		return signedAt, nil
	}
	return time.Time{}, errRejectedAuthenticodeCertificate
}

// Historical timestamp chains use the same bounded, CRB-authenticated RSA
// path verifier as historical code-signing and catalog chains.
func (d *Database) verifyHistoricalTimeChain(signer *x509.Certificate, certs []*x509.Certificate, at time.Time) bool {
	return len(d.historicalAuthenticodeChain(signer, certs, at, true)) > 0
}

func legacyCounterAttributes(attrs, mapped, signedMessage []byte, hash crypto.Hash) (time.Time, error) {
	if len(mapped) < len(attrs) || !bytes.Equal(mapped[:len(attrs)], attrs) {
		return time.Time{}, errInvalidAuthenticode
	}
	var contentType, messageDigest, signingTime bool
	var signedAt time.Time
	originalSize := len(attrs)
	for len(attrs) > 0 {
		attributeOffset := originalSize - len(attrs)
		attr, rest, err := asn1Object(attrs)
		if err != nil || !derTag(attr, asn1.ClassUniversal, asn1.TagSequence) {
			return time.Time{}, errInvalidAuthenticode
		}
		attrs = rest
		oid, rest, err := asn1Object(attr.Bytes)
		if err != nil {
			return time.Time{}, errInvalidAuthenticode
		}
		values, tail, err := asn1Object(rest)
		if err != nil || len(tail) != 0 || !derTag(values, asn1.ClassUniversal, asn1.TagSet) {
			return time.Time{}, errInvalidAuthenticode
		}
		value, tail, err := asn1Object(values.Bytes)
		if err != nil || len(tail) != 0 {
			return time.Time{}, errInvalidAuthenticode
		}
		switch {
		case derOID(oid, oidAuthenticodeContentType):
			if contentType || !(derOID(value, oidAuthenticodeData) || derOID(value, oidTimestampToken)) {
				return time.Time{}, errInvalidAuthenticode
			}
			contentType = true
		case derOID(oid, oidAuthenticodeMessageDigest):
			h := hash.New()
			h.Write(signedMessage)
			if messageDigest || value.Class != asn1.ClassUniversal || value.Tag != asn1.TagOctetString ||
				subtle.ConstantTimeCompare(value.Bytes, h.Sum(nil)) != 1 {
				return time.Time{}, errInvalidAuthenticode
			}
			messageDigest = true
		case derOID(oid, oidAuthenticodeSigningTime):
			if signingTime {
				return time.Time{}, errInvalidAuthenticode
			}
			valueOffset := attributeOffset + len(attr.FullBytes) - len(attr.Bytes) + len(oid.FullBytes) + len(values.FullBytes) - len(values.Bytes)
			var err error
			signedAt, _, err = referenceAuthenticodeTime(mapped[valueOffset:], len(value.FullBytes))
			if err != nil {
				return time.Time{}, errInvalidAuthenticode
			}
			signingTime = true
		}
	}
	if !contentType || !messageDigest || !signingTime {
		return time.Time{}, errInvalidAuthenticode
	}
	return signedAt, nil
}

func authenticodeHashOID(oid asn1.RawValue) crypto.Hash {
	for _, id := range []asn1.ObjectIdentifier{
		{1, 2, 840, 113549, 2, 5}, {1, 2, 840, 113549, 1, 1, 4},
		{1, 3, 14, 3, 2, 26}, {1, 3, 14, 3, 2, 29}, {1, 2, 840, 113549, 1, 1, 5},
		{2, 16, 840, 1, 101, 3, 4, 2, 1}, {1, 2, 840, 113549, 1, 1, 11},
		{2, 16, 840, 1, 101, 3, 4, 2, 2}, {1, 2, 840, 113549, 1, 1, 12},
		{2, 16, 840, 1, 101, 3, 4, 2, 3}, {1, 2, 840, 113549, 1, 1, 13},
	} {
		if derOID(oid, id) {
			return authenticodeHashID(id)
		}
	}
	return 0
}

func authenticodeHashID(id asn1.ObjectIdentifier) crypto.Hash {
	switch {
	case id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 2, 5}),
		id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 4}):
		return crypto.MD5
	case id.Equal(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 26}):
		return crypto.SHA1
	case id.Equal(asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}),
		id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 5}):
		return crypto.SHA1
	case id.Equal(asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}):
		return crypto.SHA256
	case id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 11}):
		return crypto.SHA256
	case id.Equal(asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 2}):
		return crypto.SHA384
	case id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 12}):
		return crypto.SHA384
	case id.Equal(asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 3}):
		return crypto.SHA512
	case id.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 13}):
		return crypto.SHA512
	}
	return 0
}

func boundedAuthenticodeDER(data []byte, depth int, nodes *int) bool {
	if depth > 48 {
		return false
	}
	for len(data) > 0 {
		if *nodes >= 131072 {
			return false
		}
		*nodes++
		value, rest, err := asn1Object(data)
		if err != nil || len(value.FullBytes) == 0 {
			return false
		}
		// A constructed BIT STRING is rejected when a certificate/key field
		// consumes it. Its arbitrary contents are not a DER child stream: do
		// not let this preliminary work bound prevent skipping that known
		// invalid certificate and authenticating a different framed entry.
		opaqueInvalidString := value.Class == asn1.ClassUniversal && value.Tag == asn1.TagBitString
		if value.IsCompound && !opaqueInvalidString && !boundedAuthenticodeDER(value.Bytes, depth+1, nodes) {
			return false
		}
		data = rest
	}
	return true
}

// indirectPEImageDigest reads SpcIndirectDataContent from the PKCS#7
// verified content. The first member identifies a PE image; the second is
// DigestInfo with the hash of the PE regions outside the certificate table.
func indirectPEImageDigest(content []byte) (crypto.Hash, []byte, error) {
	hash, digest, kind, err := authenticodeIndirectDigest(content)
	if err != nil || kind != 2 {
		return 0, nil, errInvalidAuthenticode
	}
	return hash, digest, nil
}
