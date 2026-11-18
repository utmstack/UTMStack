package signatures

import (
	"encoding/asn1"
	"errors"
)

// The packet parser supplies the grammar position for opaque identity fields.
// All other constructed data retains the same shared depth/node work limits.
// This avoids interpreting an authenticated raw name as a recursive ASN.1 tree.
func boundedCertificateIdentityDER(data []byte, depth int, nodes *int) bool {
	wire, err := readAuthenticodeCertificate(data)
	if err != nil {
		if errors.Is(err, errRejectedAuthenticodeCertificate) {
			// This framed entry is discarded without consuming its children.
			// Its bytes are bounded by the certificate reader and bag count.
			if depth > 48 || *nodes >= 131072 {
				return false
			}
			*nodes++
			return true
		}
		return boundedAuthenticodeDER(data, depth, nodes)
	}
	if depth+1 > 48 || *nodes > 131072-6 {
		return false
	}
	*nodes += 6 // Certificate/TBS envelopes, outer algorithm/OID/NULL and signature.
	for index, field := range wire.fields {
		isName := index == wire.algorithm+1 || index == wire.algorithm+3
		isUniqueID := index > wire.algorithm+4 && field.Class == asn1.ClassContextSpecific && field.IsCompound && (field.Tag == 1 || field.Tag == 2)
		if isName && derTag(field, 0, asn1.TagSequence) || isUniqueID {
			if *nodes >= 131072 {
				return false
			}
			*nodes++
			continue
		}
		if !boundedAuthenticodeDER(field.FullBytes, depth+2, nodes) {
			return false
		}
	}
	return true
}

func boundedSignerIdentityDER(data []byte, depth int, nodes *int) bool {
	if depth > 48 || *nodes > 131072-4 {
		return false
	}
	signer, tail, err := authObject(data, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return false
	}
	_, rest, err := authObject(signer.Bytes, 0, asn1.TagInteger)
	if err != nil {
		return false
	}
	identity, rest, err := authObject(rest, 0, asn1.TagSequence)
	if err != nil {
		return false
	}
	_, serial, err := authObject(identity.Bytes, 0, asn1.TagSequence)
	*nodes += 4 // SignerInfo/version/issuerAndSerialNumber/opaque issuer name.
	if err != nil || !boundedAuthenticodeDER(serial, depth+2, nodes) {
		return false
	}
	for len(rest) > 0 {
		field, next, err := asn1Object(rest)
		if err != nil {
			return false
		}
		if derTag(field, asn1.ClassContextSpecific, 1) {
			if !boundedAuthenticodeUnsignedDER(field.Bytes, depth+1, nodes) {
				return false
			}
		} else if !boundedAuthenticodeDER(field.FullBytes, depth+1, nodes) {
			return false
		}
		rest = next
	}
	return true
}

func boundedAuthenticodeUnsignedDER(data []byte, depth int, nodes *int) bool {
	if depth > 48 || *nodes >= 131072 {
		return false
	}
	*nodes++
	for len(data) > 0 {
		attr, rest, err := authObject(data, 0, asn1.TagSequence)
		if err != nil {
			return false
		}
		oid, values, err := authObject(attr.Bytes, 0, asn1.TagOID)
		if err != nil {
			return false
		}
		if derOID(oid, oidLegacyCountersignature) {
			set, tail, err := authObject(values, 0, asn1.TagSet)
			if err != nil || len(tail) != 0 || *nodes > 131072-3 {
				return false
			}
			*nodes += 3 // Attribute, identifier and SET envelope.
			if !boundedSignerIdentityDER(set.Bytes, depth+3, nodes) {
				return false
			}
		} else if !boundedAuthenticodeDER(attr.FullBytes, depth+1, nodes) {
			return false
		}
		data = rest
	}
	return true
}
