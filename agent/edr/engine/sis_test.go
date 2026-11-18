package engine

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"
)

type sisFixtureFile struct {
	kind     uint32 // record type: 0 simple, 1 per language
	fileType uint32
	data     [][]byte // one per language for kind 1
}

// sisPackage writes an EPOC package: header, languages, then records of the
// given files between an options record and an if/else/end-if block, with
// every file's data appended after the records.
func sisPackage(release6, compress bool, languages int, files []sisFixtureFile) []byte {
	le := binary.LittleEndian
	header := make([]byte, 100)
	le.PutUint32(header[4:], 0x1000006d)
	if release6 {
		le.PutUint32(header[4:], sisRelease6)
	}
	le.PutUint32(header[8:], 0x10000419)
	le.PutUint16(header[18:], uint16(languages))
	le.PutUint16(header[20:], uint16(len(files)+4))
	if !compress {
		le.PutUint16(header[36:], sisNoCompress)
	}
	le.PutUint32(header[48:], 100)
	le.PutUint32(header[52:], uint32(100+2*languages))
	b := append(header, make([]byte, 2*languages)...)
	var records [][]uint32
	var blobs [][]byte
	records = append(records, []uint32{2, 1}) // one option: names per language, then selection
	for i := 0; i < languages; i++ {
		records[0] = append(records[0], 0, 0)
	}
	records[0] = append(records[0], 0, 0, 0, 0)
	records = append(records, []uint32{3, 4, 0}) // if, with a four-byte expression
	for _, f := range files {
		r := []uint32{f.kind, f.fileType, 0, 0, 0, 0, 0}
		var lengths, pointers, originals []uint32
		for _, d := range f.data {
			stored := d
			if compress && d != nil {
				var z bytes.Buffer
				w := zlib.NewWriter(&z)
				w.Write(d)
				w.Close()
				stored = z.Bytes()
			}
			lengths = append(lengths, uint32(len(stored)))
			pointers = append(pointers, uint32(len(blobs))) // patched below
			originals = append(originals, uint32(len(d)))
			blobs = append(blobs, stored)
		}
		r = append(append(r, lengths...), pointers...)
		if release6 {
			r = append(append(r, originals...), 0, 0)
		}
		records = append(records, r)
	}
	records = append(records, []uint32{5}, []uint32{6})
	size := 0
	for _, r := range records {
		size += 4 * len(r)
	}
	dataStart := len(b) + size
	offsets := make([]uint32, len(blobs))
	for i, blob := range blobs {
		offsets[i] = uint32(dataStart)
		dataStart += len(blob)
	}
	for ri, r := range records[2 : len(records)-2] {
		f := files[ri]
		for e := range f.data {
			r[7+len(f.data)+e] = offsets[r[7+len(f.data)+e]]
		}
	}
	for _, r := range records {
		for _, v := range r {
			b = le.AppendUint32(b, v)
		}
	}
	for _, blob := range blobs {
		b = append(b, blob...)
	}
	return b
}

func TestSISPackages(t *testing.T) {
	gz := gzipFixture(t, []byte(payloadMarker))
	exact := []byte("the installed file, exactly")
	for name, c := range map[string]struct {
		release6, compress bool
	}{"release 6": {true, true}, "release 5": {false, true}, "uncompressed": {true, false}} {
		t.Run(name, func(t *testing.T) {
			files := []sisFixtureFile{
				{kind: 0, fileType: 4, data: [][]byte{nil}}, // created on installation
				{kind: 0, data: [][]byte{[]byte("readme text")}},
				{kind: 1, data: [][]byte{[]byte("english"), gz}},
			}
			b := sisPackage(c.release6, c.compress, 2, files)
			if c.compress {
				requireHidden(t, b)
			}
			expectDetected(t, payloadState(t), b, "Payload.Test")
			b = sisPackage(c.release6, c.compress, 1, []sisFixtureFile{{kind: 0, data: [][]byte{exact}}})
			expectDetected(t, payloadState(t, exact), b, "Payload.Hash0")
		})
	}
}

func TestSISHostileInput(t *testing.T) {
	base := func() []byte {
		return sisPackage(true, true, 1, []sisFixtureFile{{kind: 0, data: [][]byte{[]byte("installed file data")}}})
	}
	le := binary.LittleEndian
	record := 102 + 4*(2+1*2+4) + 4*3 // header, languages, options and if records
	cases := map[string]struct {
		mutate func([]byte) []byte
		want   error
	}{
		"file past the end":    {func(b []byte) []byte { le.PutUint32(b[record+4*8:], 1<<30); return b }, ErrMalformed},
		"records past the end": {func(b []byte) []byte { le.PutUint32(b[52:], uint32(len(b)-4)); return b }, ErrMalformed},
		"no languages":         {func(b []byte) []byte { le.PutUint16(b[18:], 0); return b }, ErrMalformed},
		"unknown record":       {func(b []byte) []byte { le.PutUint32(b[record:], 9); return b }, ErrMalformed},
		"length mismatch":      {func(b []byte) []byte { le.PutUint32(b[record+4*9:], 5); return b }, ErrMalformed},
		"corrupt data":         {func(b []byte) []byte { b[len(b)-3] ^= 0xff; return b }, ErrMalformed},
		"huge original size":   {func(b []byte) []byte { le.PutUint32(b[record+4*9:], 0xffffffff); return b }, ErrLimit},
		"too many records":     {func(b []byte) []byte { le.PutUint16(b[20:], 0xffff); return b }, ErrLimit},
		"Symbian 9":            {func(b []byte) []byte { le.PutUint32(b, sisX); return b }, ErrUnsupported},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := payloadState(t)
			expectIncomplete(t, s, c.mutate(base()), c.want)
		})
	}
}
