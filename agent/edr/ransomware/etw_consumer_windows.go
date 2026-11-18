//go:build windows

package ransomware

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/0xrawsec/golang-etw/etw"
	"golang.org/x/sys/windows"
)

// Own the native consumer lifecycle and its error state. The dependency's
// Consumer.lastError is unsynchronized; its private parser also prevents an
// external lock from covering its writes. Only the bounded fields we use are
// decoded here, through the public Windows TDH APIs.
type nativeFileConsumer struct {
	ctx      context.Context
	cancel   context.CancelFunc
	id       uintptr
	handle   syscall.Handle
	log      etw.EventTraceLogfile
	name     []uint16
	onRecord func(context.Context, *etw.EventRecord) error
	done     chan struct{}
	traceConsumerState
	closeOnce sync.Once
	closeErr  error
}

var fileConsumers = struct {
	sync.Mutex
	next uintptr
	live map[uintptr]*nativeFileConsumer
}{live: make(map[uintptr]*nativeFileConsumer)}

// Windows callback stubs cannot be freed. Allocate exactly one, rather than
// retaining one closure and a whole consumer on every service restart.
var fileRecordCallback = syscall.NewCallback(func(record *etw.EventRecord) uintptr {
	if record == nil {
		return 0
	}
	fileConsumers.Lock()
	c := fileConsumers.live[record.UserContext]
	fileConsumers.Unlock()
	if c != nil {
		c.mu.Lock()
		admit := !c.closing && c.ctx.Err() == nil
		c.mu.Unlock()
		if !admit {
			return 0
		}
		if err := c.onRecord(c.ctx, record); err != nil && c.ctx.Err() == nil {
			c.noteError(err)
		}
	}
	return 0
})

var openNativeTrace = windows.NewLazySystemDLL("advapi32.dll").NewProc("OpenTraceW")

func openFileTraceHandle(log *etw.EventTraceLogfile) (syscall.Handle, error) {
	value, _, lastError := openNativeTrace.Call(uintptr(unsafe.Pointer(log)))
	runtime.KeepAlive(log)
	// OpenTrace returns a handle, not an ULONG status. GetLastError is defined
	// only on INVALID_PROCESSTRACE_HANDLE; success can leave it nonzero.
	if value == ^uintptr(0) {
		if lastError != nil && lastError != syscall.Errno(0) {
			return 0, lastError
		}
		return 0, windows.ERROR_GEN_FAILURE
	}
	return syscall.Handle(value), nil
}

func openFileConsumer(parent context.Context, name string, onRecord func(context.Context, *etw.EventRecord) error) (*nativeFileConsumer, error) {
	ctx, cancel := context.WithCancel(parent)
	c := &nativeFileConsumer{ctx: ctx, cancel: cancel, onRecord: onRecord, done: make(chan struct{})}
	ok := false
	defer func() {
		if !ok {
			cancel()
		}
	}()
	var err error
	c.name, err = syscall.UTF16FromString(name)
	if err != nil {
		return nil, err
	}
	fileConsumers.Lock()
	if len(fileConsumers.live) >= 16 {
		fileConsumers.Unlock()
		return nil, errors.New("ransomware ETW native consumer limit reached")
	}
	if fileConsumers.next == ^uintptr(0) {
		fileConsumers.Unlock()
		return nil, errors.New("ransomware ETW consumer identity exhausted")
	}
	fileConsumers.next++
	c.id = fileConsumers.next
	fileConsumers.live[c.id] = c
	fileConsumers.Unlock()
	c.log.LoggerName = &c.name[0]
	c.log.Context = c.id
	c.log.Callback = fileRecordCallback
	// Without RAW_TIMESTAMP, ProcessTrace converts QPC timestamps to FILETIME.
	c.log.SetProcessTraceMode(etw.PROCESS_TRACE_MODE_EVENT_RECORD | etw.PROCESS_TRACE_MODE_REAL_TIME)
	c.handle, err = openFileTraceHandle(&c.log)
	if err != nil {
		fileConsumers.Lock()
		delete(fileConsumers.live, c.id)
		fileConsumers.Unlock()
		return nil, err
	}
	go func() {
		err := etw.ProcessTrace(&c.handle, 1, nil, nil)
		c.mu.Lock()
		if !c.closing && c.ctx.Err() == nil {
			if err == nil {
				err = errors.New("native ETW consumer ended unexpectedly")
			}
			if c.failure == nil {
				c.failure = err
			}
		}
		c.mu.Unlock()
		// ProcessTrace has returned, so no callback can still use this context.
		fileConsumers.Lock()
		delete(fileConsumers.live, c.id)
		fileConsumers.Unlock()
		close(c.done)
	}()
	ok = true
	return c, nil
}

