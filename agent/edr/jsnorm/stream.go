package jsnorm

// Stream feeds one JavaScript source to the same incremental tokenizer used
// for HTML script blocks. The caller owns input and output size limits.
type Stream struct {
	normalizer  *jsNormalizer
	finished    bool
	output      []byte
	unsupported bool
}

func NewStream() *Stream { return &Stream{normalizer: newJSNormalizer()} }

func (s *Stream) Feed(data []byte) {
	if s != nil && !s.finished {
		s.normalizer.feed(data)
	}
}

// Finish returns the normalized <script> file and whether a transformation
// was not modeled accurately enough to support a clean scan verdict.
func (s *Stream) Finish() ([]byte, bool) {
	if s == nil {
		return nil, true
	}
	if !s.finished {
		j := s.normalizer
		j.finish()
		s.unsupported = j.decoderIncomplete || j.packerIncomplete
		s.output = j.output()
		s.normalizer = nil
		s.finished = true
	}
	return s.output, s.unsupported
}
