package signatures

// Text detection follows the reference's classification of the first bytes
// of an object when no magic entry typed it. Every byte is in one class:
// never text (F), plain text (T), or one of two extended-character classes
// that count as text. The reference also has a UTF-8 check, but anything that
// passes it passes the ASCII check first, so UTF-8 text is typed TEXT_ASCII
// and the check is not needed.

// neverText reports the F class: 0x00-0x06, 0x0B, 0x0E-0x1A, 0x1C-0x1F and
// 0x7F. Every byte from 0x80 up is text.
func neverText(c byte) bool {
	return c <= 0x06 || c == 0x0b || c >= 0x0e && c <= 0x1a || c >= 0x1c && c <= 0x1f || c == 0x7f
}

// textType classifies buf: TEXT_ASCII when no byte is never-text, otherwise
// UTF-16 when the reference's UTF-16 check passes, otherwise BINARY_DATA.
func textType(buf []byte) FileType {
	ascii := true
	for _, c := range buf {
		if neverText(c) {
			ascii = false
			break
		}
	}
	if ascii {
		return TypeTextASCII
	}
	if t, ok := utf16Text(buf); ok {
		return t
	}
	return TypeBinaryData
}

// utf16Text is the reference's UTF-16 check. A leading FF FE or FE FF mark
// selects the byte order; without one the text is read as big-endian, so
// little-endian text without a mark is typed UTF16BE. Units are read from
// offset 2 whether or not a mark is there, up to the last complete pair. A
// unit of 0xFFFE rejects the buffer. A unit below 128 that is not plain text
// rejects an unmarked buffer at once and counts as bad in a marked one; a
// marked buffer is rejected when the bad units reach half its byte length,
// which the units after the mark can never do. Units of 128 or more are
// ignored.
func utf16Text(buf []byte) (FileType, bool) {
	if len(buf) < 2 {
		return 0, false
	}
	bigEndian, marked := true, true
	switch {
	case buf[0] == 0xff && buf[1] == 0xfe:
		bigEndian = false
	case buf[0] == 0xfe && buf[1] == 0xff:
	default:
		marked = false
	}
	bad := 0
	for i := 2; i+1 < len(buf); i += 2 {
		unit := uint16(buf[i])<<8 | uint16(buf[i+1])
		if !bigEndian {
			unit = uint16(buf[i+1])<<8 | uint16(buf[i])
		}
		if unit == 0xfffe {
			return 0, false
		}
		if unit < 128 && neverText(byte(unit)) {
			if !marked {
				return 0, false
			}
			bad++
		}
	}
	if marked && bad >= len(buf)/2 {
		return 0, false
	}
	if bigEndian {
		return TypeTextUTF16BE, true
	}
	return TypeTextUTF16LE, true
}

// isTar recognizes a tar header the reference way: at least 512 bytes whose
// octal checksum field (offset 148) equals the sum of the header bytes with
// that field counted as spaces. The GNU magic "ustar  \0" at offset 257 gives
// POSIX_TAR; any other header gives OLD_TAR.
func isTar(buf []byte) (FileType, bool) {
	if len(buf) < 512 {
		return 0, false
	}
	recorded := tarOctal(buf[148:156])
	if recorded < 0 {
		return 0, false
	}
	sum := int64(0)
	for i, c := range buf[:512] {
		if i >= 148 && i < 156 {
			c = ' '
		}
		sum += int64(c)
	}
	if sum != recorded {
		return 0, false
	}
	if string(buf[257:265]) == "ustar  \x00" {
		return TypePosixTar, true
	}
	return TypeOldTar, true
}

// tarOctal reads a numeric tar field: leading white space, then octal
// digits, ended by the field's end, NUL or white space. An all-blank field
// or one ended by another byte gives -1.
func tarOctal(field []byte) int64 {
	i := 0
	for i < len(field) && cSpace(field[i]) {
		i++
	}
	if i == len(field) {
		return -1
	}
	v := int64(0)
	for ; i < len(field) && field[i] >= '0' && field[i] <= '7'; i++ {
		v = v<<3 | int64(field[i]-'0')
	}
	if i < len(field) && field[i] != 0 && !cSpace(field[i]) {
		return -1
	}
	return v
}
