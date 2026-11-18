package watcher

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type linuxEvent struct {
	wd   int32
	mask uint32
	name string
}

func decodeLinuxEvents(data []byte) ([]linuxEvent, error) {
	var events []linuxEvent
	for len(data) > 0 {
		if len(data) < 16 {
			return nil, errors.New("short file notification header")
		}
		n := uint64(binary.NativeEndian.Uint32(data[12:16]))
		if n > uint64(len(data)-16) || n > 4096 {
			return nil, errors.New("invalid file notification name size")
		}
		name := bytes.TrimRight(data[16:16+int(n)], "\x00")
		if bytes.IndexByte(name, 0) >= 0 || bytes.IndexByte(name, '/') >= 0 || string(name) == "." || string(name) == ".." {
			return nil, errors.New("invalid file notification name")
		}
		events = append(events, linuxEvent{int32(binary.NativeEndian.Uint32(data[:4])), binary.NativeEndian.Uint32(data[4:8]), string(name)})
		data = data[16+int(n):]
	}
	return events, nil
}

func linuxPseudoPath(path string) bool {
	// These are Linux event paths even when the portable parser is tested on
	// another host. A host-native separator would misclassify their children.
	for _, root := range []string{"/proc", "/sys", "/dev", "/run"} {
		if path == root || len(path) > len(root) && path[:len(root)+1] == root+"/" {
			return true
		}
	}
	return false
}
