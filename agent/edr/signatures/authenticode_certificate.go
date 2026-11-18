package signatures

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/asn1"
	"errors"
	"math/big"
)

// A known reference rejection may be skipped inside a framed certificate bag.
// Parser/resource limitations remain distinct and must preserve uncertainty.
var errRejectedAuthenticodeCertificate = errors.New("certificate rejected by Authenticode wire policy")

func certificateWireAlgorithm(data []byte) (asn1.RawValue, []byte, error) {
	seq, tail, err := asn1Object(data)
	if err != nil {
		return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
	}
	if !derTag(seq, 0, asn1.TagSequence) {
		return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
	}
	oid, params, err := asn1Object(seq.Bytes)
	if err != nil {
		return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
	}
	if oid.Class != 0 || oid.Tag != asn1.TagOID || oid.IsCompound || len(oid.Bytes) == 0 {
		return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
	}
	if len(params) != 0 {
		value, rest, err := asn1Object(params)
		if err != nil {
			return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
		}
		if len(rest) != 0 || value.Class != 0 || value.Tag != asn1.TagNull || value.IsCompound || len(value.Bytes) != 0 {
			return asn1.RawValue{}, nil, errRejectedAuthenticodeCertificate
		}
	}
	return oid, tail, nil
}

// This wire policy is local to CRB/catalog Authenticode. Certificate metadata
// still passes Go's X.509 parser; the original signed bytes remain authoritative
// for every certificate edge. No TLS, system root or feed CMS policy changes.
type authenticodeCertificateWire struct {
	tbs       asn1.RawValue
	fields    []asn1.RawValue
	algorithm int
	hash      crypto.Hash
	signature []byte // Complete BIT STRING content, including unused-bit count.
}

func referenceCertificateRSAHash(oid asn1.RawValue) crypto.Hash {
	if derOID(oid, oidAuthenticodeRSA) || derOID(oid, asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 2}) {
		return 0 // RSA without a digest and MD2 cannot authenticate a TBS hash.
	}
	for _, id := range []asn1.ObjectIdentifier{{1, 3, 14, 3, 2, 29}, {1, 2, 840, 113549, 1, 1, 4}, {1, 2, 840, 113549, 1, 1, 5}, {1, 2, 840, 113549, 1, 1, 11}, {1, 2, 840, 113549, 1, 1, 12}, {1, 2, 840, 113549, 1, 1, 13}} {
		if derOID(oid, id) {
			return authenticodeHashID(id)
		}
	}
	if oid.Class == 0 && oid.Tag == asn1.TagOID && !oid.IsCompound && len(oid.Bytes) > 0 {
		return crypto.SHA512 // Pinned reference alternative-OID fallthrough.
	}
	return 0
}

func readAuthenticodeCertificate(data []byte) (*authenticodeCertificateWire, error) {
	if len(data) > 8<<20 {
		return nil, errInvalidAuthenticode
	}
	outer, tail, err := authObject(data, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return nil, errInvalidAuthenticode
	}
	tbs, rest, err := authObject(outer.Bytes, 0, asn1.TagSequence)
	if err != nil {
		return nil, errRejectedAuthenticodeCertificate
	}
	outerOID, rest, err := certificateWireAlgorithm(rest)
	if err != nil {
		return nil, err
	}
	sig, tail, err := asn1Object(rest)
	if err != nil || len(tail) != 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	if sig.Class != 0 || sig.Tag != asn1.TagBitString || sig.IsCompound || len(sig.Bytes) == 0 || len(sig.Bytes) > 513 {
		return nil, errRejectedAuthenticodeCertificate
	}
	out := &authenticodeCertificateWire{tbs: tbs, signature: sig.Bytes, algorithm: 1}
	for data := tbs.Bytes; len(data) > 0; {
		if len(out.fields) >= 16 {
			return nil, errRejectedAuthenticodeCertificate
		}
		field, rest, err := asn1Object(data)
		if err != nil {
			return nil, errRejectedAuthenticodeCertificate
		}
		out.fields = append(out.fields, field)
		data = rest
	}
	if len(out.fields) < 6 {
		return nil, errRejectedAuthenticodeCertificate
	}
	if out.fields[0].Class == asn1.ClassContextSpecific {
		if !derTag(out.fields[0], asn1.ClassContextSpecific, 0) {
			return nil, errRejectedAuthenticodeCertificate
		}
		v, tail, err := authObject(out.fields[0].Bytes, 0, asn1.TagInteger)
		if err != nil || len(tail) != 0 || len(v.Bytes) != 1 || v.Bytes[0] != 2 {
			return nil, errRejectedAuthenticodeCertificate
		}
		out.algorithm = 2
	}
	serial := out.fields[out.algorithm-1]
	if !derTag(serial, 0, asn1.TagInteger) || len(serial.Bytes) == 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	innerOID, tail, err := certificateWireAlgorithm(out.fields[out.algorithm].FullBytes)
	if err != nil {
		return nil, err
	}
	out.hash = referenceCertificateRSAHash(innerOID)
	if len(tail) != 0 {
		return nil, errRejectedAuthenticodeCertificate
	}
	if out.hash == 0 || out.hash != referenceCertificateRSAHash(outerOID) {
		return nil, errRejectedAuthenticodeCertificate
	}
	if !out.hash.Available() {
		return nil, errInvalidAuthenticode
	}
	return out, nil
}

