package engine

import (
	"bytes"
	"encoding/binary"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// The ZIP reader follows the reference's reader as learned from its outputs
// on damaged archives (docs/edr/native-engine/parity/zip.md):
//   - The end record is the last "PK\x05\x06" with its 22 bytes inside the
//     object, wherever it is (trailing data of any length is allowed).
//     Without one nothing is read. Its member count and directory size are
//     not used.
//   - The central directory starts exactly at the end record's offset; its
//     entries are read while their signature holds and they fit.
//   - Each entry's local header (at the entry's offset, its signature
//     checked) gives the member's flags, method and sizes; with flag 8 the
//     sizes come from the entry instead. Data that would run past the
//     object is not read; neither the CRC nor the uncompressed size is
//     checked. ZIP64 sizes/offsets are additionally resolved here, with
//     overflow-safe bounds before any slice or conversion to an index.
type zipMember struct {
	central, local   signatures.MemberMeta
	hasLocal         bool
	flags, method    uint16
	version, mtime   uint16
	crc              uint32
	compressed, size uint64
	data             int64 // start of the member's data; -1 when it cannot be read
	issue            error // invalid ZIP64 metadata must not become clean
}

// zip64Extra resolves only the sentinel-valued fields, in APPNOTE 4.5.3's
// prescribed order: uncompressed size, compressed size, local-header offset.
// No attacker-declared count or size is used for allocation.
func zip64Extra(extra []byte, size, compressed, offset uint32) (uint64, uint64, uint64, error) {
	u, c, o := uint64(size), uint64(compressed), uint64(offset)
	if size != 0xffffffff && compressed != 0xffffffff && offset != 0xffffffff {
		return u, c, o, nil
	}
	for len(extra) >= 4 {
		id, n := binary.LittleEndian.Uint16(extra), int(binary.LittleEndian.Uint16(extra[2:]))
		extra = extra[4:]
		if n > len(extra) {
			return u, c, o, ErrMalformed
		}
		if id == 1 {
			values := extra[:n]
			for _, f := range []struct {
				needed bool
				value  *uint64
			}{
				{size == 0xffffffff, &u}, {compressed == 0xffffffff, &c}, {offset == 0xffffffff, &o},
			} {
				if f.needed {
					if len(values) < 8 {
						return u, c, o, ErrMalformed
					}
					*f.value = binary.LittleEndian.Uint64(values)
					values = values[8:]
				}
			}
			return u, c, o, nil
		}
		extra = extra[n:]
	}
	return u, c, o, ErrMalformed
}

func zipMetaSize(n uint64) int64 {
	return int64(min(n, uint64(1<<63-1)))
}

// zipMembers returns the members the reference reads, at most maxFiles of
// them (ErrLimit beyond).
func zipMembers(data []byte, maxFiles int) ([]zipMember, error) {
	end := bytes.LastIndex(data[:max(0, len(data)-18)], []byte("PK\x05\x06"))
	if end < 0 {
		return nil, nil
	}
	u16 := func(at int64) uint16 { return binary.LittleEndian.Uint16(data[at:]) }
	u32 := func(at int64) uint32 { return binary.LittleEndian.Uint32(data[at:]) }
	size := int64(len(data))
	directory := uint64(u32(int64(end) + 16))
	zip64Directory := directory == 0xffffffff
	if directory == 0xffffffff {
		// A ZIP64 locator immediately precedes the ordinary end record.
		// Its target must contain the complete fixed ZIP64 end record.
		if end < 20 || u32(int64(end)-20) != 0x07064b50 {
			return nil, ErrMalformed
		}
		if u32(int64(end)-16) != 0 || u32(int64(end)-4) != 1 {
			return nil, ErrUnsupported
		}
		at := binary.LittleEndian.Uint64(data[end-12:])
		if at > uint64(len(data)) || uint64(len(data))-at < 56 {
			return nil, ErrMalformed
		}
		recordSize := binary.LittleEndian.Uint64(data[at+4:])
		if u32(int64(at)) != 0x06064b50 || recordSize < 44 || recordSize > uint64(len(data))-at-12 {
			return nil, ErrMalformed
		}
		if u32(int64(at)+16) != 0 || u32(int64(at)+20) != 0 {
			return nil, ErrUnsupported
		}
		directory = binary.LittleEndian.Uint64(data[at+48:])
	}
	if directory > uint64(size) {
		if zip64Directory {
			return nil, ErrMalformed
		}
		return nil, nil // preserve ordinary reference-compatible damaged EOCD handling
	}
	var members []zipMember
	for pos := int64(directory); pos+46 <= size && u32(pos) == 0x02014b50; {
		next := pos + 46 + int64(u16(pos+28)) + int64(u16(pos+30)) + int64(u16(pos+32))
		if next > size {
			break
		}
		if len(members) >= maxFiles {
			return nil, ErrLimit
		}
		index := len(members) + 1
		extra := data[pos+46+int64(u16(pos+28)) : pos+46+int64(u16(pos+28))+int64(u16(pos+30))]
		usize, csize, local, zip64err := zip64Extra(extra, u32(pos+24), u32(pos+20), u32(pos+42))
		m := zipMember{data: -1, issue: zip64err, central: signatures.MemberMeta{
			Name: string(data[pos+46 : pos+46+int64(u16(pos+28))]), HasName: true,
			CompressedSize: zipMetaSize(csize), Size: zipMetaSize(usize), Encrypted: u16(pos+8)&1 != 0, Index: index}}
		if zip64err == nil && local <= uint64(size) && uint64(size)-local >= 30 && u32(int64(local)) == 0x04034b50 {
			local := int64(local)
			m.version, m.mtime, m.crc = u16(local+4), u16(local+10), u32(local+14)
			names, extraLen := int64(u16(local+26)), int64(u16(local+28))
			m.hasLocal = local+30+names <= size
			m.compressed, m.size = uint64(u32(local+18)), uint64(u32(local+22))
			if start := local + 30 + names; start+extraLen <= size {
				m.size, m.compressed, _, m.issue = zip64Extra(data[start:start+extraLen], u32(local+22), u32(local+18), 0)
			} else if m.compressed == 0xffffffff || m.size == 0xffffffff {
				m.issue = ErrMalformed
			}
			if m.hasLocal {
				m.local = signatures.MemberMeta{Name: string(data[local+30 : local+30+names]), HasName: true,
					CompressedSize: zipMetaSize(m.compressed), Size: zipMetaSize(m.size), Encrypted: u16(local+6)&1 != 0, Index: index}
			}
			m.flags, m.method = u16(local+6), u16(local+8)
			if m.flags&8 != 0 {
				m.compressed, m.size, m.issue = csize, usize, zip64err
			}
			if start := local + 30 + names + extraLen; m.issue == nil && start <= size && m.compressed <= uint64(size-start) {
				m.data = start
			} else if m.issue == nil && (u32(local+18) == 0xffffffff || u32(local+22) == 0xffffffff) {
				m.issue = ErrMalformed
			}
		}
		members = append(members, m)
		pos = next
	}
	return members, nil
}