func (c *nativeFileConsumer) Stop() error {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closing = true
		c.cancel()
		c.mu.Unlock()
		err := etw.CloseTrace(c.handle)
		if err != nil && err != etw.ERROR_CTX_CLOSE_PENDING {
			c.closeErr = err
			c.noteError(err)
		}
	})
	select {
	case <-c.done:
		return c.closeErr
	case <-time.After(3 * time.Second):
		err := errors.New("native ETW consumer did not join after close")
		c.noteError(err)
		return err
	}
}

// TDH size claims are bounded by the actual native event payload. Read only a
// scalar or one path, never arbitrary event metadata/arrays or a general map.
func nativeETWProperty(record *etw.EventRecord, name string) ([]byte, error) {
	key, err := syscall.UTF16FromString(name)
	if err != nil {
		return nil, err
	}
	descriptor := etw.PropertyDataDescriptor{PropertyName: uint64(uintptr(unsafe.Pointer(&key[0]))), ArrayIndex: ^uint32(0)}
	var size uint32
	err = etw.TdhGetPropertySize(record, 0, nil, 1, &descriptor, &size)
	if err != nil {
		runtime.KeepAlive(key)
		return nil, err
	}
	if !validNativePropertySize(record.UserDataLength, size) {
		runtime.KeepAlive(key)
		return nil, errors.New("ETW property exceeds native event payload")
	}
	value := make([]byte, int(size))
	err = etw.TdhGetProperty(record, 0, nil, 1, &descriptor, size, &value[0])
	runtime.KeepAlive(key)
	return value, err
}

func nativeETWScalar(record *etw.EventRecord, name string, pointer bool) (uint64, error) {
	b, err := nativeETWProperty(record, name)
	if err != nil {
		return 0, err
	}
	if len(b) == 4 {
		return uint64(binary.LittleEndian.Uint32(b)), nil
	}
	if pointer && len(b) == 8 {
		return binary.LittleEndian.Uint64(b), nil
	}
	return 0, fmt.Errorf("unexpected native ETW %s width", name)
}

func nativeKernelFileRecord(record *etw.EventRecord) (kernelFileRecord, error) {
	r := kernelFileRecord{EventID: record.EventHeader.EventDescriptor.Id, PID: int(record.EventHeader.ProcessId)}
	var err error
	r.FileObject, err = nativeETWScalar(record, "FileObject", true)
	if err != nil {
		return r, err
	}
	var pathName string
	switch r.EventID {
	case kfCreate, kfCreateNewFile:
		pathName = "FileName"
		v, e := nativeETWScalar(record, "CreateOptions", false)
		if e != nil {
			return r, e
		}
		r.CreateOptions, r.HasCreateOptions = uint32(v), true
	case kfDeletePath, kfRenamePath:
		pathName = "FilePath"
	case kfSetInformation:
		v, e := nativeETWScalar(record, "InfoClass", false)
		if e != nil {
			return r, e
		}
		r.InfoClass, r.HasInfoClass = uint32(v), true
	case kfWrite:
		v, e := nativeETWScalar(record, "IOFlags", false)
		if e != nil {
			return r, e
		}
		r.IOFlags, r.HasIOFlags = uint32(v), true
	}
	if pathName != "" {
		b, e := nativeETWProperty(record, pathName)
		if e != nil {
			return r, e
		}
		if len(b)%2 != 0 || len(b) < 2 || b[len(b)-2] != 0 || b[len(b)-1] != 0 {
			return r, errors.New("invalid native ETW path")
		}
		units := make([]uint16, len(b)/2-1)
		for i := range units {
			units[i] = binary.LittleEndian.Uint16(b[i*2:])
			if units[i] == 0 {
				return r, errors.New("embedded NUL in native ETW path")
			}
		}
		r.Path = normalizeKernelPath(string(utf16.Decode(units)))
	}
	return r, nil
}
