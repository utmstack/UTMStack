package watcher

import (
	"encoding/binary"
	"testing"
)

func TestLinuxNotificationBoundaries(t *testing.T) {
	fixture := func(name string) []byte {
		b := make([]byte, 16)
		binary.NativeEndian.PutUint32(b, 42)
		binary.NativeEndian.PutUint32(b[4:], 8)
		binary.NativeEndian.PutUint32(b[12:], uint32(len(name)))
		return append(b, []byte(name)...)
	}
	e, err := decodeLinuxEvents(fixture("inert.txt\x00\x00"))
	if err != nil || len(e) != 1 || e[0].name != "inert.txt" || e[0].wd != 42 {
		t.Fatalf("%+v %v", e, err)
	}
	for _, b := range [][]byte{make([]byte, 15), fixture("../escape\x00"), fixture("..\x00"), fixture("a\x00b"), fixture("cut")[:17]} {
		if _, err := decodeLinuxEvents(b); err == nil {
			t.Fatalf("accepted malformed event %x", b)
		}
	}
	for _, path := range []string{"/proc", "/proc/self", "/sys/kernel", "/dev/null", "/run/user"} {
		if !linuxPseudoPath(path) {
			t.Fatal(path)
		}
	}
	for _, path := range []string{"/home/proc", "/process", "/device", "/runtime"} {
		if linuxPseudoPath(path) {
			t.Fatal(path)
		}
	}
}
