package signatures

import (
	"crypto"
	"encoding/asn1"
)

// authenticodeIndirectDigest parses the common PE/catalog indirect digest.
// Descriptive type data is ignored by the reference, but the object namespace,
// complete algorithm parameters and exact primitive digest remain mandatory.
func authenticodeIndirectDigest(content []byte) (crypto.Hash, []byte, int64, error) {
	invalid := func() (crypto.Hash, []byte, int64, error) { return 0, nil, 0, errInvalidAuthenticode }
	typeInfo, rest, err := authObject(content, 0, asn1.TagSequence)
	if err != nil {
		return invalid()
	}
	typeOID, _, err := authObject(typeInfo.Bytes, 0, asn1.TagOID)
	if err != nil {
		return invalid()
	}
	var kind int64
	switch {
	case derOID(typeOID, oidSPCPEImageData):
		kind = 2
	case derOID(typeOID, oidCatalogCAB):
		kind = 1
	default:
		return invalid()
	}
	info, tail, err := authObject(rest, 0, asn1.TagSequence)
	if err != nil || len(tail) != 0 {
		return invalid()
	}
	oid, body, err := authAlgorithm(info.Bytes)
	if err != nil {
		return invalid()
	}
	hash := referenceAuthenticodeDigest(oid)
	if hash == 0 || !hash.Available() {
		return invalid()
	}
	digest, tail, err := asn1Object(body)
	if err != nil || len(tail) != 0 || digest.Class != 0 || digest.Tag != asn1.TagOctetString || digest.IsCompound || len(digest.Bytes) != hash.Size() {
		return invalid()
	}
	return hash, digest.Bytes, kind, nil
}
