package syslog

import "testing"

func TestUDPBufferSizeSupportsFullDatagram(t *testing.T) {
	if UDPBufferSize < 65535 {
		t.Fatalf("UDPBufferSize = %d, want >= 65535 to avoid wsarecvfrom truncation", UDPBufferSize)
	}
}
