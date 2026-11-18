package signatures

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
)

// Certificate and signer identities hash raw contents, not a decoded name or
// serial number. These metadata copies must never replace the original wire
// fields used for identity matching or certificate signature verification.
func normalizeCertificateIdentity(wire *authenticodeCertificateWire, fields [][]byte) error {
	serial := wire.fields[wire.algorithm-1]
	issuer, subject := wire.algorithm+1, wire.algorithm+3
	if subject >= len(wire.fields) || !derTag(serial, 0, asn1.TagInteger) || len(serial.Bytes) == 0 {
		return errRejectedAuthenticodeCertificate
	}
	for _, index := range []int{issuer, subject} {
		name := wire.fields[index]
		if !derTag(name, 0, asn1.TagSequence) || len(name.Bytes) == 0 {
			return errRejectedAuthenticodeCertificate
		}
		// Names are opaque byte identities in this local policy. A standard
		// parser's RDN/string choices must not change that identity. Supply an
		// empty metadata name; restore raw names and optional display names below.
		fields[index] = []byte{0x30, 0}
	}
	var err error
	fields[wire.algorithm-1], err = asn1.Marshal(new(big.Int).SetBytes(serial.Bytes))
	return err
}

func certificateDisplayName(raw []byte) pkix.Name {
	var name pkix.Name
	// Display metadata is optional and never used for trust. Avoid materializing
	// an unbounded RDN list from an opaque identity; ordinary names stay legible.
	if len(raw) <= 64<<10 {
		var sequence pkix.RDNSequence
		if tail, err := asn1.Unmarshal(raw, &sequence); err == nil && len(tail) == 0 {
			name.FillFromRDNSequence(&sequence)
		}
	}
	return name
}
