package dependency

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
)

// verifyNativeEDR is a second identity check only. The signed release archive
// must be authenticated before this function is ever called on downloaded bytes.
func verifyNativeEDR(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 256<<20 {
		return errors.New("EDR executable is not a bounded regular file")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "version", "--json")
	cmd.Dir = filepath.Dir(path)
	var output boundedVersionOutput
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("EDR version check failed: %w", err)
	}
	return validateNativeEDRVersion(output.Bytes(), runtime.GOOS, runtime.GOARCH)
}

func validateNativeEDRVersion(data []byte, goos, goarch string) error {
	var info buildinfo.Info
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("invalid EDR version response: %w", err)
	}
	if info.PackageVersion != EDRVersion || info.Implementation != buildinfo.ImplementationVersion || info.OS != goos || info.Architecture != goarch {
		return fmt.Errorf("EDR identity mismatch: package=%q implementation=%q target=%s/%s", info.PackageVersion, info.Implementation, info.OS, info.Architecture)
	}
	return nil
}

type boundedVersionOutput struct{ buffer bytes.Buffer }

func (b *boundedVersionOutput) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedVersionOutput) Write(p []byte) (int, error) {
	if len(p) > (64<<10)-b.buffer.Len() {
		return 0, errors.New("EDR version response exceeds 64 KiB")
	}
	return b.buffer.Write(p)
}
