package bytecode

// PDFObject is one object in the reference PDF parser's discovery order.
// Offset points immediately after the "obj" keyword, matching ClamAV's
// pdf_obj.start. Unknown fields remain explicitly unknown rather than zero.
type PDFObject struct {
	ID          uint32
	Offset      uint32
	OffsetKnown bool
	Size        uint32
	SizeKnown   bool
	Raw         []byte
	RawKnown    bool
	Flags       uint32
	FlagsKnown  bool
}

// PDFState is a per-scan snapshot shared by PDF hook runs. The parser owns it;
// set_flags and setobjflags can update it, so it must not be shared by scans.
type PDFState struct {
	Flags              uint32
	FlagsKnown         bool
	Objects            []PDFObject
	ObjectsKnown       bool
	MutatedFlags       bool
	MutatedObjectFlags bool
}

func (s *apiState) pdfObjects() ([]PDFObject, bool) {
	if s.pdfPhase == 4 {
		return nil, true // PRE runs before discovery
	}
	if s.pdf != nil && s.pdf.ObjectsKnown {
		return s.pdf.Objects, true
	}
	return nil, false
}

func (s *apiState) invokePDF(id int, args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	if s.pdfPhase == 0 {
		switch id {
		case 81:
			return 0, nil
		case 82:
			return 0, nil
		case 88:
			return apiFailure, nil
		default:
			return apiFailure, nil
		}
	}
	if id == 88 {
		if s.pdfPhase != 2 {
			return apiFailure, nil
		}
		// In ClamAV 1.5.4, the PDF hook context is zero-initialized and
		// pdf_dumpedid is never assigned by run_pdf_hooks. The advertised
		// current-object index consequently remains zero in POSTDUMP.
		return 0, nil
	}
	if id == 78 {
		if s.pdf != nil {
			if s.pdf.FlagsKnown {
				return uint64(s.pdf.Flags), nil
			}
		} else if s.pdfPhase == 4 && s.pdfFlagsKnown {
			return uint64(s.pdfFlags), nil
		}
		return 0, ErrPDFObjects
	}
	if id == 79 {
		if s.pdf == nil {
			return 0, ErrPDFObjects
		}
		s.pdf.Flags, s.pdf.FlagsKnown, s.pdf.MutatedFlags = uint32(args[0]), true, true
		return 0, nil
	}
	if id == 81 && s.pdfPhase == 2 {
		return 0, nil // POSTDUMP maps the extracted object, not the PDF
	}
	if id == 82 && s.pdfPhase == 2 && args[1] != 0 {
		return 0, nil // pdf_getobjsize is zero in POSTDUMP
	}
	objects, known := s.pdfObjects()
	if !known {
		return 0, ErrPDFObjects
	}
	if id == 77 {
		return uint64(len(objects)), nil
	}
	if id == 80 {
		for i := range objects {
			if objects[i].ID == uint32(args[0]) {
				return uint64(i), nil
			}
		}
		return apiFailure, nil
	}
	index := uint32(args[0])
	if uint64(index) != args[0] || uint64(index) >= uint64(len(objects)) {
		if id == 81 || id == 82 {
			return 0, nil
		}
		return apiFailure, nil
	}
	obj := &objects[index]
	switch id {
	case 81:
		if !obj.SizeKnown {
			return 0, ErrPDFObjects
		}
		return uint64(obj.Size), nil
	case 82:
		if !obj.SizeKnown {
			return 0, ErrPDFObjects
		}
		amount := uint32(args[1])
		if uint64(amount) != args[1] || amount > obj.Size {
			return 0, nil
		}
		if s.pdfPhase == 2 || !obj.RawKnown || uint64(amount) > uint64(len(obj.Raw)) {
			return 0, ErrPDFObjects
		}
		if mem == nil {
			return 0, ErrRuntime
		}
		if err := bytesWork(uint64(amount)); err != nil {
			return 0, err
		}
		ptr, err := mem.alloc(int(amount), false)
		if err != nil {
			return 0, err
		}
		if err = mem.write(ptr, obj.Raw[:amount]); err != nil {
			return 0, err
		}
		mem.protect(ptr)
		return ptr, nil
	case 83:
		return uint64(obj.ID), nil
	case 84:
		if !obj.FlagsKnown {
			return 0, ErrPDFObjects
		}
		return uint64(obj.Flags), nil
	case 85:
		if s.pdf == nil {
			return 0, ErrPDFObjects
		}
		obj.Flags, obj.FlagsKnown, s.pdf.MutatedObjectFlags = uint32(args[1]), true, true
		return 0, nil
	case 86:
		if !obj.OffsetKnown {
			return 0, ErrPDFObjects
		}
		return uint64(obj.Offset), nil
	}
	return 0, ErrUnsupported
}
