package watcher

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

const (
	ReasonDataOverwrite = 0x00000001
	ReasonDataExtend    = 0x00000002
	ReasonFileCreate    = 0x00000100
	ReasonRenameNewName = 0x00002000
	ReasonClose         = 0x80000000
)

type USNChange struct {
	USN         int64
	FileRefID   uint64
	ParentRefID uint64
	Reason      uint32
	FileName    string
}

// ParseUSNBuffer decodes concatenated USN_RECORD_V2 records. The OS returns
// a leading 8-byte next-USN value before the records when reading the journal;
// callers pass the buffer AFTER that 8-byte prefix.
func ParseUSNBuffer(buf []byte) ([]USNChange, error) {
	var out []USNChange
	off := 0
	for off+60 <= len(buf) {
		recLen := int(binary.LittleEndian.Uint32(buf[off:]))
		if recLen < 60 || off+recLen > len(buf) {
			break
		}
		major := binary.LittleEndian.Uint16(buf[off+4:])
		if major != 2 {
			off += recLen
			continue // ignore non-v2 records safely
		}
		fileRef := binary.LittleEndian.Uint64(buf[off+8:])
		parentRef := binary.LittleEndian.Uint64(buf[off+16:])
		usn := int64(binary.LittleEndian.Uint64(buf[off+24:]))
		reason := binary.LittleEndian.Uint32(buf[off+40:])
		nameLen := int(binary.LittleEndian.Uint16(buf[off+56:]))
		nameOff := int(binary.LittleEndian.Uint16(buf[off+58:]))
		if off+nameOff+nameLen > len(buf) || nameLen%2 != 0 {
			return out, fmt.Errorf("corrupt USN record at offset %d", off)
		}
		u16 := make([]uint16, nameLen/2)
		for i := range u16 {
			u16[i] = binary.LittleEndian.Uint16(buf[off+nameOff+i*2:])
		}
		out = append(out, USNChange{
			USN:         usn,
			FileRefID:   fileRef,
			ParentRefID: parentRef,
			Reason:      reason,
			FileName:    string(utf16.Decode(u16)),
		})
		off += recLen
	}
	return out, nil
}
