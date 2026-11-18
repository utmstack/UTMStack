package bytecode

import "github.com/utmstack/UTMStack/agent/edr/jsnorm"

type jsNormalizerState struct {
	from   int32
	stream *jsnorm.Stream
	closed bool
}

// jsnormAPI implements the ClamBC pipe-to-JavaScript normalizer lifecycle.
// Completed files are scanned only after the bytecode run ends, as upstream
// does for the temporary jsnorm directory.
func (s *apiState) jsnormAPI(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 43 {
		from := s.pipe(args[0])
		if from == nil {
			return apiFailure, nil
		}
		if len(s.jsnorms) >= maxJSNorms {
			return 0, ErrLimit
		}
		n := len(s.jsnorms)
		s.jsnorms = append(s.jsnorms, &jsNormalizerState{from: int32(args[0]), stream: jsnorm.NewStream()})
		return uint64(n), nil
	}
	idx := int32(args[0])
	if idx < 0 || int(idx) >= len(s.jsnorms) || s.jsnorms[idx] == nil || s.jsnorms[idx].closed {
		return apiFailure, nil
	}
	state := s.jsnorms[idx]
	if id == 45 {
		output, unknown := state.stream.Finish()
		if len(output) > maxAPIOutput-s.jsOutput {
			return 0, ErrLimit
		}
		if err := bytesWork(uint64(len(output))); err != nil {
			return 0, err
		}
		s.jsOutput += len(output)
		s.jsFiles = append(s.jsFiles, output)
		s.jsUnknown = s.jsUnknown || unknown
		state.closed = true
		state.stream = nil
		return 0, nil
	}
	from := s.pipe(uint64(uint32(state.from)))
	if from == nil || from.closed {
		return apiFailure, nil
	}
	available := s.pipeReadAvail(from)
	if available == 0 {
		return apiFailure, nil
	}
	if uint64(available) > uint64(maxAPIOutput-s.jsInput) {
		return 0, ErrLimit
	}
	if err := bytesWork(uint64(available)); err != nil {
		return 0, err
	}
	var data []byte
	if from.file {
		data = s.input[from.read : from.read+available]
	} else {
		if mem == nil {
			return 0, ErrRuntime
		}
		var err error
		data, err = mem.read(from.ptr+uint64(from.read), int(available))
		if err != nil {
			return 0, err
		}
	}
	state.stream.Feed(data)
	s.jsInput += int(available)
	from.read += available
	if !from.file && from.read >= from.size && from.write >= from.size {
		from.read, from.write = 0, 0
	}
	return 0, nil
}

func (s *apiState) scanJSFiles() error {
	if len(s.jsFiles) == 0 {
		return nil
	}
	if !s.scanning || s.scanExtracted == nil && s.scanExtractedTyped == nil {
		return ErrExtraction
	}
	for _, data := range s.jsFiles {
		name, err := s.scanChild(data, -1, 0)
		if name != "" && s.childName == "" {
			s.childName = name
		}
		if err != nil {
			return err
		}
	}
	s.jsFiles = nil
	if s.jsUnknown {
		return ErrUnsupported
	}
	return nil
}
