//go:build windows

package amsi

import (
	"context"
	"encoding/binary"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

const PipeName = `\\.\pipe\utmstack_edr_amsi`

type PipeServer struct{ s *Scanner }

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{s: s} }

// Run accepts connections from the native AMSI provider DLL, one at a time.
// Protocol: request = 4-byte LE length + "appName\0" + content; reply = 1 byte
// (0 = allow, 1 = block).
func (p *PipeServer) Run(ctx context.Context) {
	name, err := windows.UTF16PtrFromString(PipeName)
	if err != nil {
		logger.Error("UTMStack EDR AMSI: bad pipe name: %v", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		h, err := windows.CreateNamedPipe(name,
			windows.PIPE_ACCESS_DUPLEX,
			windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT,
			windows.PIPE_UNLIMITED_INSTANCES,
			65536, 65536, 0, nil)
		if err != nil {
			logger.Error("UTMStack EDR AMSI: create pipe: %v", err)
			return
		}
		err = windows.ConnectNamedPipe(h, nil)
		if err != nil && err != windows.ERROR_PIPE_CONNECTED {
			windows.CloseHandle(h)
			continue
		}
		p.handle(h)
		windows.FlushFileBuffers(h)
		windows.DisconnectNamedPipe(h)
		windows.CloseHandle(h)
	}
}

func (p *PipeServer) handle(h windows.Handle) {
	var lenBuf [4]byte
	if !readFull(h, lenBuf[:]) {
		logger.Error("UTMStack EDR AMSI pipe: failed reading length")
		return
	}
	n := binary.LittleEndian.Uint32(lenBuf[:])
	logger.Debug(100, "UTMStack EDR AMSI pipe: request length=%d", n)
	if n == 0 || n > 50*1024*1024 {
		logger.Error("UTMStack EDR AMSI pipe: bad length %d", n)
		return
	}
	payload := make([]byte, n)
	if !readFull(h, payload) {
		logger.Error("UTMStack EDR AMSI pipe: failed reading %d payload bytes", n)
		return
	}
	appName, content := splitAppName(payload)
	block := p.s.ScanBuffer(appName, content)
	logger.Debug(100, "UTMStack EDR AMSI pipe: appName=%q contentLen=%d block=%v", appName, len(content), block)
	resp := []byte{0}
	if block {
		resp[0] = 1
	}
	var written uint32
	_ = windows.WriteFile(h, resp, &written, nil)
}

func readFull(h windows.Handle, buf []byte) bool {
	got := 0
	for got < len(buf) {
		var n uint32
		err := windows.ReadFile(h, buf[got:], &n, nil)
		if err != nil || n == 0 {
			return false
		}
		got += int(n)
	}
	return true
}

func splitAppName(payload []byte) (string, []byte) {
	for i, b := range payload {
		if b == 0 {
			return string(payload[:i]), payload[i+1:]
		}
	}
	return "", payload
}
