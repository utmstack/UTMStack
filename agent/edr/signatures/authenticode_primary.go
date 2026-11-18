package signatures

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/asn1"
	"errors"
)

var oidAuthenticodeData = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
var oidAuthenticodeSigningTime = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
var oidAuthenticodeRSA = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
var oidAuthenticodeContentType = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
var oidAuthenticodeMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
var oidAuthenticodeNested = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 2, 4, 1}
var oidAuthenticodeRFC3161 = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 311, 3, 3, 1}

// The pinned reference's alternative-OID reader retains its final SHA-512
// alternative when no OID matches. Keep that local compatibility behavior;
// the content digest and RSA proof must still both validate as SHA-512.
func referenceAuthenticodeDigest(id asn1.RawValue) crypto.Hash {
	if id.Class != 0 || id.Tag != asn1.TagOID || id.IsCompound || len(id.Bytes) == 0 {
		return 0
	}
	// OIW SHA1/RSA belongs to the RSA algorithm alternatives, not to the
	// content digest alternatives. In a digest position it falls through
	// to SHA-512, just like any other unrecognized nonempty OID bytes.
	if derOID(id, asn1.ObjectIdentifier{1, 3, 14, 3, 2, 29}) {
		return crypto.SHA512
	}
	if hash := authenticodeHashOID(id); hash != 0 {
		return hash
	}
	return crypto.SHA512
}

func referenceAuthenticodeRSAAlgorithm(id asn1.RawValue, hash crypto.Hash) bool {
	if derOID(id, oidAuthenticodeRSA) {
		return true
	}
	// MD2 is a recognized RSA alternative, but is not a supported content
	// digest. It must not be mistaken for the unknown-OID SHA-512 fallback.
	if derOID(id, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 2}) {
		return false
	}
	for _, candidate := range []asn1.ObjectIdentifier{{1, 3, 14, 3, 2, 29}, {1, 2, 840, 113549, 1, 1, 4}, {1, 2, 840, 113549, 1, 1, 5}, {1, 2, 840, 113549, 1, 1, 11}, {1, 2, 840, 113549, 1, 1, 12}, {1, 2, 840, 113549, 1, 1, 13}} {
		if derOID(id, candidate) {
			return authenticodeHashID(candidate) == hash
		}
	}
	return id.Class == 0 && id.Tag == asn1.TagOID && !id.IsCompound && len(id.Bytes) > 0 && hash == crypto.SHA512
}

type authenticodeAttribute struct {
	typeOID asn1.ObjectIdentifier
	value   asn1.RawValue
}

type authenticodePrimary struct {
	issuer, serial        [sha1.Size]byte
	hash                  crypto.Hash
	attributes, signature []byte
	unsigned              []authenticodeAttribute
}

type authenticodeSignedData struct {
	content      []byte
	certificates []*x509.Certificate
	primary      authenticodePrimary
}

func authObject(data []byte, class, tag int) (asn1.RawValue, []byte, error) {
	value, rest, err := asn1Object(data)
	if err != nil || !derTag(value, class, tag) {
		return value, nil, errInvalidAuthenticode
	}
	return value, rest, nil
}

func authAlgorithm(data []byte) (asn1.RawValue, []byte, error) {
	value, rest, err := authObject(data, 0, asn1.TagSequence)
	if err != nil {
		return asn1.RawValue{}, nil, err
	}
	oid, params, err := authObject(value.Bytes, 0, asn1.TagOID)
	if err != nil || oid.IsCompound || len(oid.Bytes) == 0 {
		return asn1.RawValue{}, nil, errInvalidAuthenticode
	}
	if len(params) > 0 {
		parameter, tail, err := asn1Object(params)
		if err != nil || len(tail) != 0 || parameter.Class != 0 || parameter.Tag != asn1.TagNull || parameter.IsCompound || len(parameter.Bytes) != 0 {
			return asn1.RawValue{}, nil, errInvalidAuthenticode
		}
	}
	// The reference compares the raw OID bytes, including unknown encodings.
	// Preserve them so fallback selection never silently becomes an alias.
	return oid, rest, nil
}

