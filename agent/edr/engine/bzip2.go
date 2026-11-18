package engine

import (
	"bytes"
	"io"

	"github.com/utmstack/UTMStack/agent/edr/bytecode"
)

// bzip2Member preserves physical member boundaries and damaged-stream
// publication. Later bzip2 members receive independent scans as well: stopping
// at the native scanner's first member would discard existing EDR detections.
// Joining their bytes would instead change whole-child hash/EOF decisions.
func (s *scanState) bzip2Member(data []byte, depth int) (string, error) {
	// Empty/tiny decoded members are not admitted as scan objects. Bound
	// their decoder work independently so they cannot bypass MaxFiles.
	memberLimit := s.maxFiles - s.files
	members := 0
	for {
		if members >= memberLimit {
			return "", ErrLimit
		}
		members++
		name, err, consumed := s.bzip2SingleMember(data, depth)
		if name != "" || err != nil {
			return name, err
		}
		if consumed <= 0 || consumed > len(data) {
			return "", ErrMalformed
		}
		data = data[consumed:]
		if !bytes.HasPrefix(data, []byte("BZh")) {
			return "", nil
		}
	}
}

func (s *scanState) bzip2SingleMember(data []byte, depth int) (string, error, int) {
	max, err := s.childBudget(depth)
	if err != nil {
		return "", err, 0
	}
	r := bytecode.NewBzip2MemberReader(s.ctx, data)
	defer r.Close()
	out, readErr := io.ReadAll(io.LimitReader(r, max+1))
	produced := r.DecodedBytes()
	consumed := r.CompressedBytes()
	if err := s.ctx.Err(); err != nil {
		s.spend(int(produced))
		return "", err, consumed
	}
	if produced > max {
		s.spend(int(produced))
		return "", ErrLimit, consumed
	}
	// Unflushed error bytes and read-ahead are still expansion work. Charge
	// them before admitting the published child to the remaining budget.
	s.spend(int(produced - int64(len(out))))
	if readErr != nil && len(out) < minAnalyzedObject {
		s.spend(len(out))
		return "", ErrMalformed, consumed
	}
	name, childErr := s.visit(out, depth+1)
	if name != "" {
		return name, nil, consumed
	}
	if childErr != nil {
		return "", childErr, consumed
	}
	if readErr != nil {
		return "", ErrMalformed, consumed
	}
	return "", nil, consumed
}
