package event

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrSpoolFull      = errors.New("event spool is full; waiting for relay")
	ErrSpoolBusy      = errors.New("event spool is busy")
	ErrSpoolRecord    = errors.New("event spool record must be one bounded nonempty line")
	ErrSpoolIntegrity = errors.New("event spool state needs attention")
)

const MaxSpoolRecordBytes = 1 << 20

// Spool retains an active segment and one sealed backup. Accepted means a full
// record was synced locally, not acknowledged by the remote platform. A full
// buffer refuses new records rather than overwriting accepted unread records.
// All cooperating producers and the relay use the same stable OS lock.
// Upgrade producer and relay together: v2 segment headers are not event records.
type Spool struct {
	path     string
	maxBytes int64
	mu       sync.Mutex
	stats    SpoolStats
	closed   bool
}

type SpoolStats struct {
	Accepted          uint64 `json:"accepted"`
	AcceptedBytes     uint64 `json:"accepted_bytes"`
	AppendFailures    uint64 `json:"append_failures"`
	FullRejections    uint64 `json:"full_rejections"`
	BusyRejections    uint64 `json:"busy_rejections"`
	InvalidRejections uint64 `json:"invalid_rejections"`
	WriteErrors       uint64 `json:"write_errors"`
	Rotations         uint64 `json:"rotations"`
	MaxSegmentBytes   int64  `json:"max_segment_bytes"`
	MaxRecordBytes    int    `json:"max_record_bytes"`
	Degraded          bool   `json:"degraded"`
	LegacyMode        bool   `json:"legacy_mode"`
	Closed            bool   `json:"closed"`
	LastError         string `json:"last_error,omitempty"`
}

func OpenSpool(path string, maxBytes int64) (*Spool, error) {
	if maxBytes <= 0 {
		return nil, errors.New("event spool size must be positive")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlock, err := lockSpool(ctx, path)
	if err != nil {
		return nil, err
	}
	defer unlock()
	s := &Spool{path: path, maxBytes: maxBytes}
	for _, name := range []string{path + ".1", path} {
		_, header, _, err := ensureSegment(name, name == path, false)
		if err != nil && !(name != path && os.IsNotExist(err)) {
			return nil, err
		}
		if name == path {
			s.stats.LegacyMode = header == 0
		}
	}
	if err = writeSpoolState(path+".protocol", []byte("openedr-event-spool-v2\n")); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Spool) Append(line string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if err == nil {
			s.stats.Accepted++
			s.stats.AcceptedBytes += uint64(len(line) + 1)
			return
		}
		s.stats.AppendFailures++
		switch {
		case errors.Is(err, ErrSpoolFull):
			s.stats.FullRejections++
		case errors.Is(err, ErrSpoolBusy):
			s.stats.BusyRejections++
		case errors.Is(err, ErrSpoolRecord):
			s.stats.InvalidRejections++
		default:
			s.stats.WriteErrors++
		}
		s.stats.LastError = err.Error()
		if len(s.stats.LastError) > 512 {
			s.stats.LastError = s.stats.LastError[:512]
		}
	}()
	if s.closed {
		return os.ErrClosed
	}
	if len(line) == 0 || len(line)+1 > MaxSpoolRecordBytes || int64(len(line)+1) > s.maxBytes || strings.ContainsAny(line, "\r\n") {
		return ErrSpoolRecord
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	unlock, err := lockSpool(ctx, s.path)
	if err != nil {
		return err
	}
	defer unlock()
	_, header, _, err := ensureSegment(s.path, true, false)
	if err != nil {
		return err
	}
	s.stats.LegacyMode = header == 0
	f, err := openSpoolRegular(s.path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	size := info.Size() - header
	if size < 0 {
		f.Close()
		return ErrSpoolIntegrity
	}
	if size > 0 {
		var last [1]byte
		if _, err = f.ReadAt(last[:], info.Size()-1); err != nil || last[0] != '\n' {
			f.Close()
			return fmt.Errorf("%w: partial record tail", ErrSpoolIntegrity)
		}
	}
	if size > s.maxBytes-int64(len(line)+1) {
		// A previous relay cannot understand generation headers. Until the new
		// relay converts this segment, preserve its plain bytes and refuse
		// overflow instead of rotating away from that older reader.
		if header == 0 {
			f.Close()
			return ErrSpoolFull
		}
		if err = f.Close(); err != nil {
			return err
		}
		if fi, e := os.Lstat(s.path + ".1"); e == nil {
			if !fi.Mode().IsRegular() {
				return fmt.Errorf("%w: invalid backup", ErrSpoolIntegrity)
			}
			return ErrSpoolFull
		} else if !os.IsNotExist(e) {
			return e
		}
		if err = os.Rename(s.path, s.path+".1"); err != nil {
			return err
		}
		if err = syncSpoolDirectory(filepath.Dir(s.path)); err != nil {
			return err
		}
		s.stats.Rotations++
		_, header, _, err = ensureSegment(s.path, true, true)
		if err != nil {
			return err
		}
		f, err = openSpoolRegular(s.path, os.O_RDWR, 0600)
		if err != nil {
			return err
		}
		info, err = f.Stat()
		if err != nil {
			f.Close()
			return err
		}
	}
	before := info.Size()
	if _, err = f.Seek(before, 0); err == nil {
		var n int
		n, err = f.WriteString(line + "\n")
		if err == nil && n != len(line)+1 {
			err = errors.New("short event spool write")
		}
	}
	if err != nil {
		// Ordinary failed writes restore the last complete boundary. An abrupt
		// process/machine loss can leave an uncertain tail, which is never skipped.
		rollback := f.Truncate(before)
		syncErr := f.Sync()
		closeErr := f.Close()
		return errors.Join(err, rollback, syncErr, closeErr)
	}
	err = f.Sync()
	closeErr := f.Close()
	return errors.Join(err, closeErr)
}

func (s *Spool) Snapshot() SpoolStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	stats := s.stats
	stats.Closed = s.closed
	stats.Degraded = stats.AppendFailures > 0
	stats.MaxSegmentBytes = s.maxBytes
	stats.MaxRecordBytes = MaxSpoolRecordBytes
	return stats
}
func (s *Spool) Close() error { s.mu.Lock(); defer s.mu.Unlock(); s.closed = true; return nil }
