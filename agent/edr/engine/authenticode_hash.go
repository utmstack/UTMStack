package engine

import (
	"bytes"
	"crypto"
	_ "crypto/md5"
	"encoding/binary"
)

// peAuthenticodeDigest hashes the PE bytes that the reference passes to its
// Authenticode verifier: everything before the certificate table except the
// four-byte checksum and eight-byte security directory entry. Trust is only
// considered when the declared certificate table ends at EOF. Signed files
// with other layouts continue through ordinary scanning without exemption.
// Unlike ClamAV 1.5.4's MS13-098 compatibility path, the WIN_CERTIFICATE
// entry length must describe the actual DER and padding. Malformed lengths
// can otherwise hide unsigned data from signature-based detection.
func peAuthenticodeDigest(data []byte, algorithm crypto.Hash) ([]byte, error) {
	if !algorithm.Available() {
		return nil, ErrUnsupported
	}
	img, err := parsePEImage(data)
	if err != nil {
		return nil, err
	}
	if img.certificate >= img.certificateEnd {
		return nil, ErrMalformed
	}
	peOffset := uint64(binary.LittleEndian.Uint32(data[0x3c:]))
	coff := peOffset + 4
	optional := coff + 20
	optionalSize := uint64(binary.LittleEndian.Uint16(data[coff+16:]))
	directoryStart := uint64(96)
	if binary.LittleEndian.Uint16(data[optional:]) == 0x20b {
		directoryStart = 112
	}
	checksum := optional + 64
	security := optional + directoryStart + 32
	if security+8 > optional+optionalSize || !fileSpan(data, security, 8) ||
		!fileSpan(data, checksum, 4) || checksum+4 > security {
		return nil, ErrMalformed
	}
	certOffset := uint64(binary.LittleEndian.Uint32(data[security:]))
	certSize := uint64(binary.LittleEndian.Uint32(data[security+4:]))
	if certOffset != uint64(img.certificate) || certSize == 0 ||
		!fileSpan(data, certOffset, certSize) || certOffset+certSize != uint64(len(data)) ||
		certOffset < security+8 || certOffset&7 != 0 {
		return nil, ErrMalformed
	}
	// The table is excluded from the image digest. Never let a valid
	// signature exempt extra, unsigned bytes hidden as certificate padding
	// or as a malformed second entry. A normal entry has at most seven zero
	// alignment bytes after its DER message.
	table := data[certOffset:]
	for pos := 0; pos < len(table); {
		n := authenticodeSignature(table[pos:])
		if n == 0 || binary.LittleEndian.Uint16(table[pos+4:]) != 0x0200 {
			return nil, ErrMalformed
		}
		length := int(binary.LittleEndian.Uint32(table[pos:]))
		if length&7 != 0 || length-n > 7 || !bytes.Equal(table[pos+n:pos+length], make([]byte, length-n)) {
			return nil, ErrMalformed
		}
		pos += length
	}
	h := algorithm.New()
	h.Write(data[:checksum])
	h.Write(data[checksum+4 : security])
	h.Write(data[security+8 : certOffset])
	return h.Sum(nil), nil
}

// peCatalogDigest uses the same Authenticode image regions for a catalog
// member. An unsigned PE has no certificate table, so its final region runs
// through EOF. A PE with an embedded table uses the stricter signed-file
// validator above before its digest may receive a catalog exemption.
func peCatalogDigest(data []byte, algorithm crypto.Hash) ([]byte, error) {
	if !algorithm.Available() {
		return nil, ErrUnsupported
	}
	if _, err := parsePEImage(data); err != nil {
		return nil, err
	}
	peOffset := uint64(binary.LittleEndian.Uint32(data[0x3c:]))
	coff := peOffset + 4
	optional := coff + 20
	optionalSize := uint64(binary.LittleEndian.Uint16(data[coff+16:]))
	directoryStart := uint64(96)
	if binary.LittleEndian.Uint16(data[optional:]) == 0x20b {
		directoryStart = 112
	}
	checksum := optional + 64
	security := optional + directoryStart + 32
	if security+8 > optional+optionalSize || !fileSpan(data, security, 8) ||
		!fileSpan(data, checksum, 4) || checksum+4 > security {
		return nil, ErrMalformed
	}
	certOffset := uint64(binary.LittleEndian.Uint32(data[security:]))
	certSize := uint64(binary.LittleEndian.Uint32(data[security+4:]))
	if certOffset != 0 || certSize != 0 {
		return peAuthenticodeDigest(data, algorithm)
	}
	h := algorithm.New()
	h.Write(data[:checksum])
	h.Write(data[checksum+4 : security])
	h.Write(data[security+8:])
	return h.Sum(nil), nil
}
