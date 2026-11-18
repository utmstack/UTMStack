package bytecode

const maxBzip2Memory = 64 << 20

type bzip2State struct {
	from, to int32
	stream   *streamBzip2
	reserved uint64
	closed   bool
}

// bzip2API connects the bounded Go streaming decoder to the ClamBC pipe ABI.
// Missing input and full output windows preserve their exact continuation state.
func (s *apiState) bzip2API(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 105 {
		from, to := s.pipe(args[0]), s.pipe(args[1])
		if from == nil || to == nil {
			return apiFailure, nil
		}
		if from.closed || to.closed {
			return 0, ErrRuntime
		}
		if len(s.bzip2s) >= maxBzip2s {
			return 0, ErrLimit
		}
		state := &bzip2State{from: int32(args[0]), to: int32(args[1])}
		state.stream = &streamBzip2{reserve: func(n int) error {
			if n < 0 || uint64(n) > maxBzip2Memory-s.bzip2Memory {
				return ErrLimit
			}
			s.bzip2Memory += uint64(n)
			state.reserved += uint64(n)
			return nil
		}}
		n := len(s.bzip2s)
		s.bzip2s = append(s.bzip2s, state)
		return uint64(n), nil
	}
	index := int32(args[0])
	if index < 0 || int(index) >= len(s.bzip2s) || s.bzip2s[index] == nil || s.bzip2s[index].closed {
		return apiFailure, nil
	}
	state := s.bzip2s[index]
	if id == 107 {
		state.closed = true
		state.stream = nil
		s.bzip2Memory -= state.reserved
		state.reserved = 0
		return 0, nil
	}
	from, to := s.pipe(uint64(uint32(state.from))), s.pipe(uint64(uint32(state.to)))
	if from == nil || to == nil || from.closed || to.closed || from == to {
		return 0, ErrRuntime
	}
	inAvail, outAvail := s.pipeReadAvail(from), s.pipeWriteAvail(to)
	if inAvail == 0 || outAvail == 0 {
		return apiFailure, nil
	}
	if mem == nil {
		return 0, ErrRuntime
	}
	if err := bytesWork(uint64(inAvail) + uint64(outAvail)); err != nil {
		return 0, err
	}
	var input []byte
	if from.file {
		input = s.input[from.read : from.read+inAvail]
	} else {
		var err error
		input, err = mem.read(from.ptr+uint64(from.read), int(inAvail))
		if err != nil {
			return 0, err
		}
	}
	region, offset, err := mem.span(to.ptr+uint64(to.write), int(outAvail), true)
	if err != nil {
		return 0, err
	}
	consumed, produced, status, err := state.stream.process(input, region.data[offset:offset+int(outAvail)], bytesWork)
	if err != nil {
		return 0, err
	}
	from.read += uint32(consumed)
	if !from.file && from.read >= from.size && from.write >= from.size {
		from.read, from.write = 0, 0
	}
	to.write += uint32(produced)
	return uint64(uint32(int32(status))), nil
}
