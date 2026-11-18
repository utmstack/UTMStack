package procwatch

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Linux UAPI cn_proc/connector framing. All fields use native byte order;
// the supported Linux architectures are little-endian amd64 and arm64.
const (
	connectorAck         uint32 = 0
	connectorFork        uint32 = 1
	connectorExec        uint32 = 2
	connectorExit        uint32 = 0x80000000
	maxConnectorDatagram        = 65536
)

type connectorEvent struct {
	Kind, CPU, Sequence, Ack uint32
	TimestampNS              uint64
	PID, Parent              int
	Error                    uint32
}

func connectorSubscription(port, cookie, operation uint32) []byte {
	b := make([]byte, 40)
	o := binary.NativeEndian
	o.PutUint32(b, 40)
	o.PutUint16(b[4:], 3)
	o.PutUint32(b[8:], cookie)
	o.PutUint32(b[12:], port)
	o.PutUint32(b[16:], 1)
	o.PutUint32(b[20:], 1)
	o.PutUint32(b[24:], cookie)
	o.PutUint32(b[28:], cookie)
	o.PutUint16(b[32:], 4)
	o.PutUint32(b[36:], operation)
	return b
}

func parseConnector(b []byte) ([]connectorEvent, error) {
	if len(b) > maxConnectorDatagram {
		return nil, errors.New("process connector datagram exceeds limit")
	}
	var events []connectorEvent
	o := binary.NativeEndian
	for len(b) > 0 {
		if len(b) < 16 {
			return nil, errors.New("short netlink header")
		}
		n := int(o.Uint32(b))
		if n < 16 || n > len(b) {
			return nil, errors.New("invalid netlink length")
		}
		frame := b[:n]
		if o.Uint32(frame[12:]) != 0 {
			return nil, errors.New("process connector header is not kernel-origin")
		}
		kind := o.Uint16(frame[4:])
		if kind == 2 {
			if n < 20 {
				return nil, errors.New("short netlink error")
			}
			if code := int32(o.Uint32(frame[16:])); code != 0 {
				return nil, fmt.Errorf("process connector netlink error %d", code)
			}
		} else if kind == 3 {
			if n < 36 || o.Uint32(frame[16:]) != 1 || o.Uint32(frame[20:]) != 1 {
				return nil, errors.New("invalid process connector identity/header")
			}
			count := int(o.Uint16(frame[32:]))
			if count < 20 || count != n-36 {
				return nil, errors.New("invalid process connector payload size")
			}
			data := frame[36:]
			e := connectorEvent{Kind: o.Uint32(data), CPU: o.Uint32(data[4:]), TimestampNS: o.Uint64(data[8:]), Sequence: o.Uint32(frame[24:]), Ack: o.Uint32(frame[28:])}
			switch e.Kind {
			case connectorAck:
				e.Error = o.Uint32(data[16:])
			case connectorFork:
				if count < 32 {
					return nil, errors.New("short process fork")
				}
				parent, tid, pid := int(int32(o.Uint32(data[20:]))), int(int32(o.Uint32(data[24:]))), int(int32(o.Uint32(data[28:])))
				if pid > 0 && tid == pid {
					e.PID = pid
					e.Parent = parent
				}
			case connectorExec, connectorExit:
				if count < 24 {
					return nil, errors.New("short process event")
				}
				tid, pid := int(int32(o.Uint32(data[16:]))), int(int32(o.Uint32(data[20:])))
				if pid > 0 && tid == pid {
					e.PID = pid
				}
			}
			events = append(events, e)
		} else if kind != 1 && kind != 4 {
			return nil, fmt.Errorf("unexpected netlink type %d", kind)
		}
		aligned := (n + 3) &^ 3
		if aligned > len(b) {
			if n != len(b) {
				return nil, errors.New("short netlink padding")
			}
			aligned = n
		}
		b = b[aligned:]
	}
	return events, nil
}
