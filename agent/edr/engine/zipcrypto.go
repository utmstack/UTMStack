package engine

import "hash/crc32"

// PKWARE's traditional ZIP cipher uses the internal (non-finalized) CRC32
// state. ZIP-specific passwords are attempted before container-generic ones.
type zipKeys struct{ a, b, c uint32 }

func zipCRC(state uint32, input byte) uint32 {
	return crc32.IEEETable[byte(state)^input] ^ state>>8
}

func (k *zipKeys) update(plain byte) {
	k.a = zipCRC(k.a, plain)
	k.b = (k.b+uint32(byte(k.a)))*134775813 + 1
	k.c = zipCRC(k.c, byte(k.b>>24))
}

func (k *zipKeys) decrypt(cipher byte) byte {
	t := uint16(k.c) | 2
	plain := cipher ^ byte(t*(t^1)>>8)
	k.update(plain)
	return plain
}

func (s *scanState) zipDecrypt(packed []byte, version, flags, mtime uint16, crc uint32) ([]byte, error) {
	if len(packed) < 12 {
		return nil, ErrUnsupported
	}
	if flags&0x40 != 0 { // ZIP strong encryption uses a different header.
		return nil, ErrUnsupported
	}
	var decoded []byte
	var interrupted error
	s.db.ForEachZIPPassword(func(password []byte) bool {
		if interrupted = s.ctx.Err(); interrupted != nil {
			return true
		}
		keys := zipKeys{0x12345678, 0x23456789, 0x34567890}
		for _, c := range password {
			keys.update(c)
		}
		var header [12]byte
		for i := range header {
			header[i] = keys.decrypt(packed[i])
		}
		check := uint16(crc >> 16)
		if flags&zipFlagDescriptor != 0 {
			check = mtime
		}
		if version > 20 {
			if header[11] != byte(check>>8) {
				return false
			}
		} else if uint16(header[10])|uint16(header[11])<<8 != check {
			return false
		}
		decoded = make([]byte, len(packed)-12)
		for i, c := range packed[12:] {
			if i&0x3fff == 0 {
				if interrupted = s.ctx.Err(); interrupted != nil {
					decoded = nil
					return true
				}
			}
			decoded[i] = keys.decrypt(c)
		}
		return true
	})
	if interrupted != nil {
		return nil, interrupted
	}
	if decoded == nil {
		return nil, ErrUnsupported
	}
	return decoded, nil
}
