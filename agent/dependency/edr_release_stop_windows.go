//go:build windows

package dependency

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/svc"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

// SCM can report stopped before the executable's process has released its image.
// Capture the actual service process before stopping; never follow a reused PID.
func stopEDRReleaseService(expectedImage string) error {
	manager, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer manager.Disconnect()
	service, err := manager.OpenService(config.ServiceName)
	if err != nil {
		return err
	}
	defer service.Close()
	status, err := service.Query()
	if err != nil {
		return err
	}
	if status.ProcessId == 0 {
		return svc.Stop(config.ServiceName)
	}
	process, err := windows.OpenProcess(windows.SYNCHRONIZE|windows.PROCESS_QUERY_LIMITED_INFORMATION, false, status.ProcessId)
	if err != nil {
		return fmt.Errorf("pinning EDR process before stop: %w", err)
	}
	defer windows.CloseHandle(process)
	image := make([]uint16, 32768)
	size := uint32(len(image))
	if err := windows.QueryFullProcessImageName(process, 0, &image[0], &size); err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Clean(windows.UTF16ToString(image[:size])), filepath.Clean(expectedImage)) {
		return fmt.Errorf("SCM EDR process does not match the owned update image")
	}
	if err := svc.Stop(config.ServiceName); err != nil {
		return err
	}
	state, err := windows.WaitForSingleObject(process, 60000)
	if err != nil {
		return err
	}
	if state != windows.WAIT_OBJECT_0 {
		return fmt.Errorf("EDR process did not exit after SCM stop (wait=%d)", state)
	}
	return nil
}
