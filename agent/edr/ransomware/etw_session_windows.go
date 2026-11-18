//go:build windows

package ransomware

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
	"golang.org/x/sys/windows"
)

// fileTraceSession owns only the native trace handle it created. Unlike the
// dependency's session helper, it never stops another controller on name clash.
type fileTraceSession struct {
	name     string
	handle   syscall.Handle
	provider etw.Provider
	owner    traceOwnerRecord
	store    *traceOwnerStore
	stopOnce sync.Once
	stopErr  error
}

func fileTraceProperties(name string) *etw.EventTraceProperties {
	p := etw.NewRealTimeEventTraceSessionProperties(name)
	p.Wnode.Flags = 0x00020000 // WNODE_FLAG_TRACED_GUID, required by WNODE_HEADER.
	p.LogFileMode = etw.EVENT_TRACE_REAL_TIME_MODE | etw.EVENT_TRACE_NO_PER_PROCESSOR_BUFFERING
	p.BufferSize = 64
	p.MinimumBuffers = 2
	p.MaximumBuffers = 16
	p.FlushTimer = 1
	return p
}

func startFileTrace(ctx context.Context) (*fileTraceSession, error) {
	store, err := openTraceOwnerStore(ctx)
	if err != nil {
		return nil, fmt.Errorf("ransomware ETW ownership store: %w", err)
	}
	ok := false
	defer func() {
		store.unlock()
		if !ok {
			store.Close()
		}
	}()
	if err = store.cleanupOrphans(); err != nil {
		return nil, err
	}
	var creation, exit, kernel, user windows.Filetime
	if err = windows.GetProcessTimes(windows.CurrentProcess(), &creation, &exit, &kernel, &user); err != nil {
		return nil, err
	}
	id := int64(uint64(creation.HighDateTime)<<32 | uint64(creation.LowDateTime))
	controllerGUID, err := windows.GenerateGUID()
	if err != nil {
		return nil, err
	}
	owner := traceOwnerRecord{Version: 1, PID: uint32(os.Getpid()), Creation: id, GUID: strings.ToUpper(controllerGUID.String())}
	owner.Name = traceOwnerName(owner.PID, owner.Creation)
	if err = store.persist(owner); err != nil {
		return nil, err
	}
	s := &fileTraceSession{name: owner.Name, owner: owner, store: store, provider: etw.Provider{GUID: kernelFileProviderGUID, Name: kernelFileProviderName, EnableLevel: 0xff, MatchAnyKeyword: 0x1ef0}}
	name, err := syscall.UTF16PtrFromString(s.name)
	if err != nil {
		return nil, err
	}
	properties := fileTraceProperties(s.name)
	properties.Wnode.Guid = *etw.MustParseGUIDFromString(owner.GUID)
	if err = etw.StartTrace(&s.handle, name, properties); err != nil {
		// This call did not create a native session. Remove only our exact
		// protected intent; never stop the colliding session by its name.
		_ = s.removeOwnerRecord()
		return nil, err
	}
	guid, err := etw.ParseGUID(s.provider.GUID)
	if err == nil {
		params := etw.EnableTraceParameters{Version: 2}
		err = etw.EnableTraceEx2(s.handle, guid, etw.EVENT_CONTROL_CODE_ENABLE_PROVIDER, s.provider.EnableLevel, s.provider.MatchAnyKeyword, 0, 3000, &params)
	}
	if err != nil {
		return nil, errors.Join(err, s.Stop())
	}
	ok = true
	return s, nil
}
func (s *fileTraceSession) TraceName() string         { return s.name }
func (s *fileTraceSession) Providers() []etw.Provider { return []etw.Provider{s.provider} }
func (s *fileTraceSession) Stop() error {
	s.stopOnce.Do(func() {
		defer s.store.Close()
		native, err := queryOwnedTrace(s.name, uint64(s.handle))
		if err == nil && !native.Exists {
			native, err = queryOwnedTrace(s.name, 0)
			if err == nil && native.Exists {
				err = errors.New("ETW session ownership changed before stop")
			}
		} else if err == nil {
			if native.Name != s.owner.Name || native.GUID != s.owner.GUID || native.Handle != uint64(s.handle) {
				err = errors.New("ETW session ownership changed before stop")
			} else {
				err = stopVerifiedOrphan(s.owner, native)
			}
		}
		if err == nil {
			err = s.removeOwnerRecord()
		}
		s.stopErr = err
	})
	return s.stopErr
}
func (s *fileTraceSession) removeOwnerRecord() error {
	// This may already be held by failed startup. Normal Stop acquires the
	// same lock so cleanup cannot hold a conflicting DELETE record handle.
	wasHeld := s.store.lock != nil
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.store.acquire(ctx); err != nil {
		return err
	}
	if !wasHeld {
		defer s.store.unlock()
	}
	f, err := s.store.openRecord(s.owner.filename(), false)
	if err != nil {
		return err
	}
	defer f.Close()
	record, err := readTraceRecord(f)
	if err != nil {
		return err
	}
	if record != s.owner {
		return errors.New("ETW protected ownership record changed")
	}
	return deleteTraceRecord(f)
}
func (s *fileTraceSession) Loss() error {
	p := fileTraceProperties(s.name)
	if err := etw.ControlTrace(s.handle, nil, p, etw.EVENT_TRACE_CONTROL_QUERY); err != nil {
		return err
	}
	if p.EventsLost != 0 || p.RealTimeBuffersLost != 0 || p.LogBuffersLost != 0 {
		return fmt.Errorf("ransomware ETW coverage lost: %d events, %d real-time buffers, %d log buffers", p.EventsLost, p.RealTimeBuffersLost, p.LogBuffersLost)
	}
	return nil
}
