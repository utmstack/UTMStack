package watcher

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

// buildRecord encodes one USN_RECORD_V2 with the given filename and reason.
func buildRecord(name string, reason uint32) []byte {
	nameU16 := utf16.Encode([]rune(name))
	nameBytes := make([]byte, len(nameU16)*2)
	for i, c := range nameU16 {
		binary.LittleEndian.PutUint16(nameBytes[i*2:], c)
	}
	const header = 60 // bytes before FileName in USN_RECORD_V2
	recLen := header + len(nameBytes)
	// pad record length to 8-byte alignment (as the FS does)
	for recLen%8 != 0 {
		recLen++
	}
	b := make([]byte, recLen)
	binary.LittleEndian.PutUint32(b[0:], uint32(recLen))          // RecordLength
	binary.LittleEndian.PutUint16(b[4:], 2)                       // MajorVersion
	binary.LittleEndian.PutUint16(b[6:], 0)                       // MinorVersion
	binary.LittleEndian.PutUint64(b[8:], 0x1111)                  // FileReferenceNumber
	binary.LittleEndian.PutUint64(b[16:], 0x2222)                 // ParentFileReferenceNumber
	binary.LittleEndian.PutUint64(b[24:], 42)                     // Usn
	binary.LittleEndian.PutUint64(b[32:], 0)                      // TimeStamp
	binary.LittleEndian.PutUint32(b[40:], reason)                 // Reason
	binary.LittleEndian.PutUint32(b[44:], 0)                      // SourceInfo
	binary.LittleEndian.PutUint32(b[48:], 0)                      // SecurityId
	binary.LittleEndian.PutUint32(b[52:], 0)                      // FileAttributes
	binary.LittleEndian.PutUint16(b[56:], uint16(len(nameBytes))) // FileNameLength
	binary.LittleEndian.PutUint16(b[58:], header)                 // FileNameOffset
	copy(b[header:], nameBytes)
	return b
}

func TestParseUSNBuffer(t *testing.T) {
	buf := append(buildRecord("a.exe", ReasonFileCreate), buildRecord("note.txt", ReasonDataExtend)...)
	changes, err := ParseUSNBuffer(buf)
	if err != nil {
		t.Fatalf("ParseUSNBuffer: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("got %d changes, want 2", len(changes))
	}
	if changes[0].FileName != "a.exe" || changes[0].Reason != ReasonFileCreate || changes[0].USN != 42 {
		t.Fatalf("bad first change: %+v", changes[0])
	}
	if changes[1].FileName != "note.txt" || changes[1].ParentRefID != 0x2222 {
		t.Fatalf("bad second change: %+v", changes[1])
	}
}
