package service

import (
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// writeStatusFile replaces the status as one complete JSON document. Readers
// must never confuse a truncated in-place write with a service startup state.
func writeStatusFile(data []byte) error {
	path := config.StatusFile
	temp, err := os.CreateTemp(filepath.Dir(path), ".edr-status-")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(0644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}
