package procwatch

import (
	"encoding/binary"
	"testing"
)

func connectorFixture(what, pid, tgid uint32) []byte {
	b := make([]byte, 76)
	o := binary.NativeEndian
	o.PutUint32(b, uint32(len(b)))
	o.PutUint16(b[4:], 3)
	o.PutUint32(b[16:], 1)
	o.PutUint32(b[20:], 1)
	o.PutUint32(b[24:], 7)
	o.PutUint16(b[32:], 40)
	o.PutUint32(b[36:], what)
	o.PutUint32(b[40:], 2)
	o.PutUint64(b[44:], 1234567)
	o.PutUint32(b[52:], pid)
	o.PutUint32(b[56:], tgid)
	return b
}
func TestConnectorFramingAndProcessRecords(t *testing.T) {
	b := connectorFixture(2, 42, 42)
	events, err := parseConnector(b)
	if err != nil || len(events) != 1 || events[0].PID != 42 || events[0].Kind != connectorExec || events[0].CPU != 2 || events[0].Sequence != 7 || events[0].TimestampNS != 1234567 {
		t.Fatalf("%+v %v", events, err)
	}
	fork := connectorFixture(1, 11, 11)
	binary.NativeEndian.PutUint32(fork[60:], 42)
	binary.NativeEndian.PutUint32(fork[64:], 42)
	events, err = parseConnector(append(b, fork...))
	if err != nil || len(events) != 2 || events[1].PID != 42 || events[1].Parent != 11 {
		t.Fatalf("%+v %v", events, err)
	}
	ack := connectorFixture(0, 0, 0)
	binary.NativeEndian.PutUint32(ack[28:], 123)
	events, err = parseConnector(ack)
	if err != nil || events[0].Ack != 123 || events[0].Kind != connectorAck {
		t.Fatalf("%+v %v", events, err)
	}
}
func TestConnectorRejectsTruncationAndForeignFrames(t *testing.T) {
	good := connectorFixture(2, 42, 42)
	for n := 1; n < len(good); n++ {
		if _, err := parseConnector(good[:n]); err == nil {
			t.Fatalf("accepted truncation %d", n)
		}
	}
	for _, change := range []func([]byte){
		func(b []byte) { binary.NativeEndian.PutUint32(b, 15) },
		func(b []byte) { binary.NativeEndian.PutUint32(b[16:], 3) },
		func(b []byte) { binary.NativeEndian.PutUint16(b[32:], 41) },
		func(b []byte) { binary.NativeEndian.PutUint32(b[12:], 123) },
	} {
		b := append([]byte(nil), good...)
		change(b)
		if _, err := parseConnector(b); err == nil {
			t.Fatal("accepted malformed/foreign frame")
		}
	}
	if _, err := parseConnector(make([]byte, 65537)); err == nil {
		t.Fatal("unbounded datagram")
	}
}
func TestConnectorThreadAndErrorSemantics(t *testing.T) {
	b := connectorFixture(0x80000000, 43, 42)
	events, err := parseConnector(b)
	if err != nil || len(events) != 1 || events[0].PID != 0 {
		t.Fatalf("thread exit must not remove whole process: %+v %v", events, err)
	}
	b = make([]byte, 20)
	binary.NativeEndian.PutUint32(b, 20)
	binary.NativeEndian.PutUint16(b[4:], 2)
	binary.NativeEndian.PutUint32(b[16:], 0xffffffff)
	if _, err := parseConnector(b); err == nil {
		t.Fatal("netlink error ignored")
	}
}
func TestConnectorSubscriptionCookie(t *testing.T) {
	b := connectorSubscription(1234, 77, 1)
	if len(b) != 40 || binary.NativeEndian.Uint32(b[12:]) != 1234 || binary.NativeEndian.Uint32(b[28:]) != 77 || binary.NativeEndian.Uint32(b[36:]) != 1 {
		t.Fatalf("bad subscription %x", b)
	}
}
func FuzzConnectorFraming(f *testing.F) {
	f.Add(connectorFixture(2, 42, 42))
	f.Add([]byte{})
	f.Add([]byte("broken"))
	f.Fuzz(func(t *testing.T, b []byte) { _, _ = parseConnector(b) })
}
