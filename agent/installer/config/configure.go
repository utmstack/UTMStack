package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/utmstack/UTMStack/agent/installer/utils"
)

func ConfigureService(ip, utmKey, skip, config string) error {
	bin := GetServiceBin()
	err := execBin(bin, config, ip, utmKey, skip)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("error %sing UTMStackAgent service: Check the file /logs/utmstack_agent.log for more details", config)
		}
		return fmt.Errorf("error %sing UTMStackAgent service: %v", config, err)
	}

	return err
}

func execBin(binName string, arg ...string) error {
	path := utils.GetMyPath()

	if runtime.GOOS == "linux" {
		err := os.Chmod(filepath.Join(path, binName), 0755)
		if err != nil {
			return err
		}
	}

	result, errB := utils.ExecuteWithResult(filepath.Join(path, binName), path, arg...)
	if errB {
		return fmt.Errorf("%s", result)
	}

	return nil
}
