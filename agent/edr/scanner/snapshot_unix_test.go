//go:build linux || darwin

package scanner

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"golang.org/x/sys/unix"
)

func TestSnapshotPipeReplacementDoesNotParkScanWorker(t *testing.T) {
	s, _, path, _ := scannerFixture(t, []byte("safe snapshot"))
	s.openTarget = func(p string) (*os.File, error) {
		if err := os.Rename(p, p+".saved"); err != nil {
			return nil, err
		}
		if err := unix.Mkfifo(p, 0600); err != nil {
			return nil, err
		}
		return openScannerTarget(p)
	}
	done := make(chan error, 1)
	go func() { _, _, err := s.ScanFile(path, event.SourceFileWatcher); done <- err }()
	select {
	case err := <-done:
		if !errors.Is(err, errTargetChanged) {
			t.Fatalf("pipe replacement was not rejected: %v", err)
		}
	case <-time.After(2 * time.Second):
		fd, _ := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK, 0)
		if fd >= 0 {
			_ = unix.Close(fd)
		}
		t.Fatal("FIFO replacement parked a scan worker")
	}
}
