package event

import (
	"os"
	"path/filepath"
	"sync"
)

// Spool is an append-only newline-delimited JSON event buffer with a size cap.
// When the current file exceeds maxBytes it is rotated to "<path>.1" (single
// generation; the agent relay drains fast enough that one backup suffices).
type Spool struct {
	path     string
	maxBytes int64
	mu       sync.Mutex
	f        *os.File
}

func OpenSpool(path string, maxBytes int64) (*Spool, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	return &Spool{path: path, maxBytes: maxBytes, f: f}, nil
}

func (s *Spool) Append(line string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if fi, err := s.f.Stat(); err == nil && fi.Size() >= s.maxBytes {
		if err := s.rotate(); err != nil {
			return err
		}
	}
	_, err := s.f.WriteString(line + "\n")
	return err
}

func (s *Spool) rotate() error {
	if err := s.f.Close(); err != nil {
		return err
	}
	_ = os.Rename(s.path, s.path+".1") // best-effort; overwrites prior backup
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	s.f = f
	return nil
}

func (s *Spool) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.f.Close()
}
