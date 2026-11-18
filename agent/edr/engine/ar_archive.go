package engine

import (
	"bytes"

	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Unix ar stores fixed-width ASCII headers and even-byte-aligned members.
// GNU long names use the // string table; BSD names use #1/length and occupy
// the beginning of the member's declared bytes. Debian packages are regular
// ar archives whose data members are compressed tar archives. Names remain
// metadata: nothing is written to disk or resolved against the filesystem.
// Thin archives refer to external files and are explicitly unsupported.
// https://man.freebsd.org/cgi/man.cgi?query=ar&sektion=5
// https://sourceware.org/binutils/docs/binutils/ar.html
var arMagic = []byte("!<arch>\n")
var arThinMagic = []byte("!<thin>\n")

const arMaxName = 4096

func arNumber(field []byte) (int, bool) {
	field = bytes.TrimSpace(field)
	if len(field) == 0 {
		return 0, false
	}
	n := 0
	for _, digit := range field {
		if digit < '0' || digit > '9' || n > (int(^uint(0)>>1)-int(digit-'0'))/10 {
			return 0, false
		}
		n = n*10 + int(digit-'0')
	}
	return n, true
}

// arRecord checks the declared extent before constructing any member slice.
// The returned next offset always advances, including for empty members.
func arRecord(data []byte, pos int) (name, content []byte, next int, err error) {
	if pos < len(arMagic) || pos > len(data) || len(data)-pos < 60 {
		return nil, nil, pos, ErrMalformed
	}
	header := data[pos : pos+60]
	if !bytes.Equal(header[58:60], []byte("`\n")) {
		return nil, nil, pos, ErrMalformed
	}
	size, ok := arNumber(header[48:58])
	start := pos + 60
	if !ok || size > len(data)-start {
		return nil, nil, pos, ErrMalformed
	}
	next = start + size
	if size&1 != 0 {
		if next >= len(data) || data[next] != '\n' {
			return nil, nil, pos, ErrMalformed
		}
		next++
	}
	return bytes.TrimRight(header[:16], " "), data[start : start+size], next, nil
}

// arLongNames makes forward references to GNU's string table work too. It
// walks the same bounded records without allocating a second representation;
// a malformed later member remains incomplete but cannot hide an earlier
// independently detected member in the analysis pass.
func (s *scanState) arLongNames(data []byte, maxEntries int) ([]byte, error) {
	var table []byte
	for pos, entries := len(arMagic), 0; pos < len(data); entries++ {
		if err := s.ctx.Err(); err != nil {
			return table, err
		}
		if entries >= maxEntries {
			return table, ErrLimit
		}
		name, content, next, err := arRecord(data, pos)
		if err != nil {
			return table, err
		}
		if bytes.Equal(name, []byte("//")) && table == nil {
			table = content
		}
		pos = next
	}
	return table, nil
}

func arMemberName(raw, content, table []byte) (name, payload []byte, metadata bool, err error) {
	switch {
	case bytes.Equal(raw, []byte("/")), bytes.Equal(raw, []byte("/SYM64/")), bytes.Equal(raw, []byte("//")):
		return raw, content, true, nil
	case bytes.HasPrefix(raw, []byte("#1/")):
		n, ok := arNumber(raw[3:])
		if !ok || n < 1 || n > len(content) {
			return nil, content, false, ErrMalformed
		}
		name, payload = bytes.TrimRight(content[:n], "\x00"), content[n:]
		if len(name) > arMaxName {
			return nil, payload, false, ErrLimit
		}
	case len(raw) > 1 && raw[0] == '/':
		offset, ok := arNumber(raw[1:])
		if !ok || offset >= len(table) || offset > 0 && table[offset-1] != '\n' {
			return nil, content, false, ErrMalformed
		}
		// Bound repeated references to a single pathological string-table
		// name independently of the archive's total size.
		end := bytes.IndexByte(table[offset:offset+min(len(table)-offset, arMaxName+2)], '\n')
		if end < 0 {
			if len(table)-offset > arMaxName {
				return nil, content, false, ErrLimit
			}
			return nil, content, false, ErrMalformed
		}
		name, payload = bytes.TrimSuffix(table[offset:offset+end], []byte("/")), content
		if len(name) > arMaxName {
			return nil, payload, false, ErrLimit
		}
	default:
		name, payload = bytes.TrimSuffix(raw, []byte("/")), content
	}
	if len(name) == 0 || bytes.IndexByte(name, 0) >= 0 {
		return nil, payload, false, ErrMalformed
	}
	metadata = bytes.Equal(name, []byte("__.SYMDEF")) || bytes.Equal(name, []byte("__.SYMDEF SORTED")) ||
		bytes.Equal(name, []byte("__.SYMDEF_64")) || bytes.Equal(name, []byte("__.SYMDEF_64 SORTED"))
	return name, payload, metadata, nil
}

func (s *scanState) arArchive(data []byte, depth int) (string, error) {
	if bytes.HasPrefix(data, arThinMagic) {
		return "", ErrUnsupported
	}
	if !bytes.HasPrefix(data, arMagic) {
		return "", ErrMalformed
	}
	if err := s.ctx.Err(); err != nil {
		return "", err
	}
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	maxEntries := s.maxFiles - s.files
	table, incomplete := s.arLongNames(data, maxEntries)
	for pos, entries := len(arMagic), 0; pos < len(data); entries++ {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		if entries >= maxEntries {
			return "", ErrLimit
		}
		raw, content, next, err := arRecord(data, pos)
		if err != nil {
			return "", err
		}
		pos = next
		name, payload, metadata, err := arMemberName(raw, content, table)
		if err != nil {
			incomplete = err
		}
		if metadata {
			continue
		}
		// There is no ar FileType in the compatible signature format. Keep
		// the actual recognized parent type; wildcard metadata rules can
		// still inspect member names, sizes and positions.
		member := signatures.MemberMeta{Name: string(name), HasName: len(name) != 0, Size: int64(len(payload)),
			CompressedSize: int64(len(payload)), Index: entries + 1}
		if sig, err := s.db.MatchMember(s.ctx, s.layers.TypeAt(-1), int64(len(data)), member); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
		if len(payload) == 0 {
			continue
		}
		// Visiting charges the member's actual bytes and every nested
		// decoded object. The slice itself has no expansion allocation.
		if sig, err := s.visit(payload, depth+1); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}
