package engine

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/utmstack/UTMStack/agent/edr/chm"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// chmHeader recognizes Compiled HTML Help files: the ITSF header of version 2
// or 3 with its usual length, so that ordinary data starting with the four
// letters is not mistaken for one.
func chmHeader(b []byte) bool {
	if !bytes.HasPrefix(b, []byte("ITSF")) || len(b) < 12 {
		return false
	}
	version, length := binary.LittleEndian.Uint32(b[4:]), binary.LittleEndian.Uint32(b[8:])
	return (version == 2 || version == 3) && (length == 0x58 || length == 0x60)
}

// chm analyzes a Compiled HTML Help file as the reference does
// (docs/edr/native-engine/parity/chm.md, package chm): each directory entry
// but the "::" system files and the directories, in directory order, is
// matched by its metadata (position from 0) and analyzed, from the stored
// section or the LZX-compressed one. The compressed section is charged once
// and its entries are views into it. A directory this reader cannot parse,
// or an entry whose data cannot be decoded, is incomplete.
func (s *scanState) chm(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	a, err := chm.Open(data, uint64(max(0, min(s.maxFile, s.remaining))))
	switch {
	case errors.Is(err, chm.ErrLimit):
		return "", ErrLimit
	case err != nil:
		return "", ErrMalformed
	}
	var incomplete error
	held, charged := int64(0), false
	for i, f := range a.Files {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		meta := signatures.MemberMeta{Name: f.Name, HasName: true, Size: int64(f.Length), Index: i}
		if name, err := s.db.MatchMember(s.ctx, s.layers.TypeAt(-1), int64(len(data)), meta); name != "" || err != nil {
			return name, err
		}
		if f.Length == 0 {
			continue
		}
		b, err := a.Read(f)
		if f.Section == 1 && !charged {
			charged = true
			held = int64(a.Decoded())
			s.remaining -= held
		}
		if err != nil {
			if errors.Is(err, chm.ErrLimit) {
				incomplete = ErrLimit
			} else {
				incomplete = ErrMalformed
			}
			continue
		}
		prepaid := int64(0)
		if f.Section == 1 {
			prepaid = min(held, int64(len(b)))
			held -= prepaid
		}
		if sig, err := s.visitHeld(b, depth+1, prepaid); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}
