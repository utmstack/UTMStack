package engine

import (
	"crypto"
	"encoding/asn1"
	"encoding/binary"
	"time"
)

// authenticode checks CRB blocklist entries across every embedded PE
// certificate before considering any trust exemption. A trusted result
// requires a verified PKCS#7 signer, a chain rooted in a CRB trust identity,
// and an exact PE image digest. If any certificate cannot be evaluated, the
// scan continues but cannot use an exemption to hide that uncertainty.
func (s *scanState) authenticode(data []byte) (name string, trusted bool, incomplete error) {
	if notPE(data) {
		return "", false, nil
	}
	img, err := parsePEImage(data)
	if err != nil {
		return "", false, nil // ordinary PE analysis reports parse errors
	}
	if img.certificate >= img.certificateEnd {
		// The inner WIN_CERTIFICATE length can be damaged while the outer
		// security directory and its SignedData remain readable. The
		// reference still checks CRB revocations in that case. Do not grant
		// trust: peAuthenticodeDigest requires a well-formed table.
		if der := firstDirectoryCertificateDER(data, img); len(der) != 0 {
			name, err := s.db.MatchBlockedAuthenticode(der)
			return name, false, err
		}
		return "", false, nil
	}
	var signatures [][]byte
	for pos := img.certificate; pos < img.certificateEnd; {
		if err := s.ctx.Err(); err != nil {
			return "", false, err
		}
		n := authenticodeSignature(data[pos:])
		if n == 0 || pos+n > img.certificateEnd {
			break
		}
		der := data[pos+8 : pos+n]
		name, err := s.db.MatchBlockedAuthenticode(der)
		if name != "" {
			return name, false, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
		signatures = append(signatures, der)
		pos += int((uint64(binary.LittleEndian.Uint32(data[pos:])) + 7) &^ 7)
	}
	if incomplete != nil {
		return "", false, incomplete
	}
	for _, der := range signatures {
		verified, err := s.db.VerifyTrustedAuthenticode(der, func(h crypto.Hash) ([]byte, error) {
			return peAuthenticodeDigest(data, h)
		}, time.Now().UTC())
		if verified {
			return "", true, nil
		}
		if err != nil && incomplete == nil {
			incomplete = err
		}
	}
	return "", false, incomplete
}

// firstDirectoryCertificateDER reads the first SignedData envelope from a
// complete, EOF-terminated PE security directory without trusting its inner
// WIN_CERTIFICATE length field. This is only for CRB blocklist matching;
// signed-file trust uses the stricter table and image-digest validator.
func firstDirectoryCertificateDER(data []byte, img *peImage) []byte {
	if img.certificate >= len(data) || len(data) < 64 {
		return nil
	}
	pe := uint64(binary.LittleEndian.Uint32(data[0x3c:]))
	optional := pe + 24
	if !fileSpan(data, pe+4, 20) || !fileSpan(data, optional, 2) {
		return nil
	}
	optionalSize := uint64(binary.LittleEndian.Uint16(data[pe+20:]))
	directoryStart := uint64(96)
	if binary.LittleEndian.Uint16(data[optional:]) == 0x20b {
		directoryStart = 112
	}
	security := optional + directoryStart + 32
	if security+8 > optional+optionalSize || !fileSpan(data, security, 8) {
		return nil
	}
	offset := uint64(binary.LittleEndian.Uint32(data[security:]))
	size := uint64(binary.LittleEndian.Uint32(data[security+4:]))
	if offset != uint64(img.certificate) || size < 10 || !fileSpan(data, offset, size) || offset+size != uint64(len(data)) {
		return nil
	}
	table := data[offset : offset+size]
	if binary.LittleEndian.Uint16(table[4:]) != 0x0200 || binary.LittleEndian.Uint16(table[6:]) != 2 {
		return nil
	}
	var envelope asn1.RawValue
	if _, err := asn1.Unmarshal(table[8:], &envelope); err != nil || envelope.Class != asn1.ClassUniversal || envelope.Tag != asn1.TagSequence {
		return nil
	}
	return envelope.FullBytes
}
