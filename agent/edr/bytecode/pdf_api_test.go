package bytecode

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestPDFObjectAPIsAndMutations(t *testing.T) {
	state := &PDFState{
		Flags: 5, FlagsKnown: true, ObjectsKnown: true,
		Objects: []PDFObject{
			{ID: 0x100, Offset: 17, OffsetKnown: true, Size: 8, SizeKnown: true, Raw: []byte("abcdefgh"), RawKnown: true, Flags: 3, FlagsKnown: true},
			{ID: 0x201, Offset: 56, OffsetKnown: true, Size: 4, SizeKnown: true, Raw: []byte("WXYZ"), RawKnown: true, Flags: 0, FlagsKnown: true},
		},
	}
	api := &apiState{pdfPhase: 1, pdf: state}
	apiTestCall(t, api, nil, 77, 2)
	apiTestCall(t, api, nil, 78, 5)
	apiTestCall(t, api, nil, 80, 1, 0x201)
	apiTestCall(t, api, nil, 80, apiFailure, 0x300)
	apiTestCall(t, api, nil, 81, 8, 0)
	apiTestCall(t, api, nil, 81, 0, 9)
	apiTestCall(t, api, nil, 83, 0x100, 0)
	apiTestCall(t, api, nil, 83, 0x201, 1)
	apiTestCall(t, api, nil, 83, apiFailure, 2)
	apiTestCall(t, api, nil, 84, 3, 0)
	apiTestCall(t, api, nil, 86, 56, 1)
	apiTestCall(t, api, nil, 88, apiFailure)
	apiTestCall(t, api, nil, 79, 0, 9)
	apiTestCall(t, api, nil, 78, 9)
	apiTestCall(t, api, nil, 85, 0, 1, 7)
	apiTestCall(t, api, nil, 84, 7, 1)
	if !state.MutatedFlags || !state.MutatedObjectFlags {
		t.Fatal("PDF hook mutation was not retained")
	}
	var mem memory
	ptr, err := api.invoke(context.Background(), 82, []uint64{0, 3}, &mem, func(uint64) error { return nil })
	if err != nil || ptr == 0 {
		t.Fatalf("pdf_getobj = %#x, %v", ptr, err)
	}
	b, err := mem.read(ptr, 3)
	if err != nil || !bytes.Equal(b, []byte("abc")) {
		t.Fatalf("pdf_getobj bytes = %q, %v", b, err)
	}
	if err := mem.write(ptr, []byte("x")); !errors.Is(err, ErrRuntime) {
		t.Fatalf("pdf_getobj memory must be read-only: %v", err)
	}
	apiTestCall(t, api, &mem, 82, 0, 0, 9)

	api.pdfPhase = 2
	apiTestCall(t, api, nil, 81, 0, 0)
	apiTestCall(t, api, nil, 82, 0, 0, 1)
	apiTestCall(t, api, nil, 88, 0)
	api.pdfPhase = 4
	apiTestCall(t, api, nil, 77, 0)
	apiTestCall(t, api, nil, 80, apiFailure, 0x100)
}

func TestPDFUnknownFieldsDoNotBecomeZero(t *testing.T) {
	state := &PDFState{ObjectsKnown: true, Objects: []PDFObject{{ID: 0x100}}}
	api := &apiState{pdfPhase: 1, pdf: state}
	for _, tc := range []struct {
		id   int
		args []uint64
	}{
		{78, nil}, {81, []uint64{0}}, {82, []uint64{0, 1}}, {84, []uint64{0}}, {86, []uint64{0}},
	} {
		if _, err := api.invoke(context.Background(), tc.id, tc.args, nil, func(uint64) error { return nil }); !errors.Is(err, ErrPDFObjects) {
			t.Fatalf("API %d: %v; want unknown PDF field", tc.id, err)
		}
	}
	api.pdfPhase = 2
	apiTestCall(t, api, nil, 88, 0)
}
