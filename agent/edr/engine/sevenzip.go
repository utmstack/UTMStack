package engine

import (
	"bytes"
	"errors"
	"hash/crc32"

	"github.com/utmstack/UTMStack/agent/edr/sevenzip"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// sevenZip analyzes a 7-Zip archive (package sevenzip), following the
// reference (docs/edr/native-engine/parity/sevenzip.md): each member is
// matched by its metadata (position from 0, packed size 0), then analyzed
// in archive order, each folder decoded once. Members of folders whose
// coders the reference does not decode either (BZip2, Deflate, Deflate64,
// Delta, Swap and others) are skipped as it skips them; archives it reads
// nothing of (major version, truncated or damaged header) add nothing. A
// member whose data or CRC is damaged, which the reference skips, and an
// encrypted member or header are incomplete.
func (s *scanState) sevenZip(data []byte, depth int) (string, error) {
	if depth >= s.maxDepth {
		return "", ErrLimit
	}
	lim := sevenzip.Limits{MaxFiles: max(0, s.maxFiles-s.files), MaxUnpack: uint64(max(0, min(s.maxFile, s.remaining))), Done: s.ctx.Done()}
	a, err := sevenzip.Open(data, lim)
	if errors.Is(err, sevenzip.ErrNothing) {
		return "", nil // the reference reads no member either
	}
	if err != nil {
		return "", s.sevenZipError(err)
	}
	var incomplete error
	decoded := map[int][]byte{}
	failed := map[int]error{}
	for i, f := range a.Files {
		if err := s.ctx.Err(); err != nil {
			return "", err
		}
		encrypted := f.Folder >= 0 && folderEncrypted(&a.Folders[f.Folder])
		meta := signatures.MemberMeta{Name: f.Name, HasName: true, Size: int64(f.Size), Encrypted: encrypted, Index: i}
		if name, err := s.db.MatchMember(s.ctx, s.layers.TypeAt(-1), int64(len(data)), meta); name != "" || err != nil {
			return name, err
		}
		if !f.HasStream || f.Size == 0 {
			continue
		}
		if s.files >= s.maxFiles {
			return "", ErrLimit
		}
		out, done := decoded[f.Folder]
		if !done {
			if err, ok := failed[f.Folder]; ok {
				if err != nil {
					incomplete = err
				}
				continue
			}
			b, err := a.Unpack(f.Folder)
			s.spend(len(b))
			if err != nil {
				if errors.Is(err, sevenzip.ErrUnsupported) {
					failed[f.Folder] = nil // the reference does not decode it either
					continue
				}
				failed[f.Folder] = s.sevenZipError(err)
				incomplete = failed[f.Folder]
				continue
			}
			decoded[f.Folder], out = b, b
		}
		if f.Offset+f.Size > uint64(len(out)) {
			incomplete = ErrMalformed
			continue
		}
		member := out[f.Offset : f.Offset+f.Size]
		if f.HasCRC && crc32.ChecksumIEEE(member) != f.CRC {
			// The reference skips such a member; it is reported, as
			// this engine's decoding cannot be told from damage.
			incomplete = ErrMalformed
			continue
		}
		if sig, err := s.visitHeld(member, depth+1, 0); sig != "" {
			return sig, nil
		} else if err != nil {
			incomplete = err
		}
	}
	return "", incomplete
}

func folderEncrypted(f *sevenzip.Folder) bool {
	for _, c := range f.Coders {
		if bytes.Equal(c.ID, []byte{6, 0xf1, 7, 1}) {
			return true
		}
	}
	return false
}

func (s *scanState) sevenZipError(err error) error {
	switch {
	case errors.Is(err, sevenzip.ErrAborted) && s.ctx.Err() != nil:
		return s.ctx.Err()
	case errors.Is(err, sevenzip.ErrLimit):
		return ErrLimit
	case errors.Is(err, sevenzip.ErrEncrypted):
		return ErrUnsupported
	}
	return ErrMalformed
}
