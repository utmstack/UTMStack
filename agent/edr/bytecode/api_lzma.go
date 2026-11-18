package bytecode

type lzmaState struct {
	from, to int32
	header   [13]byte
	dict     uint64
	models   uint64
	stream   *streamLZMA
	invalid  bool
	closed   bool
}

// lzmaAPI connects the Go streaming decoder to ClamBC's LZMA-alone pipe ABI.
// Dictionary and probability storage have separate shared reservation ceilings.
func (s *apiState) lzmaAPI(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 102 {
		from, to := s.pipe(args[0]), s.pipe(args[1])
		if from == nil || to == nil {
			return apiFailure, nil
		}
		if from.closed || to.closed || from == to {
			return 0, ErrRuntime
		}
		avail := s.pipeReadAvail(from)
		if avail < 13 {
			return apiFailure, nil
		}
		if mem == nil {
			return 0, ErrRuntime
		}
		if len(s.lzmas) >= maxLZMAs {
			return 0, ErrLimit
		}
		if err := bytesWork(13); err != nil {
			return 0, err
		}
		state := &lzmaState{from: int32(args[0]), to: int32(args[1])}
		if from.file {
			copy(state.header[:], s.input[from.read:from.read+13])
		} else {
			b, err := mem.read(from.ptr+uint64(from.read), 13)
			if err != nil {
				return 0, err
			}
			copy(state.header[:], b)
		}
		// ClamAV consumes the header even when the LZMA properties fail to
		// initialize. Retain the allocated ID so a later done has its slot.
		dict, models, valid := lzmaAllocation(state.header[:])
		state.invalid = !valid
		if valid {
			if dict > maxLZMADict-s.lzmaDict || models > maxLZMADict-s.lzmaModels {
				return 0, ErrLimit
			}
			var err error
			state.stream, err = newStreamLZMA(state.header[:])
			if err != nil {
				return 0, err
			}
			state.dict, state.models = dict, models
			s.lzmaDict += dict
			s.lzmaModels += models
		}
		n := len(s.lzmas)
		s.lzmas = append(s.lzmas, state)
		from.read += 13
		if !from.file && from.read >= from.size && from.write >= from.size {
			from.read, from.write = 0, 0
		}
		if state.invalid {
			return 1, nil // LZMA_RESULT_DATA_ERROR
		}
		return uint64(n), nil
	}
	index := int32(args[0])
	if index < 0 || int(index) >= len(s.lzmas) || s.lzmas[index] == nil || s.lzmas[index].closed {
		return apiFailure, nil
	}
	state := s.lzmas[index]
	if id == 104 {
		s.closeLZMA(state)
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
	if state.invalid {
		s.closeLZMA(state)
		return 1, nil
	}
	var input []byte
	if from.file {
		input = s.input[from.read : from.read+inAvail]
	} else {
		var err error
		var region *memoryRegion
		var offset int
		region, offset, err = mem.span(from.ptr+uint64(from.read), int(inAvail), false)
		if err != nil {
			return 0, err
		}
		input = region.data[offset : offset+int(inAvail)]
	}
	region, offset, err := mem.span(to.ptr+uint64(to.write), int(outAvail), true)
	if err != nil {
		return 0, err
	}
	n, m, status, err := state.stream.process(input, region.data[offset:offset+int(outAvail)], bytesWork)
	if err != nil {
		return 0, err
	}
	if err := bytesWork(uint64(n) + uint64(m)); err != nil {
		return 0, err
	}
	from.read += uint32(n)
	if !from.file && from.read >= from.size && from.write >= from.size {
		from.read, from.write = 0, 0
	}
	to.write += uint32(m)
	// The native bytecode adapter closes a failed decoder after publishing
	// its observed cursors. Stream end itself leaves the handle valid.
	if status == 1 {
		s.closeLZMA(state)
	}
	return uint64(status), nil
}

func (s *apiState) closeLZMA(state *lzmaState) {
	state.closed = true
	state.stream = nil
	s.lzmaDict -= state.dict
	s.lzmaModels -= state.models
	state.dict, state.models = 0, 0
}
