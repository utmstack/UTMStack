package services

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func execBin(binName string, arg ...string) error {
	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	if runtime.GOOS == "linux" {
		err = os.Chmod(filepath.Join(path, binName), 0755)
		if err != nil {
			return err
		}
	}

	err = utils.Execute(filepath.Join(path, binName), path, arg...)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