func certificateSequence(children ...[]byte) ([]byte, error) {
	return asn1.Marshal(asn1.RawValue{Tag: asn1.TagSequence, IsCompound: true, Bytes: bytes.Join(children, nil)})
}

func parseAuthenticodeCertificate(data []byte) (*x509.Certificate, error) {
	wire, err := readAuthenticodeCertificate(data)
	if err != nil {
		return nil, err
	}
	// Normalize only a metadata copy. Raw identities and signed bytes stay intact.
	// The reference
	// compares selected hash classes, including absent NULL and local aliases,
	// while x509.ParseCertificate requires identical encoded identifiers.
	suffix := map[crypto.Hash]int{crypto.MD5: 4, crypto.SHA1: 5, crypto.SHA256: 11, crypto.SHA384: 12, crypto.SHA512: 13}[wire.hash]
	oid, err := asn1.Marshal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, suffix})
	if err != nil {
		return nil, err
	}
	algorithm, err := certificateSequence(oid, []byte{5, 0})
	if err != nil {
		return nil, err
	}
	fields := make([][]byte, len(wire.fields))
	for i, field := range wire.fields {
		fields[i] = field.FullBytes
	}
	if wire.algorithm == 2 {
		fields[0] = []byte{0xa0, 3, 2, 1, 2} // Metadata-only canonical v3 wrapper.
	}
	if err := normalizeCertificateIdentity(wire, fields); err != nil {
		return nil, err
	}
	if err := normalizeCertificateValidity(wire, fields); err != nil {
		return nil, err
	}
	fields[wire.algorithm] = algorithm
	keyIndex := wire.algorithm + 4
	if keyIndex >= len(wire.fields) {
		return nil, errInvalidAuthenticode
	}
	publicKey, err := decodeAuthenticodePublicKey(wire.fields[keyIndex].FullBytes)
	if err != nil {
		return nil, err
	}
	metadataKey := publicKey
	if publicKey.N.Sign() == 0 || publicKey.E.Sign() == 0 || !publicKey.E.IsInt64() {
		// General X.509 supplies only names/time metadata. Its native-int key
		// constraints must not replace the already checked AV wire integers.
		metadataKey = &authenticodeRSAKey{N: big.NewInt(3), E: big.NewInt(3)}
	}
	fields[keyIndex], err = encodeAuthenticodeMetadataKey(metadataKey)
	if err != nil {
		return nil, err
	}
	policy, err := certificateExtensionPolicy(wire.fields, wire.algorithm)
	if err != nil {
		return nil, err
	}
	// Parse only identity/key/time metadata through Go's general X.509 reader.
	// Authenticode extension framing and purposes were checked independently;
	// preserve their original signed bytes instead of interpreting opaque
	// extension payloads through the unrelated PKIX extension parsers.
	fields = fields[:keyIndex+1]
	tbs, err := certificateSequence(fields...)
	if err != nil {
		return nil, err
	}
	// The reference treats BIT STRING contents as an unsigned integer. Keep
	// that integer intact when giving Go a conventional metadata-only encoding.
	content := wire.signature
	if len(content) == 0 || content[0] != 0 {
		content = append([]byte{0}, content...)
	}
	signature, err := asn1.Marshal(asn1.RawValue{Tag: asn1.TagBitString, Bytes: content})
	if err != nil {
		return nil, err
	}
	normalized, err := certificateSequence(tbs, algorithm, signature)
	if err != nil {
		return nil, err
	}
	cert, err := x509.ParseCertificate(normalized)
	if err != nil {
		return nil, err
	}
	if !cert.NotBefore.Before(cert.NotAfter) {
		return nil, errRejectedAuthenticodeCertificate
	}
	cert.Raw, cert.RawTBSCertificate = data, wire.tbs.FullBytes
	cert.RawSubjectPublicKeyInfo = wire.fields[keyIndex].FullBytes
	cert.PublicKey = publicKey.public()
	cert.RawIssuer = wire.fields[wire.algorithm+1].FullBytes
	cert.RawSubject = wire.fields[wire.algorithm+3].FullBytes
	cert.Issuer, cert.Subject = certificateDisplayName(cert.RawIssuer), certificateDisplayName(cert.RawSubject)
	policy.applyMetadata(cert)
	return cert, nil
}

func authenticodeCertificateProof(cert *x509.Certificate, key any) (valid, compatibility bool) {
	if cert == nil {
		return false, false
	}
	wire, err := readAuthenticodeCertificate(cert.Raw)
	if err != nil || !bytes.Equal(cert.RawTBSCertificate, wire.tbs.FullBytes) {
		return false, false
	}
	expectedSignature := wire.signature
	if len(expectedSignature) > 0 && expectedSignature[0] == 0 {
		expectedSignature = expectedSignature[1:]
	}
	if !bytes.Equal(cert.Signature, expectedSignature) {
		return false, false
	}
	normalized, ok := normalizeAuthenticodeRSA(key, wire.signature)
	if !ok {
		return false, false
	}
	h := wire.hash.New()
	h.Write(wire.tbs.FullBytes)
	valid = verifyAuthenticodeRSAValue(key, wire.hash, h.Sum(nil), wire.signature)
	return valid, valid && (!bytes.Equal(normalized, cert.Signature) || authenticodeCertificateHash(cert.SignatureAlgorithm) != wire.hash)
}
