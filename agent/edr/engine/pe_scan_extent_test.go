package engine

import (
	"encoding/binary"
	"errors"
	"testing"
)

func TestPEScanClipsPhysicalTailWithoutRelaxingTrustView(t *testing.T) {
	for _, size := range []int{513, 640, 767, 1023, 1024} {
		data := peValidationFixture()[:size]
		img, err := parsePEScanImage(data)
		if err != nil || img.overlay != size || len(img.sections) != 1 || img.sections[0].raw != 512 || img.sections[0].rawEnd != uint64(size) {
			t.Fatalf("size=%d scan=%+v err=%v", size, img, err)
		}
		_, strictErr := parsePEImage(data)
		if size < 1024 && !errors.Is(strictErr, ErrMalformed) || size == 1024 && strictErr != nil {
			t.Fatalf("size=%d strict view changed: %v", size, strictErr)
		}
	}
	for _, offset := range []uint32{1024, 1025, ^uint32(0)} {
		data := peValidationFixture()
		binary.LittleEndian.PutUint32(data[0x178+20:], offset)
		if _, err := parsePEScanImage(data); !errors.Is(err, ErrMalformed) {
			t.Fatalf("unmapped start accepted: offset=%d %v", offset, err)
		}
	}
	for _, security := range [][2]uint32{{640, 16}, {640, 0}, {0, 16}} {
		data := peValidationFixture()[:768]
		binary.LittleEndian.PutUint32(data[0x98+128:], security[0])
		binary.LittleEndian.PutUint32(data[0x98+132:], security[1])
		if _, err := parsePEScanImage(data); !errors.Is(err, ErrMalformed) {
			t.Fatalf("incomplete certificate-bearing layout entered unsigned view: %v", err)
		}
	}
}
