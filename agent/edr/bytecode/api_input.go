package bytecode

import (
	"bytes"
	"encoding/binary"
)

// inputSwitch models the two fmaps held by ClamAV's bytecode context. The
// extracted fmap is a snapshot of the output file at the time of the switch;
// later writes do not extend that mapping. File-pipe cursors use whichever
// fmap is current, while a pointer borrowed from the retired extracted fmap
// becomes invalid when the bytecode switches back.
func (s *apiState) inputSwitch(extracted bool, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if extracted == s.extractedInput {
		return 0, nil
	}
	if extracted {
		if !s.outputOpen || len(s.output) == 0 {
			return apiFailure, nil
		}
		if err := bytesWork(uint64(len(s.output))); err != nil {
			return 0, err
		}
		snapshot := bytes.Clone(s.output)
		if err := s.setInputSize(mem, len(snapshot)); err != nil {
			return 0, err
		}
		s.originalInput, s.savedFilePtr = s.input, s.filePtr
		s.input, s.filePtr, s.extractedInput = snapshot, 0, true
		return 0, nil
	}
	if err := s.setInputSize(mem, len(s.originalInput)); err != nil {
		return 0, err
	}
	if s.filePtr != 0 && mem != nil {
		mem.release(s.filePtr)
	}
	s.input, s.filePtr = s.originalInput, s.savedFilePtr
	s.originalInput, s.savedFilePtr, s.extractedInput = nil, 0, false
	return 0, nil
}

// The filesize global is read-only to bytecode, but the host context updates
// it when the active fmap changes. No bytecode-visible write is permitted.
func (s *apiState) setInputSize(mem *memory, size int) error {
	if s.fileSizePtr == 0 {
		return nil
	}
	if mem == nil {
		return ErrRuntime
	}
	region, offset, err := mem.span(s.fileSizePtr, 4, false)
	if err != nil || !region.readOnly {
		return ErrRuntime
	}
	binary.LittleEndian.PutUint32(region.data[offset:], uint32(size))
	return nil
}
