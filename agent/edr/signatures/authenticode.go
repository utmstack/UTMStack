package signatures

import (
	"bytes"
	"encoding/asn1"
	"errors"
	"fmt"
)

var (
	oidPKCS7SignedData   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidSPCIndirectData   = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 1, 4}
	errAuthenticodeLimit = errors.New("Authenticode verification work limit exceeded")
)

// MatchBlockedAuthenticode checks embedded X.509 certificates in one PE
// WIN_CERTIFICATE PKCS#7 entry against CRB blocklist rules. ClamAV performs
// this check before verifying the signed PE digest; an embedded certificate
// matching a trust rule, in contrast, never makes the file trusted by itself.
// The DER envelope is limited to bound ASN.1 and X.509 parsing work.
func (d *Database) MatchBlockedAuthenticode(der []byte) (string, error) {
	hasBlocklist := false
	for i := range d.certRules {
		if d.certRules[i].blocked {
			hasBlocklist = true
			break
		}
	}
	if !hasBlocklist {
		return "", nil
	}
	if len(der) > 8<<20 {
		return "", incompleteCertificate()
	}
	certificates, recognized, err := authenticodeCertificates(der, nil)
	if !recognized {
		return "", nil
	}
	if err != nil {
		if errors.Is(err, errRejectedAuthenticodeCertificate) {
			return "", nil
		}
		return "", incompleteCertificate()
	}
	var incomplete error
	duplicates := make(authenticodeDuplicateIndex)
	for _, raw := range certificates {
		cert, err := parseAuthenticodeCertificate(raw)
		if err != nil {
			if errors.Is(err, errRejectedAuthenticodeCertificate) {
				continue
			}
			// A syntactically valid SignedData envelope can contain a cert
			// unsupported by Go's X.509 parser. It must not be called clean.
			incomplete = incompleteCertificate()
			continue
		}
		identity, err := certificateIdentityFromX509(cert)
		if err != nil {
			incomplete = incompleteCertificate()
			continue
		}
		if duplicates.discard(cert) {
			continue
		}
		if name, ok := d.blockedCertificate(identity); ok {
			return name, nil
		}
	}
	return "", incomplete
}

func incompleteCertificate() error {
	return fmt.Errorf("%w: %w", ErrRuleIncomplete, unknownBecause(UnknownCertificateTrust))
}

// authenticodeCertificates parses only the SignedData prefix needed to
// reach the certificate set. The reference checks the blocklist at this
// point, before parsing signerInfos or authenticating the signed digest.
// A different PKCS#7 content type is not an Authenticode PE signature.
func authenticodeCertificates(der []byte, signedDigestOID *asn1.ObjectIdentifier) ([][]byte, bool, error) {
	outer, rest, err := asn1Object(der)
	if err != nil || len(rest) != 0 || !derTag(outer, asn1.ClassUniversal, asn1.TagSequence) {
		return nil, false, nil
	}
	contentType, rest, err := asn1Object(outer.Bytes)
	if err != nil || !derOID(contentType, oidPKCS7SignedData) {
		return nil, false, nil
	}
	content, rest, err := asn1Object(rest)
	if err != nil || len(rest) != 0 || !derTag(content, asn1.ClassContextSpecific, 0) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	signed, rest, err := asn1Object(content.Bytes)
	if err != nil || len(rest) != 0 || !derTag(signed, asn1.ClassUniversal, asn1.TagSequence) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	version, rest, err := asn1Object(signed.Bytes)
	if err != nil || !derTag(version, asn1.ClassUniversal, asn1.TagInteger) || len(version.Bytes) != 1 || version.Bytes[0] != 1 {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	digests, rest, err := asn1Object(rest)
	if err != nil || !derTag(digests, asn1.ClassUniversal, asn1.TagSet) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	algorithm, digestRest, err := asn1Object(digests.Bytes)
	if err != nil || len(digestRest) != 0 || !derTag(algorithm, asn1.ClassUniversal, asn1.TagSequence) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	algorithmOID, parameters, err := asn1Object(algorithm.Bytes)
	if err != nil || !authenticodeDigestOID(algorithmOID) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	// ClamAV accepts an absent parameter or one empty ASN.1 NULL. An
	// arbitrary trailing value must not preserve a certificate exemption.
	if len(parameters) > 0 {
		parameter, remainder, err := asn1Object(parameters)
		if err != nil || len(remainder) != 0 || parameter.Class != asn1.ClassUniversal ||
			parameter.Tag != asn1.TagNull || len(parameter.Bytes) != 0 || parameter.IsCompound {
			return nil, true, errRejectedAuthenticodeCertificate
		}
	}
	if signedDigestOID != nil {
		if remainder, err := asn1.Unmarshal(algorithmOID.FullBytes, signedDigestOID); err != nil || len(remainder) != 0 {
			return nil, true, errRejectedAuthenticodeCertificate
		}
	}
	info, rest, err := asn1Object(rest)
	if err != nil || !derTag(info, asn1.ClassUniversal, asn1.TagSequence) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	innerType, innerRest, err := asn1Object(info.Bytes)
	if err != nil || !derOID(innerType, oidSPCIndirectData) {
		// A PKCS#7 SignedData with a different payload is not a PE
		// Authenticode signature, even if it embeds a matching cert.
		return nil, false, nil
	}
	inner, innerRest, err := asn1Object(innerRest)
	if err != nil || len(innerRest) != 0 || !derTag(inner, asn1.ClassContextSpecific, 0) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	indirect, innerRest, err := asn1Object(inner.Bytes)
	if err != nil || len(innerRest) != 0 || !derTag(indirect, asn1.ClassUniversal, asn1.TagSequence) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	certSet, _, err := asn1Object(rest)
	if err != nil || !derTag(certSet, asn1.ClassContextSpecific, 0) {
		return nil, true, errRejectedAuthenticodeCertificate
	}
	var certs [][]byte
	for len(certSet.Bytes) > 0 {
		cert, remain, err := asn1Object(certSet.Bytes)
		if err != nil || !derTag(cert, asn1.ClassUniversal, asn1.TagSequence) {
			return nil, true, errRejectedAuthenticodeCertificate
		}
		certs = append(certs, cert.FullBytes)
		if len(certs) > 256 {
			return nil, true, errInvalidAuthenticode
		}
		certSet.Bytes = remain
	}
	return certs, true, nil
}

var errInvalidAuthenticode = errors.New("invalid Authenticode SignedData")

func derTag(value asn1.RawValue, class, tag int) bool {
	if value.Class != class || value.Tag != tag {
		return false
	}
	if class == asn1.ClassUniversal && (tag == asn1.TagInteger || tag == asn1.TagOID) {
		return !value.IsCompound
	}
	return value.IsCompound
}

func derOID(value asn1.RawValue, want asn1.ObjectIdentifier) bool {
	if value.Class != asn1.ClassUniversal || value.Tag != asn1.TagOID || value.IsCompound {
		return false
	}
	// Compare against a small known OID rather than decoding attacker-sized
	// identifier arcs. The reference compares these encoded content bytes.
	encoded, err := asn1.Marshal(want)
	if err != nil {
		return false
	}
	var expected asn1.RawValue
	_, err = asn1.Unmarshal(encoded, &expected)
	return err == nil && bytes.Equal(value.Bytes, expected.Bytes)
}

func authenticodeDigestOID(value asn1.RawValue) bool {
	// Prefix traversal must agree with the reference's raw alternative-OID
	// reader. Actual digest selection and RSA proof are checked separately.
	return value.Class == 0 && value.Tag == asn1.TagOID && !value.IsCompound && len(value.Bytes) > 0
}
