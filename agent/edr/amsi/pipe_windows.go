//go:build windows

package amsi

import (
	"context"
	"errors"
	"io"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/windows"
)

const PipeName = `\\.\pipe\utmstack_edr_amsi`

const pipeRequestTimeout = 30 * time.Second

type PipeServer struct {
	s       *Scanner
	mu      sync.RWMutex
	ready   bool
	failure error
}

func NewPipeServer(s *Scanner) *PipeServer { return &PipeServer{s: s} }

var ErrPipeNotReady = errors.New("AMSI local pipe is not serving")

func (p *PipeServer) Health() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.failure != nil {
		return p.failure
	}
	if !p.ready {
		return ErrPipeNotReady
	}
	return nil
}

func (p *PipeServer) setState(ready bool, err error) {
	p.mu.Lock()
	p.ready = ready
	if err != nil {
		p.failure = err
	}
	p.mu.Unlock()
}

// Run serves one bounded request at a time. Connection waits and all pipe I/O
// use overlapped operations, so shutdown or a stalled client can cancel them.
// The portable parser also accepts existing complete-buffer DLL requests.
func (p *PipeServer) Run(ctx context.Context) {
	name, err := windows.UTF16PtrFromString(PipeName)
	if err != nil {
		p.setState(false, err)
		logger.Error("UTMStack EDR AMSI: bad pipe name: %v", err)
		return
	}
	sd, err := windows.SecurityDescriptorFromString(pipeSecuritySDDL)
	if err != nil {
		p.setState(false, err)
		logger.Error("UTMStack EDR AMSI: pipe security descriptor: %v", err)
		return
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), SecurityDescriptor: sd}
	h, err := windows.CreateNamedPipe(name,
		windows.PIPE_ACCESS_DUPLEX|windows.FILE_FLAG_OVERLAPPED|windows.FILE_FLAG_FIRST_PIPE_INSTANCE,
		windows.PIPE_TYPE_BYTE|windows.PIPE_READMODE_BYTE|windows.PIPE_WAIT|windows.PIPE_REJECT_REMOTE_CLIENTS,
		1, 65536, 65536, 0, &sa)
	runtime.KeepAlive(sd)
	if err != nil {
		p.setState(false, err)
		logger.Error("UTMStack EDR AMSI: create LocalSystem-owned pipe: %v", err)
		return
	}
	defer windows.CloseHandle(h)
	p.setState(true, nil)
	defer p.setState(false, nil)
	// Retain this first instance for the service lifetime: recreating its name
	// between clients would permit a competing endpoint in that interval.
	for ctx.Err() == nil {
		_, err = pipeOperation(ctx, h, func(o *windows.Overlapped, _ *uint32) error {
			err := windows.ConnectNamedPipe(h, o)
			if err == windows.ERROR_PIPE_CONNECTED {
				return nil
			}
			return err
		})
		if err != nil && ctx.Err() == nil {
			p.setState(false, err)
			logger.Error("UTMStack EDR AMSI: local pipe stopped accepting clients: %v", err)
			return
		}
		if err == nil {
			requestCtx, cancel := context.WithTimeout(ctx, pipeRequestTimeout)
			p.handle(requestCtx, h)
			cancel()
		}
		windows.DisconnectNamedPipe(h)
	}
}

func (p *PipeServer) handle(ctx context.Context, h windows.Handle) {
	stream := pipeStream{ctx: ctx, h: h}
	req, err := readPipeRequest(stream)
	response := []byte{p.s.pipeDecision(ctx, req, err)}
	// A timed-out scan may still report its incomplete/fail-policy decision.
	// Keep delivery/peer-close bounded independently of the expired analysis.
	if ctx.Err() != nil {
		replyCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer cancel()
		stream.ctx = replyCtx
	}
	if _, err := stream.Write(response); err != nil {
		return
	}
	// FlushFileBuffers can wait forever for an uncooperative client. Instead
	// wait, within the same deadline, for the client to close after reading
	// its reply. This also prevents DisconnectNamedPipe discarding that reply.
	var end [1]byte
	_, _ = stream.Read(end[:])
}

type pipeStream struct {
	ctx context.Context
	h   windows.Handle
}

func (p pipeStream) Read(b []byte) (int, error) {
	n, err := pipeOperation(p.ctx, p.h, func(o *windows.Overlapped, n *uint32) error {
		// x/sys's race hooks require this pointer. Only GetOverlappedResult's
		// count is used, including when the operation completes immediately.
		err := windows.ReadFile(p.h, b, n, o)
		if err == nil {
			err = windows.GetOverlappedResult(p.h, o, n, false)
		}
		return err
	})
	runtime.KeepAlive(b)
	if errors.Is(err, windows.ERROR_BROKEN_PIPE) || err == nil && n == 0 {
		return int(n), io.EOF
	}
	return int(n), err
}

func (p pipeStream) Write(b []byte) (int, error) {
	n, err := pipeOperation(p.ctx, p.h, func(o *windows.Overlapped, n *uint32) error {
		err := windows.WriteFile(p.h, b, n, o)
		if err == nil {
			err = windows.GetOverlappedResult(p.h, o, n, false)
		}
		return err
	})
	runtime.KeepAlive(b)
	if err == nil && int(n) != len(b) {
		err = io.ErrShortWrite
	}
	return int(n), err
}

func pipeOperation(ctx context.Context, h windows.Handle, start func(*windows.Overlapped, *uint32) error) (uint32, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(event)
	o := windows.Overlapped{HEvent: event}
	var n uint32
	err = start(&o, &n)
	if err != windows.ERROR_IO_PENDING {
		return n, err
	}
	for {
		if err := ctx.Err(); err != nil {
			// The kernel must release the OVERLAPPED and caller's buffer before
			// either can leave scope, including when cancellation races success.
			_ = windows.CancelIoEx(h, &o)
			_ = windows.GetOverlappedResult(h, &o, &n, true)
			return n, err
		}
		state, err := windows.WaitForSingleObject(event, 50)
		if err != nil || state != uint32(windows.WAIT_TIMEOUT) && state != windows.WAIT_OBJECT_0 {
			_ = windows.CancelIoEx(h, &o)
			_ = windows.GetOverlappedResult(h, &o, &n, true)
			if err == nil {
				err = windows.ERROR_INVALID_FUNCTION
			}
			return n, err
		}
		if state == windows.WAIT_OBJECT_0 {
			err := windows.GetOverlappedResult(h, &o, &n, false)
			return n, err
		}
	}
}
