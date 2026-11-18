package bytecode

import (
	"math"
	"runtime"
)

// ClamAV 1.5.4 bytecode_api.c uses a fixed buffer, not a sliding queue.
// A consumed buffer resets only after both cursors reach its end. A file pipe
// instead reads directly from the scanned object's immutable input snapshot.
type bufferPipe struct {
	ptr          uint64
	size         uint32
	read, write  uint32
	file, closed bool
}

func pipeFileWindow() uint32 {
	// The reference uses C BUFSIZ, which is a platform C-library constant.
	// These are the native SDK values for our supported desktop targets.
	switch runtime.GOOS {
	case "windows":
		return 512
	case "darwin":
		return 1024
	default:
		return 8192
	}
}

func (s *apiState) pipe(id uint64) *bufferPipe {
	i := int32(id)
	if i < 0 || int(i) >= len(s.pipes) {
		return nil
	}
	return &s.pipes[i]
}

func (s *apiState) pipeReadAvail(p *bufferPipe) uint32 {
	if p.file {
		if uint64(p.read) >= uint64(len(s.input)) {
			return 0
		}
		return min(pipeFileWindow(), uint32(len(s.input))-p.read)
	}
	if p.write <= p.read {
		return 0
	}
	return p.write - p.read
}

func (s *apiState) pipeWriteAvail(p *bufferPipe) uint32 {
	if p.file || p.write >= p.size {
		return 0
	}
	return p.size - p.write
}

func (s *apiState) pipeAPI(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if id == 30 || id == 31 {
		if len(s.pipes) >= maxPipes {
			return 0, ErrLimit
		}
		value := uint32(args[0])
		p := bufferPipe{}
		if id == 31 {
			if uint64(value) >= uint64(len(s.input)) {
				return apiFailure, nil
			}
			p.file, p.read = true, value
		} else {
			if value == 0 || value > maxAllocation {
				return apiFailure, nil
			}
			if mem == nil {
				return 0, ErrRuntime
			}
			if err := bytesWork(uint64(value)); err != nil {
				return 0, err
			}
			ptr, err := mem.alloc(int(value), false)
			if err != nil {
				return 0, err
			}
			if err := mem.fill(ptr, 0, int(value)); err != nil {
				mem.release(ptr)
				return 0, err
			}
			p.ptr, p.size = ptr, value
		}
		n := len(s.pipes)
		s.pipes = append(s.pipes, p)
		return uint64(n), nil
	}
	p := s.pipe(args[0])
	if p == nil {
		switch id {
		case 32, 35, 33, 36:
			return 0, nil
		default:
			return apiFailure, nil
		}
	}
	// Upstream retains a slot after done but has ambiguous use-after-free
	// behavior: a released memory pipe may be mistaken for a file pipe.
	if p.closed {
		return 0, ErrRuntime
	}
	switch id {
	case 32:
		return uint64(s.pipeReadAvail(p)), nil
	case 33:
		size := uint32(args[1])
		if size == 0 || size > s.pipeReadAvail(p) {
			return 0, nil
		}
		if err := bytesWork(uint64(size)); err != nil {
			return 0, err
		}
		if p.file {
			if mem == nil {
				return 0, ErrRuntime
			}
			if s.filePtr == 0 {
				ptr, err := mem.borrowReadOnly(s.input)
				if err != nil {
					return 0, err
				}
				s.filePtr = ptr
			}
			return s.filePtr + uint64(p.read), nil
		}
		return p.ptr + uint64(p.read), nil
	case 34:
		amount := uint32(args[1])
		if p.file {
			p.read += amount // the reference uses a uint32 cursor
			return 0, nil
		}
		if p.write <= p.read {
			return apiFailure, nil
		}
		if uint64(p.read)+uint64(amount) > math.MaxUint32 {
			return 0, ErrRuntime
		}
		p.read += min(amount, p.write-p.read)
		if p.read >= p.size && p.write >= p.size {
			p.read, p.write = 0, 0
		}
		return 0, nil
	case 35:
		return uint64(s.pipeWriteAvail(p)), nil
	case 36:
		size := uint32(args[1])
		if size == 0 || size > s.pipeWriteAvail(p) {
			return 0, nil
		}
		return p.ptr + uint64(p.write), nil
	case 37:
		if p.file {
			return apiFailure, nil
		}
		amount := uint32(args[1])
		if uint64(p.write)+uint64(amount) > math.MaxUint32 {
			return 0, ErrRuntime
		}
		p.write += min(amount, p.size-p.write)
		return 0, nil
	case 38:
		p.closed = true
		if !p.file && mem != nil {
			mem.release(p.ptr)
		}
		return 0, nil
	default:
		return 0, ErrUnsupported
	}
}
