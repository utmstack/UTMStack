package engine

import (
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"encoding/binary"
	"hash/crc32"
	"io"
	"unicode/utf8"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// EGG records are little-endian and are delimited by the same four-byte
// terminator in the archive, file and block header sections. This reader
// handles ordinary non-solid, non-split member blocks. Other layouts are
// retained as incomplete until they have independent differential coverage.
const (
	eggMagic      = 0x41474745
	eggTerminator = 0x08e28222
	eggFileMagic  = 0x0a8590e3
	eggBlockMagic = 0x02b50c13
	eggNameMagic  = 0x0a8591ac
	eggWinMagic   = 0x2c86950b
	eggPosixMagic = 0x1ee922e5
	eggEncrypt    = 0x08d1470f
	eggSplit      = 0x24f5a262
	eggSolid      = 0x24e5a060
	eggComment    = 0x04c63672
)

func eggHeader(data []byte) bool {
	return len(data) >= 18 && binary.LittleEndian.Uint32(data) == eggMagic
}

type eggBlock struct {
	method byte
	packed uint32
	plain  uint32
	crc    uint32
	data   []byte
}

type eggMember struct {
	name      string
	length    uint64
	encrypted bool
	directory bool
	unknown   bool
	blocks    []eggBlock
}

func eggExtra(data []byte, at int) (magic uint32, flags byte, body []byte, next int, err error) {
	if at < 0 || at > len(data) || len(data)-at < 7 {
		return 0, 0, nil, at, ErrMalformed
	}
	magic = binary.LittleEndian.Uint32(data[at:])
	flags = data[at+4]
	at += 5
	var size uint32
	if flags&1 != 0 {
		if len(data)-at < 4 {
			return 0, 0, nil, at, ErrMalformed
		}
		size = binary.LittleEndian.Uint32(data[at:])
		at += 4
	} else {
		size = uint32(binary.LittleEndian.Uint16(data[at:]))
		at += 2
	}
	if uint64(size) > uint64(len(data)-at) {
		return 0, 0, nil, at, ErrMalformed
	}
	return magic, flags, data[at : at+int(size)], at + int(size), nil
}

func eggArchiveOf(data []byte, maxFiles int, comments *[][]byte) ([]eggMember, bool, error) {
	if !eggHeader(data) {
		return nil, false, ErrMalformed
	}
	var members []eggMember
	var unsupported, globalEncrypted bool
	var blockCount int
	at := 14 // magic, version, archive identifier, reserved
	terminated := false
	for at < len(data) {
		if len(data)-at < 4 {
			return members, unsupported, ErrMalformed
		}
		if binary.LittleEndian.Uint32(data[at:]) == eggTerminator {
			at += 4
			terminated = true
			break
		}
		magic, _, _, next, err := eggExtra(data, at)
		if err != nil {
			return members, unsupported, err
		}
		switch magic {
		case eggSolid, eggSplit:
			unsupported = true
		case eggEncrypt:
			globalEncrypted = true
		case eggComment:
			// Header extra fields are skipped by the reference parser; only
			// top-level comment records become scan representations.
		default:
			unsupported = true
		}
		at = next
	}
	if !terminated {
		return members, unsupported, ErrMalformed
	}
	for at < len(data) {
		if len(data)-at < 4 {
			return members, unsupported, ErrMalformed
		}
		magic := binary.LittleEndian.Uint32(data[at:])
		switch magic {
		case eggTerminator:
			return members, unsupported, nil
		case eggFileMagic:
			if len(members) >= maxFiles {
				return members, unsupported, ErrLimit
			}
			if len(data)-at < 16 {
				return members, unsupported, ErrMalformed
			}
			m := eggMember{length: binary.LittleEndian.Uint64(data[at+8:]), encrypted: globalEncrypted}
			at += 16
			closed := false
			for at < len(data) {
				if len(data)-at < 4 {
					return members, unsupported, ErrMalformed
				}
				if binary.LittleEndian.Uint32(data[at:]) == eggTerminator {
					at += 4
					closed = true
					break
				}
				field, flags, body, next, err := eggExtra(data, at)
				if err != nil {
					return members, unsupported, err
				}
				switch field {
				case eggNameMagic:
					if m.name != "" || len(body) == 0 {
						return members, unsupported, ErrMalformed
					}
					if flags&4 != 0 {
						m.unknown = true // filename encoding flag does not encrypt body
					}
					if flags&8 != 0 {
						if len(body) < 2 {
							return members, unsupported, ErrMalformed
						}
						body = body[2:]
						m.unknown = true // codepage conversion not yet implemented
					}
					if flags&16 != 0 {
						if len(body) < 4 {
							return members, unsupported, ErrMalformed
						}
						body = body[4:]
						m.unknown = true // relative parent-path resolution
					}
					m.name = string(body)
				case eggWinMagic:
					if len(body) < 9 {
						return members, unsupported, ErrMalformed
					}
					m.directory = m.directory || body[8]&0x40 != 0
				case eggPosixMagic:
					if len(body) < 20 {
						return members, unsupported, ErrMalformed
					}
					m.directory = m.directory || binary.LittleEndian.Uint32(body)&0xf000 == 0x4000
				case eggEncrypt:
					m.encrypted = true
				case eggComment:
					// The reference indexes file comments but does not scan
					// them as separate children. The enclosing raw pass still
					// sees their original archive bytes.
				default:
					m.unknown = true
				}
				at = next
			}
			if !closed || m.name == "" {
				return members, unsupported, ErrMalformed
			}
			members = append(members, m)
		case eggBlockMagic:
			if len(members) == 0 || len(data)-at < 22 {
				return members, unsupported, ErrMalformed
			}
			// Blocks do not consume the engine's file counter, so cap their
			// indexing memory separately from the bounded input bytes.
			if blockCount >= 65536 {
				return members, unsupported, ErrLimit
			}
			blockCount++
			b := eggBlock{
				method: data[at+4],
				plain:  binary.LittleEndian.Uint32(data[at+6:]),
				packed: binary.LittleEndian.Uint32(data[at+10:]),
				crc:    binary.LittleEndian.Uint32(data[at+14:]),
			}
			if binary.LittleEndian.Uint32(data[at+18:]) != eggTerminator {
				return members, unsupported, ErrMalformed
			}
			at += 22
			if uint64(b.packed) > uint64(len(data)-at) {
				return members, unsupported, ErrMalformed
			}
			b.data = data[at : at+int(b.packed)]
			at += int(b.packed)
			members[len(members)-1].blocks = append(members[len(members)-1].blocks, b)
		case eggComment:
			_, flags, body, next, err := eggExtra(data, at)
			if err != nil {
				return members, unsupported, err
			}
			if flags&4 != 0 || (flags&8 != 0 && !utf8.Valid(body)) {
				// The reference treats multibyte comments as UTF-8. Keep
				// undecodable and encrypted values incomplete.
				unsupported = true
			} else if comments != nil {
				// ClamAV scans a C string, ending at its first NUL.
				if end := bytes.IndexByte(body, 0); end >= 0 {
					body = body[:end]
				}
				if len(body) != 0 {
					*comments = append(*comments, body)
				}
			}
			at = next
		default:
			return members, unsupported, ErrMalformed
		}
	}
	return members, unsupported, nil // reference accepts EOF after full records
}

func (s *scanState) egg(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	var comments [][]byte
	members, unsupported, err := eggArchiveOf(data, s.maxFiles, &comments)
	if err != nil {
		return "", err
	}
	var incomplete error
	if unsupported {
		incomplete = ErrUnsupported
	}
	for _, comment := range comments {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		name, err := s.childBytes(comment, depth)
		if name != "" {
			return name, nil
		}
		if err != nil {
			incomplete = err
		}
	}
	for i, m := range members {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		var packed, plain uint64
		for _, b := range m.blocks {
			packed += uint64(b.packed)
			plain += uint64(b.plain)
		}
		if plain != m.length {
			incomplete = ErrMalformed
		}
		meta := signatures.MemberMeta{Name: m.name, HasName: true, Size: int64(min(plain, uint64(^uint64(0)>>1))), CompressedSize: int64(min(packed, uint64(^uint64(0)>>1))), Encrypted: m.encrypted, Index: i + 1}
		if name, err := s.db.MatchMember(s.ctx, signatures.TypeEGG, int64(len(data)), meta); name != "" || err != nil {
			return name, err
		}
		if m.unknown {
			incomplete = ErrUnsupported
		}
		if m.directory || len(m.blocks) == 0 {
			if !m.directory && m.length != 0 {
				incomplete = ErrMalformed
			}
			continue
		}
		if m.encrypted {
			incomplete = ErrUnsupported
			continue
		}
		max, err := s.childBudget(depth)
		if err != nil {
			return "", err
		}
		var decoded []byte
		var damaged bool
		for _, b := range m.blocks {
			if b.method > 2 {
				incomplete = ErrUnsupported
				damaged = true
				break
			}
			if b.method == 0 && b.packed != b.plain {
				incomplete = ErrMalformed
				damaged = true
				break
			}
			var reader io.Reader = bytes.NewReader(b.data)
			var closer io.Closer
			switch b.method {
			case 1:
				r := flate.NewReader(reader)
				closer, reader = r, r
			case 2:
				reader = bzip2.NewReader(reader)
			}
			part, readErr := io.ReadAll(io.LimitReader(contextReader{s.ctx, reader}, max-int64(len(decoded))+1))
			if closer != nil {
				_ = closer.Close()
			}
			decoded = append(decoded, part...)
			if int64(len(decoded)) > max {
				s.spend(len(decoded))
				return "", ErrLimit
			}
			if readErr != nil {
				incomplete = ErrMalformed
				damaged = true
				break
			}
			if uint64(len(part)) != uint64(b.plain) || crc32.ChecksumIEEE(part) != b.crc {
				incomplete = ErrMalformed
				damaged = true
			}
		}
		if len(decoded) == 0 {
			continue
		}
		name, childErr := s.childBytes(decoded, depth)
		if name != "" {
			return name, nil
		}
		if childErr != nil {
			incomplete = childErr
		} else if damaged {
			incomplete = ErrMalformed
		}
	}
	return "", incomplete
}