// Prefix failures return nil. Once the complete certificate prefix has been
// parsed, return it even when SignerInfo is invalid: catalogs may contribute
// independently authenticated certificates without contributing member hashes.
func parseAuthenticodeSignedData(der []byte, contentType asn1.ObjectIdentifier, maxCertificates int) (*authenticodeSignedData, error) {
	nodes := 6 // ContentInfo/OID/content/SignedData, certificate bag and signer SET.
	outer, tail, err := authObject(der, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	oid, rest, err := authObject(outer.Bytes, 0, asn1.TagOID)
	if err != nil || !derOID(oid, oidPKCS7SignedData) {
		return nil, errInvalidAuthenticode
	}
	wrapped, tail, err := authObject(rest, asn1.ClassContextSpecific, 0)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	signed, tail, err := authObject(wrapped.Bytes, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	version, rest, err := authObject(signed.Bytes, 0, asn1.TagInteger)
	if err != nil || len(version.Bytes) != 1 || version.Bytes[0] != 1 {
		return nil, errInvalidAuthenticode
	}
	digests, rest, err := authObject(rest, 0, asn1.TagSet)
	if err != nil {
		return nil, err
	}
	digestOID, tail, err := authAlgorithm(digests.Bytes)
	hash := referenceAuthenticodeDigest(digestOID)
	if err != nil || len(tail) != 0 || hash == 0 || !hash.Available() {
		return nil, errInvalidAuthenticode
	}
	info, rest, err := authObject(rest, 0, asn1.TagSequence)
	if err != nil {
		return nil, err
	}
	for _, field := range []asn1.RawValue{version, digests, info} {
		if !boundedAuthenticodeDER(field.FullBytes, 3, &nodes) {
			return nil, errInvalidAuthenticode
		}
	}
	innerOID, inner, err := authObject(info.Bytes, 0, asn1.TagOID)
	if err != nil || !derOID(innerOID, contentType) {
		return nil, errInvalidAuthenticode
	}
	content, tail, err := authObject(inner, asn1.ClassContextSpecific, 0)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	sequence, tail, err := authObject(content.Bytes, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	certificates, rest, err := authObject(rest, asn1.ClassContextSpecific, 0)
	if err != nil {
		return nil, err
	}
	packet := &authenticodeSignedData{content: sequence.Bytes}
	seenCertificates := 0
	duplicates := make(authenticodeDuplicateIndex)
	for data := certificates.Bytes; len(data) > 0; {
		certificate, next, err := authObject(data, 0, asn1.TagSequence)
		if err != nil || seenCertificates >= maxCertificates {
			return nil, errInvalidAuthenticode
		}
		seenCertificates++ // Rejected entries still consume the certificate budget.
		if !boundedCertificateIdentityDER(certificate.FullBytes, 4, &nodes) {
			return nil, errInvalidAuthenticode
		}
		parsed, err := parseAuthenticodeCertificate(certificate.FullBytes)
		if err != nil {
			if errors.Is(err, errRejectedAuthenticodeCertificate) {
				data = next
				continue
			}
			return nil, errInvalidAuthenticode
		}
		if !duplicates.discard(parsed) {
			packet.certificates = append(packet.certificates, parsed)
		}
		data = next
	}
	// CRLs are unused by the reference; CRB revocations are checked separately.
	if len(rest) > 0 && rest[0] == 0xa1 {
		var crls asn1.RawValue
		crls, rest, err = authObject(rest, asn1.ClassContextSpecific, 1)
		if err != nil {
			return packet, err
		}
		if !boundedAuthenticodeDER(crls.FullBytes, 3, &nodes) {
			return packet, errInvalidAuthenticode
		}
	}
	signers, tail, err := authObject(rest, 0, asn1.TagSet)
	if err != nil || len(tail) != 0 {
		return packet, errInvalidAuthenticode
	}
	signer, tail, err := authObject(signers.Bytes, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return packet, errInvalidAuthenticode
	}
	if !boundedSignerIdentityDER(signer.FullBytes, 4, &nodes) {
		return packet, errInvalidAuthenticode
	}
	primary, err := parseAuthenticodePrimary(signer.Bytes, hash, contentType, packet.content)
	packet.primary = primary
	return packet, err
}

func parseAuthenticodePrimary(data []byte, hash crypto.Hash, contentType asn1.ObjectIdentifier, content []byte) (authenticodePrimary, error) {
	var out authenticodePrimary
	version, rest, err := authObject(data, 0, asn1.TagInteger)
	if err != nil || len(version.Bytes) != 1 || version.Bytes[0] != 1 {
		return out, errInvalidAuthenticode
	}
	identity, rest, err := authObject(rest, 0, asn1.TagSequence)
	if err != nil {
		return out, err
	}
	issuer, serialData, err := authObject(identity.Bytes, 0, asn1.TagSequence)
	if err != nil {
		return out, err
	}
	serial, tail, err := authObject(serialData, 0, asn1.TagInteger)
	if err != nil || len(tail) != 0 || len(serial.Bytes) == 0 {
		return out, errInvalidAuthenticode
	}
	out.issuer, out.serial = sha1.Sum(issuer.Bytes), sha1.Sum(serial.Bytes)
	id, rest, err := authAlgorithm(rest)
	if err != nil || referenceAuthenticodeDigest(id) != hash {
		return out, errInvalidAuthenticode
	}
	attrs, rest, err := authObject(rest, asn1.ClassContextSpecific, 0)
	if err != nil {
		return out, err
	}
	if err = verifyAuthenticodeAttributes(attrs.Bytes, hash, contentType, content); err != nil {
		return out, err
	}
	algorithm, rest, err := authAlgorithm(rest)
	if err != nil || !derOID(algorithm, oidAuthenticodeRSA) {
		return out, errInvalidAuthenticode
	}
	signature, rest, err := asn1Object(rest)
	if err != nil || signature.Class != 0 || signature.Tag != asn1.TagOctetString || signature.IsCompound || len(signature.Bytes) < 128 || len(signature.Bytes) > 513 {
		return out, errInvalidAuthenticode
	}
	out.hash, out.attributes, out.signature = hash, attrs.FullBytes, signature.Bytes
	if len(rest) > 0 {
		unsigned, tail, err := authObject(rest, asn1.ClassContextSpecific, 1)
		if err != nil || len(tail) != 0 {
			return out, errInvalidAuthenticode
		}
		out.unsigned, err = parseAuthenticodeUnsigned(unsigned.Bytes)
		if err != nil {
			return out, err
		}
	}
	return out, nil
}

func verifyAuthenticodeAttributes(data []byte, hash crypto.Hash, contentType asn1.ObjectIdentifier, content []byte) error {
	var haveType, haveDigest bool
	count := 0
	for len(data) > 0 {
		count++
		if count > 4096 {
			return errInvalidAuthenticode
		}
		attr, rest, err := authObject(data, 0, asn1.TagSequence)
		if err != nil {
			return err
		}
		data = rest
		oid, values, err := authObject(attr.Bytes, 0, asn1.TagOID)
		if err != nil {
			return err
		}
		isType, isDigest := derOID(oid, oidAuthenticodeContentType), derOID(oid, oidAuthenticodeMessageDigest)
		if !isType && !isDigest {
			continue
		}
		set, tail, err := authObject(values, 0, asn1.TagSet)
		if err != nil || len(tail) != 0 {
			return errInvalidAuthenticode
		}
		value, tail, err := asn1Object(set.Bytes)
		if err != nil || len(tail) != 0 {
			return errInvalidAuthenticode
		}
		if isType {
			if haveType || !derOID(value, contentType) {
				return errInvalidAuthenticode
			}
			haveType = true
		} else {
			if haveDigest || value.Class != 0 || value.Tag != asn1.TagOctetString || value.IsCompound || len(value.Bytes) != hash.Size() {
				return errInvalidAuthenticode
			}
			h := hash.New()
			h.Write(content)
			if subtle.ConstantTimeCompare(h.Sum(nil), value.Bytes) != 1 {
				return errInvalidAuthenticode
			}
			haveDigest = true
		}
	}
	if !haveType || !haveDigest {
		return errInvalidAuthenticode
	}
	return nil
}

func parseAuthenticodeUnsigned(data []byte) ([]authenticodeAttribute, error) {
	var attrs []authenticodeAttribute
	seen := make(map[string]bool)
	for len(data) > 0 {
		if len(attrs) >= 4096 {
			return nil, errInvalidAuthenticode
		}
		attribute, rest, err := authObject(data, 0, asn1.TagSequence)
		if err != nil {
			return nil, err
		}
		data = rest
		oid, values, err := authObject(attribute.Bytes, 0, asn1.TagOID)
		if err != nil {
			return nil, err
		}
		var id asn1.ObjectIdentifier
		for _, recognized := range []asn1.ObjectIdentifier{oidLegacyCountersignature, oidAuthenticodeNested, oidAuthenticodeRFC3161} {
			if derOID(oid, recognized) {
				id = recognized
				break
			}
		}
		entry := authenticodeAttribute{typeOID: id}
		if len(id) > 0 {
			key := id.String()
			if seen[key] {
				return nil, errInvalidAuthenticode
			}
			seen[key] = true
			value, tail, err := authObject(values, 0, asn1.TagSet)
			if err != nil || len(tail) != 0 {
				return nil, errInvalidAuthenticode
			}
			entry.value = value
		}
		attrs = append(attrs, entry)
	}
	return attrs, nil
}

func authenticodeSignatureDigest(hash crypto.Hash, attributes []byte) []byte {
	h := hash.New()
	// SignedAttributes is hashed with the SET tag, retaining all original
	// length and content octets. Re-marshalling or sorting changes the proof.
	h.Write([]byte{0x31})
	h.Write(attributes[1:])
	return h.Sum(nil)
}

func verifyAuthenticodeRSA(key any, hash crypto.Hash, digest, signature []byte) bool {
	if len(signature) < 128 || len(signature) > 513 {
		return false
	}
	return verifyAuthenticodeRSAValue(key, hash, digest, signature)
}

// These low-level helpers prove SignerInfo RSA only. Runtime trust selection
// uses the resolved store's authenticated purposes and insertion order.
func (d *Database) verifiedAuthenticodePrimarySigners(packet *authenticodeSignedData) []*x509.Certificate {
	verified, _ := d.authenticodePrimarySignersChecked(packet)
	return verified
}

func (d *Database) authenticodePrimarySignersChecked(packet *authenticodeSignedData) ([]*x509.Certificate, error) {
	if packet == nil || packet.primary.hash == 0 || len(packet.primary.attributes) < 2 {
		return nil, nil
	}
	certs := append([]*x509.Certificate(nil), packet.certificates...)
	for _, trusted := range d.catalogCerts {
		if trusted.codeSign {
			certs = append(certs, trusted.cert)
		}
	}
	seen := make(map[[sha256.Size]byte]bool)
	var verified []*x509.Certificate
	checked := 0
	uncertain := false
	digest := authenticodeSignatureDigest(packet.primary.hash, packet.primary.attributes)
	for _, cert := range certs {
		fingerprint := sha256.Sum256(cert.Raw)
		if seen[fingerprint] {
			continue
		}
		seen[fingerprint] = true
		identity, err := certificateIdentityFromX509(cert)
		if err != nil {
			uncertain = true
			continue
		}
		if identity.serial != packet.primary.serial {
			continue
		}
		issuer, tail, err := asn1Object(cert.RawIssuer)
		if err != nil || len(tail) != 0 || sha1.Sum(issuer.Bytes) != packet.primary.issuer {
			continue
		}
		checked++
		if checked > 64 {
			return nil, errAuthenticodeLimit
		}
		if _, blocked := d.blockedCertificate(identity); blocked {
			continue
		}
		key, ok := authenticodeKey(cert.PublicKey)
		if ok && verifyAuthenticodeRSA(key, packet.primary.hash, digest, packet.primary.signature) {
			verified = append(verified, cert)
		}
	}
	if len(verified) == 0 && uncertain {
		return nil, errInvalidAuthenticode
	}
	return verified, nil
}
