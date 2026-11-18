package bytecode

type inflateState struct {
	from, to int32
	stream   *streamInflater
	closed   bool
}

// inflateAPI connects the Go-owned incremental DEFLATE decoder to the
// ClamBC pipe API. Data errors search for a full-flush marker and expose
// zlib's status to the bytecode caller. Dictionary requests return
// Z_NEED_DICT; ClamBC provides no setter for the requested dictionary.
func (s *apiState) inflateAPI(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 39 {
		fromID, toID := int32(args[0]), int32(args[1])
		from, to := s.pipe(args[0]), s.pipe(args[1])
		if from == nil || to == nil {
			return apiFailure, nil
		}
		if from.closed || to.closed {
			return 0, ErrRuntime
		}
		if len(s.inflates) >= maxInflates {
			return 0, ErrLimit
		}
		stream, valid := newStreamInflater(int32(args[2]))
		if !valid {
			// The reference grows its ID table before inflateInit2 checks
			// windowBits; a failed initialization consumes one slot.
			s.inflates = append(s.inflates, nil)
			return apiFailure, nil
		}
		n := len(s.inflates)
		s.inflates = append(s.inflates, &inflateState{from: fromID, to: toID, stream: stream})
		return uint64(n), nil
	}
	index := int32(args[0])
	if index < 0 || int(index) >= len(s.inflates) || s.inflates[index] == nil || s.inflates[index].closed {
		return apiFailure, nil
	}
	state := s.inflates[index]
	if id == 41 {
		state.closed = true
		state.stream = nil
		return 0, nil
	}
	from, to := s.pipe(uint64(uint32(state.from))), s.pipe(uint64(uint32(state.to)))
	if from == nil || to == nil || from.closed || to.closed {
		return 0, ErrRuntime
	}
	if from == to {
		return 0, ErrRuntime // overlapping zlib input/output is undefined
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
	r, off, err := mem.span(to.ptr+uint64(to.write), int(outAvail), true)
	if err != nil {
		return 0, err
	}
	consumed, produced, status, err := state.stream.processWithRecovery(input, r.data[off:off+int(outAvail)])
	if err != nil {
		return 0, err
	}
	if from.file {
		from.read += uint32(consumed)
	} else {
		from.read += uint32(consumed)
		if from.read >= from.size && from.write >= from.size {
			from.read, from.write = 0, 0
		}
	}
	to.write += uint32(produced)
	if status == inflateStreamEnd {
		state.closed = true // reference calls inflate_done on stream end
		state.stream = nil
	}
	return uint64(uint32(int32(status))), nil
}
