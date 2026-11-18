package engine

import (
	"context"
	"encoding/binary"
)

// word6Directory follows the Word 6 macro-directory reader in ClamAV 1.5.4
// (libclamav/vba_extract.c, word_read_fib through cli_wm_readdir). It returns
// the packed 24-byte records without allocating from an untrusted count.
// The table's length bounds dispatch of record IDs, not the data read by a
// dispatched record. In particular, a macro list may extend past that length.
// The reference discards an unreadable directory without exposing any macro;
// modern Word streams can have unrelated values in these legacy FIB fields.
// This is a behavioral implementation; it does not execute WordBasic tokens.
func word6Directory(ctx context.Context, data []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(data) < 0x120 {
		return nil, nil
	}
	le := binary.LittleEndian
	offset, length := le.Uint32(data[0x118:]), le.Uint32(data[0x11c:])
	if length == 0 {
		return nil, nil
	}
	// Both reference additions are uint32 expressions before conversion to
	// off_t. Preserve that wrap without turning it into an unchecked Go slice.
	at, end := uint64(offset+1), uint64(offset+length)
	take := func(n uint64) ([]byte, bool) {
		if !fileSpan(data, at, n) {
			return nil, false
		}
		b := data[at : at+n]
		at += n
		return b, true
	}
	count16 := func() (uint16, bool) {
		b, ok := take(2)
		if !ok {
			return 0, false
		}
		return le.Uint16(b), true
	}
	for at < end {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		id, ok := take(1)
		if !ok {
			return nil, nil
		}
		switch id[0] {
		case 0x01:
			count, ok := count16()
			if !ok {
				return nil, nil
			}
			records, ok := take(uint64(count) * 24)
			if !ok {
				return nil, nil
			}
			return records, nil
		case 0x03:
			count, ok := take(1)
			if !ok {
				return nil, nil
			}
			at += uint64(count[0]) * 14
			count, ok = take(1)
			if !ok {
				return nil, nil
			}
			n := count[0]
			if n == 0 {
				more, ok := take(2)
				if !ok {
					return nil, nil
				}
				if more[0] != 2 {
					at -= 2
					break
				}
				n = more[1]
			}
			if n != 0 {
				at += uint64(n)*4 + 1
			}
		case 0x05:
			count, ok := count16()
			if !ok {
				return nil, nil
			}
			at += uint64(count) * 12
		case 0x10:
			size, ok := count16()
			if !ok {
				return nil, nil
			}
			unicode := size == 0xffff
			if unicode {
				size, ok = count16()
				if !ok {
					return nil, nil
				}
			}
			// The pinned reader subtracts the whole declared size after
			// one name, so it visits at most one name even when more fit.
			if int16(size) > 0 {
				name, ok := take(1)
				if !ok {
					return nil, nil
				}
				if unicode {
					at += uint64(name[0])*2 + 3
				} else {
					at += uint64(name[0]) + 2
				}
			}
		case 0x11:
			count, ok := count16()
			if !ok {
				return nil, nil
			}
			for i := 0; i < int(count); i++ {
				if i&255 == 0 {
					if err := ctx.Err(); err != nil {
						return nil, err
					}
				}
				at += 2 // internal name ID
				name, ok := take(1)
				if !ok {
					return nil, nil
				}
				at += uint64(name[0]) + 1
			}
		default: // includes the 0x40 and 0x12 terminators
			return nil, nil
		}
	}
	return nil, nil
}

// word6Macros decrypts each bounded macro and scans it as an OLE2 buffer,
// exactly as vba_scandata does. This deliberately does not recursively type
// the tokens, normalize them as text, or apply file hashes and EOF offsets.
func (s *scanState) word6Macros(data []byte, depth int) (string, error) {
	records, err := word6Directory(s.ctx, data)
	if err != nil {
		return "", err
	}
	var incomplete error
	for len(records) != 0 {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		record := records[:24]
		records = records[24:]
		length := uint64(binary.LittleEndian.Uint32(record[12:]))
		offset := uint64(binary.LittleEndian.Uint32(record[20:]))
		if length == 0 {
			continue
		}
		if !fileSpan(data, offset, length) {
			incomplete = ErrMalformed
			continue // a later valid macro can still detect
		}
		if depth > s.maxDepth || s.files >= s.maxFiles || length > uint64(max(s.maxFile, 0)) || length > uint64(max(s.remaining, 0)) {
			return "", ErrLimit
		}
		source := make([]byte, int(length))
		key := record[1]
		for i, b := range data[offset : offset+length] {
			if i&4095 == 0 {
				if err := s.ctx.Err(); err != nil {
					return "", err
				}
			}
			source[i] = b ^ key
		}
		s.spend(len(source))
		s.files++
		if sig, err := s.vbaModule(source, depth); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}
