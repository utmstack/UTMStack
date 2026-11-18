package fanotify

import (
	"encoding/binary"
	"errors"
)

// Decode only interprets kernel metadata; it never operates on descriptor
// numbers, so malformed-frame tests and fuzzing cannot touch process handles.
func decodeEvents(data []byte) (events []rawEvent, err error) {
	for len(data) > 0 {
		if len(events) >= 256 || len(data) < 24 {
			return events, errors.New("fanotify truncated or oversized metadata batch")
		}
		length := int(binary.NativeEndian.Uint32(data))
		metadata := int(binary.NativeEndian.Uint16(data[6:]))
		if length < 24 || length > len(data) || metadata < 24 || metadata > length {
			return events, errors.New("fanotify invalid metadata lengths")
		}
		raw := rawEvent{fd: int(int32(binary.NativeEndian.Uint32(data[16:]))), pid: int(int32(binary.NativeEndian.Uint32(data[20:]))), pidfd: -1, mask: binary.NativeEndian.Uint64(data[8:])}
		events = append(events, raw)
		current := &events[len(events)-1]
		if data[4] != 3 {
			return events, errors.New("fanotify metadata version mismatch")
		}
		for pos := metadata; pos < length; {
			if pos+4 > length {
				return events, errors.New("fanotify truncated information header")
			}
			infoLen := int(binary.NativeEndian.Uint16(data[pos+2:]))
			if infoLen < 4 || infoLen > length-pos {
				return events, errors.New("fanotify invalid information length")
			}
			if data[pos] == 4 {
				if infoLen != 8 {
					return events, errors.New("fanotify invalid pidfd record")
				}
				fd := int(int32(binary.NativeEndian.Uint32(data[pos+4:])))
				current.pidfds = append(current.pidfds, fd)
				if len(current.pidfds) != 1 {
					return events, errors.New("fanotify duplicate pidfd record")
				}
				current.pidfd = fd
			}
			pos += infoLen
		}
		data = data[length:]
	}
	return events, nil
}
