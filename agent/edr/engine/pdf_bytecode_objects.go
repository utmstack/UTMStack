package engine

import (
	"bytes"
	"strconv"
)

type pdfReferenceObject struct {
	id    uint32
	start int
	end   int
}

// pdfReferenceObjects reconstructs the discovery order, IDs and starts of
// ordinary indirect objects for PDF bytecode hooks. ClamAV 1.5.4's
// pdf_findobj scans raw bytes for "obj", then resumes after the first
// following "endobj". This can count a header inside a comment, while a
// header inside an earlier object's string or stream is swallowed by that
// object's end search. Object flags stay unknown in this fallback.
// Source: https://github.com/Cisco-Talos/clamav/blob/clamav-1.5.4/libclamav/pdf.c
func pdfReferenceObjects(data []byte) ([]pdfReferenceObject, bool) {
	const maxObjects = 64 * 1024
	var objects []pdfReferenceObject
	for cursor := 0; cursor < len(data); {
		at := bytes.Index(data[cursor:], []byte("obj"))
		if at < 0 {
			return objects, true
		}
		at += cursor
		cursor = at + len("obj")
		if at == 0 || !pdfWhite(data[at-1]) {
			continue
		}
		// The tagged parser reads two decimal tokens backwards from the
		// whitespace immediately before "obj". Require exact bounded tokens
		// here; malformed or ambiguous syntax must not become a known ID.
		i := at - 1
		for i >= 0 && pdfWhite(data[i]) {
			i--
		}
		genEnd := i + 1
		for i >= 0 && data[i] >= '0' && data[i] <= '9' {
			i--
		}
		genStart := i + 1
		if genStart == genEnd || i < 0 || !pdfWhite(data[i]) {
			return nil, false
		}
		for i >= 0 && pdfWhite(data[i]) {
			i--
		}
		objEnd := i + 1
		for i >= 0 && data[i] >= '0' && data[i] <= '9' {
			i--
		}
		objStart := i + 1
		if objStart == objEnd || i >= 0 && !pdfWhite(data[i]) {
			return nil, false
		}
		object, err := strconv.ParseUint(string(data[objStart:objEnd]), 10, 24)
		if err != nil {
			return nil, false
		}
		generation, err := strconv.ParseUint(string(data[genStart:genEnd]), 10, 8)
		if err != nil {
			return nil, false
		}
		end := bytes.Index(data[cursor:], []byte("endobj"))
		if end < 0 || len(objects) >= maxObjects {
			return nil, false
		}
		objects = append(objects, pdfReferenceObject{id: uint32(object<<8 | generation), start: cursor, end: cursor + end})
		cursor += end + len("endobj")
	}
	return objects, true
}
