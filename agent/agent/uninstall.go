package agent

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

func UninstallAll() error {
	// Use the current executable path - the agent uninstalls itself
	if err := exec.Run(os.Args[0], fs.GetExecutablePath(), "uninstall"); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
