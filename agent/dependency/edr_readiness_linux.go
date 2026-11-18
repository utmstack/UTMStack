//go:build linux

package dependency

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"golang.org/x/sys/unix"
)

type linuxEDRReadySensor struct {
	Supported bool   `json:"supported"`
	Requested bool   `json:"requested"`
	State     string `json:"state"`
	Error     string `json:"error"`
}

// A running analysis engine alone cannot prove that configured permission
// protection survived the update. Bind the gate to the local configuration as
// well as the fresh, exact-release status already verified by checkReady.
func checkEDRPlatformReadiness(base string, data []byte) error {
	var status struct {
		Sensors map[string]linuxEDRReadySensor `json:"sensor_states"`
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return fmt.Errorf("decoding Linux protection readiness: %w", err)
	}
	// Nonblocking open plus an opened-file type check keeps a substituted FIFO
	// from parking the updater. Approved regular-file configuration symlinks
	// retain the existing configuration loader's behavior.
	f, err := os.OpenFile(filepath.Join(base, "edr.json"), os.O_RDONLY|unix.O_NONBLOCK, 0)
	if os.IsNotExist(err) {
		if status.Sensors["file_access_monitor"].Requested || status.Sensors["file_prevention"].Requested || status.Sensors["network_enforcement"].Requested {
			return errors.New("Linux requested protection has no local configuration")
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("opening Linux readiness configuration: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > 1<<20 {
		return errors.New("Linux readiness configuration is not a bounded regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(f, (1<<20)+1))
	if err != nil {
		return err
	}
	if len(raw) > 1<<20 {
		return errors.New("Linux readiness configuration exceeds 1 MiB")
	}
	cfg := edrconfig.Default()
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return fmt.Errorf("decoding Linux readiness configuration: %w", err)
	}
	if err = cfg.LinuxProtection.Validate(); err != nil {
		return err
	}
	if !cfg.Enabled {
		return nil
	}
	var required []string
	if cfg.LinuxProtection.Fanotify.Enabled {
		if !cfg.Sensors.FileWatcherOn() {
			return errors.New("configured Linux file protection requires the file watcher")
		}
		required = append(required, "file_access_monitor")
		if cfg.LinuxProtection.Effective().Fanotify.Mode != "notify" {
			required = append(required, "file_prevention")
		}
	}
	if cfg.LinuxProtection.NetworkEnforcement && cfg.Blocklist.Enabled && cfg.Blocklist.Enforce {
		required = append(required, "network_enforcement")
	}
	for _, name := range required {
		s, ok := status.Sensors[name]
		if !ok || !s.Supported || !s.Requested || s.State != "active" || s.Error != "" {
			return fmt.Errorf("configured Linux %s is not ready: state=%q supported=%t requested=%t error=%q", name, s.State, s.Supported, s.Requested, s.Error)
		}
	}
	return nil
}
